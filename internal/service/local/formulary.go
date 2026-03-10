package local

import (
	"context"
	"fmt"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/services/formulary/formularyservice"
)

// formularyService implements service.FormularyService by calling the real
// FormularyService directly (no gRPC).
type formularyService struct {
	svc *formularyservice.FormularyService
}

// NewFormularyService creates a local adapter for formulary operations.
func NewFormularyService(svc *formularyservice.FormularyService) service.FormularyService {
	return &formularyService{svc: svc}
}

// --- Drug lookup ---

func (f *formularyService) SearchMedications(_ context.Context, query, category string, page, perPage int) (*service.MedicationListResponse, error) {
	result := f.svc.SearchMedications(query, category, page, perPage)
	return &service.MedicationListResponse{
		Medications: toMedDetailSlice(result.Medications),
		Page:        result.Page,
		PerPage:     result.PerPage,
		Total:       result.Total,
		TotalPages:  totalPages(result.Total, result.PerPage),
	}, nil
}

func (f *formularyService) GetMedication(_ context.Context, code string) (*service.MedicationDetail, error) {
	med, err := f.svc.GetMedication(code)
	if err != nil {
		return nil, err
	}
	return toMedDetail(med), nil
}

func (f *formularyService) ListMedicationsByCategory(_ context.Context, category string, page, perPage int) (*service.MedicationListResponse, error) {
	result := f.svc.ListMedicationsByCategory(category, page, perPage)
	return &service.MedicationListResponse{
		Medications: toMedDetailSlice(result.Medications),
		Page:        result.Page,
		PerPage:     result.PerPage,
		Total:       result.Total,
		TotalPages:  totalPages(result.Total, result.PerPage),
	}, nil
}

// --- Safety checks ---

func (f *formularyService) CheckInteractions(_ context.Context, req *service.CheckInteractionsRequest) (*service.CheckInteractionsResponse, error) {
	result := f.svc.CheckInteractions(req.MedicationCodes, req.AllergyCodes, req.SiteID)

	interactions := make([]service.InteractionDetail, 0, len(result.Interactions))
	for _, rule := range result.Interactions {
		interactions = append(interactions, service.InteractionDetail{
			Severity:       rule.Severity,
			Type:           rule.Type,
			Description:    rule.Description,
			MedicationA:    rule.MedicationA,
			MedicationB:    rule.MedicationB,
			Source:         rule.Source,
			ClinicalEffect: rule.ClinicalEffect,
			Recommendation: rule.Recommendation,
		})
	}

	alerts := make([]service.AllergyAlertDTO, 0, len(result.AllergyAlerts))
	for _, a := range result.AllergyAlerts {
		alerts = append(alerts, service.AllergyAlertDTO{
			Severity:             a.Severity,
			AllergyCode:          a.AllergyCode,
			MedicationCode:       a.MedicationCode,
			Description:          a.Description,
			CrossReactivityClass: a.CrossReactivityClass,
		})
	}

	warnings := make([]service.DosingWarningDTO, 0, len(result.DosingWarnings))
	for _, w := range result.DosingWarnings {
		warnings = append(warnings, service.DosingWarningDTO{
			MedicationCode: w.MedicationCode,
			Warning:        w.Warning,
			Severity:       w.Severity,
		})
	}

	var stockSummary *service.StockSummaryDTO
	if len(result.StockItems) > 0 {
		items := make([]service.StockItemDTO, 0, len(result.StockItems))
		for _, item := range result.StockItems {
			items = append(items, service.StockItemDTO{
				MedicationCode: item.MedicationCode,
				Available:      item.Available,
				Quantity:       item.Quantity,
				Unit:           item.Unit,
			})
		}
		stockSummary = &service.StockSummaryDTO{Items: items}
	}

	return &service.CheckInteractionsResponse{
		Interactions:   interactions,
		AllergyAlerts:  alerts,
		DosingWarnings: warnings,
		StockSummary:   stockSummary,
		OverallRisk:    result.OverallRisk,
	}, nil
}

func (f *formularyService) CheckAllergyConflicts(_ context.Context, req *service.CheckAllergyConflictsRequest) (*service.CheckAllergyConflictsResponse, error) {
	result := f.svc.CheckAllergyConflicts(req.MedicationCodes, req.AllergyCodes)

	alerts := make([]service.AllergyAlertDTO, 0, len(result.Alerts))
	for _, a := range result.Alerts {
		alerts = append(alerts, service.AllergyAlertDTO{
			Severity:             a.Severity,
			AllergyCode:          a.AllergyCode,
			MedicationCode:       a.MedicationCode,
			Description:          a.Description,
			CrossReactivityClass: a.CrossReactivityClass,
		})
	}

	return &service.CheckAllergyConflictsResponse{
		Alerts: alerts,
		Safe:   result.Safe,
	}, nil
}

// --- Dosing (stub) ---

func (f *formularyService) ValidateDosing(_ context.Context, req *service.ValidateDosingRequest) (*service.ValidateDosingResponse, error) {
	result, configured := f.svc.ValidateDosing(
		req.MedicationCode, req.DoseValue, req.DoseUnit,
		req.Frequency, req.Route, req.PatientWeightKg,
	)

	resp := &service.ValidateDosingResponse{
		Configured: configured,
	}
	if result != nil {
		resp.Valid = result.Valid
		resp.Message = result.Message
	} else {
		resp.Message = "Dosing engine not configured"
	}
	return resp, nil
}

func (f *formularyService) GetDosingOptions(_ context.Context, medicationCode string, patientWeightKg float64) (*service.GetDosingOptionsResponse, error) {
	opts, configured := f.svc.GetDosingOptions(medicationCode, patientWeightKg)

	resp := &service.GetDosingOptionsResponse{
		Configured: configured,
	}
	if !configured {
		resp.Message = "Dosing engine not configured"
	}
	dtos := make([]service.DosingOptionDTO, 0, len(opts))
	for _, o := range opts {
		dtos = append(dtos, service.DosingOptionDTO{
			DoseValue:  o.DoseValue,
			DoseUnit:   o.DoseUnit,
			Frequency:  o.Frequency,
			Route:      o.Route,
			Indication: o.Indication,
		})
	}
	resp.Options = dtos
	return resp, nil
}

func (f *formularyService) GenerateSchedule(_ context.Context, req *service.GenerateScheduleRequest) (*service.GenerateScheduleResponse, error) {
	entries, configured := f.svc.GenerateSchedule(
		req.MedicationCode, req.DoseValue, req.DoseUnit,
		req.Frequency, req.StartTime, req.DurationDays,
	)

	resp := &service.GenerateScheduleResponse{
		Configured: configured,
	}
	if !configured {
		resp.Message = "Dosing engine not configured"
	}
	dtos := make([]service.ScheduleEntryDTO, 0, len(entries))
	for _, e := range entries {
		dtos = append(dtos, service.ScheduleEntryDTO{
			Time:      e.Time,
			DoseValue: e.DoseValue,
			DoseUnit:  e.DoseUnit,
			Note:      e.Note,
		})
	}
	resp.Entries = dtos
	return resp, nil
}

// --- Stock management ---

func (f *formularyService) GetStockLevel(_ context.Context, siteID, medicationCode string) (*service.StockLevelResponse, error) {
	sl, err := f.svc.GetStockLevel(siteID, medicationCode)
	if err != nil {
		return nil, err
	}
	return &service.StockLevelResponse{
		SiteID:               sl.SiteID,
		MedicationCode:       sl.MedicationCode,
		Quantity:             sl.Quantity,
		Unit:                 sl.Unit,
		LastUpdated:          sl.LastUpdated,
		EarliestExpiry:       sl.EarliestExpiry,
		DailyConsumptionRate: sl.DailyConsumptionRate,
	}, nil
}

func (f *formularyService) UpdateStockLevel(_ context.Context, req *service.UpdateStockLevelRequest) (*service.UpdateStockLevelResponse, error) {
	err := f.svc.UpdateStockLevel(req.SiteID, req.MedicationCode, req.Quantity, req.Unit, req.Reason, req.UpdatedBy)
	if err != nil {
		return nil, err
	}
	return &service.UpdateStockLevelResponse{
		Success:     true,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (f *formularyService) RecordDelivery(_ context.Context, req *service.FormularyDeliveryRequest) (*service.FormularyDeliveryResponse, error) {
	deliveryID := fmt.Sprintf("dlv-%d", time.Now().UnixNano())
	items := make([]formularyservice.DeliveryItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, formularyservice.DeliveryItemInput{
			MedicationCode: it.MedicationCode,
			Quantity:       it.Quantity,
			Unit:           it.Unit,
			BatchNumber:    it.BatchNumber,
			ExpiryDate:     it.ExpiryDate,
		})
	}
	recorded, err := f.svc.RecordDelivery(req.SiteID, req.ReceivedBy, req.DeliveryDate, deliveryID, items)
	if err != nil {
		return nil, err
	}
	return &service.FormularyDeliveryResponse{
		DeliveryID:    deliveryID,
		ItemsRecorded: recorded,
	}, nil
}

func (f *formularyService) GetStockPrediction(_ context.Context, siteID, medicationCode string) (*service.StockPredictionResponse, error) {
	pred, err := f.svc.GetStockPrediction(siteID, medicationCode)
	if err != nil {
		return nil, err
	}
	return &service.StockPredictionResponse{
		DaysRemaining:     pred.DaysRemaining,
		RiskLevel:         pred.RiskLevel,
		EarliestExpiry:    pred.EarliestExpiry,
		ExpiringQuantity:  pred.ExpiringQuantity,
		RecommendedAction: pred.RecommendedAction,
	}, nil
}

func (f *formularyService) GetRedistributionSuggestions(_ context.Context, medicationCode string) (*service.FormularyRedistributionResponse, error) {
	suggestions, err := f.svc.GetRedistributionSuggestions(medicationCode)
	if err != nil {
		return nil, err
	}
	dtos := make([]service.FormularyRedistributionSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		dtos = append(dtos, service.FormularyRedistributionSuggestion{
			FromSite:          s.FromSite,
			ToSite:            s.ToSite,
			SuggestedQuantity: s.SuggestedQuantity,
			Rationale:         s.Rationale,
			FromSiteQuantity:  s.FromSiteQuantity,
			ToSiteQuantity:    s.ToSiteQuantity,
		})
	}
	return &service.FormularyRedistributionResponse{Suggestions: dtos}, nil
}

// --- Formulary metadata ---

func (f *formularyService) GetFormularyInfo(_ context.Context) (*service.FormularyInfoResponse, error) {
	info := f.svc.GetFormularyInfo()
	return &service.FormularyInfoResponse{
		Version:               info.Version,
		TotalMedications:      info.TotalMedications,
		TotalInteractions:     info.TotalInteractions,
		LastUpdated:           info.LastUpdated,
		Categories:            info.Categories,
		DosingEngineAvailable: info.DosingEngineAvailable,
	}, nil
}

// --- Helpers ---

func toMedDetail(r *formularyservice.MedicationRecord) *service.MedicationDetail {
	if r == nil {
		return nil
	}
	return &service.MedicationDetail{
		Code:              r.Code,
		Display:           r.Display,
		Form:              r.Form,
		Route:             r.Route,
		Category:          r.Category,
		Available:         true, // matches gRPC server behaviour
		WHOEssential:      r.WHOEssential,
		TherapeuticClass:  r.TherapeuticClass,
		CommonFrequencies: r.CommonFrequencies,
		Strength:          r.Strength,
		Unit:              r.Unit,
	}
}

func toMedDetailSlice(recs []*formularyservice.MedicationRecord) []service.MedicationDetail {
	out := make([]service.MedicationDetail, 0, len(recs))
	for _, r := range recs {
		if d := toMedDetail(r); d != nil {
			out = append(out, *d)
		}
	}
	return out
}

func totalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	tp := total / perPage
	if total%perPage != 0 {
		tp++
	}
	return tp
}
