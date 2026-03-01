package service

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type formularyAdapter struct {
	pool *grpcclient.Pool
}

func NewFormularyService(pool *grpcclient.Pool) FormularyService {
	return &formularyAdapter{pool: pool}
}

func (f *formularyAdapter) client() (formularyv1.FormularyServiceClient, error) {
	conn, err := f.pool.Conn("formulary")
	if err != nil {
		return nil, fmt.Errorf("formulary service unavailable: %w", err)
	}
	return formularyv1.NewFormularyServiceClient(conn), nil
}

func (f *formularyAdapter) SearchMedications(ctx context.Context, query, category string, page, perPage int) (*MedicationListResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.SearchMedications(ctx, &formularyv1.SearchMedicationsRequest{
		Query:    query,
		Category: category,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, err
	}
	return &MedicationListResponse{
		Medications: toMedicationDetails(resp.Medications),
		Page:        paginationPage(resp.Pagination),
		PerPage:     paginationPerPage(resp.Pagination),
		Total:       paginationTotal(resp.Pagination),
		TotalPages:  paginationTotalPages(resp.Pagination),
	}, nil
}

func (f *formularyAdapter) GetMedication(ctx context.Context, code string) (*MedicationDetail, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetMedication(ctx, &formularyv1.GetMedicationRequest{Code: code})
	if err != nil {
		return nil, err
	}
	return toMedicationDetail(resp.Medication), nil
}

func (f *formularyAdapter) ListMedicationsByCategory(ctx context.Context, category string, page, perPage int) (*MedicationListResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ListMedicationsByCategory(ctx, &formularyv1.ListMedicationsByCategoryRequest{
		Category: category,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, err
	}
	return &MedicationListResponse{
		Medications: toMedicationDetails(resp.Medications),
		Page:        paginationPage(resp.Pagination),
		PerPage:     paginationPerPage(resp.Pagination),
		Total:       paginationTotal(resp.Pagination),
		TotalPages:  paginationTotalPages(resp.Pagination),
	}, nil
}

func (f *formularyAdapter) CheckInteractions(ctx context.Context, req *CheckInteractionsRequest) (*CheckInteractionsResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.CheckInteractions(ctx, &formularyv1.CheckInteractionsRequest{
		MedicationCodes: req.MedicationCodes,
		PatientId:       req.PatientID,
		AllergyCodes:    req.AllergyCodes,
		SiteId:          req.SiteID,
	})
	if err != nil {
		return nil, err
	}
	return toCheckInteractionsResponse(resp), nil
}

func (f *formularyAdapter) CheckAllergyConflicts(ctx context.Context, req *CheckAllergyConflictsRequest) (*CheckAllergyConflictsResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.CheckAllergyConflicts(ctx, &formularyv1.CheckAllergyConflictsRequest{
		MedicationCodes: req.MedicationCodes,
		AllergyCodes:    req.AllergyCodes,
	})
	if err != nil {
		return nil, err
	}
	return toCheckAllergyConflictsResponse(resp), nil
}

func (f *formularyAdapter) ValidateDosing(ctx context.Context, req *ValidateDosingRequest) (*ValidateDosingResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ValidateDosing(ctx, &formularyv1.ValidateDosingRequest{
		MedicationCode:  req.MedicationCode,
		DoseValue:       req.DoseValue,
		DoseUnit:        req.DoseUnit,
		Frequency:       req.Frequency,
		Route:           req.Route,
		PatientWeightKg: req.PatientWeightKg,
	})
	if err != nil {
		return nil, err
	}
	return &ValidateDosingResponse{
		Valid:      resp.Valid,
		Message:    resp.Message,
		Configured: resp.Configured,
	}, nil
}

func (f *formularyAdapter) GetDosingOptions(ctx context.Context, medicationCode string, patientWeightKg float64) (*GetDosingOptionsResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetDosingOptions(ctx, &formularyv1.GetDosingOptionsRequest{
		MedicationCode:  medicationCode,
		PatientWeightKg: patientWeightKg,
	})
	if err != nil {
		return nil, err
	}
	opts := make([]DosingOptionDTO, 0, len(resp.Options))
	for _, o := range resp.Options {
		opts = append(opts, DosingOptionDTO{
			DoseValue:  o.DoseValue,
			DoseUnit:   o.DoseUnit,
			Frequency:  o.Frequency,
			Route:      o.Route,
			Indication: o.Indication,
		})
	}
	return &GetDosingOptionsResponse{
		Options:    opts,
		Configured: resp.Configured,
		Message:    resp.Message,
	}, nil
}

func (f *formularyAdapter) GenerateSchedule(ctx context.Context, req *GenerateScheduleRequest) (*GenerateScheduleResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GenerateSchedule(ctx, &formularyv1.GenerateScheduleRequest{
		MedicationCode: req.MedicationCode,
		DoseValue:      req.DoseValue,
		DoseUnit:       req.DoseUnit,
		Frequency:      req.Frequency,
		StartTime:      req.StartTime,
		DurationDays:   int32(req.DurationDays),
	})
	if err != nil {
		return nil, err
	}
	entries := make([]ScheduleEntryDTO, 0, len(resp.Entries))
	for _, e := range resp.Entries {
		entries = append(entries, ScheduleEntryDTO{
			Time:      e.Time,
			DoseValue: e.DoseValue,
			DoseUnit:  e.DoseUnit,
			Note:      e.Note,
		})
	}
	return &GenerateScheduleResponse{
		Entries:    entries,
		Configured: resp.Configured,
		Message:    resp.Message,
	}, nil
}

func (f *formularyAdapter) GetStockLevel(ctx context.Context, siteID, medicationCode string) (*StockLevelResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetStockLevel(ctx, &formularyv1.GetStockLevelRequest{
		SiteId:         siteID,
		MedicationCode: medicationCode,
	})
	if err != nil {
		return nil, err
	}
	return &StockLevelResponse{
		SiteID:               resp.SiteId,
		MedicationCode:       resp.MedicationCode,
		Quantity:             int(resp.Quantity),
		Unit:                 resp.Unit,
		LastUpdated:          resp.LastUpdated,
		EarliestExpiry:       resp.EarliestExpiry,
		DailyConsumptionRate: resp.DailyConsumptionRate,
	}, nil
}

func (f *formularyAdapter) UpdateStockLevel(ctx context.Context, req *UpdateStockLevelRequest) (*UpdateStockLevelResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.UpdateStockLevel(ctx, &formularyv1.UpdateStockLevelRequest{
		SiteId:         req.SiteID,
		MedicationCode: req.MedicationCode,
		Quantity:       int32(req.Quantity),
		Unit:           req.Unit,
		Reason:         req.Reason,
		UpdatedBy:      req.UpdatedBy,
	})
	if err != nil {
		return nil, err
	}
	return &UpdateStockLevelResponse{
		Success:     resp.Success,
		LastUpdated: resp.LastUpdated,
	}, nil
}

func (f *formularyAdapter) RecordDelivery(ctx context.Context, req *FormularyDeliveryRequest) (*FormularyDeliveryResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	items := make([]*formularyv1.DeliveryItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &formularyv1.DeliveryItem{
			MedicationCode: it.MedicationCode,
			Quantity:       int32(it.Quantity),
			Unit:           it.Unit,
			BatchNumber:    it.BatchNumber,
			ExpiryDate:     it.ExpiryDate,
		})
	}
	resp, err := c.RecordDelivery(ctx, &formularyv1.RecordDeliveryRequest{
		SiteId:       req.SiteID,
		Items:        items,
		ReceivedBy:   req.ReceivedBy,
		DeliveryDate: req.DeliveryDate,
	})
	if err != nil {
		return nil, err
	}
	return &FormularyDeliveryResponse{
		DeliveryID:    resp.DeliveryId,
		ItemsRecorded: int(resp.ItemsRecorded),
	}, nil
}

func (f *formularyAdapter) GetStockPrediction(ctx context.Context, siteID, medicationCode string) (*StockPredictionResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetStockPrediction(ctx, &formularyv1.GetStockPredictionRequest{
		SiteId:         siteID,
		MedicationCode: medicationCode,
	})
	if err != nil {
		return nil, err
	}
	return &StockPredictionResponse{
		DaysRemaining:     int(resp.DaysRemaining),
		RiskLevel:         resp.RiskLevel,
		EarliestExpiry:    resp.EarliestExpiry,
		ExpiringQuantity:  int(resp.ExpiringQuantity),
		RecommendedAction: resp.RecommendedAction,
	}, nil
}

func (f *formularyAdapter) GetRedistributionSuggestions(ctx context.Context, medicationCode string) (*FormularyRedistributionResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetRedistributionSuggestions(ctx, &formularyv1.GetRedistributionSuggestionsRequest{
		MedicationCode: medicationCode,
	})
	if err != nil {
		return nil, err
	}
	suggestions := make([]FormularyRedistributionSuggestion, 0, len(resp.Suggestions))
	for _, s := range resp.Suggestions {
		suggestions = append(suggestions, FormularyRedistributionSuggestion{
			FromSite:          s.FromSite,
			ToSite:            s.ToSite,
			SuggestedQuantity: int(s.SuggestedQuantity),
			Rationale:         s.Rationale,
			FromSiteQuantity:  int(s.FromSiteQuantity),
			ToSiteQuantity:    int(s.ToSiteQuantity),
		})
	}
	return &FormularyRedistributionResponse{Suggestions: suggestions}, nil
}

func (f *formularyAdapter) GetFormularyInfo(ctx context.Context) (*FormularyInfoResponse, error) {
	c, err := f.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetFormularyInfo(ctx, &formularyv1.GetFormularyInfoRequest{})
	if err != nil {
		return nil, err
	}
	return &FormularyInfoResponse{
		Version:               resp.Version,
		TotalMedications:      int(resp.TotalMedications),
		TotalInteractions:     int(resp.TotalInteractions),
		LastUpdated:           resp.LastUpdated,
		Categories:            resp.Categories,
		DosingEngineAvailable: resp.DosingEngineAvailable,
	}, nil
}

// --- Proto → DTO converters ---

func toMedicationDetail(m *formularyv1.Medication) *MedicationDetail {
	if m == nil {
		return nil
	}
	return &MedicationDetail{
		Code:              m.Code,
		Display:           m.Display,
		Form:              m.Form,
		Route:             m.Route,
		Category:          m.Category,
		Available:         m.Available,
		WHOEssential:      m.WhoEssential,
		TherapeuticClass:  m.TherapeuticClass,
		CommonFrequencies: m.CommonFrequencies,
		Strength:          m.Strength,
		Unit:              m.Unit,
	}
}

func toMedicationDetails(meds []*formularyv1.Medication) []MedicationDetail {
	out := make([]MedicationDetail, 0, len(meds))
	for _, m := range meds {
		if d := toMedicationDetail(m); d != nil {
			out = append(out, *d)
		}
	}
	return out
}

func toCheckInteractionsResponse(resp *formularyv1.CheckInteractionsResponse) *CheckInteractionsResponse {
	interactions := make([]InteractionDetail, 0, len(resp.Interactions))
	for _, i := range resp.Interactions {
		interactions = append(interactions, InteractionDetail{
			Severity:       i.Severity,
			Type:           i.Type,
			Description:    i.Description,
			MedicationA:    i.MedicationA,
			MedicationB:    i.MedicationB,
			Source:         i.Source,
			ClinicalEffect:  i.ClinicalEffect,
			Recommendation: i.Recommendation,
		})
	}
	alerts := make([]AllergyAlertDTO, 0, len(resp.AllergyAlerts))
	for _, a := range resp.AllergyAlerts {
		alerts = append(alerts, AllergyAlertDTO{
			Severity:             a.Severity,
			AllergyCode:          a.AllergyCode,
			MedicationCode:       a.MedicationCode,
			Description:          a.Description,
			CrossReactivityClass: a.CrossReactivityClass,
		})
	}
	warnings := make([]DosingWarningDTO, 0, len(resp.DosingWarnings))
	for _, w := range resp.DosingWarnings {
		warnings = append(warnings, DosingWarningDTO{
			MedicationCode: w.MedicationCode,
			Warning:        w.Warning,
			Severity:       w.Severity,
		})
	}
	var stockSummary *StockSummaryDTO
	if resp.StockSummary != nil {
		items := make([]StockItemDTO, 0, len(resp.StockSummary.Items))
		for _, s := range resp.StockSummary.Items {
			items = append(items, StockItemDTO{
				MedicationCode: s.MedicationCode,
				Available:      s.Available,
				Quantity:       int(s.Quantity),
				Unit:           s.Unit,
			})
		}
		stockSummary = &StockSummaryDTO{Items: items}
	}
	return &CheckInteractionsResponse{
		Interactions:   interactions,
		AllergyAlerts:  alerts,
		DosingWarnings: warnings,
		StockSummary:   stockSummary,
		OverallRisk:    resp.OverallRisk,
	}
}

func toCheckAllergyConflictsResponse(resp *formularyv1.CheckAllergyConflictsResponse) *CheckAllergyConflictsResponse {
	alerts := make([]AllergyAlertDTO, 0, len(resp.Alerts))
	for _, a := range resp.Alerts {
		alerts = append(alerts, AllergyAlertDTO{
			Severity:             a.Severity,
			AllergyCode:          a.AllergyCode,
			MedicationCode:       a.MedicationCode,
			Description:          a.Description,
			CrossReactivityClass: a.CrossReactivityClass,
		})
	}
	return &CheckAllergyConflictsResponse{
		Alerts: alerts,
		Safe:   resp.Safe,
	}
}

func paginationPage(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 1
	}
	return int(pg.Page)
}

func paginationPerPage(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 25
	}
	return int(pg.PerPage)
}

func paginationTotal(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 0
	}
	return int(pg.Total)
}

func paginationTotalPages(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 0
	}
	return int(pg.TotalPages)
}
