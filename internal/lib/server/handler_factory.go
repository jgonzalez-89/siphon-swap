package server

import (
	"context"
	"cryptoswap/internal/lib/logger"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func NewHandlerFactory(ctx context.Context, logger logger.Logger) *handlerFactory {
	return &handlerFactory{
		ctx:    ctx,
		logger: logger,
	}
}

type handlerFactory struct {
	logger logger.Logger
	ctx    context.Context
}

func (f *handlerFactory) New(handler any, registerFunc RegisterFunc, swaggerF SwaggerFunc) Handler {
	swagger, err := swaggerF()
	if err != nil {
		f.logger.Fatalf(f.ctx, "error getting swagger: %v", err)
		return Handler{}
	}
	return Handler{
		Handler:      handler,
		RegisterFunc: registerFunc,
		Swagger:      swagger,
	}
}

type RegisterFunc func(router gin.IRouter, handler any)

type SwaggerFunc func() (*openapi3.T, error)

type Handler struct {
	Handler      any
	RegisterFunc RegisterFunc
	Swagger      *openapi3.T
}
