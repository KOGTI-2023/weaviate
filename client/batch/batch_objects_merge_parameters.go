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

package batch

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewBatchObjectsMergeParams creates a new BatchObjectsMergeParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewBatchObjectsMergeParams() *BatchObjectsMergeParams {
	return &BatchObjectsMergeParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewBatchObjectsMergeParamsWithTimeout creates a new BatchObjectsMergeParams object
// with the ability to set a timeout on a request.
func NewBatchObjectsMergeParamsWithTimeout(timeout time.Duration) *BatchObjectsMergeParams {
	return &BatchObjectsMergeParams{
		timeout: timeout,
	}
}

// NewBatchObjectsMergeParamsWithContext creates a new BatchObjectsMergeParams object
// with the ability to set a context for a request.
func NewBatchObjectsMergeParamsWithContext(ctx context.Context) *BatchObjectsMergeParams {
	return &BatchObjectsMergeParams{
		Context: ctx,
	}
}

// NewBatchObjectsMergeParamsWithHTTPClient creates a new BatchObjectsMergeParams object
// with the ability to set a custom HTTPClient for a request.
func NewBatchObjectsMergeParamsWithHTTPClient(client *http.Client) *BatchObjectsMergeParams {
	return &BatchObjectsMergeParams{
		HTTPClient: client,
	}
}

/*
BatchObjectsMergeParams contains all the parameters to send to the API endpoint

	for the batch objects merge operation.

	Typically these are written to a http.Request.
*/
type BatchObjectsMergeParams struct {

	// Body.
	Body BatchObjectsMergeBody

	/* ConsistencyLevel.

	   Determines how many replicas must acknowledge a request before it is considered successful
	*/
	ConsistencyLevel *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the batch objects merge params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *BatchObjectsMergeParams) WithDefaults() *BatchObjectsMergeParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the batch objects merge params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *BatchObjectsMergeParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the batch objects merge params
func (o *BatchObjectsMergeParams) WithTimeout(timeout time.Duration) *BatchObjectsMergeParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the batch objects merge params
func (o *BatchObjectsMergeParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the batch objects merge params
func (o *BatchObjectsMergeParams) WithContext(ctx context.Context) *BatchObjectsMergeParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the batch objects merge params
func (o *BatchObjectsMergeParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the batch objects merge params
func (o *BatchObjectsMergeParams) WithHTTPClient(client *http.Client) *BatchObjectsMergeParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the batch objects merge params
func (o *BatchObjectsMergeParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the batch objects merge params
func (o *BatchObjectsMergeParams) WithBody(body BatchObjectsMergeBody) *BatchObjectsMergeParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the batch objects merge params
func (o *BatchObjectsMergeParams) SetBody(body BatchObjectsMergeBody) {
	o.Body = body
}

// WithConsistencyLevel adds the consistencyLevel to the batch objects merge params
func (o *BatchObjectsMergeParams) WithConsistencyLevel(consistencyLevel *string) *BatchObjectsMergeParams {
	o.SetConsistencyLevel(consistencyLevel)
	return o
}

// SetConsistencyLevel adds the consistencyLevel to the batch objects merge params
func (o *BatchObjectsMergeParams) SetConsistencyLevel(consistencyLevel *string) {
	o.ConsistencyLevel = consistencyLevel
}

// WriteToRequest writes these params to a swagger request
func (o *BatchObjectsMergeParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if o.ConsistencyLevel != nil {

		// query param consistency_level
		var qrConsistencyLevel string

		if o.ConsistencyLevel != nil {
			qrConsistencyLevel = *o.ConsistencyLevel
		}
		qConsistencyLevel := qrConsistencyLevel
		if qConsistencyLevel != "" {

			if err := r.SetQueryParam("consistency_level", qConsistencyLevel); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}