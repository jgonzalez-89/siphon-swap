package apierrors

import "net/http"

var (
	InternalServer = ErrorDefinition{
		Code:    http.StatusInternalServerError,
		message: "Internal server error",
	}
	BadRequest = ErrorDefinition{
		Code:    http.StatusBadRequest,
		message: "Bad request",
	}
	NotFound = ErrorDefinition{
		Code:    http.StatusNotFound,
		message: "Not found",
	}
)
