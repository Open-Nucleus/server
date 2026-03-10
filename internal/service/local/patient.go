// Package local provides in-process service adapters for the monolith binary.
// Each adapter implements the service.* interfaces by calling business logic
// directly, bypassing the gRPC transport layer.
package local

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/patient/pipeline"
)

// patientService implements service.PatientService by calling the write
// pipeline and SQLite index directly (no gRPC).
type patientService struct {
	pw  *pipeline.Writer
	idx sqliteindex.Index
	git gitstore.Store
}

// NewPatientService creates a local adapter for patient operations.
func NewPatientService(pw *pipeline.Writer, idx sqliteindex.Index, git gitstore.Store) service.PatientService {
	return &patientService{pw: pw, idx: idx, git: git}
}

// --- Mutation context helpers ---

func mutCtxFromHTTP(ctx context.Context) pipeline.MutationContext {
	mc := pipeline.MutationContext{Timestamp: time.Now().UTC()}
	claims := model.ClaimsFromContext(ctx)
	if claims != nil {
		mc.PractitionerID = claims.Subject
		mc.NodeID = claims.Node
		mc.SiteID = claims.Site
	}
	return mc
}

func writeResultToResponse(result *pipeline.WriteResult) *service.WriteResponse {
	wr := &service.WriteResponse{
		Resource: fhirJSONToMap(result.FHIRJson),
	}
	wr.Git = &service.GitMeta{Commit: result.CommitHash, Message: result.CommitMsg}
	return wr
}

func fhirJSONToMap(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// --- Patients ---

func (s *patientService) ListPatients(_ context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
	opts := sqliteindex.PatientListOpts{
		PaginationOpts: fhir.PaginationOpts{
			Page:    req.Page,
			PerPage: req.PerPage,
			Sort:    req.Sort,
		},
		Gender:        req.Gender,
		BirthDateFrom: req.BirthDateFrom,
		BirthDateTo:   req.BirthDateTo,
		SiteID:        req.SiteID,
		ActiveOnly:    req.Status != "all",
	}

	rows, pg, err := s.idx.ListPatients(opts)
	if err != nil {
		return nil, fmt.Errorf("patient: query failed: %w", err)
	}

	patients := make([]any, len(rows))
	for i, row := range rows {
		patients[i] = fhirJSONToMap(s.readFHIR(fhir.ResourcePatient, "", row.ID))
	}

	resp := &service.ListPatientsResponse{
		Patients: patients,
		Page:     req.Page,
		PerPage:  req.PerPage,
	}
	if pg != nil {
		resp.Total = pg.Total
		resp.TotalPages = pg.TotalPages
		resp.Page = pg.Page
		resp.PerPage = pg.PerPage
	}
	return resp, nil
}

func (s *patientService) GetPatient(_ context.Context, patientID string) (*service.PatientBundle, error) {
	bundle, err := s.idx.GetPatientBundle(patientID)
	if err != nil {
		return nil, fmt.Errorf("patient: query failed: %w", err)
	}
	if bundle == nil {
		return nil, fmt.Errorf("patient %s not found", patientID)
	}

	encSlice := make([]any, len(bundle.Encounters))
	for i, r := range bundle.Encounters {
		encSlice[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceEncounter, patientID, r.ID))
	}
	obsSlice := make([]any, len(bundle.Observations))
	for i, r := range bundle.Observations {
		obsSlice[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceObservation, patientID, r.ID))
	}
	condSlice := make([]any, len(bundle.Conditions))
	for i, r := range bundle.Conditions {
		condSlice[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceCondition, patientID, r.ID))
	}
	medSlice := make([]any, len(bundle.MedicationRequests))
	for i, r := range bundle.MedicationRequests {
		medSlice[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceMedicationRequest, patientID, r.ID))
	}
	allergySlice := make([]any, len(bundle.AllergyIntolerances))
	for i, r := range bundle.AllergyIntolerances {
		allergySlice[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceAllergyIntolerance, patientID, r.ID))
	}
	flagSlice := make([]any, len(bundle.Flags))
	for i, r := range bundle.Flags {
		flagSlice[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceFlag, patientID, r.ID))
	}

	result := &service.PatientBundle{
		Patient:              fhirJSONToMap(s.readFHIR(fhir.ResourcePatient, "", patientID)),
		Encounters:           encSlice,
		Observations:         obsSlice,
		Conditions:           condSlice,
		MedicationRequests:   medSlice,
		AllergyIntolerances:  allergySlice,
		Flags:                flagSlice,
	}
	return result, nil
}

func (s *patientService) SearchPatients(_ context.Context, query string, page, perPage int) (*service.ListPatientsResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}
	rows, pg, err := s.idx.SearchPatients(query, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: search failed: %w", err)
	}

	patients := make([]any, len(rows))
	for i, row := range rows {
		patients[i] = fhirJSONToMap(s.readFHIR(fhir.ResourcePatient, "", row.ID))
	}

	resp := &service.ListPatientsResponse{
		Patients: patients,
		Page:     page,
		PerPage:  perPage,
	}
	if pg != nil {
		resp.Total = pg.Total
		resp.TotalPages = pg.TotalPages
	}
	return resp, nil
}

func (s *patientService) CreatePatient(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourcePatient, "", body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) UpdatePatient(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpUpdate, fhir.ResourcePatient, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) DeletePatient(ctx context.Context, patientID string) (*service.WriteResponse, error) {
	result, err := s.pw.Delete(ctx, fhir.ResourcePatient, patientID, patientID, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	wr := &service.WriteResponse{}
	wr.Git = &service.GitMeta{Commit: result.CommitHash, Message: result.CommitMsg}
	return wr, nil
}

func (s *patientService) MatchPatients(_ context.Context, req *service.MatchPatientsRequest) (*service.MatchPatientsResponse, error) {
	birthYear := ""
	if len(req.BirthDateApprox) >= 4 {
		birthYear = req.BirthDateApprox[:4]
	}

	candidates, err := s.idx.GetMatchCandidates(req.FamilyName, birthYear)
	if err != nil {
		return nil, fmt.Errorf("patient: match query failed: %w", err)
	}

	threshold := req.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	var matches []service.PatientMatch
	for _, cand := range candidates {
		score := 0.0
		var factors []string

		if toLower(req.FamilyName) == toLower(cand.FamilyName) {
			score += 0.30
			factors = append(factors, "family_name_exact")
		}

		if req.Gender != "" && toLower(req.Gender) == toLower(cand.Gender) {
			score += 0.10
			factors = append(factors, "gender")
		}

		if birthYear != "" && len(cand.BirthDate) >= 4 && cand.BirthDate[:4] == birthYear {
			score += 0.10
			factors = append(factors, "birth_year")
		}

		if score >= threshold {
			matches = append(matches, service.PatientMatch{
				PatientID:    cand.ID,
				Confidence:   score,
				MatchFactors: factors,
			})
		}
	}

	return &service.MatchPatientsResponse{Matches: matches}, nil
}

func (s *patientService) GetPatientHistory(_ context.Context, patientID string) (*service.PatientHistoryResponse, error) {
	pathPrefix := fhir.PatientDirPath(patientID)
	commits, err := s.git.LogPath(pathPrefix, 100)
	if err != nil {
		return nil, fmt.Errorf("patient: git log failed: %w", err)
	}

	entries := make([]service.HistoryEntry, len(commits))
	for i, c := range commits {
		parsed, _ := gitstore.ParseCommitMessage(c.Message)
		entries[i] = service.HistoryEntry{
			CommitHash:   c.Hash,
			Timestamp:    c.Timestamp.Format(time.RFC3339),
			Author:       parsed.Author,
			Node:         parsed.NodeID,
			Site:         parsed.SiteID,
			Operation:    parsed.Operation,
			ResourceType: parsed.ResourceType,
			ResourceID:   parsed.ResourceID,
			Message:      c.Message,
		}
	}
	return &service.PatientHistoryResponse{Entries: entries}, nil
}

func (s *patientService) GetPatientTimeline(_ context.Context, patientID string) (*service.PatientTimelineResponse, error) {
	opts := fhir.PaginationOpts{Page: 1, PerPage: 100}
	events, _, err := s.idx.GetTimeline(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: timeline query failed: %w", err)
	}

	result := make([]any, len(events))
	for i, e := range events {
		result[i] = map[string]any{
			"event_type":  e.EventType,
			"resource_id": e.ResourceID,
			"date":        e.Date,
		}
	}
	return &service.PatientTimelineResponse{Events: result}, nil
}

// --- Encounters ---

func (s *patientService) ListEncounters(_ context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}
	rows, pg, err := s.idx.ListEncounters(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) GetEncounter(_ context.Context, patientID, encounterID string) (any, error) {
	row, err := s.idx.GetEncounter(patientID, encounterID)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("encounter %s not found", encounterID)
	}
	return fhirJSONToMap(s.readFHIR(fhir.ResourceEncounter, patientID, row.ID)), nil
}

func (s *patientService) CreateEncounter(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceEncounter, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) UpdateEncounter(ctx context.Context, patientID, encounterID string, body json.RawMessage) (*service.WriteResponse, error) {
	_ = encounterID // pipeline extracts ID from body
	result, err := s.pw.Write(ctx, fhir.OpUpdate, fhir.ResourceEncounter, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Observations ---

func (s *patientService) ListObservations(_ context.Context, patientID string, filters service.ObservationFilters, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := sqliteindex.ObservationListOpts{
		PaginationOpts: fhir.PaginationOpts{Page: page, PerPage: perPage},
		Code:           filters.Code,
		Category:       filters.Category,
		DateFrom:       filters.DateFrom,
		DateTo:         filters.DateTo,
		EncounterID:    filters.EncounterID,
	}
	rows, pg, err := s.idx.ListObservations(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) GetObservation(_ context.Context, patientID, observationID string) (any, error) {
	row, err := s.idx.GetObservation(patientID, observationID)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("observation %s not found", observationID)
	}
	return fhirJSONToMap(s.readFHIR(fhir.ResourceObservation, patientID, row.ID)), nil
}

func (s *patientService) CreateObservation(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceObservation, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Conditions ---

func (s *patientService) ListConditions(_ context.Context, patientID string, filters service.ConditionFilters, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := sqliteindex.ConditionListOpts{
		PaginationOpts: fhir.PaginationOpts{Page: page, PerPage: perPage},
		ClinicalStatus: filters.ClinicalStatus,
		Category:       filters.Category,
		Code:           filters.Code,
	}
	rows, pg, err := s.idx.ListConditions(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) CreateCondition(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceCondition, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) UpdateCondition(ctx context.Context, patientID, conditionID string, body json.RawMessage) (*service.WriteResponse, error) {
	_ = conditionID
	result, err := s.pw.Write(ctx, fhir.OpUpdate, fhir.ResourceCondition, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Medication Requests ---

func (s *patientService) ListMedicationRequests(_ context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}
	rows, pg, err := s.idx.ListMedicationRequests(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) CreateMedicationRequest(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceMedicationRequest, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) UpdateMedicationRequest(ctx context.Context, patientID, medicationRequestID string, body json.RawMessage) (*service.WriteResponse, error) {
	_ = medicationRequestID
	result, err := s.pw.Write(ctx, fhir.OpUpdate, fhir.ResourceMedicationRequest, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Allergy Intolerances ---

func (s *patientService) ListAllergyIntolerances(_ context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}
	rows, pg, err := s.idx.ListAllergyIntolerances(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) CreateAllergyIntolerance(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceAllergyIntolerance, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) UpdateAllergyIntolerance(ctx context.Context, patientID, allergyIntoleranceID string, body json.RawMessage) (*service.WriteResponse, error) {
	_ = allergyIntoleranceID
	result, err := s.pw.Write(ctx, fhir.OpUpdate, fhir.ResourceAllergyIntolerance, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Immunizations ---

func (s *patientService) ListImmunizations(_ context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}
	rows, pg, err := s.idx.ListImmunizations(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) GetImmunization(_ context.Context, patientID, immunizationID string) (any, error) {
	row, err := s.idx.GetImmunization(patientID, immunizationID)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("immunization %s not found", immunizationID)
	}
	return fhirJSONToMap(s.readFHIR(fhir.ResourceImmunization, patientID, row.ID)), nil
}

func (s *patientService) CreateImmunization(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceImmunization, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Procedures ---

func (s *patientService) ListProcedures(_ context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}
	rows, pg, err := s.idx.ListProcedures(patientID, opts)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return s.clinicalListFromRows(rows, pg, page, perPage), nil
}

func (s *patientService) GetProcedure(_ context.Context, patientID, procedureID string) (any, error) {
	row, err := s.idx.GetProcedure(patientID, procedureID)
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("procedure %s not found", procedureID)
	}
	return fhirJSONToMap(s.readFHIR(fhir.ResourceProcedure, patientID, row.ID)), nil
}

func (s *patientService) CreateProcedure(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, fhir.ResourceProcedure, patientID, body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Generic top-level resources (Practitioner, Organization, Location) ---

func (s *patientService) ListResources(_ context.Context, resourceType string, page, perPage int) (*service.ClinicalListResponse, error) {
	opts := fhir.PaginationOpts{Page: page, PerPage: perPage}

	switch resourceType {
	case fhir.ResourcePractitioner:
		rows, pg, err := s.idx.ListPractitioners(opts)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		return s.clinicalListFromRows(rows, pg, page, perPage), nil
	case fhir.ResourceOrganization:
		rows, pg, err := s.idx.ListOrganizations(opts)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		return s.clinicalListFromRows(rows, pg, page, perPage), nil
	case fhir.ResourceLocation:
		rows, pg, err := s.idx.ListLocations(opts)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		return s.clinicalListFromRows(rows, pg, page, perPage), nil
	case fhir.ResourceMeasureReport:
		rows, pg, err := s.idx.ListMeasureReports(opts)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		return s.clinicalListFromRows(rows, pg, page, perPage), nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func (s *patientService) GetResource(_ context.Context, resourceType, resourceID string) (any, error) {
	var patientID string

	switch resourceType {
	case fhir.ResourcePatient:
		row, err := s.idx.GetPatient(resourceID)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		if row == nil {
			return nil, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}
	case fhir.ResourceEncounter:
		row, err := s.idx.GetEncounterByID(resourceID)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		if row == nil {
			return nil, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}
		patientID = row.PatientID
	case fhir.ResourceObservation:
		row, err := s.idx.GetObservationByID(resourceID)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		if row == nil {
			return nil, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}
		patientID = row.PatientID
	case fhir.ResourcePractitioner:
		row, err := s.idx.GetPractitioner(resourceID)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		if row == nil {
			return nil, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}
		_ = row
	case fhir.ResourceOrganization:
		row, err := s.idx.GetOrganization(resourceID)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		if row == nil {
			return nil, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}
		_ = row
	case fhir.ResourceLocation:
		row, err := s.idx.GetLocation(resourceID)
		if err != nil {
			return nil, fmt.Errorf("patient: %w", err)
		}
		if row == nil {
			return nil, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}
		_ = row
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return fhirJSONToMap(s.readFHIR(resourceType, patientID, resourceID)), nil
}

func (s *patientService) CreateResource(ctx context.Context, resourceType string, body json.RawMessage) (*service.WriteResponse, error) {
	result, err := s.pw.Write(ctx, fhir.OpCreate, resourceType, "", body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

func (s *patientService) UpdateResource(ctx context.Context, resourceType, resourceID string, body json.RawMessage) (*service.WriteResponse, error) {
	_ = resourceID
	result, err := s.pw.Write(ctx, fhir.OpUpdate, resourceType, "", body, mutCtxFromHTTP(ctx))
	if err != nil {
		return nil, fmt.Errorf("patient: %w", err)
	}
	return writeResultToResponse(result), nil
}

// --- Helpers ---

// readFHIR reads FHIR JSON from Git for a given resource, decrypting if needed.
func (s *patientService) readFHIR(resourceType, patientID, resourceID string) []byte {
	data, _ := s.git.Read(fhir.GitPath(resourceType, patientID, resourceID))
	if decrypted, err := s.pw.DecryptFromGit(patientID, data); err == nil {
		return decrypted
	}
	return data
}

func newClinicalList(resources []any, pg *fhir.Pagination, page, perPage int) *service.ClinicalListResponse {
	resp := &service.ClinicalListResponse{
		Resources: resources,
		Page:      page,
		PerPage:   perPage,
	}
	if pg != nil {
		resp.Total = pg.Total
		resp.TotalPages = pg.TotalPages
		resp.Page = pg.Page
		resp.PerPage = pg.PerPage
	}
	return resp
}

// clinicalListFromRows builds a ClinicalListResponse from typed row slices.
// Uses type switch to read FHIR JSON from Git for each known row type.
func (s *patientService) clinicalListFromRows(rows any, pg *fhir.Pagination, page, perPage int) *service.ClinicalListResponse {
	var resources []any
	switch v := rows.(type) {
	case []*fhir.EncounterRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceEncounter, r.PatientID, r.ID))
		}
	case []*fhir.ObservationRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceObservation, r.PatientID, r.ID))
		}
	case []*fhir.ConditionRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceCondition, r.PatientID, r.ID))
		}
	case []*fhir.MedicationRequestRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceMedicationRequest, r.PatientID, r.ID))
		}
	case []*fhir.AllergyIntoleranceRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceAllergyIntolerance, r.PatientID, r.ID))
		}
	case []*fhir.ImmunizationRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceImmunization, r.PatientID, r.ID))
		}
	case []*fhir.ProcedureRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceProcedure, r.PatientID, r.ID))
		}
	case []*fhir.PractitionerRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourcePractitioner, "", r.ID))
		}
	case []*fhir.OrganizationRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceOrganization, "", r.ID))
		}
	case []*fhir.LocationRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceLocation, "", r.ID))
		}
	case []*fhir.MeasureReportRow:
		resources = make([]any, len(v))
		for i, r := range v {
			resources[i] = fhirJSONToMap(s.readFHIR(fhir.ResourceMeasureReport, "", r.ID))
		}
	}
	return newClinicalList(resources, pg, page, perPage)
}

// --- Crypto-erasure ---

func (s *patientService) ErasePatient(_ context.Context, patientID string) (*service.EraseResponse, error) {
	// 1. Destroy the patient's encryption key (makes Git data permanently unreadable)
	if s.pw != nil {
		if err := s.pw.DestroyPatientKey(patientID); err != nil {
			return nil, fmt.Errorf("erase: destroy key: %w", err)
		}
	}

	// 2. Delete all index rows for this patient from SQLite
	if err := s.idx.DeletePatientData(patientID); err != nil {
		return nil, fmt.Errorf("erase: delete index: %w", err)
	}

	return &service.EraseResponse{
		Erased:    true,
		PatientID: patientID,
	}, nil
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}
