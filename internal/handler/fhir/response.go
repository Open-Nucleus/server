package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"

	pkgfhir "github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// WriteFHIRResource writes a raw FHIR JSON resource with appropriate headers.
func WriteFHIRResource(w http.ResponseWriter, status int, fhirJSON json.RawMessage) {
	versionID, lastUpdated := extractMetaForHeaders(fhirJSON)
	if versionID != "" {
		w.Header().Set("ETag", fmt.Sprintf(`W/"%s"`, versionID))
	}
	if lastUpdated != "" {
		w.Header().Set("Last-Modified", lastUpdated)
	}
	w.Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(fhirJSON)
}

// WriteFHIRBundle writes a FHIR Bundle response.
func WriteFHIRBundle(w http.ResponseWriter, status int, bundleJSON []byte) {
	w.Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(bundleJSON)
}

// WriteFHIRError writes an OperationOutcome error response.
func WriteFHIRError(w http.ResponseWriter, httpStatus int, issueCode, diagnostics string) {
	outcome, err := pkgfhir.NewOperationOutcome([]pkgfhir.OutcomeIssue{
		{
			Severity:    issueSeverity(httpStatus),
			Code:        issueCode,
			Diagnostics: diagnostics,
		},
	})
	if err != nil {
		http.Error(w, `{"resourceType":"OperationOutcome","issue":[{"severity":"fatal","code":"exception","diagnostics":"internal error"}]}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
	w.WriteHeader(httpStatus)
	w.Write(outcome)
}

// WriteFHIRCreated writes a 201 response with Location header.
func WriteFHIRCreated(w http.ResponseWriter, resourceType, id string, fhirJSON json.RawMessage) {
	w.Header().Set("Location", fmt.Sprintf("/fhir/%s/%s", resourceType, id))
	WriteFHIRResource(w, http.StatusCreated, fhirJSON)
}

// CheckConditional checks If-None-Match / If-Modified-Since headers.
// Returns true if a 304 was sent (caller should return immediately).
func CheckConditional(w http.ResponseWriter, r *http.Request, versionID, lastUpdated string) bool {
	if versionID != "" {
		inm := r.Header.Get("If-None-Match")
		etag := fmt.Sprintf(`W/"%s"`, versionID)
		if inm == etag {
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}
	return false
}

// extractMetaForHeaders parses meta.versionId and meta.lastUpdated from FHIR JSON.
func extractMetaForHeaders(fhirJSON json.RawMessage) (versionID, lastUpdated string) {
	var resource struct {
		Meta struct {
			VersionID   string `json:"versionId"`
			LastUpdated string `json:"lastUpdated"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(fhirJSON, &resource); err != nil {
		return "", ""
	}
	return resource.Meta.VersionID, resource.Meta.LastUpdated
}

// issueSeverity returns FHIR severity based on HTTP status.
func issueSeverity(httpStatus int) string {
	if httpStatus >= 500 {
		return "fatal"
	}
	return "error"
}

// errorCodeToFHIR maps internal error substrings to FHIR issue-type and HTTP status.
var errorCodeToFHIR = map[string]struct {
	HTTPStatus int
	IssueCode  string
}{
	"not found":   {http.StatusNotFound, "not-found"},
	"NotFound":    {http.StatusNotFound, "not-found"},
	"not-found":   {http.StatusNotFound, "not-found"},
	"unavailable": {http.StatusServiceUnavailable, "transient"},
}

// MapServiceError maps a service error to an appropriate FHIR OperationOutcome response.
func MapServiceError(w http.ResponseWriter, err error) {
	msg := err.Error()
	for pattern, mapping := range errorCodeToFHIR {
		if contains(msg, pattern) {
			WriteFHIRError(w, mapping.HTTPStatus, mapping.IssueCode, msg)
			return
		}
	}
	WriteFHIRError(w, http.StatusInternalServerError, "exception", msg)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
