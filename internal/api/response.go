package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the unified error format returned by all API endpoints.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes a unified error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: message,
		Code:  status,
	})
}

// WriteErrorWithDetails writes a unified error response with extra details.
func WriteErrorWithDetails(w http.ResponseWriter, status int, message, details string) {
	WriteJSON(w, status, ErrorResponse{
		Error:   message,
		Code:    status,
		Details: details,
	})
}

// DecodeJSON decodes the request body into v.
func DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
