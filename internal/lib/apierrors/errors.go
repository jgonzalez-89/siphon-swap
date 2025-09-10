package apierrors

import (
	"fmt"
)

const (
	errorTmpl = "ErrorCode (%d): %s"
)

type ErrorDefinition struct {
	Code    int
	message string
}

type ApiError struct {
	ErrorDefinition
	err error
}

func (e *ApiError) Error() string {
	return fmt.Sprintf(errorTmpl+" reason: %s", e.Code, e.message, e.err.Error())
}

func NewApiError(def ErrorDefinition, err error) *ApiError {
	return &ApiError{
		ErrorDefinition: def,
		err:             err,
	}
}
