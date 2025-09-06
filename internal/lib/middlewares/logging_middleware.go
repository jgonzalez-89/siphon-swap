package middlewares

import (
	"cryptoswap/internal/lib/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs the end of the request with detailed information
func LoggingMiddleware(logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		startTime := time.Now()
		logger.Infof(ctx, "[%s] %s", ctx.Request.Method, ctx.Request.URL.Path)
		ctx.Next()

		duration := time.Since(startTime)

		// Don't log static files
		if strings.HasPrefix(ctx.Request.URL.Path, "/api") {
			// Log with appropriate level and detailed information
			logger.Infof(ctx, "[%s] %s %s - %d ms",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				ctx.Writer.Status(),
				duration.Milliseconds(),
			)
		}
	})
}
