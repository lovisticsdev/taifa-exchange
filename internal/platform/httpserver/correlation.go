package httpserver

import (
	"context"
	"net/http"
	"strings"

	"taifa-exchange/internal/platform/ids"
)

const correlationIDHeader = "X-Correlation-ID"

type correlationIDContextKey struct{}

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := strings.TrimSpace(r.Header.Get(correlationIDHeader))
		if correlationID == "" {
			correlationID = ids.NewCorrelationID()
		}

		w.Header().Set(correlationIDHeader, correlationID)

		ctx := context.WithValue(r.Context(), correlationIDContextKey{}, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	correlationID, ok := ctx.Value(correlationIDContextKey{}).(string)
	if !ok {
		return ""
	}

	return correlationID
}
