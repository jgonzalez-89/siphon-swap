package server

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/gin-middleware"
)

type ServerConfig struct {
	Port string
}

func NewServerBuilder(router *gin.Engine, config ServerConfig) *serverBuilder {
	return &serverBuilder{
		handlers: []Handler{},
		router:   router,
		config:   config,
	}
}

type serverBuilder struct {
	handlers    []Handler
	middlewares []gin.HandlerFunc
	router      *gin.Engine
	config      ServerConfig
}

func (b *serverBuilder) WithHandlers(handlerFunc ...Handler) *serverBuilder {
	b.handlers = append(b.handlers, handlerFunc...)
	return b
}

func (b *serverBuilder) WithMiddlewares(handlerFunc ...gin.HandlerFunc) *serverBuilder {
	b.middlewares = append(b.middlewares, handlerFunc...)
	return b
}

func (b *serverBuilder) Build() *http.Server {
	for _, middleware := range b.middlewares {
		b.router.Use(middleware)
	}

	for _, handler := range b.handlers {
		r := b.router.Group("/")
		r.Use(ginmiddleware.OapiRequestValidator(handler.Swagger))
		handler.RegisterFunc(b.router.Group("/"), handler.Handler)
	}

	return &http.Server{
		Addr:    ":" + b.config.Port,
		Handler: b.router,
	}
}
