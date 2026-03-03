package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	pkgfhir "github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// Everything handles GET /fhir/Patient/{id}/$everything.
// Returns a searchset Bundle with the patient and all related resources.
func (h *FHIRHandler) Everything(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		WriteFHIRError(w, http.StatusBadRequest, "invalid", "Patient ID is required")
		return
	}

	bundle, err := h.patientSvc.GetPatient(r.Context(), patientID)
	if err != nil {
		MapServiceError(w, err)
		return
	}

	var entries []pkgfhir.BundleEntry

	// Patient entry (search mode = match)
	if bundle.Patient != nil {
		patientJSON, err := json.Marshal(bundle.Patient)
		if err == nil {
			entries = append(entries, pkgfhir.BundleEntry{
				FullURL:    fmt.Sprintf("/fhir/Patient/%s", patientID),
				Resource:   patientJSON,
				SearchMode: "match",
			})
		}
	}

	// Related resources (search mode = include)
	addEntries := func(resources []any, resourceType string) {
		for _, res := range resources {
			resJSON, err := json.Marshal(res)
			if err != nil {
				continue
			}
			resID := extractResourceID(resJSON)
			entries = append(entries, pkgfhir.BundleEntry{
				FullURL:    fmt.Sprintf("/fhir/%s/%s", resourceType, resID),
				Resource:   resJSON,
				SearchMode: "include",
			})
		}
	}

	addEntries(bundle.Encounters, pkgfhir.ResourceEncounter)
	addEntries(bundle.Observations, pkgfhir.ResourceObservation)
	addEntries(bundle.Conditions, pkgfhir.ResourceCondition)
	addEntries(bundle.MedicationRequests, pkgfhir.ResourceMedicationRequest)
	addEntries(bundle.AllergyIntolerances, pkgfhir.ResourceAllergyIntolerance)
	addEntries(bundle.Flags, pkgfhir.ResourceFlag)

	total := len(entries)
	links := []pkgfhir.BundleLink{
		{Relation: "self", URL: fmt.Sprintf("/fhir/Patient/%s/$everything", patientID)},
	}

	bundleJSON, err := pkgfhir.NewSearchBundle(total, entries, links)
	if err != nil {
		WriteFHIRError(w, http.StatusInternalServerError, "exception", "Failed to build Bundle")
		return
	}

	WriteFHIRBundle(w, http.StatusOK, bundleJSON)
}
