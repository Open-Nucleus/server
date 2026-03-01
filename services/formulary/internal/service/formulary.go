package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/FibrinLab/open-nucleus/services/formulary/internal/dosing"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/store"
)

// FormularyService contains the core business logic for the formulary.
type FormularyService struct {
	drugDB       *store.DrugDB
	interactions *store.InteractionIndex
	stock        *store.StockStore
	dosing       dosing.Engine
	version      string
	lastUpdated  string
}

// New creates a new FormularyService.
func New(drugDB *store.DrugDB, interactions *store.InteractionIndex, stock *store.StockStore, dosingEngine dosing.Engine) *FormularyService {
	return &FormularyService{
		drugDB:       drugDB,
		interactions: interactions,
		stock:        stock,
		dosing:       dosingEngine,
		version:      "1.0.0",
		lastUpdated:  time.Now().UTC().Format(time.RFC3339),
	}
}

// --- Drug lookup ---

type SearchResult struct {
	Medications []*store.MedicationRecord
	Total       int
	Page        int
	PerPage     int
}

func (s *FormularyService) SearchMedications(query, category string, page, perPage int) *SearchResult {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}
	meds, total := s.drugDB.Search(query, category, page, perPage)
	return &SearchResult{Medications: meds, Total: total, Page: page, PerPage: perPage}
}

func (s *FormularyService) GetMedication(code string) (*store.MedicationRecord, error) {
	med, ok := s.drugDB.Get(code)
	if !ok {
		return nil, fmt.Errorf("medication %q not found", code)
	}
	return med, nil
}

func (s *FormularyService) ListMedicationsByCategory(category string, page, perPage int) *SearchResult {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}
	meds, total := s.drugDB.ListByCategory(category, page, perPage)
	return &SearchResult{Medications: meds, Total: total, Page: page, PerPage: perPage}
}

// --- Interaction checking ---

type InteractionCheckResult struct {
	Interactions   []*store.InteractionRule
	AllergyAlerts  []store.AllergyMatch
	DosingWarnings []DosingWarningItem
	StockItems     []StockCheckItem
	OverallRisk    string
}

type DosingWarningItem struct {
	MedicationCode string
	Warning        string
	Severity       string
}

type StockCheckItem struct {
	MedicationCode string
	Available      bool
	Quantity       int
	Unit           string
}

func (s *FormularyService) CheckInteractions(medicationCodes, allergyCodes []string, siteID string) *InteractionCheckResult {
	result := &InteractionCheckResult{
		OverallRisk: "safe",
	}

	// 1. Pair-wise interaction check
	for i := 0; i < len(medicationCodes); i++ {
		for j := i + 1; j < len(medicationCodes); j++ {
			if rule := s.interactions.CheckPair(medicationCodes[i], medicationCodes[j]); rule != nil {
				result.Interactions = append(result.Interactions, rule)
				result.OverallRisk = maxRisk(result.OverallRisk, rule.Severity)
			}
		}
	}

	// 2. Class-level interaction check
	for i := 0; i < len(medicationCodes); i++ {
		classRules := s.interactions.CheckClass(medicationCodes[i])
		for _, rule := range classRules {
			// Only add if the other medication in the rule is in our list
			for j := 0; j < len(medicationCodes); j++ {
				if i == j {
					continue
				}
				otherUpper := strings.ToUpper(medicationCodes[j])
				if strings.ToUpper(rule.MedicationA) == otherUpper || strings.ToUpper(rule.MedicationB) == otherUpper {
					// Avoid duplicates from pair check
					if !containsInteraction(result.Interactions, rule) {
						result.Interactions = append(result.Interactions, rule)
						result.OverallRisk = maxRisk(result.OverallRisk, rule.Severity)
					}
				}
			}
		}
	}

	// 3. Allergy check
	if len(allergyCodes) > 0 {
		result.AllergyAlerts = s.interactions.CheckAllergies(medicationCodes, allergyCodes)
		for _, alert := range result.AllergyAlerts {
			result.OverallRisk = maxRisk(result.OverallRisk, alert.Severity)
		}
	}

	// 4. Stock check (if site provided)
	if siteID != "" {
		for _, code := range medicationCodes {
			sl, err := s.stock.Get(siteID, code)
			if err == nil {
				result.StockItems = append(result.StockItems, StockCheckItem{
					MedicationCode: code,
					Available:      sl.Quantity > 0,
					Quantity:       sl.Quantity,
					Unit:           sl.Unit,
				})
			}
		}
	}

	return result
}

// --- Allergy conflict checking ---

type AllergyConflictResult struct {
	Alerts []store.AllergyMatch
	Safe   bool
}

func (s *FormularyService) CheckAllergyConflicts(medicationCodes, allergyCodes []string) *AllergyConflictResult {
	matches := s.interactions.CheckAllergies(medicationCodes, allergyCodes)
	return &AllergyConflictResult{
		Alerts: matches,
		Safe:   len(matches) == 0,
	}
}

// --- Dosing (stub) ---

func (s *FormularyService) ValidateDosing(medicationCode string, doseValue float64, doseUnit, frequency, route string, patientWeightKg float64) (*dosing.ValidationResult, bool) {
	result, _ := s.dosing.ValidateDose(medicationCode, doseValue, doseUnit, frequency, route, patientWeightKg)
	return result, s.dosing.Available()
}

func (s *FormularyService) GetDosingOptions(medicationCode string, patientWeightKg float64) ([]dosing.DosingOption, bool) {
	opts, _ := s.dosing.GetOptions(medicationCode, patientWeightKg)
	return opts, s.dosing.Available()
}

func (s *FormularyService) GenerateSchedule(medicationCode string, doseValue float64, doseUnit, frequency, startTime string, durationDays int) ([]dosing.ScheduleEntry, bool) {
	entries, _ := s.dosing.GenerateSchedule(medicationCode, doseValue, doseUnit, frequency, startTime, durationDays)
	return entries, s.dosing.Available()
}

// --- Stock management ---

func (s *FormularyService) GetStockLevel(siteID, medicationCode string) (*store.StockLevel, error) {
	return s.stock.Get(siteID, medicationCode)
}

func (s *FormularyService) UpdateStockLevel(siteID, medicationCode string, quantity int, unit, reason, updatedBy string) error {
	sl := &store.StockLevel{
		SiteID:         siteID,
		MedicationCode: medicationCode,
		Quantity:       quantity,
		Unit:           unit,
	}
	return s.stock.Upsert(sl)
}

type DeliveryItemInput struct {
	MedicationCode string
	Quantity       int
	Unit           string
	BatchNumber    string
	ExpiryDate     string
}

func (s *FormularyService) RecordDelivery(siteID, receivedBy, deliveryDate, deliveryID string, items []DeliveryItemInput) (int, error) {
	recorded := 0
	for _, item := range items {
		sl, err := s.stock.Get(siteID, item.MedicationCode)
		if err != nil {
			return recorded, err
		}
		sl.Quantity += item.Quantity
		sl.Unit = item.Unit
		if item.ExpiryDate != "" {
			if sl.EarliestExpiry == "" || item.ExpiryDate < sl.EarliestExpiry {
				sl.EarliestExpiry = item.ExpiryDate
			}
		}
		if err := s.stock.Upsert(sl); err != nil {
			return recorded, err
		}
		recorded++
	}

	if err := s.stock.RecordDelivery(deliveryID, siteID, receivedBy, deliveryDate, recorded); err != nil {
		return recorded, err
	}

	return recorded, nil
}

type StockPrediction struct {
	DaysRemaining     int
	RiskLevel         string
	EarliestExpiry    string
	ExpiringQuantity  int
	RecommendedAction string
}

func (s *FormularyService) GetStockPrediction(siteID, medicationCode string) (*StockPrediction, error) {
	sl, err := s.stock.Get(siteID, medicationCode)
	if err != nil {
		return nil, err
	}

	pred := &StockPrediction{
		EarliestExpiry: sl.EarliestExpiry,
	}

	// Calculate days remaining from consumption rate
	if sl.DailyConsumptionRate > 0 {
		pred.DaysRemaining = int(float64(sl.Quantity) / sl.DailyConsumptionRate)
	} else if sl.Quantity > 0 {
		pred.DaysRemaining = 365 // No consumption data, assume long supply
	}

	// Check expiry
	if sl.EarliestExpiry != "" {
		expiryTime, err := time.Parse("2006-01-02", sl.EarliestExpiry)
		if err == nil {
			daysToExpiry := int(time.Until(expiryTime).Hours() / 24)
			if daysToExpiry < pred.DaysRemaining {
				pred.DaysRemaining = daysToExpiry
			}
			if daysToExpiry < 30 {
				pred.ExpiringQuantity = sl.Quantity
			}
		}
	}

	// Risk classification
	switch {
	case sl.Quantity == 0:
		pred.RiskLevel = "critical"
		pred.RecommendedAction = "Immediate resupply required"
	case pred.DaysRemaining < 14:
		pred.RiskLevel = "high"
		pred.RecommendedAction = "Order resupply urgently"
	case pred.DaysRemaining < 30:
		pred.RiskLevel = "moderate"
		pred.RecommendedAction = "Plan resupply within 2 weeks"
	default:
		pred.RiskLevel = "low"
		pred.RecommendedAction = "Stock levels adequate"
	}

	return pred, nil
}

type RedistributionSuggestion struct {
	FromSite          string
	ToSite            string
	SuggestedQuantity int
	Rationale         string
	FromSiteQuantity  int
	ToSiteQuantity    int
}

func (s *FormularyService) GetRedistributionSuggestions(medicationCode string) ([]RedistributionSuggestion, error) {
	levels, err := s.stock.ListByMedication(medicationCode)
	if err != nil {
		return nil, err
	}

	var surplus, shortage []*store.StockLevel
	for _, sl := range levels {
		pred, err := s.GetStockPrediction(sl.SiteID, sl.MedicationCode)
		if err != nil {
			continue
		}
		if pred.DaysRemaining > 90 {
			surplus = append(surplus, sl)
		} else if pred.DaysRemaining < 14 {
			shortage = append(shortage, sl)
		}
	}

	var suggestions []RedistributionSuggestion
	for _, short := range shortage {
		for _, surp := range surplus {
			// Suggest transferring up to half the surplus
			transferQty := surp.Quantity / 2
			if transferQty < 1 {
				continue
			}
			suggestions = append(suggestions, RedistributionSuggestion{
				FromSite:          surp.SiteID,
				ToSite:            short.SiteID,
				SuggestedQuantity: transferQty,
				Rationale:         fmt.Sprintf("Transfer from surplus site (>90 days supply) to shortage site (<14 days supply)"),
				FromSiteQuantity:  surp.Quantity,
				ToSiteQuantity:    short.Quantity,
			})
		}
	}

	return suggestions, nil
}

// --- Formulary metadata ---

type FormularyInfo struct {
	Version              string
	TotalMedications     int
	TotalInteractions    int
	LastUpdated          string
	Categories           []string
	DosingEngineAvailable bool
}

func (s *FormularyService) GetFormularyInfo() *FormularyInfo {
	return &FormularyInfo{
		Version:               s.version,
		TotalMedications:      s.drugDB.Count(),
		TotalInteractions:     s.interactions.Count(),
		LastUpdated:           s.lastUpdated,
		Categories:            s.drugDB.Categories(),
		DosingEngineAvailable: s.dosing.Available(),
	}
}

// --- Helpers ---

func maxRisk(current, new string) string {
	riskOrder := map[string]int{
		"safe":     0,
		"low":      1,
		"moderate": 2,
		"high":     3,
		"critical": 4,
	}
	if riskOrder[new] > riskOrder[current] {
		return new
	}
	return current
}

func containsInteraction(list []*store.InteractionRule, rule *store.InteractionRule) bool {
	for _, r := range list {
		if r.MedicationA == rule.MedicationA && r.MedicationB == rule.MedicationB {
			return true
		}
	}
	return false
}
