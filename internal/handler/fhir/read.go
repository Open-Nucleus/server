package fhir

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Read returns a handler for GET /fhir/{Type}/{id}.
func (h *FHIRHandler) Read(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			WriteFHIRError(w, http.StatusBadRequest, "invalid", "Resource ID is required")
			return
		}

		disp := h.dispatchers[resourceType]
		if disp == nil || disp.Read == nil {
			WriteFHIRError(w, http.StatusNotFound, "not-supported",
				"Resource type not supported: "+resourceType)
			return
		}

		resource, err := disp.Read(r.Context(), id)
		if err != nil {
			MapServiceError(w, err)
			return
		}

		versionID, lastUpdated := extractMetaForHeaders(resource)
		if CheckConditional(w, r, versionID, lastUpdated) {
			return
		}

		WriteFHIRResource(w, http.StatusOK, resource)
	}
}
