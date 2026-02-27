package server

import (
	"context"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
)

func (s *Server) ValidateToken(_ context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, errCode, err := s.svc.ValidateToken(req.Token)
	if err != nil {
		return &authv1.ValidateTokenResponse{
			Valid:     false,
			ErrorCode: errCode,
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid: true,
		Claims: &authv1.TokenClaims{
			Sub:         claims.Subject,
			DeviceId:    claims.DeviceID,
			NodeId:      claims.NodeID,
			SiteId:      claims.SiteID,
			Role:        claims.Role,
			Permissions: claims.Permissions,
			SiteScope:   claims.SiteScope,
			Jti:         claims.ID,
		},
	}, nil
}

func (s *Server) CheckPermission(_ context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionResponse, error) {
	allowed, reason, err := s.svc.CheckPermission(req.Token, req.Permission, req.TargetSiteId)
	if err != nil {
		return &authv1.CheckPermissionResponse{
			Allowed: false,
			Reason:  err.Error(),
		}, nil
	}
	return &authv1.CheckPermissionResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}
