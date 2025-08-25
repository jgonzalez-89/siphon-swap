package middlewares

import (
	"cryptoswap/internal/lib/constants"
	"cryptoswap/internal/lib/logger"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// LoggingMiddleware logs the end of the request with detailed information
func LoggingMiddleware(logger logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := constants.NewContextWithRequestId()
			r = r.WithContext(ctx)
			startTime := time.Now()

			// Create a custom response writer to capture status code and response size
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code
			}

			next.ServeHTTP(wrappedWriter, r)

			// Don't log static files
			if strings.HasPrefix(r.URL.Path, "/api") {
				// Log with appropriate level and detailed information
				logger.Infof(ctx, "✅ [%s] %s %s - %d ms",
					r.Method,
					r.URL.Path,
					http.StatusText(wrappedWriter.statusCode),
					time.Since(startTime).Milliseconds())
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
