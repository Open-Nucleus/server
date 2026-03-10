package local

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/services/auth/authservice"
)

// authSvc implements service.AuthService by calling
// authservice.AuthService directly (no gRPC).
type authSvc struct {
	real *authservice.AuthService
}

// NewAuthService creates a local adapter for auth operations.
func NewAuthService(real *authservice.AuthService) service.AuthService {
	return &authSvc{real: real}
}

func (a *authSvc) Login(_ context.Context, req *service.LoginRequest) (*service.LoginResponse, error) {
	sig := []byte(req.ChallengeResponse.Signature)

	accessToken, refreshToken, expiresAt, roleDef, siteID, err := a.real.Authenticate(
		req.DeviceID,
		sig,
		req.PractitionerID,
	)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return &service.LoginResponse{
		Token:        accessToken,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
		Role: service.RoleDTO{
			Code:        roleDef.Code,
			Display:     roleDef.Display,
			Permissions: roleDef.Permissions,
		},
		SiteID: siteID,
		NodeID: a.real.NodeID(),
	}, nil
}

func (a *authSvc) Refresh(_ context.Context, refreshToken string) (*service.RefreshResponse, error) {
	newAccess, newRefresh, expiresAt, err := a.real.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return &service.RefreshResponse{
		Token:        newAccess,
		ExpiresAt:    expiresAt,
		RefreshToken: newRefresh,
	}, nil
}

func (a *authSvc) Logout(_ context.Context, token string) error {
	if err := a.real.Logout(token); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	return nil
}

func (a *authSvc) Whoami(ctx context.Context) (*service.WhoamiResponse, error) {
	// Extract identity from the JWT claims in context, matching the
	// gRPC server behaviour (GetCurrentIdentity returns node identity
	// when no per-request claims are available).
	claims := model.ClaimsFromContext(ctx)
	if claims != nil {
		return &service.WhoamiResponse{
			Subject: claims.Subject,
			NodeID:  claims.Node,
			SiteID:  claims.Site,
			Role: service.RoleDTO{
				Code:        claims.Role,
				Permissions: claims.Permissions,
			},
		}, nil
	}

	// Fallback: return the node's own identity (same as gRPC server).
	return &service.WhoamiResponse{
		Subject: "system",
		NodeID:  a.real.NodeID(),
	}, nil
}
