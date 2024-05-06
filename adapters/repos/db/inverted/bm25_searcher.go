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
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	enterrors "github.com/weaviate/weaviate/entities/errors"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/sroar"
	"github.com/weaviate/weaviate/adapters/repos/db/helpers"
	"github.com/weaviate/weaviate/adapters/repos/db/inverted/stopwords"
	"github.com/weaviate/weaviate/adapters/repos/db/lsmkv"
	"github.com/weaviate/weaviate/adapters/repos/db/priorityqueue"
	"github.com/weaviate/weaviate/adapters/repos/db/propertyspecific"
	"github.com/weaviate/weaviate/entities/inverted"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
	"github.com/weaviate/weaviate/entities/searchparams"
	"github.com/weaviate/weaviate/entities/storobj"
)

type BM25Searcher struct {
	config         schema.BM25Config
	store          *lsmkv.Store
	getClass       func(string) *models.Class
	classSearcher  ClassSearcher // to allow recursive searches on ref-props
	propIndices    propertyspecific.Indices
	propLenTracker propLengthRetriever
	logger         logrus.FieldLogger
	shardVersion   uint16
}

type propLengthRetriever interface {
	PropertyMean(prop string) (float32, error)
}

func NewBM25Searcher(config schema.BM25Config, store *lsmkv.Store,
	getClass func(string) *models.Class, propIndices propertyspecific.Indices,
	classSearcher ClassSearcher, propLenTracker propLengthRetriever,
	logger logrus.FieldLogger, shardVersion uint16,
) *BM25Searcher {
	return &BM25Searcher{
		config:         config,
		store:          store,
		getClass:       getClass,
		propIndices:    propIndices,
		classSearcher:  classSearcher,
		propLenTracker: propLenTracker,
		logger:         logger.WithField("action", "bm25_search"),
		shardVersion:   shardVersion,
	}
}

func (b *BM25Searcher) extractTermInformation(class *models.Class, params searchparams.KeywordRanking) (map[string][]string, map[string][]int, map[string][]string, map[string]float32, float64, error) {
	var stopWordDetector *stopwords.Detector
	var err error
	if class.InvertedIndexConfig != nil && class.InvertedIndexConfig.Stopwords != nil {
		stopWordDetector, err = stopwords.NewDetectorFromConfig(*(class.InvertedIndexConfig.Stopwords))
	}
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	queryTermsByTokenization := map[string][]string{}
	duplicateBoostsByTokenization := map[string][]int{}
	propNamesByTokenization := map[string][]string{}
	propertyBoosts := make(map[string]float32, len(params.Properties))

	for _, tokenization := range helpers.Tokenizations {
		queryTerms, dupBoosts := helpers.TokenizeAndCountDuplicates(tokenization, params.Query)
		queryTermsByTokenization[tokenization] = queryTerms
		duplicateBoostsByTokenization[tokenization] = dupBoosts

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
			return nil, nil, nil, nil, 0, err
		}
		averagePropLength += float64(propMean)

		prop, err := schema.GetPropertyByName(class, property)
		if err != nil {
			return nil, nil, nil, nil, 0, err
		}

		switch dt, _ := schema.AsPrimitive(prop.DataType); dt {
		case schema.DataTypeText, schema.DataTypeTextArray:
			if _, exists := propNamesByTokenization[prop.Tokenization]; !exists {
				return nil, nil, nil, nil, 0, fmt.Errorf("cannot handle tokenization '%v' of property '%s'",
					prop.Tokenization, prop.Name)
			}
			propNamesByTokenization[prop.Tokenization] = append(propNamesByTokenization[prop.Tokenization], property)
		default:
			return nil, nil, nil, nil, 0, fmt.Errorf("cannot handle datatype '%v' of property '%s'", dt, prop.Name)
		}
	}
	averagePropLength = averagePropLength / float64(len(params.Properties))

	return queryTermsByTokenization, duplicateBoostsByTokenization, propNamesByTokenization, propertyBoosts, averagePropLength, nil
}

func (b *BM25Searcher) BM25F(ctx context.Context, filterDocIds helpers.AllowList,
	className schema.ClassName, limit int, keywordRanking searchparams.KeywordRanking,
) ([]*storobj.Object, []float32, error) {
	// WEAVIATE-471 - If a property is not searchable, return an error
	for _, property := range keywordRanking.Properties {
		if !PropertyHasSearchableIndex(b.getClass(className.String()), property) {
			return nil, nil, inverted.NewMissingSearchableIndexError(property)
		}
	}
	class := b.getClass(className.String())
	if class == nil {
		return nil, nil, fmt.Errorf("could not find class %s in schema", className)
	}

	objs, scores, err := b.wand(ctx, filterDocIds, class, keywordRanking, limit)
	if err != nil {
		return nil, nil, errors.Wrap(err, "wand")
	}
	return objs, scores, nil
}

func (b *BM25Searcher) GetPropertyLengthTracker() *JsonPropertyLengthTracker {
	return b.propLenTracker.(*JsonPropertyLengthTracker)
}

func (b *BM25Searcher) wand(
	ctx context.Context, filterDocIds helpers.AllowList, class *models.Class, params searchparams.KeywordRanking, limit int,
) ([]*storobj.Object, []float32, error) {
	useWandDisk := os.Getenv("USE_WAND_DISK") == "true"
	useWandDiskForced := os.Getenv("USE_WAND_DISK") == "force"
	validateWandDisk := os.Getenv("USE_WAND_DISK") == "validate"
	validateWandDiskForced := os.Getenv("USE_WAND_DISK") == "validate-force"
	if useWandDisk || useWandDiskForced {
		return b.wandDiskMem(ctx, filterDocIds, class, params, limit, useWandDiskForced)
	} else if validateWandDisk || validateWandDiskForced {
		return b.validateWand(ctx, filterDocIds, class, params, limit, validateWandDiskForced)
	} else {
		return b.wandMem(ctx, filterDocIds, class, params, limit)
	}
}

func (b *BM25Searcher) validateWand(
	ctx context.Context, filterDocIds helpers.AllowList, class *models.Class, params searchparams.KeywordRanking, limit int, useWandDiskForced bool,
) ([]*storobj.Object, []float32, error) {
	objsD, scoresD, errD := b.wandDiskMem(ctx, filterDocIds, class, params, limit, useWandDiskForced)
	objsM, scoresM, errM := b.wandMem(ctx, filterDocIds, class, params, limit)
	if errD != nil {
		return nil, nil, errD
	}
	if errM != nil {
		return nil, nil, errM
	}

	var err error
	// compare results and scores
	if len(objsD) != len(objsM) {
		fmt.Printf("different number of results: disk %d, mem %d", len(objsD), len(objsM))
		err = fmt.Errorf("different number of results: disk %d, mem %d", len(objsD), len(objsM))
	}
	if len(scoresD) != len(scoresM) {
		fmt.Printf("different number of scores: disk %d, mem %d", len(objsD), len(objsM))
		err = fmt.Errorf("different number of scores: disk %d, mem %d", len(scoresD), len(scoresM))
	}

	if err != nil {
		for i := range objsM {
			if len(objsD) <= i && len(objsM) <= i {
				fmt.Printf("disk %v,%v mem %v,%v\n", scoresD[i], objsD[i].ID(), scoresM[i], objsM[i].ID())
			}
		}
		return nil, nil, err
	}

	for i := range objsM {

		if objsD[i].ID() != objsM[i].ID() {
			err = fmt.Errorf("different IDs at index %d: disk %v,%v mem %v,%v", i, scoresD[i], objsD[i].ID(), scoresM[i], objsM[i].ID())
			fmt.Printf("different IDs at index %d: disk %v,%v mem %v,%v\n", i, scoresD[i], objsD[i].ID(), scoresM[i], objsM[i].ID())
			break
		}
		if scoresD[i] != scoresM[i] {
			err = fmt.Errorf("different scores at index %d: disk %v,%v mem %v,%v", i, scoresD[i], objsD[i].ID(), scoresM[i], objsM[i].ID())
			fmt.Printf("different scores at index %d: disk %v,%v mem %v,%v\n", i, scoresD[i], objsD[i].ID(), scoresM[i], objsM[i].ID())
			break
		}
	}

	if err != nil {
		for i := range objsM {
			fmt.Printf("disk %v,%v mem %v,%v\n", scoresD[i], objsD[i].ID(), scoresM[i], objsM[i].ID())
		}
		return nil, nil, err
	}

	return objsM, scoresM, errM
}

// global var to store wand times
/*
var (
	wandMemTimes     []float64 = make([]float64, 5)
	wandMemCounter   int       = 0
	wandMemLastClass string
	wandMemStats     []float64 = make([]float64, 1)
)
*/

func (b *BM25Searcher) wandMem(ctx context.Context, filterDocIds helpers.AllowList, class *models.Class, params searchparams.KeywordRanking, limit int) ([]*storobj.Object, []float32, error) {
	// start timer
	// startTime := float64(time.Now().UnixNano()) / 1e6

	N := float64(b.store.Bucket(helpers.ObjectsBucketLSM).Count())

	// stopword filtering for word tokenization
	queryTermsByTokenization, duplicateBoostsByTokenization, propNamesByTokenization, propertyBoosts, averagePropLength, err := b.extractTermInformation(class, params)
	if err != nil {
		return nil, nil, err
	}

	// wandDiskTimes[wandTimesId] += float64(time.Now().UnixNano())/1e6 - startTime
	// wandTimesId++
	// startTime = float64(time.Now().UnixNano()) / 1e6

	return b.wandMemScoring(queryTermsByTokenization, duplicateBoostsByTokenization, propNamesByTokenization, propertyBoosts, averagePropLength, N, filterDocIds, params, limit)
}

func (b *BM25Searcher) wandMemScoring(queryTermsByTokenization map[string][]string, duplicateBoostsByTokenization map[string][]int, propNamesByTokenization map[string][]string, propertyBoosts map[string]float32, averagePropLength float64, N float64, filterDocIds helpers.AllowList, params searchparams.KeywordRanking, limit int) ([]*storobj.Object, []float32, error) {
	/*
		f wandMemLastClass == "" {
			wandMemLastClass = string(class.Class)
		}
		if wandMemLastClass != string(class.Class) {
			fmt.Printf(" MEM,%v", wandMemLastClass)
			sum := 0.
			for i := 0; i < len(wandMemTimes); i++ {
				fmt.Printf(",%8.2f", wandMemTimes[i]/float64(wandMemCounter))
				sum += wandMemTimes[i] / float64(wandMemCounter)
			}
			fmt.Printf(",%8.2f", sum)

			// for i := 0; i < len(wandMemStats); i++ {
			//	fmt.Printf(",%8.2f", wandMemStats[i]/float64(wandMemCounter))
			// }
			fmt.Printf("\n")
			wandMemCounter = 0
			wandMemLastClass = string(class.Class)
			wandMemTimes = make([]float64, len(wandMemTimes))
			wandMemStats = make([]float64, len(wandMemStats))
		}

		wandMemCounter++
		wandTimesId := 0

		// start timer
		startTime := float64(time.Now().UnixNano()) / 1e6
	*/

	// 100 is a reasonable expected capacity for the total number of terms to query.
	results := make([]term, 0, 100)
	indices := make([]map[uint64]int, 0, 100)

	eg := enterrors.NewErrorGroupWrapper(b.logger)
	eg.SetLimit(_NUMCPU)

	var resultsLock sync.Mutex

	for tokenization, propNames := range propNamesByTokenization {
		propNames := propNames
		if len(propNames) > 0 {
			for queryTermId, queryTerm := range queryTermsByTokenization[tokenization] {
				tokenization := tokenization
				queryTerm := queryTerm
				queryTermId := queryTermId

				dupBoost := duplicateBoostsByTokenization[tokenization][queryTermId]

				eg.Go(func() (err error) {
					termResult, docIndices, termErr := b.createTerm(N, filterDocIds, queryTerm, propNames, propertyBoosts, dupBoost, params.AdditionalExplanations)
					if termErr != nil {
						err = termErr
						return
					}
					resultsLock.Lock()
					results = append(results, termResult)
					indices = append(indices, docIndices)
					resultsLock.Unlock()
					return
				}, "query_term", queryTerm, "prop_names", propNames, "has_filter", filterDocIds != nil)
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}
	// all results. Sum up the length of the results from all terms to get an upper bound of how many results there are
	if limit == 0 {
		for _, ind := range indices {
			limit += len(ind)
		}
	}

	// wandMemTimes[wandTimesId] += float64(time.Now().UnixNano())/1e6 - startTime
	// wandTimesId++
	// startTime = float64(time.Now().UnixNano()) / 1e6

	// the results are needed in the original order to be able to locate frequency/property length for the top-results
	resultsOriginalOrder := Terms{}
	resultsOriginalOrder.list = make([]term, len(results))
	copy(resultsOriginalOrder.list, results)

	resultsTerms := Terms{}
	resultsTerms.list = results

	topKHeap := b.getTopKHeap(limit, resultsTerms, averagePropLength)

	// wandMemTimes[wandTimesId] += float64(time.Now().UnixNano())/1e6 - startTime
	// wandTimesId++
	// startTime = float64(time.Now().UnixNano()) / 1e6

	objects, scores, err := b.getTopKObjects(topKHeap, resultsOriginalOrder, indices, params.AdditionalExplanations)

	// wandMemTimes[wandTimesId] += float64(time.Now().UnixNano())/1e6 - startTime
	// wandTimesId++

	return objects, scores, err
}

func (b *BM25Searcher) removeStopwordsFromQueryTerms(queryTerms []string,
	duplicateBoost []int, detector *stopwords.Detector,
) ([]string, []int) {
	if detector == nil || len(queryTerms) == 0 {
		return queryTerms, duplicateBoost
	}

	i := 0
WordLoop:
	for {
		if i == len(queryTerms) {
			return queryTerms, duplicateBoost
		}
		queryTerm := queryTerms[i]
		if detector.IsStopword(queryTerm) {
			queryTerms[i] = queryTerms[len(queryTerms)-1]
			queryTerms = queryTerms[:len(queryTerms)-1]
			duplicateBoost[i] = duplicateBoost[len(duplicateBoost)-1]
			duplicateBoost = duplicateBoost[:len(duplicateBoost)-1]

			continue WordLoop
		}

		i++
	}
}

func (b *BM25Searcher) getTopKObjects(topKHeap *priorityqueue.Queue[any],
	results Terms, indices []map[uint64]int, additionalExplanations bool,
) ([]*storobj.Object, []float32, error) {
	objectsBucket := b.store.Bucket(helpers.ObjectsBucketLSM)
	if objectsBucket == nil {
		return nil, nil, errors.Errorf("objects bucket not found")
	}

	objects := make([]*storobj.Object, 0, topKHeap.Len())
	scores := make([]float32, 0, topKHeap.Len())

	buf := make([]byte, 8)
	for topKHeap.Len() > 0 {
		res := topKHeap.Pop()
		binary.LittleEndian.PutUint64(buf, res.ID)
		objectByte, err := objectsBucket.GetBySecondary(0, buf)
		if err != nil {
			return nil, nil, err
		}

		// If there is a crash and WAL recovery, the inverted index may have objects that are not in the objects bucket.
		// This is an issue that needs to be fixed, but for now we need to reduce the huge amount of log messages that
		// are generated by this issue. Logging the first time we encounter a missing object in a query still resulted
		// in a huge amount of log messages and it will happen on all queries, so we not log at all for now.
		// The user has already been alerted about ppossible data loss when the WAL recovery happened.
		// TODO: consider deleting these entries from the inverted index and alerting the user
		if len(objectByte) == 0 {
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
			for j, result := range results.list {
				if termIndex, ok := indices[j][res.ID]; ok {
					queryTerm := result.queryTerm
					if len(result.data) <= termIndex {
						b.logger.Warnf(
							"Skipping object explanation in BM25: term index %v is out of range for query term %v, length %d, id %v",
							termIndex, queryTerm, len(result.data), res.ID)
						continue
					}
					obj.Object.Additional["BM25F_"+queryTerm+"_frequency"] = result.data[termIndex].frequency
					obj.Object.Additional["BM25F_"+queryTerm+"_propLength"] = result.data[termIndex].propLength
				}
			}
		}
		objects = append(objects, obj)
		scores = append(scores, res.Dist)

	}
	return objects, scores, nil
}

func (b *BM25Searcher) getTopKHeap(limit int, results Terms, averagePropLength float64,
) *priorityqueue.Queue[any] {
	topKHeap := priorityqueue.NewMinScoreAndId[any](limit)
	worstDist := float64(-10000) // tf score can be negative
	sort.Sort(results)
	for {
		if results.completelyExhausted() || results.pivot(worstDist) {
			return topKHeap
		}

		id, score := results.scoreNext(averagePropLength, b.config)

		if topKHeap.Len() < limit || topKHeap.Top().Dist < float32(score) {
			topKHeap.Insert(id, float32(score))
			for topKHeap.Len() > limit {
				topKHeap.Pop()
			}
			// only update the worst distance when the queue is full, otherwise results can be missing if the first
			// entry that is checked already has a very high score
			if topKHeap.Len() >= limit {
				worstDist = float64(topKHeap.Top().Dist)
			}
		}
	}
}

func (b *BM25Searcher) createTerm(N float64, filterDocIds helpers.AllowList, query string,
	propertyNames []string, propertyBoosts map[string]float32, duplicateTextBoost int,
	additionalExplanations bool,
) (term, map[uint64]int, error) {
	termResult := term{queryTerm: query}
	filteredDocIDs := sroar.NewBitmap() // to build the global n if there is a filter

	allMsAndProps := make(AllMapPairsAndPropName, 0, len(propertyNames))
	for _, propName := range propertyNames {

		bucket := b.store.Bucket(helpers.BucketSearchableFromPropNameLSM(propName))
		if bucket == nil {
			return termResult, nil, fmt.Errorf("could not find bucket for property %v", propName)
		}
		preM, err := bucket.MapList([]byte(query))
		if err != nil {
			return termResult, nil, err
		}

		var m []lsmkv.MapPair
		if filterDocIds != nil {
			m = make([]lsmkv.MapPair, 0, len(preM))
			for _, val := range preM {
				docID := binary.BigEndian.Uint64(val.Key)
				if filterDocIds.Contains(docID) {
					m = append(m, val)
				} else {
					filteredDocIDs.Set(docID)
				}
			}
		} else {
			m = preM
		}
		if len(m) == 0 {
			continue
		}

		allMsAndProps = append(allMsAndProps, MapPairsAndPropName{MapPairs: m, propname: propName})
	}

	// sort ascending, this code has two effects
	// 1) We can skip writing the indices from the last property to the map (see next comment). Therefore, having the
	//    biggest property at the end will save us most writes on average
	// 2) For the first property all entries are new, and we can create the map with the respective size. When choosing
	//    the second-biggest entry as the first property we save additional allocations later
	sort.Sort(allMsAndProps)
	if len(allMsAndProps) > 2 {
		allMsAndProps[len(allMsAndProps)-2], allMsAndProps[0] = allMsAndProps[0], allMsAndProps[len(allMsAndProps)-2]
	}

	var docMapPairs []docPointerWithScore = nil
	var docMapPairsIndices map[uint64]int = nil
	for i, mAndProps := range allMsAndProps {
		m := mAndProps.MapPairs
		propName := mAndProps.propname

		// The indices are needed for two things:
		// a) combining the results of different properties
		// b) Retrieve additional information that helps to understand the results when debugging. The retrieval is done
		//    in a later step, after it is clear which objects are the most relevant
		//
		// When b) is not needed the results from the last property do not need to be added to the index-map as there
		// won't be any follow-up combinations.
		includeIndicesForLastElement := false
		if additionalExplanations || i < len(allMsAndProps)-1 {
			includeIndicesForLastElement = true
		}

		// only create maps/slices if we know how many entries there are
		if docMapPairs == nil {
			docMapPairs = make([]docPointerWithScore, 0, len(m))
			docMapPairsIndices = make(map[uint64]int, len(m))
			for k, val := range m {
				if len(val.Value) < 8 {
					b.logger.Warnf("Skipping pair in BM25: MapPair.Value should be 8 bytes long, but is %d.", len(val.Value))
					continue
				}
				freqBits := binary.LittleEndian.Uint32(val.Value[0:4])
				propLenBits := binary.LittleEndian.Uint32(val.Value[4:8])
				docMapPairs = append(docMapPairs,
					docPointerWithScore{
						id:         binary.BigEndian.Uint64(val.Key),
						frequency:  math.Float32frombits(freqBits) * propertyBoosts[propName],
						propLength: math.Float32frombits(propLenBits),
					})
				if includeIndicesForLastElement {
					docMapPairsIndices[binary.BigEndian.Uint64(val.Key)] = k
				}
			}
		} else {
			for _, val := range m {
				if len(val.Value) < 8 {
					b.logger.Warnf("Skipping pair in BM25: MapPair.Value should be 8 bytes long, but is %d.", len(val.Value))
					continue
				}
				key := binary.BigEndian.Uint64(val.Key)
				ind, ok := docMapPairsIndices[key]
				freqBits := binary.LittleEndian.Uint32(val.Value[0:4])
				propLenBits := binary.LittleEndian.Uint32(val.Value[4:8])
				if ok {
					if ind >= len(docMapPairs) {
						// the index is not valid anymore, but the key is still in the map
						b.logger.Warnf("Skipping pair in BM25: Index %d is out of range for key %d, length %d.", ind, key, len(docMapPairs))
						continue
					}
					if ind < len(docMapPairs) && docMapPairs[ind].id != key {
						b.logger.Warnf("Skipping pair in BM25: id at %d in doc map pairs, %d, differs from current key, %d", ind, docMapPairs[ind].id, key)
						continue
					}

					docMapPairs[ind].propLength += math.Float32frombits(propLenBits)
					docMapPairs[ind].frequency += math.Float32frombits(freqBits) * propertyBoosts[propName]
				} else {
					docMapPairs = append(docMapPairs,
						docPointerWithScore{
							id:         binary.BigEndian.Uint64(val.Key),
							frequency:  math.Float32frombits(freqBits) * propertyBoosts[propName],
							propLength: math.Float32frombits(propLenBits),
						})
					if includeIndicesForLastElement {
						docMapPairsIndices[binary.BigEndian.Uint64(val.Key)] = len(docMapPairs) - 1 // current last entry
					}
				}
			}
		}
	}
	if docMapPairs == nil {
		termResult.exhausted = true
		return termResult, docMapPairsIndices, nil
	}
	termResult.data = docMapPairs

	n := float64(len(docMapPairs))
	if filterDocIds != nil {
		n += float64(filteredDocIDs.GetCardinality())
	}
	termResult.idf = math.Log(float64(1)+(N-n+0.5)/(n+0.5)) * float64(duplicateTextBoost)

	// catch special case where there are no results and would panic termResult.data[0].id
	// related to #4125
	if len(termResult.data) == 0 {
		termResult.posPointer = 0
		termResult.idPointer = 0
		termResult.exhausted = true
		return termResult, docMapPairsIndices, nil
	}

	termResult.posPointer = 0
	termResult.idPointer = termResult.data[0].id
	return termResult, docMapPairsIndices, nil
}

type term struct {
	// doubles as max impact (with tf=1, the max impact would be 1*idf), if there
	// is a boost for a queryTerm, simply apply it here once
	idf float64

	idPointer  uint64
	posPointer uint64
	data       []docPointerWithScore
	exhausted  bool
	queryTerm  string
}

func (t *term) scoreAndAdvance(averagePropLength float64, config schema.BM25Config) (uint64, float64) {
	id := t.idPointer
	pair := t.data[t.posPointer]
	freq := float64(pair.frequency)
	tf := freq / (freq + config.K1*(1-config.B+config.B*float64(pair.propLength)/averagePropLength))

	// advance
	t.posPointer++
	if t.posPointer >= uint64(len(t.data)) {
		t.exhausted = true
	} else {
		t.idPointer = t.data[t.posPointer].id
	}

	return id, tf * t.idf
}

func (t *term) advanceAtLeast(minID uint64) {
	for t.idPointer < minID {
		t.posPointer++
		if t.posPointer >= uint64(len(t.data)) {
			t.exhausted = true
			return
		}
		t.idPointer = t.data[t.posPointer].id
	}
}

type Terms struct {
	list []term
}

func (t Terms) completelyExhausted() bool {
	for i := range t.list {
		if !t.list[i].exhausted {
			return false
		}
	}
	return true
}

func (t Terms) pivot(minScore float64) bool {
	minID, pivotPoint, abort := t.findMinID(minScore)
	if abort {
		return true
	}
	if pivotPoint == 0 {
		return false
	}

	t.advanceAllAtLeast(minID)
	sort.Sort(t)
	return false
}

func (t Terms) advanceAllAtLeast(minID uint64) {
	for i := range t.list {
		t.list[i].advanceAtLeast(minID)
	}
}

func (t Terms) findMinID(minScore float64) (uint64, int, bool) {
	cumScore := float64(0)

	for i, term := range t.list {
		if term.exhausted {
			continue
		}
		cumScore += term.idf
		if cumScore >= minScore {
			return term.idPointer, i, false
		}
	}

	return 0, 0, true
}

func (t Terms) findFirstNonExhausted() (int, bool) {
	for i := range t.list {
		if !t.list[i].exhausted {
			return i, true
		}
	}

	return -1, false
}

func (t Terms) scoreNext(averagePropLength float64, config schema.BM25Config) (uint64, float64) {
	pos, ok := t.findFirstNonExhausted()
	if !ok {
		// done, nothing left to score
		return 0, 0
	}

	id := t.list[pos].idPointer
	var cumScore float64
	for i := pos; i < len(t.list); i++ {
		if t.list[i].idPointer != id || t.list[i].exhausted {
			continue
		}
		_, score := t.list[i].scoreAndAdvance(averagePropLength, config)
		cumScore += score
	}

	sort.Sort(t) // pointer was advanced in scoreAndAdvance

	return id, cumScore
}

// provide sort interface
func (t Terms) Len() int {
	return len(t.list)
}

func (t Terms) Less(i, j int) bool {
	return t.list[i].idPointer < t.list[j].idPointer
}

func (t Terms) Swap(i, j int) {
	t.list[i], t.list[j] = t.list[j], t.list[i]
}

type MapPairsAndPropName struct {
	propname string
	MapPairs []lsmkv.MapPair
}

type AllMapPairsAndPropName []MapPairsAndPropName

// provide sort interface
func (m AllMapPairsAndPropName) Len() int {
	return len(m)
}

func (m AllMapPairsAndPropName) Less(i, j int) bool {
	return len(m[i].MapPairs) < len(m[j].MapPairs)
}

func (m AllMapPairsAndPropName) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func PropertyHasSearchableIndex(class *models.Class, tentativePropertyName string) bool {
	if class == nil {
		return false
	}

	propertyName := strings.Split(tentativePropertyName, "^")[0]
	p, err := schema.GetPropertyByName(class, propertyName)
	if err != nil {
		return false
	}
	return HasSearchableIndex(p)
}
