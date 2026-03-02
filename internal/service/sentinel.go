package service

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	sentinelv1 "github.com/FibrinLab/open-nucleus/gen/proto/sentinel/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type sentinelAdapter struct {
	pool *grpcclient.Pool
}

func NewSentinelService(pool *grpcclient.Pool) SentinelService {
	return &sentinelAdapter{pool: pool}
}

func (s *sentinelAdapter) client() (sentinelv1.SentinelServiceClient, error) {
	conn, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("sentinel service unavailable: %w", err)
	}
	return sentinelv1.NewSentinelServiceClient(conn), nil
}

func (s *sentinelAdapter) ListAlerts(ctx context.Context, page, perPage int) (*AlertListResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ListAlerts(ctx, &sentinelv1.ListAlertsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, err
	}
	return &AlertListResponse{
		Alerts:     toAlertDetails(resp.Alerts),
		Page:       paginationPage(resp.Pagination),
		PerPage:    paginationPerPage(resp.Pagination),
		Total:      paginationTotal(resp.Pagination),
		TotalPages: paginationTotalPages(resp.Pagination),
	}, nil
}

func (s *sentinelAdapter) GetAlertSummary(ctx context.Context) (*AlertSummaryResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetAlertSummary(ctx, &sentinelv1.GetAlertSummaryRequest{})
	if err != nil {
		return nil, err
	}
	return &AlertSummaryResponse{
		Total:          int(resp.Total),
		Critical:       int(resp.Critical),
		Warning:        int(resp.Warning),
		Info:           int(resp.Info),
		Unacknowledged: int(resp.Unacknowledged),
	}, nil
}

func (s *sentinelAdapter) GetAlert(ctx context.Context, alertID string) (*AlertDetail, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetAlert(ctx, &sentinelv1.GetAlertRequest{AlertId: alertID})
	if err != nil {
		return nil, err
	}
	return toAlertDetail(resp.Alert), nil
}

func (s *sentinelAdapter) AcknowledgeAlert(ctx context.Context, alertID string) (*AlertDetail, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.AcknowledgeAlert(ctx, &sentinelv1.AcknowledgeAlertRequest{
		AlertId: alertID,
	})
	if err != nil {
		return nil, err
	}
	return toAlertDetail(resp.Alert), nil
}

func (s *sentinelAdapter) DismissAlert(ctx context.Context, alertID, reason string) (*AlertDetail, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.DismissAlert(ctx, &sentinelv1.DismissAlertRequest{
		AlertId: alertID,
		Reason:  reason,
	})
	if err != nil {
		return nil, err
	}
	return toAlertDetail(resp.Alert), nil
}

// --- Proto → DTO converters ---

func toAlertDetail(a *sentinelv1.Alert) *AlertDetail {
	if a == nil {
		return nil
	}
	return &AlertDetail{
		ID:             a.Id,
		Type:           a.Type,
		Severity:       a.Severity,
		Status:         a.Status,
		Title:          a.Title,
		Description:    a.Description,
		PatientID:      a.PatientId,
		CreatedAt:      a.CreatedAt,
		AcknowledgedAt: a.AcknowledgedAt,
		AcknowledgedBy: a.AcknowledgedBy,
	}
}

func toAlertDetails(alerts []*sentinelv1.Alert) []AlertDetail {
	out := make([]AlertDetail, 0, len(alerts))
	for _, a := range alerts {
		if d := toAlertDetail(a); d != nil {
			out = append(out, *d)
		}
	}
	return out
}
