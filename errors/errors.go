// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "unsafe" // required for go:linkname
)

// New returns an error with the supplied message.
func New(message string) error {
	return &Status{code: 2, msg: message}
}

// -----------------------------------------------------------------------------
// STATUS

type Status struct {
	code Code
	msg  string
}

func (e *Status) Error() string {
	return e.msg
}

func (e *Status) Code() Code {
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

func (e *Status) HttpCode() int {
	return httpCode[e.code]
}

// -----------------------------------------------------------------------------
// ERROR CODES

type Code uint32

const (
	OK              Code = iota // ok
	CANCELLED                   // operation cancelled by caller
	UNKNOWN                     // unknown error
	INVALID                     // invalid argument from caller
	DEADLINE                    // operation deadline exceeded
	NOTFOUND                    // entity not found
	EXISTS                      // entity already exists
	PERMISSION                  // permission denied, caller does not have permission
	EXHAUSTED                   // resource exhausted beyond allowable limit
	FAILED                      // operation failed preconditions
	ABORTED                     // operation aborted by the system
	RANGE                       // operation out of valid range
	UNIMPLEMENTED               // operation not implemented or not supported
	INTERNAL                    // internal system error
	UNAVAILABLE                 // service unavailable, try again later
	DATALOSS                    // unrecoverable data loss or corruption
	UNAUTHENTICATED             // caller is not authenticated
)

var statusText = map[Code]string{
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

var httpCode = map[Code]int{
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

var postgresCode = map[string]Code{
	"00":    OK,
	"01":    ABORTED,
	"02":    NOTFOUND,
	"03":    UNAVAILABLE,
	"08":    UNAVAILABLE,
	"0A":    UNIMPLEMENTED,
	"0L":    PERMISSION,
	"0P":    UNIMPLEMENTED,
	"20":    INVALID,
	"21":    EXISTS,
	"22":    INVALID,
	"23":    INVALID,
	"26":    EXISTS,
	"27":    INTERNAL,
	"28":    UNAUTHENTICATED,
	"53":    EXHAUSTED,
	"54":    EXHAUSTED,
	"55":    FAILED,
	"23505": EXISTS,
	"25P03": DEADLINE,
	"42501": PERMISSION,
	"42P01": UNIMPLEMENTED,
	"57014": CANCELLED,
}

func (c Code) String() string {
	return statusText[c]
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
// DATABASE ERROR METHODS

func Postgres(err error, message string) error {
	if e, ok := err.(*pgx.PgError); ok {
		message += ": " + e.Detail
		if code, ok := postgresCode[e.Code]; ok {
			return &Status{code: code, msg: message}
		}
		if code, ok := postgresCode[e.Code[:2]]; ok {
			return &Status{code: code, msg: message}
		}
	}
	return Internal(message)
}

// -----------------------------------------------------------------------------
// MESSAGE FORMATTING

func Msg(format string, a ...any) string {
	if len(a) == 0 {
		return format
	}
	return fmt.Sprintf(format, a...)
}

// -----------------------------------------------------------------------------
// STUBS

//go:noescape
//go:linkname Unwrap errors.Unwrap
func Unwrap(err error) error

//go:noescape
//go:linkname Is errors.Is
func Is(err, target error) bool

//go:noescape
//go:linkname As errors.As
func As(err error, target interface{}) bool

//go:noescape
//go:linkname Join errors.Join
func Join(errs ...error) error
