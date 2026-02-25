package handler

import (
	"net/http"

	"github.com/FibrinLab/open-nucleus/internal/model"
)

// Stub returns a handler that responds with 501 Not Implemented.
func Stub(w http.ResponseWriter, r *http.Request) {
	model.NotImplementedError(w)
}

// StubHandler returns an http.HandlerFunc for stub endpoints.
func StubHandler() http.HandlerFunc {
	return Stub
}
