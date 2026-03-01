package server

import (
	"context"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
)

func (s *Server) CheckInteractions(_ context.Context, req *formularyv1.CheckInteractionsRequest) (*formularyv1.CheckInteractionsResponse, error) {
	result := s.svc.CheckInteractions(req.MedicationCodes, req.AllergyCodes, req.SiteId)

	resp := &formularyv1.CheckInteractionsResponse{
		OverallRisk: result.OverallRisk,
	}

	for _, rule := range result.Interactions {
		resp.Interactions = append(resp.Interactions, &formularyv1.Interaction{
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

	for _, alert := range result.AllergyAlerts {
		resp.AllergyAlerts = append(resp.AllergyAlerts, &formularyv1.AllergyAlert{
			Severity:             alert.Severity,
			AllergyCode:          alert.AllergyCode,
			MedicationCode:       alert.MedicationCode,
			Description:          alert.Description,
			CrossReactivityClass: alert.CrossReactivityClass,
		})
	}

	for _, w := range result.DosingWarnings {
		resp.DosingWarnings = append(resp.DosingWarnings, &formularyv1.DosingWarning{
			MedicationCode: w.MedicationCode,
			Warning:        w.Warning,
			Severity:       w.Severity,
		})
	}

	if len(result.StockItems) > 0 {
		resp.StockSummary = &formularyv1.StockSummary{}
		for _, item := range result.StockItems {
			resp.StockSummary.Items = append(resp.StockSummary.Items, &formularyv1.StockItem{
				MedicationCode: item.MedicationCode,
				Available:      item.Available,
				Quantity:       int32(item.Quantity),
				Unit:           item.Unit,
			})
		}
	}

	return resp, nil
}

func (s *Server) CheckAllergyConflicts(_ context.Context, req *formularyv1.CheckAllergyConflictsRequest) (*formularyv1.CheckAllergyConflictsResponse, error) {
	result := s.svc.CheckAllergyConflicts(req.MedicationCodes, req.AllergyCodes)

	resp := &formularyv1.CheckAllergyConflictsResponse{
		Safe: result.Safe,
	}

	for _, alert := range result.Alerts {
		resp.Alerts = append(resp.Alerts, &formularyv1.AllergyAlert{
			Severity:             alert.Severity,
			AllergyCode:          alert.AllergyCode,
			MedicationCode:       alert.MedicationCode,
			Description:          alert.Description,
			CrossReactivityClass: alert.CrossReactivityClass,
		})
	}

	return resp, nil
}
