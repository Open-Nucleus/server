package service

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	sentinelv1 "github.com/FibrinLab/open-nucleus/gen/proto/sentinel/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type supplyAdapter struct {
	pool *grpcclient.Pool
}

func NewSupplyService(pool *grpcclient.Pool) SupplyService {
	return &supplyAdapter{pool: pool}
}

func (s *supplyAdapter) client() (sentinelv1.SentinelServiceClient, error) {
	conn, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("supply service unavailable: %w", err)
	}
	return sentinelv1.NewSentinelServiceClient(conn), nil
}

func (s *supplyAdapter) GetInventory(ctx context.Context, page, perPage int) (*InventoryListResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetInventory(ctx, &sentinelv1.GetInventoryRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, err
	}
	return &InventoryListResponse{
		Items:      toInventoryItemDetails(resp.Items),
		Page:       paginationPage(resp.Pagination),
		PerPage:    paginationPerPage(resp.Pagination),
		Total:      paginationTotal(resp.Pagination),
		TotalPages: paginationTotalPages(resp.Pagination),
	}, nil
}

func (s *supplyAdapter) GetInventoryItem(ctx context.Context, itemCode string) (*InventoryItemDetail, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetInventoryItem(ctx, &sentinelv1.GetInventoryItemRequest{
		ItemCode: itemCode,
	})
	if err != nil {
		return nil, err
	}
	return toInventoryItemDetail(resp.Item), nil
}

func (s *supplyAdapter) RecordDelivery(ctx context.Context, req *RecordDeliveryRequest) (*RecordDeliveryResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	items := make([]*sentinelv1.DeliveryItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &sentinelv1.DeliveryItem{
			ItemCode:    it.ItemCode,
			Quantity:    int32(it.Quantity),
			Unit:        it.Unit,
			BatchNumber: it.BatchNumber,
			ExpiryDate:  it.ExpiryDate,
		})
	}
	resp, err := c.RecordDelivery(ctx, &sentinelv1.RecordDeliveryRequest{
		SiteId:       req.SiteID,
		Items:        items,
		ReceivedBy:   req.ReceivedBy,
		DeliveryDate: req.DeliveryDate,
	})
	if err != nil {
		return nil, err
	}
	return &RecordDeliveryResponse{
		DeliveryID:    resp.DeliveryId,
		ItemsRecorded: int(resp.ItemsRecorded),
	}, nil
}

func (s *supplyAdapter) GetPredictions(ctx context.Context) (*PredictionsResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetPredictions(ctx, &sentinelv1.GetPredictionsRequest{})
	if err != nil {
		return nil, err
	}
	preds := make([]SupplyPrediction, 0, len(resp.Predictions))
	for _, p := range resp.Predictions {
		preds = append(preds, SupplyPrediction{
			ItemCode:               p.ItemCode,
			Display:                p.Display,
			CurrentQuantity:        int(p.CurrentQuantity),
			PredictedDaysRemaining: int(p.PredictedDaysRemaining),
			RiskLevel:              p.RiskLevel,
			RecommendedAction:      p.RecommendedAction,
		})
	}
	return &PredictionsResponse{Predictions: preds}, nil
}

func (s *supplyAdapter) GetRedistribution(ctx context.Context) (*RedistributionResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetRedistribution(ctx, &sentinelv1.GetRedistributionRequest{})
	if err != nil {
		return nil, err
	}
	suggestions := make([]RedistributionSuggestion, 0, len(resp.Suggestions))
	for _, s := range resp.Suggestions {
		suggestions = append(suggestions, RedistributionSuggestion{
			ItemCode:          s.ItemCode,
			FromSite:          s.FromSite,
			ToSite:            s.ToSite,
			SuggestedQuantity: int(s.SuggestedQuantity),
			Rationale:         s.Rationale,
		})
	}
	return &RedistributionResponse{Suggestions: suggestions}, nil
}

// --- Proto → DTO converters ---

func toInventoryItemDetail(item *sentinelv1.InventoryItem) *InventoryItemDetail {
	if item == nil {
		return nil
	}
	return &InventoryItemDetail{
		ItemCode:     item.ItemCode,
		Display:      item.Display,
		Quantity:     int(item.Quantity),
		Unit:         item.Unit,
		SiteID:       item.SiteId,
		LastUpdated:  item.LastUpdated,
		ReorderLevel: int(item.ReorderLevel),
	}
}

func toInventoryItemDetails(items []*sentinelv1.InventoryItem) []InventoryItemDetail {
	out := make([]InventoryItemDetail, 0, len(items))
	for _, item := range items {
		if d := toInventoryItemDetail(item); d != nil {
			out = append(out, *d)
		}
	}
	return out
}
