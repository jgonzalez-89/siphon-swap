package middlewares

import (
	"cryptoswap/internal/lib/constants"
	"cryptoswap/internal/lib/ids"
	"cryptoswap/internal/lib/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs the end of the request with detailed information
func LoggingMiddleware(logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		setRequestId(ctx)

		startTime := time.Now()
		logger.Infof(ctx, "Incoming [%s] request to %s", ctx.Request.Method, ctx.Request.URL.Path)
		ctx.Next()

		duration := time.Since(startTime)

		logger.Infof(ctx, "Finished [%s] request to %s with %d status code in %d ms",
			ctx.Request.Method,
			ctx.Request.URL.Path,
			ctx.Writer.Status(),
			duration.Milliseconds(),
		)
	})
}

func setRequestId(ctx *gin.Context) {
	if _, ok := ctx.Get(constants.RequestId); !ok {
		ctx.Set(constants.RequestId, ids.NewRequestId())
	}
}
