package server

import (
	"context"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	"github.com/FibrinLab/open-nucleus/pkg/auth"
)

func (s *Server) RegisterDevice(_ context.Context, req *authv1.RegisterDeviceRequest) (*authv1.RegisterDeviceResponse, error) {
	device, err := s.svc.RegisterDevice(
		req.PublicKey,
		req.PractitionerId,
		req.SiteId,
		req.DeviceName,
		req.Role,
		req.BootstrapSecret,
	)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RegisterDeviceResponse{Device: deviceToProto(device)}, nil
}

func (s *Server) GetChallenge(_ context.Context, req *authv1.GetChallengeRequest) (*authv1.GetChallengeResponse, error) {
	nonce, expiresAt, err := s.svc.GetChallenge(req.DeviceId)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.GetChallengeResponse{
		Challenge: &authv1.ChallengeResponse{
			Nonce:     nonce,
			ExpiresAt: timestamppb(expiresAt),
		},
	}, nil
}

func (s *Server) Authenticate(_ context.Context, req *authv1.AuthenticateRequest) (*authv1.AuthenticateResponse, error) {
	// The client sends: device_id, signature (Ed25519 sig of nonce), practitioner_id
	// We need to extract the nonce from the store and verify
	accessToken, refreshToken, expiresAt, roleDef, siteID, err := s.svc.AuthenticateWithNonce(
		req.DeviceId,
		req.Signature[:32], // nonce bytes (first 32 bytes)
		req.Signature,       // full signature
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Role:         roleToProto(roleDef),
		SiteId:       siteID,
		NodeId:       s.svc.NodeID(),
	}, nil
}

func (s *Server) RefreshToken(_ context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	access, refresh, expiresAt, err := s.svc.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RefreshTokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *Server) Logout(_ context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	err := s.svc.Logout(req.Token)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.LogoutResponse{}, nil
}

func (s *Server) GetCurrentIdentity(_ context.Context, _ *authv1.GetCurrentIdentityRequest) (*authv1.GetCurrentIdentityResponse, error) {
	// In production, this would extract claims from the request context
	// For now, return node identity
	return &authv1.GetCurrentIdentityResponse{
		Subject:  "system",
		DeviceId: "",
		NodeId:   s.svc.NodeID(),
		SiteId:   "",
	}, nil
}

// --- Helpers ---

func roleToProto(r auth.RoleDefinition) *authv1.RoleInfo {
	return &authv1.RoleInfo{
		Code:        r.Code,
		Display:     r.Display,
		Permissions: r.Permissions,
		SiteScope:   r.SiteScope,
	}
}
