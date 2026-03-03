package fhir

import (
	"context"
	"encoding/json"
	"fmt"

	pkgfhir "github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

// ResourceDispatch maps service calls for a specific FHIR resource type.
type ResourceDispatch struct {
	Read      func(ctx context.Context, resourceID string) (json.RawMessage, error)
	Search    func(ctx context.Context, params *FHIRSearchParams) ([]json.RawMessage, int, int, int, error) // resources, page, perPage, total
	Create    func(ctx context.Context, body json.RawMessage) (json.RawMessage, string, error)             // resource, ID
	Update    func(ctx context.Context, resourceID string, body json.RawMessage) (json.RawMessage, error)
	Delete    func(ctx context.Context, resourceID string) error
	ReadPerm  string
	WritePerm string
}

// buildDispatchers creates the dispatch table for all FHIR resource types.
func buildDispatchers(svc service.PatientService) map[string]*ResourceDispatch {
	d := map[string]*ResourceDispatch{}

	// Common read via generic GetResource
	genericRead := func(resourceType string) func(context.Context, string) (json.RawMessage, error) {
		return func(ctx context.Context, id string) (json.RawMessage, error) {
			result, err := svc.GetResource(ctx, resourceType, id)
			if err != nil {
				return nil, err
			}
			return marshalResource(result)
		}
	}

	// --- Patient ---
	d[pkgfhir.ResourcePatient] = &ResourceDispatch{
		ReadPerm:  "patient:read",
		WritePerm: "patient:write",
		Read:      genericRead(pkgfhir.ResourcePatient),
		Search: func(ctx context.Context, params *FHIRSearchParams) ([]json.RawMessage, int, int, int, error) {
			resp, err := svc.ListPatients(ctx, &service.ListPatientsRequest{
				Page:          params.Page,
				PerPage:       params.Count,
				Gender:        params.Filters["gender"],
				BirthDateFrom: params.Filters["birthdate"],
				SiteID:        params.Filters["_site"],
			})
			if err != nil {
				return nil, 0, 0, 0, err
			}
			resources, err := marshalSlice(resp.Patients)
			return resources, resp.Page, resp.PerPage, resp.Total, err
		},
		Create: func(ctx context.Context, body json.RawMessage) (json.RawMessage, string, error) {
			resp, err := svc.CreatePatient(ctx, body)
			if err != nil {
				return nil, "", err
			}
			raw, err := marshalResource(resp.Resource)
			return raw, extractID(resp.Resource), err
		},
		Update: func(ctx context.Context, id string, body json.RawMessage) (json.RawMessage, error) {
			resp, err := svc.UpdatePatient(ctx, id, body)
			if err != nil {
				return nil, err
			}
			return marshalResource(resp.Resource)
		},
		Delete: func(ctx context.Context, id string) error {
			_, err := svc.DeletePatient(ctx, id)
			return err
		},
	}

	// --- Patient-scoped types ---
	patientScopedTypes := []struct {
		Type      string
		ReadPerm  string
		WritePerm string
		List      func(ctx context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error)
		Create    func(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error)
		Update    func(ctx context.Context, patientID, resourceID string, body json.RawMessage) (*service.WriteResponse, error)
	}{
		{
			Type: pkgfhir.ResourceEncounter, ReadPerm: "encounter:read", WritePerm: "encounter:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListEncounters(ctx, pid, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateEncounter(ctx, pid, body)
			},
			Update: func(ctx context.Context, pid, rid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.UpdateEncounter(ctx, pid, rid, body)
			},
		},
		{
			Type: pkgfhir.ResourceObservation, ReadPerm: "observation:read", WritePerm: "observation:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListObservations(ctx, pid, service.ObservationFilters{}, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateObservation(ctx, pid, body)
			},
		},
		{
			Type: pkgfhir.ResourceCondition, ReadPerm: "condition:read", WritePerm: "condition:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListConditions(ctx, pid, service.ConditionFilters{}, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateCondition(ctx, pid, body)
			},
			Update: func(ctx context.Context, pid, rid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.UpdateCondition(ctx, pid, rid, body)
			},
		},
		{
			Type: pkgfhir.ResourceMedicationRequest, ReadPerm: "medication:read", WritePerm: "medication:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListMedicationRequests(ctx, pid, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateMedicationRequest(ctx, pid, body)
			},
			Update: func(ctx context.Context, pid, rid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.UpdateMedicationRequest(ctx, pid, rid, body)
			},
		},
		{
			Type: pkgfhir.ResourceAllergyIntolerance, ReadPerm: "allergy:read", WritePerm: "allergy:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListAllergyIntolerances(ctx, pid, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateAllergyIntolerance(ctx, pid, body)
			},
			Update: func(ctx context.Context, pid, rid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.UpdateAllergyIntolerance(ctx, pid, rid, body)
			},
		},
		{
			Type: pkgfhir.ResourceImmunization, ReadPerm: "encounter:read", WritePerm: "encounter:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListImmunizations(ctx, pid, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateImmunization(ctx, pid, body)
			},
		},
		{
			Type: pkgfhir.ResourceProcedure, ReadPerm: "encounter:read", WritePerm: "encounter:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				return svc.ListProcedures(ctx, pid, page, perPage)
			},
			Create: func(ctx context.Context, pid string, body json.RawMessage) (*service.WriteResponse, error) {
				return svc.CreateProcedure(ctx, pid, body)
			},
		},
		{
			Type: pkgfhir.ResourceFlag, ReadPerm: "alert:read", WritePerm: "alert:write",
			List: func(ctx context.Context, pid string, page, perPage int) (*service.ClinicalListResponse, error) {
				// Flag search — ListFlags doesn't exist at service layer for patient, so use ListResources as fallback
				// For FHIR search we delegate to list encounters-style call
				return nil, fmt.Errorf("Flag search not yet supported via FHIR API")
			},
		},
	}

	for _, ps := range patientScopedTypes {
		rt := ps.Type
		rp := ps.ReadPerm
		wp := ps.WritePerm
		listFn := ps.List
		createFn := ps.Create
		updateFn := ps.Update

		disp := &ResourceDispatch{
			ReadPerm:  rp,
			WritePerm: wp,
			Read:      genericRead(rt),
		}

		if listFn != nil {
			disp.Search = func(ctx context.Context, params *FHIRSearchParams) ([]json.RawMessage, int, int, int, error) {
				resp, err := listFn(ctx, params.Patient, params.Page, params.Count)
				if err != nil {
					return nil, 0, 0, 0, err
				}
				resources, err := marshalSlice(resp.Resources)
				return resources, resp.Page, resp.PerPage, resp.Total, err
			}
		}

		if createFn != nil {
			disp.Create = func(ctx context.Context, body json.RawMessage) (json.RawMessage, string, error) {
				patientID, err := pkgfhir.ExtractPatientReference(body)
				if err != nil {
					return nil, "", fmt.Errorf("failed to extract patient reference: %w", err)
				}
				if patientID == "" {
					return nil, "", fmt.Errorf("patient reference required for %s", rt)
				}
				resp, err := createFn(ctx, patientID, body)
				if err != nil {
					return nil, "", err
				}
				raw, err := marshalResource(resp.Resource)
				return raw, extractID(resp.Resource), err
			}
		}

		if updateFn != nil {
			disp.Update = func(ctx context.Context, resourceID string, body json.RawMessage) (json.RawMessage, error) {
				patientID, err := pkgfhir.ExtractPatientReference(body)
				if err != nil {
					return nil, fmt.Errorf("failed to extract patient reference: %w", err)
				}
				if patientID == "" {
					return nil, fmt.Errorf("patient reference required for %s", rt)
				}
				resp, err := updateFn(ctx, patientID, resourceID, body)
				if err != nil {
					return nil, err
				}
				return marshalResource(resp.Resource)
			}
		}

		d[rt] = disp
	}

	// --- Top-level types (Practitioner, Organization, Location) ---
	topLevelTypes := []struct {
		Type string
	}{
		{pkgfhir.ResourcePractitioner},
		{pkgfhir.ResourceOrganization},
		{pkgfhir.ResourceLocation},
	}

	for _, tl := range topLevelTypes {
		rt := tl.Type
		d[rt] = &ResourceDispatch{
			ReadPerm:  "patient:read",
			WritePerm: "patient:write",
			Read:      genericRead(rt),
			Search: func(ctx context.Context, params *FHIRSearchParams) ([]json.RawMessage, int, int, int, error) {
				resp, err := svc.ListResources(ctx, rt, params.Page, params.Count)
				if err != nil {
					return nil, 0, 0, 0, err
				}
				resources, err := marshalSlice(resp.Resources)
				return resources, resp.Page, resp.PerPage, resp.Total, err
			},
			Create: func(ctx context.Context, body json.RawMessage) (json.RawMessage, string, error) {
				resp, err := svc.CreateResource(ctx, rt, body)
				if err != nil {
					return nil, "", err
				}
				raw, err := marshalResource(resp.Resource)
				return raw, extractID(resp.Resource), err
			},
			Update: func(ctx context.Context, id string, body json.RawMessage) (json.RawMessage, error) {
				resp, err := svc.UpdateResource(ctx, rt, id, body)
				if err != nil {
					return nil, err
				}
				return marshalResource(resp.Resource)
			},
		}
	}

	// --- MeasureReport (SystemScoped, generic RPCs) ---
	d[pkgfhir.ResourceMeasureReport] = &ResourceDispatch{
		ReadPerm:  "alert:read",
		WritePerm: "alert:write",
		Read:      genericRead(pkgfhir.ResourceMeasureReport),
		Search: func(ctx context.Context, params *FHIRSearchParams) ([]json.RawMessage, int, int, int, error) {
			resp, err := svc.ListResources(ctx, pkgfhir.ResourceMeasureReport, params.Page, params.Count)
			if err != nil {
				return nil, 0, 0, 0, err
			}
			resources, err := marshalSlice(resp.Resources)
			return resources, resp.Page, resp.PerPage, resp.Total, err
		},
		Create: func(ctx context.Context, body json.RawMessage) (json.RawMessage, string, error) {
			resp, err := svc.CreateResource(ctx, pkgfhir.ResourceMeasureReport, body)
			if err != nil {
				return nil, "", err
			}
			raw, err := marshalResource(resp.Resource)
			return raw, extractID(resp.Resource), err
		},
	}

	// --- StructureDefinition (read-only, served from profile registry) ---
	d[pkgfhir.ResourceStructureDefinition] = &ResourceDispatch{
		ReadPerm: "patient:read",
		Read: func(ctx context.Context, id string) (json.RawMessage, error) {
			// Look up profile by name (id) and generate StructureDefinition
			for _, def := range pkgfhir.AllProfileDefs() {
				if def.Name == id {
					data, err := pkgfhir.GenerateStructureDefinition(def)
					if err != nil {
						return nil, fmt.Errorf("failed to generate StructureDefinition: %w", err)
					}
					return data, nil
				}
			}
			return nil, fmt.Errorf("StructureDefinition %s not found", id)
		},
		Search: func(ctx context.Context, params *FHIRSearchParams) ([]json.RawMessage, int, int, int, error) {
			defs := pkgfhir.AllProfileDefs()
			var resources []json.RawMessage
			for _, def := range defs {
				data, err := pkgfhir.GenerateStructureDefinition(def)
				if err != nil {
					continue
				}
				resources = append(resources, data)
			}
			return resources, 1, len(resources), len(resources), nil
		},
	}

	// --- Read-only types (Provenance, DetectedIssue, SupplyDelivery) ---
	d[pkgfhir.ResourceProvenance] = &ResourceDispatch{
		ReadPerm: "patient:read",
		Read:     genericRead(pkgfhir.ResourceProvenance),
	}
	d[pkgfhir.ResourceDetectedIssue] = &ResourceDispatch{
		ReadPerm: "alert:read",
		Read:     genericRead(pkgfhir.ResourceDetectedIssue),
	}
	d[pkgfhir.ResourceSupplyDelivery] = &ResourceDispatch{
		ReadPerm: "supply:read",
		Read:     genericRead(pkgfhir.ResourceSupplyDelivery),
	}

	return d
}

// marshalResource converts any service response to json.RawMessage.
func marshalResource(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, fmt.Errorf("resource not found")
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}
	return data, nil
}

// marshalSlice converts a slice of any service responses to json.RawMessage slice.
func marshalSlice(items []any) ([]json.RawMessage, error) {
	result := make([]json.RawMessage, len(items))
	for i, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal resource at index %d: %w", i, err)
		}
		result[i] = data
	}
	return result, nil
}

// extractID extracts the "id" field from a service response map.
func extractID(v any) string {
	if m, ok := v.(map[string]any); ok {
		if id, ok := m["id"].(string); ok {
			return id
		}
	}
	return ""
}
