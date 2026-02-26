package patient_test

import (
	"context"
	"os"
	"testing"
	"time"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/config"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/server"
	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testServer(t *testing.T) *server.Server {
	t.Helper()

	// Temp git repo
	gitDir, err := os.MkdirTemp("", "patient-test-git-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(gitDir) })

	git, err := gitstore.NewStore(gitDir, "test", "test@test.com")
	if err != nil {
		t.Fatal(err)
	}

	// In-memory SQLite
	idx, err := sqliteindex.NewIndex(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { idx.Close() })

	cfg := &config.Config{
		GRPCPort: 0,
		Matching: config.MatchingConfig{
			DefaultThreshold: 0.7,
			MaxResults:       10,
			FuzzyMaxDistance: 2,
		},
	}

	pw := pipeline.NewWriter(git, idx, 5*time.Second)
	return server.NewServer(cfg, pw, idx, git)
}

func mutCtx() *patientv1.MutationContext {
	return &patientv1.MutationContext{
		PractitionerId: "dr-test",
		NodeId:         "node-1",
		SiteId:         "clinic-1",
		Timestamp:      timestamppb.Now(),
	}
}

func TestCreatePatient_FullRoundtrip(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// Create patient
	createResp, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: []byte(`{
			"resourceType": "Patient",
			"name": [{"family": "Ibrahim", "given": ["Fatima"]}],
			"gender": "female",
			"birthDate": "1990-01-15"
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if createResp.Patient == nil {
		t.Fatal("expected patient in response")
	}
	patientID := createResp.Patient.Id
	if patientID == "" {
		t.Fatal("expected non-empty patient ID")
	}
	if createResp.Git == nil || createResp.Git.CommitHash == "" {
		t.Fatal("expected git commit info")
	}

	// Get patient
	getResp, err := srv.GetPatient(ctx, &patientv1.GetPatientRequest{PatientId: patientID})
	if err != nil {
		t.Fatal(err)
	}
	if getResp.Patient.Id != patientID {
		t.Errorf("get returned ID %s, want %s", getResp.Patient.Id, patientID)
	}

	// List patients
	listResp, err := srv.ListPatients(ctx, &patientv1.ListPatientsRequest{
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(listResp.Patients) != 1 {
		t.Errorf("expected 1 patient, got %d", len(listResp.Patients))
	}
}

func TestCreateEncounter_FullRoundtrip(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// First create a patient
	createResp, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: []byte(`{
			"resourceType": "Patient",
			"name": [{"family": "Test", "given": ["User"]}],
			"gender": "male",
			"birthDate": "1985-06-20"
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	patientID := createResp.Patient.Id

	// Create encounter
	encResp, err := srv.CreateEncounter(ctx, &patientv1.CreateEncounterRequest{
		PatientId: patientID,
		FhirJson: []byte(`{
			"resourceType": "Encounter",
			"status": "finished",
			"class": {"code": "AMB", "system": "http://terminology.hl7.org/CodeSystem/v3-ActCode"},
			"subject": {"reference": "Patient/` + patientID + `"},
			"period": {"start": "2026-01-15T09:00:00Z", "end": "2026-01-15T10:00:00Z"}
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if encResp.Encounter == nil {
		t.Fatal("expected encounter in response")
	}
	if encResp.Git == nil || encResp.Git.CommitHash == "" {
		t.Fatal("expected git commit info")
	}

	// List encounters
	listResp, err := srv.ListEncounters(ctx, &patientv1.ListEncountersRequest{
		PatientId:  patientID,
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(listResp.Encounters) != 1 {
		t.Errorf("expected 1 encounter, got %d", len(listResp.Encounters))
	}
}

func TestDeletePatient_SoftDelete(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// Create
	createResp, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: []byte(`{
			"resourceType": "Patient",
			"name": [{"family": "ToDelete", "given": ["User"]}],
			"gender": "male",
			"birthDate": "1990"
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	patientID := createResp.Patient.Id

	// Delete (soft delete)
	delResp, err := srv.DeletePatient(ctx, &patientv1.DeletePatientRequest{
		PatientId: patientID,
		Context:   mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if delResp.Git == nil {
		t.Fatal("expected git commit info")
	}

	// Patient should still exist (soft delete) but with active=false
	getResp, err := srv.GetPatient(ctx, &patientv1.GetPatientRequest{PatientId: patientID})
	if err != nil {
		t.Fatal(err)
	}
	if getResp.Patient == nil {
		t.Fatal("soft-deleted patient should still be retrievable")
	}

	// Active-only list should exclude the deleted patient
	listResp, err := srv.ListPatients(ctx, &patientv1.ListPatientsRequest{
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	// ActiveOnly defaults to true (when status != "all")
	if len(listResp.Patients) != 0 {
		t.Errorf("expected 0 active patients, got %d", len(listResp.Patients))
	}
}

func TestCreateBatch_Atomic(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// Create a patient first
	createResp, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: []byte(`{
			"resourceType": "Patient",
			"name": [{"family": "Batch", "given": ["Test"]}],
			"gender": "female",
			"birthDate": "1995"
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	patientID := createResp.Patient.Id

	// Atomic batch with one invalid resource
	_, err = srv.CreateBatch(ctx, &patientv1.CreateBatchRequest{
		PatientId: patientID,
		Resources: []*commonv1.FHIRResource{
			{
				ResourceType: "Encounter",
				JsonPayload: []byte(`{
					"resourceType": "Encounter",
					"status": "finished",
					"class": {"code": "AMB"},
					"subject": {"reference": "Patient/` + patientID + `"},
					"period": {"start": "2026-01-15"}
				}`),
			},
			{
				ResourceType: "Encounter",
				JsonPayload: []byte(`{"resourceType": "Encounter"}`), // Invalid: missing required fields
			},
		},
		Context: mutCtx(),
		Atomic:  true,
	})
	if err == nil {
		t.Fatal("expected error for atomic batch with invalid resource")
	}

	// Verify nothing was written
	listResp, err := srv.ListEncounters(ctx, &patientv1.ListEncountersRequest{
		PatientId:  patientID,
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(listResp.Encounters) != 0 {
		t.Errorf("expected 0 encounters after atomic batch failure, got %d", len(listResp.Encounters))
	}
}

func TestRebuildIndex(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// Create some resources
	createResp, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: []byte(`{
			"resourceType": "Patient",
			"name": [{"family": "Rebuild", "given": ["Test"]}],
			"gender": "male",
			"birthDate": "1990"
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}
	patientID := createResp.Patient.Id

	_, err = srv.CreateEncounter(ctx, &patientv1.CreateEncounterRequest{
		PatientId: patientID,
		FhirJson: []byte(`{
			"resourceType": "Encounter",
			"status": "finished",
			"class": {"code": "AMB"},
			"subject": {"reference": "Patient/` + patientID + `"},
			"period": {"start": "2026-01-15"}
		}`),
		Context: mutCtx(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Rebuild index
	rebuildResp, err := srv.RebuildIndex(ctx, &patientv1.RebuildIndexRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if rebuildResp.ResourcesIndexed < 2 {
		t.Errorf("expected at least 2 resources indexed, got %d", rebuildResp.ResourcesIndexed)
	}
	if rebuildResp.GitHead == "" {
		t.Error("expected non-empty git HEAD")
	}
}

func TestHealthCheck(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	resp, err := srv.Health(ctx, &patientv1.HealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %s, want ok", resp.Status)
	}
}

func TestMatchPatient_WeightedScoring(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// Create several patients
	patients := []string{
		`{"resourceType":"Patient","name":[{"family":"Ibrahim","given":["Fatima"]}],"gender":"female","birthDate":"1990-01-15"}`,
		`{"resourceType":"Patient","name":[{"family":"Ibrahim","given":["Aisha"]}],"gender":"female","birthDate":"1992-03-20"}`,
		`{"resourceType":"Patient","name":[{"family":"Okafor","given":["Chidi"]}],"gender":"male","birthDate":"1985-06-10"}`,
	}
	for _, p := range patients {
		_, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
			FhirJson: []byte(p),
			Context:  mutCtx(),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Match: should find both Ibrahims
	matchResp, err := srv.MatchPatients(ctx, &patientv1.MatchPatientsRequest{
		FamilyName: "Ibrahim",
		GivenNames: []string{"Fatima"},
		Gender:     "female",
		Threshold:  0.3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matchResp.Matches) < 1 {
		t.Error("expected at least 1 match for Ibrahim")
	}
	// First match should have higher confidence (exact given name match)
	if len(matchResp.Matches) >= 2 && matchResp.Matches[0].Confidence <= matchResp.Matches[1].Confidence {
		t.Error("first match should have higher confidence")
	}
}

func TestCreatePatient_ValidationError(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	// Missing required fields
	_, err := srv.CreatePatient(ctx, &patientv1.CreatePatientRequest{
		FhirJson: []byte(`{"resourceType": "Patient"}`),
		Context:  mutCtx(),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected INVALID_ARGUMENT, got %s", st.Code())
	}
}

func TestCheckIndexHealth(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	resp, err := srv.CheckIndexHealth(ctx, &patientv1.CheckIndexHealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	// Empty repo should be healthy (both empty)
	if !resp.Healthy {
		t.Errorf("expected healthy, got message: %s", resp.Message)
	}
}
