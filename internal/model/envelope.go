package model

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard API response wrapper per spec section 11.1.
type Envelope struct {
	Status     string      `json:"status"`
	Data       any         `json:"data,omitempty"`
	Error      *ErrorBody  `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Warnings   []Warning   `json:"warnings,omitempty"`
	Git        *GitInfo    `json:"git,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type Warning struct {
	Severity              string `json:"severity"`
	Type                  string `json:"type"`
	Description           string `json:"description"`
	InteractingMedication string `json:"interacting_medication,omitempty"`
	Source                string `json:"source,omitempty"`
}

type GitInfo struct {
	Commit  string `json:"commit"`
	Message string `json:"message"`
}

type Meta struct {
	RequestID  string `json:"request_id"`
	DurationMS int64  `json:"duration_ms"`
	NodeID     string `json:"node_id"`
}

// JSON writes the envelope as a JSON response.
func JSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(env)
}

// Success writes a success envelope.
func Success(w http.ResponseWriter, status int, data any) {
	JSON(w, status, Envelope{Status: "success", Data: data})
}

// SuccessWithPagination writes a success envelope with pagination.
func SuccessWithPagination(w http.ResponseWriter, data any, pg *Pagination) {
	JSON(w, http.StatusOK, Envelope{Status: "success", Data: data, Pagination: pg})
}

// ErrorResponse writes an error envelope.
func ErrorResponse(w http.ResponseWriter, httpStatus int, code, message string, details any) {
	JSON(w, httpStatus, Envelope{
		Status: "error",
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
