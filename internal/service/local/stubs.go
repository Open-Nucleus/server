package local

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/service"
)

// stubSentinelService implements service.SentinelService with "not available" responses.
// Sentinel is a Python process; when running standalone, connect via gRPC instead.
type stubSentinelService struct{}

func NewStubSentinelService() service.SentinelService {
	return &stubSentinelService{}
}

func (s *stubSentinelService) ListAlerts(_ context.Context, page, perPage int) (*service.AlertListResponse, error) {
	return &service.AlertListResponse{
		Alerts:  []service.AlertDetail{},
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *stubSentinelService) GetAlertSummary(_ context.Context) (*service.AlertSummaryResponse, error) {
	return &service.AlertSummaryResponse{}, nil
}

func (s *stubSentinelService) GetAlert(_ context.Context, alertID string) (*service.AlertDetail, error) {
	return nil, fmt.Errorf("sentinel agent not connected: alert %s", alertID)
}

func (s *stubSentinelService) AcknowledgeAlert(_ context.Context, alertID string) (*service.AlertDetail, error) {
	return nil, fmt.Errorf("sentinel agent not connected: alert %s", alertID)
}

func (s *stubSentinelService) DismissAlert(_ context.Context, alertID, _ string) (*service.AlertDetail, error) {
	return nil, fmt.Errorf("sentinel agent not connected: alert %s", alertID)
}

// stubSupplyService implements service.SupplyService with "not available" responses.
type stubSupplyService struct{}

func NewStubSupplyService() service.SupplyService {
	return &stubSupplyService{}
}

func (s *stubSupplyService) GetInventory(_ context.Context, page, perPage int) (*service.InventoryListResponse, error) {
	return &service.InventoryListResponse{
		Items:   []service.InventoryItemDetail{},
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *stubSupplyService) GetInventoryItem(_ context.Context, itemCode string) (*service.InventoryItemDetail, error) {
	return nil, fmt.Errorf("sentinel agent not connected: item %s", itemCode)
}

func (s *stubSupplyService) RecordDelivery(_ context.Context, _ *service.RecordDeliveryRequest) (*service.RecordDeliveryResponse, error) {
	return nil, fmt.Errorf("sentinel agent not connected")
}

func (s *stubSupplyService) GetPredictions(_ context.Context) (*service.PredictionsResponse, error) {
	return &service.PredictionsResponse{
		Predictions: []service.SupplyPrediction{},
	}, nil
}

func (s *stubSupplyService) GetRedistribution(_ context.Context) (*service.RedistributionResponse, error) {
	return &service.RedistributionResponse{
		Suggestions: []service.RedistributionSuggestion{},
	}, nil
}
