package service

import (
	"context"
	"encoding/json"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// patientAdapter adapts the Patient gRPC client to the PatientService interface.
type patientAdapter struct {
	pool *grpcclient.Pool
}

func NewPatientService(pool *grpcclient.Pool) PatientService {
	return &patientAdapter{pool: pool}
}

func (p *patientAdapter) client() (patientv1.PatientServiceClient, error) {
	conn, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return patientv1.NewPatientServiceClient(conn), nil
}

// mutCtxFromHTTP extracts mutation context from the HTTP request context (JWT claims).
func mutCtxFromHTTP(ctx context.Context) *patientv1.MutationContext {
	mc := &patientv1.MutationContext{
		Timestamp: timestamppb.Now(),
	}
	claims := model.ClaimsFromContext(ctx)
	if claims != nil {
		mc.PractitionerId = claims.Subject
		mc.NodeId = claims.Node
		mc.SiteId = claims.Site
	}
	return mc
}

func (p *patientAdapter) ListPatients(ctx context.Context, req *ListPatientsRequest) (*ListPatientsResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListPatients(ctx, &patientv1.ListPatientsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(req.Page),
			PerPage: int32(req.PerPage),
			Sort:    req.Sort,
		},
		Gender:        req.Gender,
		BirthDateFrom: req.BirthDateFrom,
		BirthDateTo:   req.BirthDateTo,
		SiteId:        req.SiteID,
		Status:        req.Status,
		HasAlerts:     req.HasAlerts,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	patients := make([]any, len(resp.Patients))
	for i, r := range resp.Patients {
		patients[i] = fhirResourceToMap(r)
	}

	total, totalPages, page, perPage := 0, 0, req.Page, req.PerPage
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
		page = int(resp.Pagination.Page)
		perPage = int(resp.Pagination.PerPage)
	}

	return &ListPatientsResponse{
		Patients:   patients,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) GetPatient(ctx context.Context, patientID string) (*PatientBundle, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetPatient(ctx, &patientv1.GetPatientRequest{PatientId: patientID})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	bundle := &PatientBundle{
		Patient:              fhirResourceToMap(resp.Patient),
		Encounters:           fhirResourcesToSlice(resp.Encounters),
		Observations:         fhirResourcesToSlice(resp.Observations),
		Conditions:           fhirResourcesToSlice(resp.Conditions),
		MedicationRequests:   fhirResourcesToSlice(resp.MedicationRequests),
		AllergyIntolerances:  fhirResourcesToSlice(resp.AllergyIntolerances),
		Flags:                fhirResourcesToSlice(resp.Flags),
	}
	return bundle, nil
}

func (p *patientAdapter) SearchPatients(ctx context.Context, query string, page, perPage int) (*ListPatientsResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.SearchPatients(ctx, &patientv1.SearchPatientsRequest{
		Query: query,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	patients := make([]any, len(resp.Patients))
	for i, r := range resp.Patients {
		patients[i] = fhirResourceToMap(r)
	}

	total, totalPages := 0, 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ListPatientsResponse{
		Patients:   patients,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) CreatePatient(ctx context.Context, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: body,
		Context:  mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Patient),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) UpdatePatient(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.UpdatePatient(ctx, &patientv1.UpdatePatientRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Patient),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) DeletePatient(ctx context.Context, patientID string) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.DeletePatient(ctx, &patientv1.DeletePatientRequest{
		PatientId: patientID,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) MatchPatients(ctx context.Context, req *MatchPatientsRequest) (*MatchPatientsResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.MatchPatients(ctx, &patientv1.MatchPatientsRequest{
		FamilyName:      req.FamilyName,
		GivenNames:      req.GivenNames,
		Gender:          req.Gender,
		BirthDateApprox: req.BirthDateApprox,
		District:        req.District,
		Threshold:       req.Threshold,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	matches := make([]PatientMatch, len(resp.Matches))
	for i, m := range resp.Matches {
		matches[i] = PatientMatch{
			PatientID:    m.PatientId,
			Confidence:   float64(m.Confidence),
			MatchFactors: m.MatchFactors,
		}
	}
	return &MatchPatientsResponse{Matches: matches}, nil
}

func (p *patientAdapter) GetPatientHistory(ctx context.Context, patientID string) (*PatientHistoryResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetPatientHistory(ctx, &patientv1.GetPatientHistoryRequest{PatientId: patientID})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	entries := make([]HistoryEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		entries[i] = HistoryEntry{
			CommitHash:   e.CommitHash,
			Timestamp:    e.Timestamp,
			Author:       e.Author,
			Operation:    e.Operation,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceId,
			Message:      e.Message,
		}
	}
	return &PatientHistoryResponse{Entries: entries}, nil
}

func (p *patientAdapter) GetPatientTimeline(ctx context.Context, patientID string) (*PatientTimelineResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetPatientTimeline(ctx, &patientv1.GetPatientTimelineRequest{PatientId: patientID})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	events := make([]any, len(resp.Events))
	for i, e := range resp.Events {
		events[i] = map[string]any{
			"event_type":  e.EventType,
			"resource_id": e.ResourceId,
			"date":        e.Date,
		}
	}
	return &PatientTimelineResponse{Events: events}, nil
}

// --- Encounters ---

func (p *patientAdapter) ListEncounters(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListEncounters(ctx, &patientv1.ListEncountersRequest{
		PatientId: patientID,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := make([]any, len(resp.Encounters))
	for i, r := range resp.Encounters {
		resources[i] = fhirResourceToMap(r)
	}

	total, totalPages := 0, 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ClinicalListResponse{
		Resources:  resources,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) GetEncounter(ctx context.Context, patientID, encounterID string) (any, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetEncounter(ctx, &patientv1.GetEncounterRequest{
		PatientId:   patientID,
		EncounterId: encounterID,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	return fhirResourceToMap(resp.Encounter), nil
}

func (p *patientAdapter) CreateEncounter(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateEncounter(ctx, &patientv1.CreateEncounterRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Encounter),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) UpdateEncounter(ctx context.Context, patientID, encounterID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.UpdateEncounter(ctx, &patientv1.UpdateEncounterRequest{
		PatientId:   patientID,
		EncounterId: encounterID,
		FhirJson:    body,
		Context:     mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Encounter),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Observations ---

func (p *patientAdapter) ListObservations(ctx context.Context, patientID string, filters ObservationFilters, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListObservations(ctx, &patientv1.ListObservationsRequest{
		PatientId: patientID,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
		Code:        filters.Code,
		Category:    filters.Category,
		DateFrom:    filters.DateFrom,
		DateTo:      filters.DateTo,
		EncounterId: filters.EncounterID,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := make([]any, len(resp.Observations))
	for i, r := range resp.Observations {
		resources[i] = fhirResourceToMap(r)
	}

	total, totalPages := 0, 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ClinicalListResponse{
		Resources:  resources,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) GetObservation(ctx context.Context, patientID, observationID string) (any, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetObservation(ctx, &patientv1.GetObservationRequest{
		PatientId:     patientID,
		ObservationId: observationID,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	return fhirResourceToMap(resp.Observation), nil
}

func (p *patientAdapter) CreateObservation(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateObservation(ctx, &patientv1.CreateObservationRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Observation),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Conditions ---

func (p *patientAdapter) ListConditions(ctx context.Context, patientID string, filters ConditionFilters, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListConditions(ctx, &patientv1.ListConditionsRequest{
		PatientId: patientID,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
		ClinicalStatus: filters.ClinicalStatus,
		Category:       filters.Category,
		Code:           filters.Code,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := make([]any, len(resp.Conditions))
	for i, r := range resp.Conditions {
		resources[i] = fhirResourceToMap(r)
	}

	total, totalPages := 0, 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ClinicalListResponse{
		Resources:  resources,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) CreateCondition(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateCondition(ctx, &patientv1.CreateConditionRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Condition),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) UpdateCondition(ctx context.Context, patientID, conditionID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.UpdateCondition(ctx, &patientv1.UpdateConditionRequest{
		PatientId:   patientID,
		ConditionId: conditionID,
		FhirJson:    body,
		Context:     mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.Condition),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Medication Requests ---

func (p *patientAdapter) ListMedicationRequests(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListMedicationRequests(ctx, &patientv1.ListMedicationRequestsRequest{
		PatientId: patientID,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := make([]any, len(resp.MedicationRequests))
	for i, r := range resp.MedicationRequests {
		resources[i] = fhirResourceToMap(r)
	}

	total, totalPages := 0, 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ClinicalListResponse{
		Resources:  resources,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) CreateMedicationRequest(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateMedicationRequest(ctx, &patientv1.CreateMedicationRequestRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.MedicationRequest),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) UpdateMedicationRequest(ctx context.Context, patientID, medicationRequestID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.UpdateMedicationRequest(ctx, &patientv1.UpdateMedicationRequestRequest{
		PatientId:           patientID,
		MedicationRequestId: medicationRequestID,
		FhirJson:            body,
		Context:             mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.MedicationRequest),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Allergy Intolerances ---

func (p *patientAdapter) ListAllergyIntolerances(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListAllergyIntolerances(ctx, &patientv1.ListAllergyIntolerancesRequest{
		PatientId: patientID,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := make([]any, len(resp.AllergyIntolerances))
	for i, r := range resp.AllergyIntolerances {
		resources[i] = fhirResourceToMap(r)
	}

	total, totalPages := 0, 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ClinicalListResponse{
		Resources:  resources,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) CreateAllergyIntolerance(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateAllergyIntolerance(ctx, &patientv1.CreateAllergyIntoleranceRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.AllergyIntolerance),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) UpdateAllergyIntolerance(ctx context.Context, patientID, allergyIntoleranceID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.UpdateAllergyIntolerance(ctx, &patientv1.UpdateAllergyIntoleranceRequest{
		PatientId:              patientID,
		AllergyIntoleranceId:   allergyIntoleranceID,
		FhirJson:               body,
		Context:                mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{
		Resource: fhirResourceToMap(resp.AllergyIntolerance),
	}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Immunizations ---

func (p *patientAdapter) ListImmunizations(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListImmunizations(ctx, &patientv1.ListImmunizationsRequest{
		PatientId:  patientID,
		Pagination: &commonv1.PaginationRequest{Page: int32(page), PerPage: int32(perPage)},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := fhirResourcesToSlice(resp.Immunizations)
	pg := resp.Pagination
	totalPages := 1
	if pg != nil && pg.PerPage > 0 {
		totalPages = (int(pg.Total) + int(pg.PerPage) - 1) / int(pg.PerPage)
	}
	return &ClinicalListResponse{
		Resources:  resources,
		Page:       int(pg.GetPage()),
		PerPage:    int(pg.GetPerPage()),
		Total:      int(pg.GetTotal()),
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) GetImmunization(ctx context.Context, patientID, immunizationID string) (any, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetImmunization(ctx, &patientv1.GetImmunizationRequest{
		PatientId:      patientID,
		ImmunizationId: immunizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return fhirResourceToMap(resp.Immunization), nil
}

func (p *patientAdapter) CreateImmunization(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateImmunization(ctx, &patientv1.CreateImmunizationRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{Resource: fhirResourceToMap(resp.Immunization)}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Procedures ---

func (p *patientAdapter) ListProcedures(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListProcedures(ctx, &patientv1.ListProceduresRequest{
		PatientId:  patientID,
		Pagination: &commonv1.PaginationRequest{Page: int32(page), PerPage: int32(perPage)},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := fhirResourcesToSlice(resp.Procedures)
	pg := resp.Pagination
	totalPages := 1
	if pg != nil && pg.PerPage > 0 {
		totalPages = (int(pg.Total) + int(pg.PerPage) - 1) / int(pg.PerPage)
	}
	return &ClinicalListResponse{
		Resources:  resources,
		Page:       int(pg.GetPage()),
		PerPage:    int(pg.GetPerPage()),
		Total:      int(pg.GetTotal()),
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) GetProcedure(ctx context.Context, patientID, procedureID string) (any, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetProcedure(ctx, &patientv1.GetProcedureRequest{
		PatientId:   patientID,
		ProcedureId: procedureID,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return fhirResourceToMap(resp.Procedure), nil
}

func (p *patientAdapter) CreateProcedure(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateProcedure(ctx, &patientv1.CreateProcedureRequest{
		PatientId: patientID,
		FhirJson:  body,
		Context:   mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{Resource: fhirResourceToMap(resp.Procedure)}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Generic top-level resources ---

func (p *patientAdapter) ListResources(ctx context.Context, resourceType string, page, perPage int) (*ClinicalListResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListResources(ctx, &patientv1.ListResourcesRequest{
		ResourceType: resourceType,
		Pagination:   &commonv1.PaginationRequest{Page: int32(page), PerPage: int32(perPage)},
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	resources := fhirResourcesToSlice(resp.Resources)
	pg := resp.Pagination
	totalPages := 1
	if pg != nil && pg.PerPage > 0 {
		totalPages = (int(pg.Total) + int(pg.PerPage) - 1) / int(pg.PerPage)
	}
	return &ClinicalListResponse{
		Resources:  resources,
		Page:       int(pg.GetPage()),
		PerPage:    int(pg.GetPerPage()),
		Total:      int(pg.GetTotal()),
		TotalPages: totalPages,
	}, nil
}

func (p *patientAdapter) GetResource(ctx context.Context, resourceType, resourceID string) (any, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetResource(ctx, &patientv1.GetResourceRequest{
		ResourceType: resourceType,
		ResourceId:   resourceID,
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return fhirResourceToMap(resp.Resource), nil
}

func (p *patientAdapter) CreateResource(ctx context.Context, resourceType string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateResource(ctx, &patientv1.CreateResourceRequest{
		ResourceType: resourceType,
		FhirJson:     body,
		Context:      mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{Resource: fhirResourceToMap(resp.Resource)}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

func (p *patientAdapter) UpdateResource(ctx context.Context, resourceType, resourceID string, body json.RawMessage) (*WriteResponse, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.UpdateResource(ctx, &patientv1.UpdateResourceRequest{
		ResourceType: resourceType,
		ResourceId:   resourceID,
		FhirJson:     body,
		Context:      mutCtxFromHTTP(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}

	wr := &WriteResponse{Resource: fhirResourceToMap(resp.Resource)}
	if resp.Git != nil {
		wr.Git = &GitMeta{Commit: resp.Git.CommitHash, Message: resp.Git.Message}
	}
	return wr, nil
}

// --- Helpers ---

// fhirResourceToMap converts a proto FHIRResource to a map for JSON serialization.
func fhirResourceToMap(r *commonv1.FHIRResource) any {
	if r == nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(r.JsonPayload, &m); err != nil {
		return map[string]any{
			"resourceType": r.ResourceType,
			"id":           r.Id,
		}
	}
	return m
}

// fhirResourcesToSlice converts a slice of proto FHIRResources.
func fhirResourcesToSlice(resources []*commonv1.FHIRResource) []any {
	result := make([]any, len(resources))
	for i, r := range resources {
		result[i] = fhirResourceToMap(r)
	}
	return result
}

func (p *patientAdapter) ErasePatient(_ context.Context, _ string) (*EraseResponse, error) {
	return nil, fmt.Errorf("crypto-erasure not available via gRPC adapter")
}
