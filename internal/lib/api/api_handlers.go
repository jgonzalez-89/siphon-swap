package api

import (
	"cryptoswap/internal/lib/apierrors"

	"github.com/gin-gonic/gin"
)

type ResponseHandler interface {
	OK(c *gin.Context, status int, data any)
	Error(c *gin.Context, err *apierrors.ApiError)
}

type responseManager struct{}

func NewResponseManager() ResponseHandler {
	return &responseManager{}
}

func (r *responseManager) OK(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func (r *responseManager) Error(c *gin.Context, err *apierrors.ApiError) {
	c.JSON(err.Code, gin.H{"error": err.Error()})
}
