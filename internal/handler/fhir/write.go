package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Create returns a handler for POST /fhir/{Type}.
func (h *FHIRHandler) Create(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		disp := h.dispatchers[resourceType]
		if disp == nil || disp.Create == nil {
			WriteFHIRError(w, http.StatusMethodNotAllowed, "not-supported",
				"Create not supported for: "+resourceType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", "Failed to read request body")
			return
		}

		// Validate resourceType in body matches URL
		if err := validateResourceType(body, resourceType); err != nil {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", err.Error())
			return
		}

		resource, id, err := disp.Create(r.Context(), body)
		if err != nil {
			MapServiceError(w, err)
			return
		}

		WriteFHIRCreated(w, resourceType, id, resource)
	}
}

// Update returns a handler for PUT /fhir/{Type}/{id}.
func (h *FHIRHandler) Update(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", "Resource ID is required")
			return
		}

		disp := h.dispatchers[resourceType]
		if disp == nil || disp.Update == nil {
			WriteFHIRError(w, http.StatusMethodNotAllowed, "not-supported",
				"Update not supported for: "+resourceType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", "Failed to read request body")
			return
		}

		// Validate resourceType in body matches URL
		if err := validateResourceType(body, resourceType); err != nil {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", err.Error())
			return
		}

		// Validate body ID matches URL ID
		if err := validateBodyID(body, id); err != nil {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", err.Error())
			return
		}

		resource, err := disp.Update(r.Context(), id, body)
		if err != nil {
			MapServiceError(w, err)
			return
		}

		WriteFHIRResource(w, http.StatusOK, resource)
	}
}

// Delete returns a handler for DELETE /fhir/{Type}/{id}.
func (h *FHIRHandler) Delete(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", "Resource ID is required")
			return
		}

		disp := h.dispatchers[resourceType]
		if disp == nil || disp.Delete == nil {
			WriteFHIRError(w, http.StatusMethodNotAllowed, "not-supported",
				"Delete not supported for: "+resourceType)
			return
		}

		if err := disp.Delete(r.Context(), id); err != nil {
			MapServiceError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// validateResourceType ensures the body's resourceType matches the URL resource type.
func validateResourceType(body []byte, expected string) error {
	var partial struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(body, &partial); err != nil {
		return fmt.Errorf("invalid JSON body")
	}
	if partial.ResourceType == "" {
		return fmt.Errorf("resourceType is required in request body")
	}
	if partial.ResourceType != expected {
		return fmt.Errorf("resourceType mismatch: body has %q but URL is %q",
			partial.ResourceType, expected)
	}
	return nil
}

// validateBodyID ensures the body's id matches the URL id (for updates).
func validateBodyID(body []byte, expectedID string) error {
	var partial struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &partial); err != nil {
		return fmt.Errorf("invalid JSON body")
	}
	if partial.ID != "" && partial.ID != expectedID {
		return fmt.Errorf("id mismatch: body has %q but URL has %q",
			partial.ID, expectedID)
	}
	return nil
}
