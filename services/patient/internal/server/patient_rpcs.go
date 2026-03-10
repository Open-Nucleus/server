package server

import (
	"context"
	"strings"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListPatients(ctx context.Context, req *patientv1.ListPatientsRequest) (*patientv1.ListPatientsResponse, error) {
	opts := sqliteindex.PatientListOpts{
		PaginationOpts: paginationFromProto(req.Pagination),
		Gender:         req.Gender,
		BirthDateFrom:  req.BirthDateFrom,
		BirthDateTo:    req.BirthDateTo,
		SiteID:         req.SiteId,
		ActiveOnly:     req.Status != "all",
	}

	rows, pg, err := s.idx.ListPatients(opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListPatientsResponse{
		Pagination: paginationToProto(pg),
	}
	for _, row := range rows {
		resp.Patients = append(resp.Patients, s.patientRowToProto(row))
	}
	return resp, nil
}

func (s *Server) GetPatient(ctx context.Context, req *patientv1.GetPatientRequest) (*patientv1.GetPatientResponse, error) {
	bundle, err := s.idx.GetPatientBundle(req.PatientId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if bundle == nil {
		return nil, status.Errorf(codes.NotFound, "patient %s not found", req.PatientId)
	}

	pid := req.PatientId
	resp := &patientv1.GetPatientResponse{
		Patient: s.patientRowToProto(bundle.Patient),
	}
	for _, e := range bundle.Encounters {
		resp.Encounters = append(resp.Encounters, &commonv1.FHIRResource{ResourceType: fhir.ResourceEncounter, Id: e.ID, JsonPayload: s.readFHIR(fhir.ResourceEncounter, pid, e.ID)})
	}
	for _, o := range bundle.Observations {
		resp.Observations = append(resp.Observations, &commonv1.FHIRResource{ResourceType: fhir.ResourceObservation, Id: o.ID, JsonPayload: s.readFHIR(fhir.ResourceObservation, pid, o.ID)})
	}
	for _, c := range bundle.Conditions {
		resp.Conditions = append(resp.Conditions, &commonv1.FHIRResource{ResourceType: fhir.ResourceCondition, Id: c.ID, JsonPayload: s.readFHIR(fhir.ResourceCondition, pid, c.ID)})
	}
	for _, m := range bundle.MedicationRequests {
		resp.MedicationRequests = append(resp.MedicationRequests, &commonv1.FHIRResource{ResourceType: fhir.ResourceMedicationRequest, Id: m.ID, JsonPayload: s.readFHIR(fhir.ResourceMedicationRequest, pid, m.ID)})
	}
	for _, a := range bundle.AllergyIntolerances {
		resp.AllergyIntolerances = append(resp.AllergyIntolerances, &commonv1.FHIRResource{ResourceType: fhir.ResourceAllergyIntolerance, Id: a.ID, JsonPayload: s.readFHIR(fhir.ResourceAllergyIntolerance, pid, a.ID)})
	}
	for _, f := range bundle.Flags {
		resp.Flags = append(resp.Flags, &commonv1.FHIRResource{ResourceType: fhir.ResourceFlag, Id: f.ID, JsonPayload: s.readFHIR(fhir.ResourceFlag, pid, f.ID)})
	}
	return resp, nil
}

func (s *Server) GetPatientBundle(ctx context.Context, req *patientv1.GetPatientBundleRequest) (*patientv1.GetPatientBundleResponse, error) {
	bundle, err := s.idx.GetPatientBundle(req.PatientId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if bundle == nil {
		return nil, status.Errorf(codes.NotFound, "patient %s not found", req.PatientId)
	}

	pid := req.PatientId
	resp := &patientv1.GetPatientBundleResponse{
		Patient: s.patientRowToProto(bundle.Patient),
	}
	for _, e := range bundle.Encounters {
		resp.Encounters = append(resp.Encounters, &commonv1.FHIRResource{ResourceType: fhir.ResourceEncounter, Id: e.ID, JsonPayload: s.readFHIR(fhir.ResourceEncounter, pid, e.ID)})
	}
	for _, o := range bundle.Observations {
		resp.Observations = append(resp.Observations, &commonv1.FHIRResource{ResourceType: fhir.ResourceObservation, Id: o.ID, JsonPayload: s.readFHIR(fhir.ResourceObservation, pid, o.ID)})
	}
	for _, c := range bundle.Conditions {
		resp.Conditions = append(resp.Conditions, &commonv1.FHIRResource{ResourceType: fhir.ResourceCondition, Id: c.ID, JsonPayload: s.readFHIR(fhir.ResourceCondition, pid, c.ID)})
	}
	for _, m := range bundle.MedicationRequests {
		resp.MedicationRequests = append(resp.MedicationRequests, &commonv1.FHIRResource{ResourceType: fhir.ResourceMedicationRequest, Id: m.ID, JsonPayload: s.readFHIR(fhir.ResourceMedicationRequest, pid, m.ID)})
	}
	for _, a := range bundle.AllergyIntolerances {
		resp.AllergyIntolerances = append(resp.AllergyIntolerances, &commonv1.FHIRResource{ResourceType: fhir.ResourceAllergyIntolerance, Id: a.ID, JsonPayload: s.readFHIR(fhir.ResourceAllergyIntolerance, pid, a.ID)})
	}
	for _, f := range bundle.Flags {
		resp.Flags = append(resp.Flags, &commonv1.FHIRResource{ResourceType: fhir.ResourceFlag, Id: f.ID, JsonPayload: s.readFHIR(fhir.ResourceFlag, pid, f.ID)})
	}
	return resp, nil
}

func (s *Server) CreatePatient(ctx context.Context, req *patientv1.CreatePatientRequest) (*patientv1.CreatePatientResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourcePatient, "", req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreatePatientResponse{
		Patient: toFHIRResource(fhir.ResourcePatient, result.ResourceID, result.FHIRJson),
		Git:     toGitCommitInfo(result),
	}, nil
}

func (s *Server) UpdatePatient(ctx context.Context, req *patientv1.UpdatePatientRequest) (*patientv1.UpdatePatientResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, fhir.ResourcePatient, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdatePatientResponse{
		Patient: toFHIRResource(fhir.ResourcePatient, result.ResourceID, result.FHIRJson),
		Git:     toGitCommitInfo(result),
	}, nil
}

func (s *Server) DeletePatient(ctx context.Context, req *patientv1.DeletePatientRequest) (*patientv1.DeletePatientResponse, error) {
	result, err := s.pipeline.Delete(ctx, fhir.ResourcePatient, req.PatientId, req.PatientId, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.DeletePatientResponse{
		Git: toGitCommitInfo(result),
	}, nil
}

func (s *Server) SearchPatients(ctx context.Context, req *patientv1.SearchPatientsRequest) (*patientv1.SearchPatientsResponse, error) {
	opts := paginationFromProto(req.Pagination)
	rows, pg, err := s.idx.SearchPatients(req.Query, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}

	resp := &patientv1.SearchPatientsResponse{
		Pagination: paginationToProto(pg),
	}
	for _, row := range rows {
		resp.Patients = append(resp.Patients, s.patientRowToProto(row))
	}
	return resp, nil
}

func (s *Server) MatchPatients(ctx context.Context, req *patientv1.MatchPatientsRequest) (*patientv1.MatchPatientsResponse, error) {
	// Extract birth year
	birthYear := ""
	if len(req.BirthDateApprox) >= 4 {
		birthYear = req.BirthDateApprox[:4]
	}

	candidates, err := s.idx.GetMatchCandidates(req.FamilyName, birthYear)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "match query failed: %v", err)
	}

	threshold := req.Threshold
	if threshold == 0 {
		threshold = s.cfg.Matching.DefaultThreshold
	}

	maxResults := s.cfg.Matching.MaxResults
	reqFamilyLower := toLower(req.FamilyName)
	reqGivenLower := make([]string, len(req.GivenNames))
	for i, g := range req.GivenNames {
		reqGivenLower[i] = toLower(g)
	}

	var matches []*patientv1.PatientMatch
	for _, cand := range candidates {
		score := 0.0
		var factors []string

		candFamilyLower := toLower(cand.FamilyName)

		// Family name exact (0.30)
		if reqFamilyLower == candFamilyLower {
			score += 0.30
			factors = append(factors, "family_name_exact")
		} else if levenshtein(reqFamilyLower, candFamilyLower) <= s.cfg.Matching.FuzzyMaxDistance {
			score += 0.20
			factors = append(factors, "family_name_fuzzy")
		} else if soundex(reqFamilyLower) == soundex(candFamilyLower) {
			score += 0.20
			factors = append(factors, "family_name_soundex")
		}

		// Given name matching
		candGiven := parseGivenNames(cand.GivenNames)
		for _, cg := range candGiven {
			cgLower := toLower(cg)
			for _, rg := range reqGivenLower {
				if rg == cgLower {
					score += 0.15
					factors = append(factors, "given_name_exact")
					goto doneGiven
				}
				if levenshtein(rg, cgLower) <= s.cfg.Matching.FuzzyMaxDistance {
					score += 0.10
					factors = append(factors, "given_name_fuzzy")
					goto doneGiven
				}
			}
		}
	doneGiven:

		// Gender (0.10)
		if req.Gender != "" && toLower(req.Gender) == toLower(cand.Gender) {
			score += 0.10
			factors = append(factors, "gender")
		}

		// Birth year (0.10)
		if birthYear != "" && len(cand.BirthDate) >= 4 && cand.BirthDate[:4] == birthYear {
			score += 0.10
			factors = append(factors, "birth_year")
		}

		// District (0.05) — reads address data from Git
		if candJSON, _ := s.git.Read(fhir.GitPath(fhir.ResourcePatient, "", cand.ID)); req.District != "" && strings.Contains(string(candJSON), req.District) {
			score += 0.05
			factors = append(factors, "district")
		}

		if score >= threshold {
			matches = append(matches, &patientv1.PatientMatch{
				PatientId:    cand.ID,
				Confidence:   score,
				MatchFactors: factors,
			})
		}
	}

	// Sort by confidence descending
	for i := range len(matches) {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Confidence > matches[i].Confidence {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Limit results
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return &patientv1.MatchPatientsResponse{Matches: matches}, nil
}

func (s *Server) GetPatientHistory(ctx context.Context, req *patientv1.GetPatientHistoryRequest) (*patientv1.GetPatientHistoryResponse, error) {
	pathPrefix := fhir.PatientDirPath(req.PatientId)
	commits, err := s.git.LogPath(pathPrefix, 100)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "git log failed: %v", err)
	}

	resp := &patientv1.GetPatientHistoryResponse{}
	for _, c := range commits {
		parsed, _ := gitstore.ParseCommitMessage(c.Message)
		resp.Entries = append(resp.Entries, &patientv1.HistoryEntry{
			CommitHash:   c.Hash,
			Timestamp:    c.Timestamp.Format("2006-01-02T15:04:05Z"),
			Author:       parsed.Author,
			Node:         parsed.NodeID,
			Site:         parsed.SiteID,
			Operation:    parsed.Operation,
			ResourceType: parsed.ResourceType,
			ResourceId:   parsed.ResourceID,
			Message:      c.Message,
		})
	}
	return resp, nil
}

func (s *Server) GetPatientTimeline(ctx context.Context, req *patientv1.GetPatientTimelineRequest) (*patientv1.GetPatientTimelineResponse, error) {
	opts := paginationFromProto(req.Pagination)
	events, pg, err := s.idx.GetTimeline(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "timeline query failed: %v", err)
	}

	resp := &patientv1.GetPatientTimelineResponse{
		Pagination: paginationToProto(pg),
	}
	for _, e := range events {
		resp.Events = append(resp.Events, &patientv1.TimelineEvent{
			EventType:  e.EventType,
			ResourceId: e.ResourceID,
			Date:       e.Date,
			FhirJson:   s.readFHIR(timelineEventToResourceType(e.EventType), req.PatientId, e.ResourceID),
		})
	}
	return resp, nil
}
