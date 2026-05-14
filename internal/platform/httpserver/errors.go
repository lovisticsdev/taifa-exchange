package httpserver

import (
	"net/http"
)

const (
	ErrorCodeInvalidJSON  = "INVALID_JSON"
	ErrorCodeValidation   = "VALIDATION_ERROR"
	ErrorCodeNotFound     = "NOT_FOUND"
	ErrorCodeConflict     = "CONFLICT"
	ErrorCodeUnauthorized = "UNAUTHORIZED"
	ErrorCodeForbidden    = "FORBIDDEN"
	ErrorCodeInternal     = "INTERNAL_ERROR"
)

type ErrorBody struct {
	Code          string `json:"code"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}

type ErrorEnvelope struct {
	CorrelationID string    `json:"correlation_id"`
	Error         ErrorBody `json:"error"`
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	correlationID := CorrelationIDFromContext(r.Context())

	WriteJSON(w, r, statusCode, ErrorEnvelope{
		CorrelationID: correlationID,
		Error: ErrorBody{
			Code:          code,
			CorrelationID: correlationID,
			Message:       message,
		},
	})
}
