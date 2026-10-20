package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Headers already sent, best effort fallback
			return
		}
	}
}

func WriteError(w http.ResponseWriter, statusCode int, code string, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	WriteJSON(w, statusCode, resp)
}

func ReadJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("empty request body")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body contains multiple json objects")
	}

	return nil
}
