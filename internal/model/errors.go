package model

import "net/http"

// Error codes from spec section 11.2.
const (
	ErrValidation          = "VALIDATION_ERROR"
	ErrInvalidFHIR         = "INVALID_FHIR_RESOURCE"
	ErrAuthRequired        = "AUTH_REQUIRED"
	ErrTokenExpired        = "TOKEN_EXPIRED"
	ErrTokenRevoked        = "TOKEN_REVOKED"
	ErrInsufficientPerms   = "INSUFFICIENT_PERMISSIONS"
	ErrSiteScopeViolation  = "SITE_SCOPE_VIOLATION"
	ErrResourceNotFound    = "RESOURCE_NOT_FOUND"
	ErrMergeConflict       = "MERGE_CONFLICT"
	ErrDuplicateResource   = "DUPLICATE_RESOURCE"
	ErrClinicalSafetyBlock = "CLINICAL_SAFETY_BLOCK"
	ErrRateLimited         = "RATE_LIMITED"
	ErrInternal            = "INTERNAL_ERROR"
	ErrGitWriteFailed      = "GIT_WRITE_FAILED"
	ErrSQLiteIndexFailed   = "SQLITE_INDEX_FAILED"
	ErrServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrNotImplemented      = "NOT_IMPLEMENTED"
	ErrConsentRequired     = "CONSENT_REQUIRED"
)

// ErrorHTTPStatus maps error codes to HTTP status codes.
var ErrorHTTPStatus = map[string]int{
	ErrValidation:          http.StatusBadRequest,
	ErrInvalidFHIR:         http.StatusBadRequest,
	ErrAuthRequired:        http.StatusUnauthorized,
	ErrTokenExpired:        http.StatusUnauthorized,
	ErrTokenRevoked:        http.StatusUnauthorized,
	ErrInsufficientPerms:   http.StatusForbidden,
	ErrSiteScopeViolation:  http.StatusForbidden,
	ErrResourceNotFound:    http.StatusNotFound,
	ErrMergeConflict:       http.StatusConflict,
	ErrDuplicateResource:   http.StatusConflict,
	ErrClinicalSafetyBlock: http.StatusUnprocessableEntity,
	ErrRateLimited:         http.StatusTooManyRequests,
	ErrInternal:            http.StatusInternalServerError,
	ErrGitWriteFailed:      http.StatusInternalServerError,
	ErrSQLiteIndexFailed:   http.StatusInternalServerError,
	ErrServiceUnavailable:  http.StatusServiceUnavailable,
	ErrNotImplemented:      http.StatusNotImplemented,
	ErrConsentRequired:     http.StatusForbidden,
}

// WriteError writes a typed error response.
func WriteError(w http.ResponseWriter, code, message string, details any) {
	status, ok := ErrorHTTPStatus[code]
	if !ok {
		status = http.StatusInternalServerError
	}
	ErrorResponse(w, status, code, message, details)
}

// NotImplementedError writes a 501 Not Implemented response.
func NotImplementedError(w http.ResponseWriter) {
	WriteError(w, ErrNotImplemented, "This endpoint is not yet implemented", nil)
}
