//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package inverted

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/adapters/repos/db/inverted/stopwords"
	"github.com/weaviate/weaviate/adapters/repos/db/lsmkv"
	"github.com/weaviate/weaviate/adapters/repos/db/priorityqueue"
	"golang.org/x/sync/errgroup"

	"github.com/weaviate/weaviate/entities/models"

	"github.com/weaviate/weaviate/adapters/repos/db/helpers"
	"github.com/weaviate/weaviate/entities/schema"
	"github.com/weaviate/weaviate/entities/searchparams"
	"github.com/weaviate/weaviate/entities/storobj"
)

// global var to store wand times
var (
	wandDiskTimes     []float64 = make([]float64, 5)
	wandDiskCounter   int       = 0
	wandDiskLastClass string    = ""
	wandDiskStats     []float64 = make([]float64, 1)
)

func (b *BM25Searcher) wandDisk(
	ctx context.Context, filterDocIds helpers.AllowList, class *models.Class, params searchparams.KeywordRanking, limit int,
) ([]*storobj.Object, []float32, error) {
	if wandDiskLastClass == "" {
		wandDiskLastClass = string(class.Class)
	}
	if wandDiskLastClass != string(class.Class) {
		fmt.Printf("DISK,%v", wandDiskLastClass)
		sum := 0.
		for i := 0; i < len(wandDiskTimes); i++ {
			fmt.Printf(",%8.2f", wandDiskTimes[i]/float64(wandDiskCounter))
			sum += wandDiskTimes[i] / float64(wandDiskCounter)
		}
		fmt.Printf(",%8.2f", sum)
		//for i := 0; i < len(wandDiskStats); i++ {
		//	fmt.Printf(",%8.2f", wandDiskStats[i]/float64(wandDiskCounter))
		//}
		fmt.Printf("\n")
		wandDiskTimes = make([]float64, len(wandDiskTimes))
		wandDiskStats = make([]float64, len(wandDiskStats))
		wandDiskCounter = 0
		wandDiskLastClass = string(class.Class)
	}

	wandDiskCounter++
	wandTimesId := 0

	// start timer
	startTime := float64(time.Now().UnixNano()) / 1e6

	N := float64(b.store.Bucket(helpers.ObjectsBucketLSM).Count())

	var stopWordDetector *stopwords.Detector
	if class.InvertedIndexConfig != nil && class.InvertedIndexConfig.Stopwords != nil {
		var err error
		stopWordDetector, err = stopwords.NewDetectorFromConfig(*(class.InvertedIndexConfig.Stopwords))
		if err != nil {
			return nil, nil, err
		}
	}

	queryTermsByTokenization := map[string][]string{}
	duplicateBoostsByTokenization := map[string][]int{}
	propNamesByTokenization := map[string][]string{}
	propertyBoosts := make(map[string]float32, len(params.Properties))

	for _, tokenization := range helpers.Tokenizations {
		queryTerms, dupBoosts := helpers.TokenizeAndCountDuplicates(tokenization, params.Query)
		queryTermsByTokenization[tokenization] = queryTerms
		duplicateBoostsByTokenization[tokenization] = dupBoosts

		// stopword filtering for word tokenization
		if tokenization == models.PropertyTokenizationWord {
			queryTerms, dupBoosts = b.removeStopwordsFromQueryTerms(queryTermsByTokenization[tokenization],
				duplicateBoostsByTokenization[tokenization], stopWordDetector)
			queryTermsByTokenization[tokenization] = queryTerms
			duplicateBoostsByTokenization[tokenization] = dupBoosts
		}

		propNamesByTokenization[tokenization] = make([]string, 0)
	}

	averagePropLength := 0.
	for _, propertyWithBoost := range params.Properties {
		property := propertyWithBoost
		propBoost := 1
		if strings.Contains(propertyWithBoost, "^") {
			property = strings.Split(propertyWithBoost, "^")[0]
			boostStr := strings.Split(propertyWithBoost, "^")[1]
			propBoost, _ = strconv.Atoi(boostStr)
		}
		propertyBoosts[property] = float32(propBoost)

		propMean, err := b.GetPropertyLengthTracker().PropertyMean(property)
		if err != nil {
			return nil, nil, err
		}
		averagePropLength += float64(propMean)

		prop, err := schema.GetPropertyByName(class, property)
		if err != nil {
			return nil, nil, err
		}

		switch dt, _ := schema.AsPrimitive(prop.DataType); dt {
		case schema.DataTypeText, schema.DataTypeTextArray:
			if _, exists := propNamesByTokenization[prop.Tokenization]; !exists {
				return nil, nil, fmt.Errorf("cannot handle tokenization '%v' of property '%s'",
					prop.Tokenization, prop.Name)
			}
			propNamesByTokenization[prop.Tokenization] = append(propNamesByTokenization[prop.Tokenization], property)
		default:
			return nil, nil, fmt.Errorf("cannot handle datatype '%v' of property '%s'", dt, prop.Name)
		}
	}

	wandDiskTimes[wandTimesId] += float64(time.Now().UnixNano())/1e6 - startTime
	wandTimesId++
	startTime = float64(time.Now().UnixNano()) / 1e6

	averagePropLength = averagePropLength / float64(len(params.Properties))

	allSegments, propertySizes, err := lsmkv.GetAllSegments(b.store, propNamesByTokenization, queryTermsByTokenization)
	if err != nil {
		return nil, nil, err
	}

	wandDiskTimes[wandTimesId] += float64(time.Now().UnixNano())/1e6 - startTime
	wandTimesId++
	startTime = float64(time.Now().UnixNano()) / 1e6

	allObjects := make([][]*storobj.Object, len(allSegments))
	allScores := make([][]float32, len(allSegments))

	var eg errgroup.Group
	eg.SetLimit(_NUMCPU)

	currentBucket := 0
	for segment, propName := range allSegments {
		segment := segment
		propName := propName
		myCurrentBucket := currentBucket
		currentBucket++
		eg.Go(func() (err error) {
			terms := make([]*lsmkv.Term, 0, len(queryTermsByTokenization[models.PropertyTokenizationWord]))
			for i, term := range queryTermsByTokenization[models.PropertyTokenizationWord] {
				// pass i to the closure
				i := i
				term := term
				duplicateTextBoost := duplicateBoostsByTokenization[models.PropertyTokenizationWord][i]

				singleTerms, err := segment.WandTerm([]byte(term), N, float64(duplicateTextBoost), float64(propertyBoosts[propName]), propertySizes[term])
				if err == nil {
					terms = append(terms, singleTerms)
				}
			}

			flatTerms := terms
			wandDiskStats[0] += float64(len(queryTermsByTokenization[models.PropertyTokenizationWord]))

			resultsOriginalOrder := make(lsmkv.Terms, len(flatTerms))
			copy(resultsOriginalOrder, flatTerms)

			topKHeap, _ := lsmkv.GetTopKHeap(limit, flatTerms, averagePropLength, b.config)
			objects, scores, err := b.getTopKObjectsDisk(topKHeap, resultsOriginalOrder, params.AdditionalExplanations)

			for _, term := range flatTerms {
				term.ClearData()
			}

			allObjects[myCurrentBucket] = objects
			allScores[myCurrentBucket] = scores

			if err != nil {
				return err
			}
			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, nil, err
	}

	// merge the results from the different buckets
	objects, scores := b.rankMultiBucket(allObjects, allScores, limit)

	return objects, scores, nil
}

func (b *BM25Searcher) getTopKObjectsDisk(topKHeap *priorityqueue.Queue[any],
	results lsmkv.Terms, additionalExplanations bool,
) ([]*storobj.Object, []float32, error) {
	objectsBucket := b.store.Bucket(helpers.ObjectsBucketLSM)
	if objectsBucket == nil {
		return nil, nil, errors.Errorf("objects bucket not found")
	}

	objects := make([]*storobj.Object, 0, topKHeap.Len())
	scores := make([]float32, 0, topKHeap.Len())

	// If there is a crash and WAL recovery, the inverted index may have objects that are not in the objects bucket.
	// This is an issue that needs to be fixed, but for now we need to reduce the huge amount of log messages that
	// are generated by this issue. Therefore, we only log the first time we encounter a missing object.
	// TODO: consider deleting these entries from the inverted index and alerting the user
	// Related to #4125
	loggedMissingObjects := false

	buf := make([]byte, 8)
	for topKHeap.Len() > 0 {
		res := topKHeap.Pop()
		binary.LittleEndian.PutUint64(buf, res.ID)
		objectByte, err := objectsBucket.GetBySecondary(0, buf)
		if err != nil {
			return nil, nil, err
		}

		if len(objectByte) == 0 {
			if !loggedMissingObjects {
				b.logger.Warnf("Skipping object in BM25: object with id %v has a length of 0 bytes (not found).", res.ID)
				b.logger.Warnf("This is likely due to a partial WAL recovery. Further occurrences of this message for this query will be suppressed.")
				loggedMissingObjects = true
			}
			continue
		}

		obj, err := storobj.FromBinary(objectByte)
		if err != nil {
			return nil, nil, err
		}

		if additionalExplanations {
			// add score explanation
			if obj.AdditionalProperties() == nil {
				obj.Object.Additional = make(map[string]interface{})
			}
			for _, result := range results {
				queryTerm := result.QueryTerm
				obj.Object.Additional["BM25F_"+queryTerm+"_frequency"] = result.Data.Frequency
				obj.Object.Additional["BM25F_"+queryTerm+"_propLength"] = result.Data.PropLength
			}
		}
		objects = append(objects, obj)
		scores = append(scores, res.Dist)

	}
	return objects, scores, nil
}

func (b *BM25Searcher) rankMultiBucket(allObjects [][]*storobj.Object, allScores [][]float32, limit int) ([]*storobj.Object, []float32) {
	if len(allObjects) == 1 {
		return allObjects[0], allScores[0]
	}

	// allObjects and allScores are ordered by reverse score already
	// we need to merge them and keep the top K

	// merge allObjects and allScores
	mergedObjects := make([]*storobj.Object, limit)
	mergedScores := make([]float32, limit)
	mergedPos := limit - 1

	bucketPosition := make([]int, len(allObjects))

	for i := range bucketPosition {
		bucketPosition[i] = len(allObjects[i]) - 1
	}

	// iterate by bucket, bet the one with the highest score and add it to the merged list
	for {
		// find the best score
		bestScore := float32(-1)
		bestScoreIndex := -1
		lowestDocID := ""

		for i := range allObjects {
			if bucketPosition[i] >= 0 {
				if allScores[i][bucketPosition[i]] > bestScore {
					bestScore = allScores[i][bucketPosition[i]]
					bestScoreIndex = i
					lowestDocID = allObjects[i][bucketPosition[i]].ID().String()
				} else if allScores[i][bucketPosition[i]] == bestScore {
					uuid2 := allObjects[i][bucketPosition[i]].ID().String()
					res := strings.Compare(uuid2, lowestDocID)
					if res < 0 {
						bestScoreIndex = i
						lowestDocID = uuid2
					}
				}
			}
		}

		// if we found a score, add it to the merged list
		if bestScoreIndex != -1 {
			mergedObjects[mergedPos] = allObjects[bestScoreIndex][bucketPosition[bestScoreIndex]]
			mergedScores[mergedPos] = allScores[bestScoreIndex][bucketPosition[bestScoreIndex]]
			bucketPosition[bestScoreIndex]--
			mergedPos--
		}

		// if we didn't find any score, we are done
		if bestScoreIndex == -1 || mergedPos < 0 {
			break
		}
	}

	// if the merged list is smaller than the limit, we need to remove the empty slots
	if mergedPos > limit {
		mergedObjects = mergedObjects[mergedPos+1:]
		mergedScores = mergedScores[mergedPos+1:]
	}

	return mergedObjects, mergedScores
}