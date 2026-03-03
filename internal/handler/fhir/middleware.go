package fhir

import (
	"net/http"
	"strings"
)

// ContentNegotiation is middleware that enforces FHIR content type requirements.
// XML is rejected with 406; JSON variants and */* are accepted.
func ContentNegotiation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "" && !acceptsJSON(accept) {
			WriteFHIRError(w, http.StatusNotAcceptable, "not-supported",
				"XML not supported, use application/fhir+json")
			return
		}
		w.Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// acceptsJSON returns true if the Accept header allows JSON responses.
func acceptsJSON(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		mt := strings.TrimSpace(strings.Split(part, ";")[0])
		switch mt {
		case "application/fhir+json", "application/json", "*/*", "":
			return true
		}
	}
	return false
}
