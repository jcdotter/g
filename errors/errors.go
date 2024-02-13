// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/grpg/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New returns an error with the supplied message.
func New(message string) error {
	return &Status{code: 500, msg: message}
}

// -----------------------------------------------------------------------------
// STATUS

type Status struct {
	code uint32
	msg  string
}

func (e *Status) Error() string {
	return e.msg
}

func (e *Status) Code() uint32 {
	return e.code
}

func (e *Status) Status() string {
	return statusText[e.code]
}

func (e *Status) String() string {
	return e.Status() + ": " + e.msg
}

func (e *Status) GprcError() error {
	return status.Error(codes.Code(e.code), e.msg)
}

func (e *Status) GprcStatus() *status.Status {
	return status.New(codes.Code(e.code), e.msg)
}

func (e *Status) HttpError(w http.ResponseWriter) {
	http.Error(w, e.String(), httpCode[e.code])
}

// -----------------------------------------------------------------------------
// ERROR CODES

const (
	OK              uint32 = iota // ok
	CANCELLED                     // operation cancelled by caller
	UNKNOWN                       // unknown error
	INVALID                       // invalid argument from caller
	DEADLINE                      // operation deadline exceeded
	NOTFOUND                      // entity not found
	EXISTS                        // entity already exists
	PERMISSION                    // permission denied, caller does not have permission
	EXHAUSTED                     // resource exhausted beyond allowable limit
	FAILED                        // operation failed preconditions
	ABORTED                       // operation aborted by the system
	RANGE                         // operation out of valid range
	UNIMPLEMENTED                 // operation not implemented or not supported
	INTERNAL                      // internal system error
	UNAVAILABLE                   // service unavailable, try again later
	DATALOSS                      // unrecoverable data loss or corruption
	UNAUTHENTICATED               // caller is not authenticated
)

var statusText = map[uint32]string{
	OK:              "OK",
	CANCELLED:       "CANCELLED",
	UNKNOWN:         "UNKNOWN",
	INVALID:         "INVALID",
	DEADLINE:        "DEADLINE",
	NOTFOUND:        "NOTFOUND",
	EXISTS:          "EXISTS",
	PERMISSION:      "PERMISSION",
	EXHAUSTED:       "EXHAUSTED",
	FAILED:          "FAILED",
	ABORTED:         "ABORTED",
	RANGE:           "RANGE",
	UNIMPLEMENTED:   "UNIMPLEMENTED",
	INTERNAL:        "INTERNAL",
	UNAVAILABLE:     "UNAVAILABLE",
	DATALOSS:        "DATALOSS",
	UNAUTHENTICATED: "UNAUTHENTICATED",
}

var httpCode = map[uint32]int{
	OK:              http.StatusOK,
	CANCELLED:       http.StatusRequestTimeout,
	UNKNOWN:         http.StatusInternalServerError,
	INVALID:         http.StatusBadRequest,
	DEADLINE:        http.StatusGatewayTimeout,
	NOTFOUND:        http.StatusNotFound,
	EXISTS:          http.StatusConflict,
	PERMISSION:      http.StatusForbidden,
	EXHAUSTED:       http.StatusTooManyRequests,
	FAILED:          http.StatusPreconditionFailed,
	ABORTED:         http.StatusConflict,
	RANGE:           http.StatusBadRequest,
	UNIMPLEMENTED:   http.StatusNotImplemented,
	INTERNAL:        http.StatusInternalServerError,
	UNAVAILABLE:     http.StatusServiceUnavailable,
	DATALOSS:        http.StatusInternalServerError,
	UNAUTHENTICATED: http.StatusUnauthorized,
}

func StatusText(code uint32) string {
	return statusText[code]
}

// -----------------------------------------------------------------------------
// SERVER ERROR METHODS

// Cancelled returns an error representing the cancellation of an operation.
func Cancelled(message string) error {
	return &Status{code: CANCELLED, msg: message}
}

// Unknown returns an error representing an unknown error.
func Unknown(message string) error {
	return &Status{code: UNKNOWN, msg: message}
}

// Invalid returns an error representing an invalid argument.
func Invalid(message string) error {
	return &Status{code: INVALID, msg: message}
}

// Deadline returns an error representing a deadline exceeded.
func Deadline(message string) error {
	return &Status{code: DEADLINE, msg: message}
}

// NotFound returns an error representing an entity not found.
func NotFound(message string) error {
	return &Status{code: NOTFOUND, msg: message}
}

// Exists returns an error representing an entity already exists.
func Exists(message string) error {
	return &Status{code: EXISTS, msg: message}
}

// Permission returns an error representing a permission denied.
func Permission(message string) error {
	return &Status{code: PERMISSION, msg: message}
}

// Exhausted returns an error representing a resource exhausted.
func Exhausted(message string) error {
	return &Status{code: EXHAUSTED, msg: message}
}

// Failed returns an error representing an operation failed preconditions.
func Failed(message string) error {
	return &Status{code: FAILED, msg: message}
}

// Aborted returns an error representing an operation aborted by the system.
func Aborted(message string) error {
	return &Status{code: ABORTED, msg: message}
}

// Range returns an error representing an operation out of valid range.
func Range(message string) error {
	return &Status{code: RANGE, msg: message}
}

// Unimplemented returns an error representing an operation not implemented or not supported.
func Unimplemented(message string) error {
	return &Status{code: UNIMPLEMENTED, msg: message}
}

// Internal returns an error representing an internal system error.
func Internal(message string) error {
	return &Status{code: INTERNAL, msg: message}
}

// Unavailable returns an error representing a service unavailable.
func Unavailable(message string) error {
	return &Status{code: UNAVAILABLE, msg: message}
}

// DataLoss returns an error representing unrecoverable data loss or corruption.
func DataLoss(message string) error {
	return &Status{code: DATALOSS, msg: message}
}

// Unauthenticated returns an error representing a caller not authenticated.
func Unauthenticated(message string) error {
	return &Status{code: UNAUTHENTICATED, msg: message}
}

// -----------------------------------------------------------------------------
// INTERNAL ERROR METHODS

var ()
