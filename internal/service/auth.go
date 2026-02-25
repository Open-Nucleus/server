package service

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

// authAdapter adapts the Auth gRPC client to the AuthService interface.
type authAdapter struct {
	pool *grpcclient.Pool
}

func NewAuthService(pool *grpcclient.Pool) AuthService {
	return &authAdapter{pool: pool}
}

func (a *authAdapter) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	_, err := a.pool.Conn("auth")
	if err != nil {
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}

	// In Phase 1, we call the gRPC service.
	// Since the backend doesn't exist yet, this will return an error
	// that the handler translates to SERVICE_UNAVAILABLE.
	return nil, fmt.Errorf("auth service unavailable: backend not connected")
}

func (a *authAdapter) Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, error) {
	_, err := a.pool.Conn("auth")
	if err != nil {
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}
	return nil, fmt.Errorf("auth service unavailable: backend not connected")
}

func (a *authAdapter) Logout(ctx context.Context, token string) error {
	_, err := a.pool.Conn("auth")
	if err != nil {
		return fmt.Errorf("auth service unavailable: %w", err)
	}
	return fmt.Errorf("auth service unavailable: backend not connected")
}

func (a *authAdapter) Whoami(ctx context.Context) (*WhoamiResponse, error) {
	_, err := a.pool.Conn("auth")
	if err != nil {
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}
	return nil, fmt.Errorf("auth service unavailable: backend not connected")
}
