package httpserver

import (
	"log/slog"
	"net/http"
)

func RecovererMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error(
						"panic recovered",
						"panic", recovered,
						"method", r.Method,
						"path", r.URL.Path,
						"correlation_id", CorrelationIDFromContext(r.Context()),
					)

					WriteError(
						w,
						r,
						http.StatusInternalServerError,
						ErrorCodeInternal,
						"An internal error occurred.",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
