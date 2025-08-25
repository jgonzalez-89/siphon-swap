package middlewares

import (
	"cryptoswap/internal/lib/constants"
	"cryptoswap/internal/lib/logger"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func LoggingMiddleware(logger logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx := constants.NewContextWithRequestId()
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

			// No loguear archivos estáticos
			if r.URL.Path != "/" && r.URL.Path != "/favicon.ico" {
				logger.Infof(ctx, "%s %s %d %v", r.Method, r.URL.Path,
					r.Response.StatusCode, time.Since(start))
			}
		})
	}
}
