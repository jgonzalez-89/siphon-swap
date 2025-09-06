package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterFunc func(router gin.IRouter, handler any, options any)

type Handler struct {
	Handler      any
	Options      any
	RegisterFunc RegisterFunc
}

type ServerConfig struct {
	Port string
}

func NewServerBuilder(router *gin.Engine, config ServerConfig) *serverBuilder {
	return &serverBuilder{
		handlers: []Handler{},
		router:   router,
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
		handler.RegisterFunc(b.router, handler.Handler, handler.Options)
	}

	return &http.Server{
		Addr:    ":" + b.config.Port,
		Handler: b.router,
	}
}
