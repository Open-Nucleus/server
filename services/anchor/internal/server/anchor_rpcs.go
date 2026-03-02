package server

import (
	"context"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
)

func (s *Server) GetStatus(_ context.Context, _ *anchorv1.GetAnchorStatusRequest) (*anchorv1.GetAnchorStatusResponse, error) {
	result, err := s.svc.GetStatus()
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.GetAnchorStatusResponse{
		State:          result.State,
		LastAnchorId:   result.LastAnchorID,
		LastAnchorTime: result.LastAnchorTime,
		MerkleRoot:     result.MerkleRoot,
		NodeDid:        result.NodeDID,
		QueueDepth:     int32(result.QueueDepth),
		Backend:        result.Backend,
	}, nil
}

func (s *Server) TriggerAnchor(_ context.Context, req *anchorv1.TriggerAnchorRequest) (*anchorv1.TriggerAnchorResponse, error) {
	result, err := s.svc.TriggerAnchor(req.Manual)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.TriggerAnchorResponse{
		AnchorId:   result.AnchorID,
		State:      result.State,
		MerkleRoot: result.MerkleRoot,
		GitHead:    result.GitHead,
		Skipped:    result.Skipped,
		Message:    result.Message,
	}, nil
}

func (s *Server) VerifyAnchor(_ context.Context, req *anchorv1.VerifyAnchorRequest) (*anchorv1.VerifyAnchorResponse, error) {
	result, err := s.svc.Verify(req.CommitHash)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.VerifyAnchorResponse{
		Verified:   result.Verified,
		AnchorId:   result.AnchorID,
		MerkleRoot: result.MerkleRoot,
		AnchoredAt: result.AnchoredAt,
		CommitHash: result.CommitHash,
		State:      result.State,
	}, nil
}

func (s *Server) GetHistory(_ context.Context, req *anchorv1.GetAnchorHistoryRequest) (*anchorv1.GetAnchorHistoryResponse, error) {
	page, perPage := paginationFromProto(req.Pagination)
	records, total, err := s.svc.GetHistory(page, perPage)
	if err != nil {
		return nil, mapError(err)
	}

	protoRecords := make([]*anchorv1.AnchorRecord, 0, len(records))
	for _, r := range records {
		protoRecords = append(protoRecords, &anchorv1.AnchorRecord{
			AnchorId:   r.AnchorID,
			MerkleRoot: r.MerkleRoot,
			GitHead:    r.GitHead,
			State:      r.State,
			Timestamp:  r.Timestamp,
			Backend:    r.Backend,
			TxId:       r.TxID,
			NodeDid:    r.NodeDID,
		})
	}

	return &anchorv1.GetAnchorHistoryResponse{
		Records:    protoRecords,
		Pagination: paginationToProto(page, perPage, total),
	}, nil
}

func paginationFromProto(pg *commonv1.PaginationRequest) (int, int) {
	page, perPage := 1, 25
	if pg != nil {
		if pg.Page > 0 {
			page = int(pg.Page)
		}
		if pg.PerPage > 0 {
			perPage = int(pg.PerPage)
		}
	}
	return page, perPage
}

func paginationToProto(page, perPage, total int) *commonv1.PaginationResponse {
	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	return &commonv1.PaginationResponse{
		Page:       int32(page),
		PerPage:    int32(perPage),
		Total:      int32(total),
		TotalPages: int32(totalPages),
	}
}
