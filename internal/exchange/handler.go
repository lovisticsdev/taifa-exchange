package exchange

import (
	"log/slog"
	"net/http"
	"strings"

	"taifa-exchange/internal/platform/httpserver"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	var request AuthorizeRequest

	if err := httpserver.DecodeJSON(r, &request); err != nil {
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeInvalidJSON,
			"Request body must be valid JSON.",
		)
		return
	}

	if h == nil || h.service == nil {
		httpserver.WriteError(
			w,
			r,
			http.StatusServiceUnavailable,
			httpserver.ErrorCodeInternal,
			"Exchange authorization service is unavailable.",
		)
		return
	}

	result, err := h.service.Authorize(r.Context(), AuthorizeInput{
		Token:         bearerTokenFromRequest(r),
		CorrelationID: httpserver.CorrelationIDFromContext(r.Context()),
		Request:       request,
	})
	if err != nil {
		h.logger.Error(
			"exchange authorization failed",
			"error", err,
			"correlation_id", httpserver.CorrelationIDFromContext(r.Context()),
		)

		httpserver.WriteError(
			w,
			r,
			http.StatusInternalServerError,
			httpserver.ErrorCodeInternal,
			"Exchange could not complete authorization.",
		)
		return
	}

	if result.DecisionRecord.Decision == DecisionAllow {
		httpserver.WriteData(w, r, http.StatusOK, result.Response)
		return
	}

	httpserver.WriteError(
		w,
		r,
		result.HTTPStatusCode,
		result.ErrorCode,
		result.ErrorMessage,
	)
}

func bearerTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}

	parts := strings.Fields(header)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
