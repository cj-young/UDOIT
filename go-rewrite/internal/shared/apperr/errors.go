package apperr

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeValidation         Code = "VALIDATION"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodePreconditionFailed Code = "PRECONDITION_FAILED"
	CodeInternal           Code = "INTERNAL"
)

type AppError struct {
	Code    Code
	Reason  string
	Message string
	Details map[string]any
	op      string
	cause   error
}

type Option func(*AppError)

func (a *AppError) Error() string {
	msg := a.Message

	if a.op != "" {
		msg = a.op + ": " + msg
	}

	if a.cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, a.cause)
	}

	return msg
}

func (a *AppError) Unwrap() error {
	return a.cause
}

func (a *AppError) Op() string {
	return a.op
}

func New(code Code, reason, message string, opts ...Option) *AppError {
	e := &AppError{
		Code:    code,
		Reason:  reason,
		Message: message,
		Details: make(map[string]any),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func WithOp(op string) Option {
	return func(e *AppError) {
		e.op = op
	}
}

func WithCause(cause error) Option {
	return func(e *AppError) {
		e.cause = cause
	}
}

func WithDetails(details map[string]any) Option {
	return func(e *AppError) {
		e.Details = details
	}
}

func Internal(op string, cause error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Reason:  "internal_error",
		Message: "an unexpected error occurred",
		Details: make(map[string]any),
		op:      op,
		cause:   cause,
	}
}

func IsCode(err error, code Code) bool {
	e, ok := errors.AsType[*AppError](err)
	return ok && e.Code == code
}

func IsAppError(err error) bool {
	_, ok := errors.AsType[*AppError](err)
	return ok
}