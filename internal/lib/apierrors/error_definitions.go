package apierrors

import "net/http"

var (
	InternalServerError = ErrorDefinition{
		Code:    http.StatusInternalServerError,
		message: "Internal server error",
	}
	BadRequestError = ErrorDefinition{
		Code:    http.StatusBadRequest,
		message: "Bad request",
	}
	NotFoundError = ErrorDefinition{
		Code:    http.StatusNotFound,
		message: "Not found",
	}
)
