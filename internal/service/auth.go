package service

import (
	"context"
	"fmt"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

// authAdapter adapts the Auth gRPC client to the AuthService interface.
type authAdapter struct {
	pool *grpcclient.Pool
}

func NewAuthService(pool *grpcclient.Pool) AuthService {
	return &authAdapter{pool: pool}
}

func (a *authAdapter) client() (authv1.AuthServiceClient, error) {
	conn, err := a.pool.Conn("auth")
	if err != nil {
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}
	return authv1.NewAuthServiceClient(conn), nil
}

func (a *authAdapter) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.Authenticate(ctx, &authv1.AuthenticateRequest{
		DeviceId:       req.DeviceID,
		Signature:      []byte(req.ChallengeResponse.Signature),
		PractitionerId: req.PractitionerID,
	})
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return &LoginResponse{
		Token:        resp.AccessToken,
		ExpiresAt:    resp.ExpiresAt,
		RefreshToken: resp.RefreshToken,
		Role: RoleDTO{
			Code:        resp.Role.GetCode(),
			Display:     resp.Role.GetDisplay(),
			Permissions: resp.Role.GetPermissions(),
		},
		SiteID: resp.SiteId,
		NodeID: resp.NodeId,
	}, nil
}

func (a *authAdapter) Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.RefreshToken(ctx, &authv1.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return &RefreshResponse{
		Token:        resp.AccessToken,
		ExpiresAt:    resp.ExpiresAt,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (a *authAdapter) Logout(ctx context.Context, token string) error {
	c, err := a.client()
	if err != nil {
		return err
	}

	_, err = c.Logout(ctx, &authv1.LogoutRequest{Token: token})
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	return nil
}

func (a *authAdapter) Whoami(ctx context.Context) (*WhoamiResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetCurrentIdentity(ctx, &authv1.GetCurrentIdentityRequest{})
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return &WhoamiResponse{
		Subject: resp.Subject,
		NodeID:  resp.NodeId,
		SiteID:  resp.SiteId,
		Role: RoleDTO{
			Code:        resp.Role.GetCode(),
			Display:     resp.Role.GetDisplay(),
			Permissions: resp.Role.GetPermissions(),
		},
	}, nil
}
