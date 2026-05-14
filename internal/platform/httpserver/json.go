package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type DataEnvelope struct {
	CorrelationID string `json:"correlation_id"`
	Data          any    `json:"data"`
}

func DecodeJSON(r *http.Request, destination any) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	return nil
}

func WriteData(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	WriteJSON(w, r, statusCode, DataEnvelope{
		CorrelationID: CorrelationIDFromContext(r.Context()),
		Data:          data,
	})
}

func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
