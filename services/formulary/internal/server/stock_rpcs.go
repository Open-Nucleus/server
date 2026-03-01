package server

import (
	"context"
	"fmt"
	"time"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/service"
)

func (s *Server) GetStockLevel(_ context.Context, req *formularyv1.GetStockLevelRequest) (*formularyv1.GetStockLevelResponse, error) {
	sl, err := s.svc.GetStockLevel(req.SiteId, req.MedicationCode)
	if err != nil {
		return nil, mapError(err)
	}
	return &formularyv1.GetStockLevelResponse{
		SiteId:               sl.SiteID,
		MedicationCode:       sl.MedicationCode,
		Quantity:             int32(sl.Quantity),
		Unit:                 sl.Unit,
		LastUpdated:          sl.LastUpdated,
		EarliestExpiry:       sl.EarliestExpiry,
		DailyConsumptionRate: sl.DailyConsumptionRate,
	}, nil
}

func (s *Server) UpdateStockLevel(_ context.Context, req *formularyv1.UpdateStockLevelRequest) (*formularyv1.UpdateStockLevelResponse, error) {
	err := s.svc.UpdateStockLevel(req.SiteId, req.MedicationCode, int(req.Quantity), req.Unit, req.Reason, req.UpdatedBy)
	if err != nil {
		return nil, mapError(err)
	}
	return &formularyv1.UpdateStockLevelResponse{
		Success:     true,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) RecordDelivery(_ context.Context, req *formularyv1.RecordDeliveryRequest) (*formularyv1.RecordDeliveryResponse, error) {
	deliveryID := fmt.Sprintf("dlv-%d", time.Now().UnixNano())
	items := make([]service.DeliveryItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, service.DeliveryItemInput{
			MedicationCode: it.MedicationCode,
			Quantity:       int(it.Quantity),
			Unit:           it.Unit,
			BatchNumber:    it.BatchNumber,
			ExpiryDate:     it.ExpiryDate,
		})
	}
	recorded, err := s.svc.RecordDelivery(req.SiteId, req.ReceivedBy, req.DeliveryDate, deliveryID, items)
	if err != nil {
		return nil, mapError(err)
	}
	return &formularyv1.RecordDeliveryResponse{
		DeliveryId:    deliveryID,
		ItemsRecorded: int32(recorded),
	}, nil
}

func (s *Server) GetStockPrediction(_ context.Context, req *formularyv1.GetStockPredictionRequest) (*formularyv1.GetStockPredictionResponse, error) {
	pred, err := s.svc.GetStockPrediction(req.SiteId, req.MedicationCode)
	if err != nil {
		return nil, mapError(err)
	}
	return &formularyv1.GetStockPredictionResponse{
		DaysRemaining:     int32(pred.DaysRemaining),
		RiskLevel:         pred.RiskLevel,
		EarliestExpiry:    pred.EarliestExpiry,
		ExpiringQuantity:  int32(pred.ExpiringQuantity),
		RecommendedAction: pred.RecommendedAction,
	}, nil
}

func (s *Server) GetRedistributionSuggestions(_ context.Context, req *formularyv1.GetRedistributionSuggestionsRequest) (*formularyv1.GetRedistributionSuggestionsResponse, error) {
	suggestions, err := s.svc.GetRedistributionSuggestions(req.MedicationCode)
	if err != nil {
		return nil, mapError(err)
	}
	resp := &formularyv1.GetRedistributionSuggestionsResponse{}
	for _, s := range suggestions {
		resp.Suggestions = append(resp.Suggestions, &formularyv1.RedistributionSuggestion{
			FromSite:          s.FromSite,
			ToSite:            s.ToSite,
			SuggestedQuantity: int32(s.SuggestedQuantity),
			Rationale:         s.Rationale,
			FromSiteQuantity:  int32(s.FromSiteQuantity),
			ToSiteQuantity:    int32(s.ToSiteQuantity),
		})
	}
	return resp, nil
}
