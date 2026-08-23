package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Logger logs each HTTP request with method, path, status, and duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		r = r.WithContext(ctx)
		rw.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(rw, r)

		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestID,
			"ip", realIP(r),
		)
	})
}

// statusRecorder wraps ResponseWriter to capture the status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}
