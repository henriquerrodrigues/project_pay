package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"project_pay/cmd/internal/traceid"
)

// statusRecorder captura o status code escrito pelo handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// RequestLogger gera um trace ID (UUID v4) por request, injeta no context
// e no header X-Trace-Id, e loga método/rota/status/duração ao final.
// Envolve o mux inteiro.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tid := uuid.NewString()

		w.Header().Set("X-Trace-Id", tid)
		ctx := traceid.NewContext(r.Context(), tid)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(ctx))

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"trace_id", tid,
		)
	})
}
