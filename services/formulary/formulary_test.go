package formulary_test

import (
	"context"
	"testing"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	"github.com/FibrinLab/open-nucleus/services/formulary/formularytest"
)

func setup(t *testing.T) *formularytest.Env {
	t.Helper()
	tmpDir := t.TempDir()
	return formularytest.Start(t, tmpDir)
}

// --- Drug Lookup Tests ---

func TestSearchMedications_All(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.SearchMedications(context.Background(), &formularyv1.SearchMedicationsRequest{
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 50},
	})
	if err != nil {
		t.Fatalf("SearchMedications: %v", err)
	}
	if len(resp.Medications) != 20 {
		t.Errorf("expected 20 medications, got %d", len(resp.Medications))
	}
	if resp.Pagination.Total != 20 {
		t.Errorf("expected total=20, got %d", resp.Pagination.Total)
	}
}

func TestSearchMedications_Query(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.SearchMedications(context.Background(), &formularyv1.SearchMedicationsRequest{
		Query:      "amox",
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatalf("SearchMedications: %v", err)
	}
	if len(resp.Medications) != 1 {
		t.Errorf("expected 1 match for 'amox', got %d", len(resp.Medications))
	}
	if len(resp.Medications) > 0 && resp.Medications[0].Code != "J01CA04" {
		t.Errorf("expected J01CA04, got %s", resp.Medications[0].Code)
	}
}

func TestSearchMedications_ByCategory(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.SearchMedications(context.Background(), &formularyv1.SearchMedicationsRequest{
		Category:   "antibiotic",
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatalf("SearchMedications: %v", err)
	}
	if len(resp.Medications) != 4 {
		t.Errorf("expected 4 antibiotics, got %d", len(resp.Medications))
	}
}

func TestSearchMedications_Pagination(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.SearchMedications(context.Background(), &formularyv1.SearchMedicationsRequest{
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 5},
	})
	if err != nil {
		t.Fatalf("SearchMedications: %v", err)
	}
	if len(resp.Medications) != 5 {
		t.Errorf("expected 5 on page 1, got %d", len(resp.Medications))
	}
	if resp.Pagination.TotalPages != 4 {
		t.Errorf("expected 4 total pages, got %d", resp.Pagination.TotalPages)
	}
}

func TestGetMedication_Found(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetMedication(context.Background(), &formularyv1.GetMedicationRequest{
		Code: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetMedication: %v", err)
	}
	if resp.Medication.Display != "Amoxicillin" {
		t.Errorf("expected Amoxicillin, got %s", resp.Medication.Display)
	}
	if !resp.Medication.WhoEssential {
		t.Error("expected WHO essential = true")
	}
	if resp.Medication.TherapeuticClass != "Penicillin" {
		t.Errorf("expected Penicillin class, got %s", resp.Medication.TherapeuticClass)
	}
}

func TestGetMedication_NotFound(t *testing.T) {
	env := setup(t)
	_, err := env.Client.GetMedication(context.Background(), &formularyv1.GetMedicationRequest{
		Code: "XXXXXX",
	})
	if err == nil {
		t.Fatal("expected error for unknown code")
	}
}

func TestListMedicationsByCategory(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.ListMedicationsByCategory(context.Background(), &formularyv1.ListMedicationsByCategoryRequest{
		Category:   "analgesic",
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatalf("ListMedicationsByCategory: %v", err)
	}
	if len(resp.Medications) != 3 { // Paracetamol, Diclofenac, Morphine
		t.Errorf("expected 3 analgesics, got %d", len(resp.Medications))
	}
}

// --- Interaction Tests ---

func TestCheckInteractions_MethotrexateTrimethoprim(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckInteractions(context.Background(), &formularyv1.CheckInteractionsRequest{
		MedicationCodes: []string{"L04AX03", "J01EA01"},
	})
	if err != nil {
		t.Fatalf("CheckInteractions: %v", err)
	}
	if len(resp.Interactions) == 0 {
		t.Fatal("expected at least 1 interaction for MTX+TMP")
	}
	found := false
	for _, ix := range resp.Interactions {
		if ix.Severity == "high" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected high-severity interaction for MTX+TMP")
	}
	if resp.OverallRisk != "high" {
		t.Errorf("expected overall_risk=high, got %s", resp.OverallRisk)
	}
}

func TestCheckInteractions_Safe(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckInteractions(context.Background(), &formularyv1.CheckInteractionsRequest{
		MedicationCodes: []string{"N02BE01", "A02BC01"}, // Paracetamol + Omeprazole
	})
	if err != nil {
		t.Fatalf("CheckInteractions: %v", err)
	}
	if len(resp.Interactions) != 0 {
		t.Errorf("expected 0 interactions, got %d", len(resp.Interactions))
	}
	if resp.OverallRisk != "safe" {
		t.Errorf("expected overall_risk=safe, got %s", resp.OverallRisk)
	}
}

func TestCheckInteractions_QTProlongation(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckInteractions(context.Background(), &formularyv1.CheckInteractionsRequest{
		MedicationCodes: []string{"P01BA02", "C01BD01"}, // HCQ + Amiodarone
	})
	if err != nil {
		t.Fatalf("CheckInteractions: %v", err)
	}
	if len(resp.Interactions) == 0 {
		t.Fatal("expected QT interaction")
	}
	if resp.OverallRisk != "high" {
		t.Errorf("expected high risk, got %s", resp.OverallRisk)
	}
}

func TestCheckInteractions_WarfarinNSAID(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckInteractions(context.Background(), &formularyv1.CheckInteractionsRequest{
		MedicationCodes: []string{"B01AA03", "M01AB05"}, // Warfarin + Diclofenac
	})
	if err != nil {
		t.Fatalf("CheckInteractions: %v", err)
	}
	if len(resp.Interactions) == 0 {
		t.Fatal("expected Warfarin+NSAID interaction")
	}
	if resp.OverallRisk != "high" {
		t.Errorf("expected high risk, got %s", resp.OverallRisk)
	}
}

func TestCheckInteractions_MultiDrug(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckInteractions(context.Background(), &formularyv1.CheckInteractionsRequest{
		MedicationCodes: []string{"L04AX03", "J01EA01", "M01AB05"}, // MTX + TMP + Diclofenac
	})
	if err != nil {
		t.Fatalf("CheckInteractions: %v", err)
	}
	// Should find at least 2 interactions: MTX+TMP and MTX+NSAID
	if len(resp.Interactions) < 2 {
		t.Errorf("expected at least 2 interactions, got %d", len(resp.Interactions))
	}
}

func TestCheckInteractions_WithAllergies(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckInteractions(context.Background(), &formularyv1.CheckInteractionsRequest{
		MedicationCodes: []string{"J01CA04"}, // Amoxicillin
		AllergyCodes:    []string{"91936005"}, // Penicillin allergy
	})
	if err != nil {
		t.Fatalf("CheckInteractions: %v", err)
	}
	if len(resp.AllergyAlerts) == 0 {
		t.Fatal("expected allergy alert for penicillin allergy + amoxicillin")
	}
	if resp.OverallRisk != "high" {
		t.Errorf("expected high risk for allergy, got %s", resp.OverallRisk)
	}
}

// --- Allergy Tests ---

func TestCheckAllergyConflicts_PenicillinAmoxicillin(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckAllergyConflicts(context.Background(), &formularyv1.CheckAllergyConflictsRequest{
		MedicationCodes: []string{"J01CA04"}, // Amoxicillin (J01C* = penicillin)
		AllergyCodes:    []string{"91936005"}, // Penicillin allergy
	})
	if err != nil {
		t.Fatalf("CheckAllergyConflicts: %v", err)
	}
	if resp.Safe {
		t.Error("expected unsafe for penicillin allergy + amoxicillin")
	}
	if len(resp.Alerts) == 0 {
		t.Fatal("expected at least 1 alert")
	}
	if resp.Alerts[0].Severity != "high" {
		t.Errorf("expected high severity, got %s", resp.Alerts[0].Severity)
	}
}

func TestCheckAllergyConflicts_PenicillinCephalosporin(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckAllergyConflicts(context.Background(), &formularyv1.CheckAllergyConflictsRequest{
		MedicationCodes: []string{"J01DB01"}, // Cefalexin (J01DB* = cephalosporin)
		AllergyCodes:    []string{"91936005"}, // Penicillin allergy
	})
	if err != nil {
		t.Fatalf("CheckAllergyConflicts: %v", err)
	}
	if resp.Safe {
		t.Error("expected unsafe for penicillin allergy + cephalosporin (cross-reactivity)")
	}
	if len(resp.Alerts) == 0 {
		t.Fatal("expected cross-reactivity alert")
	}
	if resp.Alerts[0].CrossReactivityClass != "beta-lactam" {
		t.Errorf("expected beta-lactam cross-reactivity class, got %s", resp.Alerts[0].CrossReactivityClass)
	}
}

func TestCheckAllergyConflicts_Safe(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.CheckAllergyConflicts(context.Background(), &formularyv1.CheckAllergyConflictsRequest{
		MedicationCodes: []string{"N02BE01"}, // Paracetamol
		AllergyCodes:    []string{"91936005"}, // Penicillin allergy
	})
	if err != nil {
		t.Fatalf("CheckAllergyConflicts: %v", err)
	}
	if !resp.Safe {
		t.Error("expected safe for paracetamol with penicillin allergy")
	}
}

// --- Dosing Tests (stub) ---

func TestValidateDosing_NotConfigured(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.ValidateDosing(context.Background(), &formularyv1.ValidateDosingRequest{
		MedicationCode: "J01CA04",
		DoseValue:      500,
		DoseUnit:       "mg",
		Frequency:      "TID",
	})
	if err != nil {
		t.Fatalf("ValidateDosing: %v", err)
	}
	if resp.Configured {
		t.Error("expected configured=false for stub engine")
	}
	if resp.Message == "" {
		t.Error("expected message about not configured")
	}
}

func TestGetDosingOptions_NotConfigured(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetDosingOptions(context.Background(), &formularyv1.GetDosingOptionsRequest{
		MedicationCode: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetDosingOptions: %v", err)
	}
	if resp.Configured {
		t.Error("expected configured=false")
	}
}

func TestGenerateSchedule_NotConfigured(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GenerateSchedule(context.Background(), &formularyv1.GenerateScheduleRequest{
		MedicationCode: "J01CA04",
		DoseValue:      500,
		DoseUnit:       "mg",
		Frequency:      "TID",
		DurationDays:   7,
	})
	if err != nil {
		t.Fatalf("GenerateSchedule: %v", err)
	}
	if resp.Configured {
		t.Error("expected configured=false")
	}
}

// --- Stock Tests ---

func TestStockLevel_DefaultEmpty(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetStockLevel(context.Background(), &formularyv1.GetStockLevelRequest{
		SiteId:         "site-A",
		MedicationCode: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetStockLevel: %v", err)
	}
	if resp.Quantity != 0 {
		t.Errorf("expected 0 quantity for new stock, got %d", resp.Quantity)
	}
}

func TestUpdateAndGetStockLevel(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Update stock
	_, err := env.Client.UpdateStockLevel(ctx, &formularyv1.UpdateStockLevelRequest{
		SiteId:         "site-A",
		MedicationCode: "J01CA04",
		Quantity:       100,
		Unit:           "capsules",
		Reason:         "initial stock",
		UpdatedBy:      "dr-test",
	})
	if err != nil {
		t.Fatalf("UpdateStockLevel: %v", err)
	}

	// Get stock
	resp, err := env.Client.GetStockLevel(ctx, &formularyv1.GetStockLevelRequest{
		SiteId:         "site-A",
		MedicationCode: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetStockLevel: %v", err)
	}
	if resp.Quantity != 100 {
		t.Errorf("expected 100, got %d", resp.Quantity)
	}
	if resp.Unit != "capsules" {
		t.Errorf("expected capsules, got %s", resp.Unit)
	}
}

func TestRecordDelivery(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	resp, err := env.Client.RecordDelivery(ctx, &formularyv1.RecordDeliveryRequest{
		SiteId:       "site-A",
		ReceivedBy:   "nurse-1",
		DeliveryDate: "2026-03-01",
		Items: []*formularyv1.DeliveryItem{
			{MedicationCode: "J01CA04", Quantity: 200, Unit: "capsules", BatchNumber: "BATCH001", ExpiryDate: "2027-06-01"},
			{MedicationCode: "N02BE01", Quantity: 500, Unit: "tablets", BatchNumber: "BATCH002", ExpiryDate: "2027-12-01"},
		},
	})
	if err != nil {
		t.Fatalf("RecordDelivery: %v", err)
	}
	if resp.ItemsRecorded != 2 {
		t.Errorf("expected 2 items recorded, got %d", resp.ItemsRecorded)
	}
	if resp.DeliveryId == "" {
		t.Error("expected delivery ID")
	}

	// Verify stock updated
	stock, err := env.Client.GetStockLevel(ctx, &formularyv1.GetStockLevelRequest{
		SiteId:         "site-A",
		MedicationCode: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetStockLevel after delivery: %v", err)
	}
	if stock.Quantity != 200 {
		t.Errorf("expected 200 after delivery, got %d", stock.Quantity)
	}
}

func TestStockPrediction_Critical(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Stock at 0 = critical
	resp, err := env.Client.GetStockPrediction(ctx, &formularyv1.GetStockPredictionRequest{
		SiteId:         "site-A",
		MedicationCode: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetStockPrediction: %v", err)
	}
	if resp.RiskLevel != "critical" {
		t.Errorf("expected critical risk for 0 stock, got %s", resp.RiskLevel)
	}
}

func TestRedistribution(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetRedistributionSuggestions(context.Background(), &formularyv1.GetRedistributionSuggestionsRequest{
		MedicationCode: "J01CA04",
	})
	if err != nil {
		t.Fatalf("GetRedistributionSuggestions: %v", err)
	}
	// With no stock data, should return empty suggestions
	if len(resp.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions with no stock data, got %d", len(resp.Suggestions))
	}
}

// --- Formulary Metadata Tests ---

func TestGetFormularyInfo(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetFormularyInfo(context.Background(), &formularyv1.GetFormularyInfoRequest{})
	if err != nil {
		t.Fatalf("GetFormularyInfo: %v", err)
	}
	if resp.TotalMedications != 20 {
		t.Errorf("expected 20 medications, got %d", resp.TotalMedications)
	}
	if resp.TotalInteractions < 17 {
		t.Errorf("expected at least 17 interactions, got %d", resp.TotalInteractions)
	}
	if resp.DosingEngineAvailable {
		t.Error("expected dosing engine unavailable in stub mode")
	}
	if len(resp.Categories) == 0 {
		t.Error("expected at least 1 category")
	}
}

// --- Health Test ---

func TestHealth(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.Health(context.Background(), &formularyv1.HealthRequest{})
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("expected healthy, got %s", resp.Status)
	}
	if resp.MedicationsLoaded != 20 {
		t.Errorf("expected 20 medications loaded, got %d", resp.MedicationsLoaded)
	}
}
