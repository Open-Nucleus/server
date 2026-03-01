package patienttest

import (
	"fmt"
	"net"
	"time"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/config"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/server"
	"google.golang.org/grpc"
)

// StartStandalone boots an in-process Patient Service without requiring *testing.T.
// Returns the environment and a cleanup function.
func StartStandalone(tmpDir string) (*Env, func(), error) {
	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	git, err := gitstore.NewStore(tmpDir+"/patient-data", "smoke", "smoke@test.local")
	if err != nil {
		return nil, cleanup, fmt.Errorf("patient gitstore: %w", err)
	}

	idx, err := sqliteindex.NewIndex(tmpDir + "/patient.db")
	if err != nil {
		return nil, cleanup, fmt.Errorf("patient sqlite: %w", err)
	}
	cleanups = append(cleanups, func() { idx.Close() })

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
		return nil, cleanup, fmt.Errorf("patient listen: %w", err)
	}

	go func() { _ = grpcServer.Serve(lis) }()
	cleanups = append(cleanups, func() { grpcServer.Stop() })

	return &Env{Addr: lis.Addr().String()}, cleanup, nil
}
