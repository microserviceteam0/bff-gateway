package apperr

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("access denied")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrInternal           = errors.New("internal server error")
	ErrTimeout            = errors.New("request timeout")
)
