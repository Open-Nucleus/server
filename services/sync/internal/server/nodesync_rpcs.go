package server

import (
	"context"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Handshake(_ context.Context, _ *syncv1.HandshakeRequest) (*syncv1.HandshakeResponse, error) {
	// Will be implemented when node-to-node sync is active
	return nil, status.Error(codes.Unimplemented, "node-to-node handshake not yet implemented")
}

func (s *Server) RequestPack(_ *syncv1.RequestPackRequest, _ syncv1.NodeSyncService_RequestPackServer) error {
	return status.Error(codes.Unimplemented, "node-to-node pack request not yet implemented")
}

func (s *Server) SendPack(_ syncv1.NodeSyncService_SendPackServer) error {
	return status.Error(codes.Unimplemented, "node-to-node pack send not yet implemented")
}
