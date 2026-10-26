package middleware

import (
	"net/http"
	"time"

	"github.com/PhamVanPhuc2k2/manage-company/pkg/logger"
	"github.com/rs/zerolog"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(base zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := chimw.GetReqID(r.Context())

			l := base.With().Str("request_id", requestID).Str("method", r.Method).Str("path", r.URL.Path).Logger()

			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			w.Header().Set("X-Request-Id", requestID)

			next.ServeHTTP(ww, r.WithContext(logger.WithContext(r.Context(), l)))

			event := l.Info()
			if ww.Status() >= 500 {
				event = l.Error()
			} else if ww.Status() >= 400 {
				event = l.Warn()
			}

			event.Int("status", ww.Status()).Int("bytes", ww.BytesWritten()).Dur("duration", time.Since(start)).Str("remote_ip", r.RemoteAddr).Msg("HTTP request")
		})
	}
}
