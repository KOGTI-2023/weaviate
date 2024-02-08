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

// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// BatchStats The summary of a nodes batch queue congestion status.
//
// swagger:model BatchStats
type BatchStats struct {

	// How many objects are in the processing step before being added to the batch queue.
	CurrentlyProcessedObjects int64 `json:"currentlyProcessedObjects"`

	// How many objects are currently in the batch queue.
	QueueLength *int64 `json:"queueLength,omitempty"`

	// How many objects are approximately processed from the batch queue per second.
	RatePerSecond int64 `json:"ratePerSecond"`
}

// Validate validates this batch stats
func (m *BatchStats) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this batch stats based on context it is used
func (m *BatchStats) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *BatchStats) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *BatchStats) UnmarshalBinary(b []byte) error {
	var res BatchStats
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
