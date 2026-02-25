package grpcclient

import (
	"context"
	"fmt"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Pool manages gRPC connections to all backend services.
type Pool struct {
	conns map[string]*grpc.ClientConn
}

// NewPool dials all configured backend services.
func NewPool(cfg config.GRPCConfig) (*Pool, error) {
	targets := map[string]string{
		"auth":      cfg.AuthService,
		"patient":   cfg.PatientService,
		"sync":      cfg.SyncService,
		"formulary": cfg.FormularyService,
		"anchor":    cfg.AnchorService,
		"sentinel":  cfg.SentinelAgent,
	}

	pool := &Pool{conns: make(map[string]*grpc.ClientConn)}

	for name, addr := range targets {
		if addr == "" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
		conn, err := grpc.DialContext(ctx, addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		cancel()
		if err != nil {
			// Non-fatal: services may not be running yet.
			// Handlers will get SERVICE_UNAVAILABLE when calling.
			pool.conns[name] = nil
			continue
		}
		pool.conns[name] = conn
	}

	return pool, nil
}

// Conn returns the connection for a named service.
func (p *Pool) Conn(name string) (*grpc.ClientConn, error) {
	conn, ok := p.conns[name]
	if !ok || conn == nil {
		return nil, fmt.Errorf("service %q is not available", name)
	}
	return conn, nil
}

// Close shuts down all connections.
func (p *Pool) Close() {
	for _, conn := range p.conns {
		if conn != nil {
			conn.Close()
		}
	}
}

// RequestTimeout returns a context with the configured request timeout.
func RequestTimeout(cfg config.GRPCConfig) (context.Context, context.CancelFunc) {
	timeout := cfg.RequestTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}
