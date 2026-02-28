// Package patienttest exports helpers for spinning up an in-process Patient
// Service for integration/E2E tests.
package patienttest

import (
	"net"
	"testing"
	"time"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/config"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/server"
	"google.golang.org/grpc"
)

// Env holds the running Patient Service test environment.
type Env struct {
	Addr string
}

// Start boots an in-process Patient Service on a dynamic port.
func Start(t *testing.T, tmpDir string) *Env {
	t.Helper()

	git, err := gitstore.NewStore(tmpDir+"/patient-data", "test", "test@test.local")
	if err != nil {
		t.Fatal(err)
	}

	idx, err := sqliteindex.NewIndex(tmpDir + "/patient.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { idx.Close() })

	cfg := &config.Config{
		GRPCPort: 0,
		Matching: config.MatchingConfig{
			DefaultThreshold: 0.7,
			MaxResults:       10,
			FuzzyMaxDistance: 2,
		},
	}

	pw := pipeline.NewWriter(git, idx, 5*time.Second)
	srv := server.NewServer(cfg, pw, idx, git)

	grpcServer := grpc.NewServer()
	patientv1.RegisterPatientServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })

	return &Env{Addr: lis.Addr().String()}
}
