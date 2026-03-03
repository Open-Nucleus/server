package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"

	pkgfhir "github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// Search returns a handler for GET /fhir/{Type} (search-type interaction).
func (h *FHIRHandler) Search(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		disp := h.dispatchers[resourceType]
		if disp == nil || disp.Search == nil {
			WriteFHIRError(w, http.StatusNotFound, "not-supported",
				"Search not supported for: "+resourceType)
			return
		}

		params := ParseFHIRSearchParams(r)

		// Patient-scoped resources require a patient parameter
		def := pkgfhir.GetResourceDef(resourceType)
		if def != nil && def.Scope == pkgfhir.PatientScoped && params.Patient == "" {
			WriteFHIRError(w, http.StatusBadRequest, "required",
				fmt.Sprintf("Search parameter 'patient' or 'subject' is required for %s", resourceType))
			return
		}

		resources, page, perPage, total, err := disp.Search(r.Context(), params)
		if err != nil {
			MapServiceError(w, err)
			return
		}

		// Build Bundle entries
		entries := make([]pkgfhir.BundleEntry, len(resources))
		for i, res := range resources {
			resID := extractResourceID(res)
			entries[i] = pkgfhir.BundleEntry{
				FullURL:    fmt.Sprintf("/fhir/%s/%s", resourceType, resID),
				Resource:   res,
				SearchMode: "match",
			}
		}

		// Pagination links
		totalPages := 1
		if perPage > 0 {
			totalPages = (total + perPage - 1) / perPage
		}
		pg := &pkgfhir.Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		}
		links := pkgfhir.PaginationToLinks(pg, "/fhir/"+resourceType)

		bundle, err := pkgfhir.NewSearchBundle(total, entries, links)
		if err != nil {
			WriteFHIRError(w, http.StatusInternalServerError, "exception", "Failed to build Bundle")
			return
		}

		WriteFHIRBundle(w, http.StatusOK, bundle)
	}
}

// extractResourceID extracts the "id" field from FHIR JSON.
func extractResourceID(fhirJSON []byte) string {
	var r struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return ""
	}
	return r.ID
}
