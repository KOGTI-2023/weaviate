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

package offloads

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

	"github.com/weaviate/weaviate/entities/models"
)

// NewOffloadParams creates a new OffloadParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewOffloadParams() *OffloadParams {
	return &OffloadParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewOffloadParamsWithTimeout creates a new OffloadParams object
// with the ability to set a timeout on a request.
func NewOffloadParamsWithTimeout(timeout time.Duration) *OffloadParams {
	return &OffloadParams{
		timeout: timeout,
	}
}

// NewOffloadParamsWithContext creates a new OffloadParams object
// with the ability to set a context for a request.
func NewOffloadParamsWithContext(ctx context.Context) *OffloadParams {
	return &OffloadParams{
		Context: ctx,
	}
}

// NewOffloadParamsWithHTTPClient creates a new OffloadParams object
// with the ability to set a custom HTTPClient for a request.
func NewOffloadParamsWithHTTPClient(client *http.Client) *OffloadParams {
	return &OffloadParams{
		HTTPClient: client,
	}
}

/*
OffloadParams contains all the parameters to send to the API endpoint

	for the offload operation.

	Typically these are written to a http.Request.
*/
type OffloadParams struct {

	/* Backend.

	   Offload backend name e.g. filesystem, gcs, s3.
	*/
	Backend string

	// Body.
	Body *models.OffloadRequest

	/* Class.

	   Class (name) offloaded tenants belong to
	*/
	Class string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the offload params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *OffloadParams) WithDefaults() *OffloadParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the offload params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *OffloadParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the offload params
func (o *OffloadParams) WithTimeout(timeout time.Duration) *OffloadParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the offload params
func (o *OffloadParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the offload params
func (o *OffloadParams) WithContext(ctx context.Context) *OffloadParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the offload params
func (o *OffloadParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the offload params
func (o *OffloadParams) WithHTTPClient(client *http.Client) *OffloadParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the offload params
func (o *OffloadParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBackend adds the backend to the offload params
func (o *OffloadParams) WithBackend(backend string) *OffloadParams {
	o.SetBackend(backend)
	return o
}

// SetBackend adds the backend to the offload params
func (o *OffloadParams) SetBackend(backend string) {
	o.Backend = backend
}

// WithBody adds the body to the offload params
func (o *OffloadParams) WithBody(body *models.OffloadRequest) *OffloadParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the offload params
func (o *OffloadParams) SetBody(body *models.OffloadRequest) {
	o.Body = body
}

// WithClass adds the class to the offload params
func (o *OffloadParams) WithClass(class string) *OffloadParams {
	o.SetClass(class)
	return o
}

// SetClass adds the class to the offload params
func (o *OffloadParams) SetClass(class string) {
	o.Class = class
}

// WriteToRequest writes these params to a swagger request
func (o *OffloadParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param backend
	if err := r.SetPathParam("backend", o.Backend); err != nil {
		return err
	}
	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	// path param class
	if err := r.SetPathParam("class", o.Class); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
