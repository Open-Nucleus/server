package server

import (
	"context"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
)

func (s *Server) GetStatus(_ context.Context, _ *syncv1.GetStatusRequest) (*syncv1.GetStatusResponse, error) {
	state, syncID, peer, pending := s.engine.GetStatus()
	return &syncv1.GetStatusResponse{
		State:          string(state),
		LastSync:       s.engine.LastSyncTime(),
		PendingChanges: int32(pending),
		NodeId:         s.engine.NodeID(),
		SiteId:         s.engine.SiteID(),
		CurrentSyncId:  syncID,
		CurrentPeer:    peer,
	}, nil
}

func (s *Server) ListPeers(_ context.Context, _ *syncv1.ListPeersRequest) (*syncv1.ListPeersResponse, error) {
	peers, err := s.peers.List()
	if err != nil {
		return nil, mapError(err)
	}

	protoPeers := make([]*syncv1.PeerInfo, len(peers))
	for i, p := range peers {
		protoPeers[i] = &syncv1.PeerInfo{
			NodeId:    p.NodeID,
			SiteId:    p.SiteID,
			LastSeen:  p.LastSeen,
			State:     "offline",
			Trusted:   p.Trusted,
			Transport: p.Transport,
			TheirHead: p.TheirHead,
		}
	}
	return &syncv1.ListPeersResponse{Peers: protoPeers}, nil
}

func (s *Server) TriggerSync(_ context.Context, req *syncv1.TriggerSyncRequest) (*syncv1.TriggerSyncResponse, error) {
	syncID, err := s.engine.TriggerSync(req.TargetNode)
	if err != nil {
		return nil, mapError(err)
	}
	return &syncv1.TriggerSyncResponse{
		SyncId: syncID,
		State:  "completed",
	}, nil
}

func (s *Server) CancelSync(_ context.Context, req *syncv1.CancelSyncRequest) (*syncv1.CancelSyncResponse, error) {
	cancelled := s.engine.CancelSync(req.SyncId)
	return &syncv1.CancelSyncResponse{Cancelled: cancelled}, nil
}

func (s *Server) GetHistory(_ context.Context, req *syncv1.GetSyncHistoryRequest) (*syncv1.GetSyncHistoryResponse, error) {
	page := int(req.GetPagination().GetPage())
	perPage := int(req.GetPagination().GetPerPage())
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	entries, total, err := s.history.List(perPage, offset)
	if err != nil {
		return nil, mapError(err)
	}

	protoEntries := make([]*syncv1.SyncHistoryEntry, len(entries))
	for i, e := range entries {
		protoEntries[i] = &syncv1.SyncHistoryEntry{
			SyncId:            e.ID,
			PeerNode:          e.PeerNode,
			Transport:         e.Transport,
			Direction:         e.Direction,
			State:             e.State,
			StartedAt:         e.StartedAt,
			CompletedAt:       e.CompletedAt,
			ResourcesSent:     int32(e.ResourcesSent),
			ResourcesReceived: int32(e.ResourcesReceived),
			ConflictsDetected: int32(e.ConflictsDetected),
			LocalHeadBefore:   e.LocalHeadBefore,
			LocalHeadAfter:    e.LocalHeadAfter,
			ErrorMessage:      e.ErrorMessage,
		}
	}

	totalPages := (total + perPage - 1) / perPage
	return &syncv1.GetSyncHistoryResponse{
		Entries: protoEntries,
		Pagination: &commonv1.PaginationResponse{
			Page:       int32(page),
			PerPage:    int32(perPage),
			Total:      int32(total),
			TotalPages: int32(totalPages),
		},
	}, nil
}

func (s *Server) TrustPeer(_ context.Context, req *syncv1.TrustPeerRequest) (*syncv1.TrustPeerResponse, error) {
	if err := s.peers.Trust(req.NodeId); err != nil {
		return nil, mapError(err)
	}
	peer, err := s.peers.Get(req.NodeId)
	if err != nil {
		return nil, mapError(err)
	}
	return &syncv1.TrustPeerResponse{
		Peer: &syncv1.PeerInfo{
			NodeId:  peer.NodeID,
			SiteId:  peer.SiteID,
			Trusted: true,
		},
	}, nil
}

func (s *Server) UntrustPeer(_ context.Context, req *syncv1.UntrustPeerRequest) (*syncv1.UntrustPeerResponse, error) {
	if err := s.peers.Untrust(req.NodeId); err != nil {
		return nil, mapError(err)
	}
	peer, err := s.peers.Get(req.NodeId)
	if err != nil {
		return nil, mapError(err)
	}
	return &syncv1.UntrustPeerResponse{
		Peer: &syncv1.PeerInfo{
			NodeId:  peer.NodeID,
			SiteId:  peer.SiteID,
			Trusted: false,
		},
	}, nil
}
