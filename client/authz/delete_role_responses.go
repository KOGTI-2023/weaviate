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

package authz

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/weaviate/weaviate/entities/models"
)

// DeleteRoleReader is a Reader for the DeleteRole structure.
type DeleteRoleReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteRoleReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDeleteRoleNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewDeleteRoleBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewDeleteRoleUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewDeleteRoleForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewDeleteRoleInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDeleteRoleNoContent creates a DeleteRoleNoContent with default headers values
func NewDeleteRoleNoContent() *DeleteRoleNoContent {
	return &DeleteRoleNoContent{}
}

/*
DeleteRoleNoContent describes a response with status code 204, with default header values.

Successfully deleted.
*/
type DeleteRoleNoContent struct {
}

// IsSuccess returns true when this delete role no content response has a 2xx status code
func (o *DeleteRoleNoContent) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this delete role no content response has a 3xx status code
func (o *DeleteRoleNoContent) IsRedirect() bool {
	return false
}

// IsClientError returns true when this delete role no content response has a 4xx status code
func (o *DeleteRoleNoContent) IsClientError() bool {
	return false
}

// IsServerError returns true when this delete role no content response has a 5xx status code
func (o *DeleteRoleNoContent) IsServerError() bool {
	return false
}

// IsCode returns true when this delete role no content response a status code equal to that given
func (o *DeleteRoleNoContent) IsCode(code int) bool {
	return code == 204
}

// Code gets the status code for the delete role no content response
func (o *DeleteRoleNoContent) Code() int {
	return 204
}

func (o *DeleteRoleNoContent) Error() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleNoContent ", 204)
}

func (o *DeleteRoleNoContent) String() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleNoContent ", 204)
}

func (o *DeleteRoleNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteRoleBadRequest creates a DeleteRoleBadRequest with default headers values
func NewDeleteRoleBadRequest() *DeleteRoleBadRequest {
	return &DeleteRoleBadRequest{}
}

/*
DeleteRoleBadRequest describes a response with status code 400, with default header values.

Bad request
*/
type DeleteRoleBadRequest struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this delete role bad request response has a 2xx status code
func (o *DeleteRoleBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this delete role bad request response has a 3xx status code
func (o *DeleteRoleBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this delete role bad request response has a 4xx status code
func (o *DeleteRoleBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this delete role bad request response has a 5xx status code
func (o *DeleteRoleBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this delete role bad request response a status code equal to that given
func (o *DeleteRoleBadRequest) IsCode(code int) bool {
	return code == 400
}

// Code gets the status code for the delete role bad request response
func (o *DeleteRoleBadRequest) Code() int {
	return 400
}

func (o *DeleteRoleBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleBadRequest  %+v", 400, o.Payload)
}

func (o *DeleteRoleBadRequest) String() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleBadRequest  %+v", 400, o.Payload)
}

func (o *DeleteRoleBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteRoleUnauthorized creates a DeleteRoleUnauthorized with default headers values
func NewDeleteRoleUnauthorized() *DeleteRoleUnauthorized {
	return &DeleteRoleUnauthorized{}
}

/*
DeleteRoleUnauthorized describes a response with status code 401, with default header values.

Unauthorized or invalid credentials.
*/
type DeleteRoleUnauthorized struct {
}

// IsSuccess returns true when this delete role unauthorized response has a 2xx status code
func (o *DeleteRoleUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this delete role unauthorized response has a 3xx status code
func (o *DeleteRoleUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this delete role unauthorized response has a 4xx status code
func (o *DeleteRoleUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this delete role unauthorized response has a 5xx status code
func (o *DeleteRoleUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this delete role unauthorized response a status code equal to that given
func (o *DeleteRoleUnauthorized) IsCode(code int) bool {
	return code == 401
}

// Code gets the status code for the delete role unauthorized response
func (o *DeleteRoleUnauthorized) Code() int {
	return 401
}

func (o *DeleteRoleUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleUnauthorized ", 401)
}

func (o *DeleteRoleUnauthorized) String() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleUnauthorized ", 401)
}

func (o *DeleteRoleUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteRoleForbidden creates a DeleteRoleForbidden with default headers values
func NewDeleteRoleForbidden() *DeleteRoleForbidden {
	return &DeleteRoleForbidden{}
}

/*
DeleteRoleForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type DeleteRoleForbidden struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this delete role forbidden response has a 2xx status code
func (o *DeleteRoleForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this delete role forbidden response has a 3xx status code
func (o *DeleteRoleForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this delete role forbidden response has a 4xx status code
func (o *DeleteRoleForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this delete role forbidden response has a 5xx status code
func (o *DeleteRoleForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this delete role forbidden response a status code equal to that given
func (o *DeleteRoleForbidden) IsCode(code int) bool {
	return code == 403
}

// Code gets the status code for the delete role forbidden response
func (o *DeleteRoleForbidden) Code() int {
	return 403
}

func (o *DeleteRoleForbidden) Error() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleForbidden  %+v", 403, o.Payload)
}

func (o *DeleteRoleForbidden) String() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleForbidden  %+v", 403, o.Payload)
}

func (o *DeleteRoleForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteRoleInternalServerError creates a DeleteRoleInternalServerError with default headers values
func NewDeleteRoleInternalServerError() *DeleteRoleInternalServerError {
	return &DeleteRoleInternalServerError{}
}

/*
DeleteRoleInternalServerError describes a response with status code 500, with default header values.

An error has occurred while trying to fulfill the request. Most likely the ErrorResponse will contain more information about the error.
*/
type DeleteRoleInternalServerError struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this delete role internal server error response has a 2xx status code
func (o *DeleteRoleInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this delete role internal server error response has a 3xx status code
func (o *DeleteRoleInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this delete role internal server error response has a 4xx status code
func (o *DeleteRoleInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this delete role internal server error response has a 5xx status code
func (o *DeleteRoleInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this delete role internal server error response a status code equal to that given
func (o *DeleteRoleInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the delete role internal server error response
func (o *DeleteRoleInternalServerError) Code() int {
	return 500
}

func (o *DeleteRoleInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleInternalServerError  %+v", 500, o.Payload)
}

func (o *DeleteRoleInternalServerError) String() string {
	return fmt.Sprintf("[DELETE /authz/roles/{id}][%d] deleteRoleInternalServerError  %+v", 500, o.Payload)
}

func (o *DeleteRoleInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}