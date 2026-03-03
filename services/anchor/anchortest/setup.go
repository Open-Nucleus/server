package anchortest

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"testing"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/server"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/service"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "modernc.org/sqlite"
)

// Env holds the test environment for an anchor service.
type Env struct {
	Addr   string
	Client anchorv1.AnchorServiceClient
	Svc    *service.AnchorService
}

// Start boots an anchor service in-process for testing.
func Start(t *testing.T, tmpDir string) *Env {
	t.Helper()

	env, cleanup, err := boot(tmpDir)
	if err != nil {
		t.Fatalf("start anchor: %v", err)
	}
	t.Cleanup(cleanup)

	conn, err := grpc.NewClient(env.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial anchor: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	env.Client = anchorv1.NewAnchorServiceClient(conn)
	return env
}

func boot(tmpDir string) (*Env, func(), error) {
	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	// SQLite for queue.
	dbPath := filepath.Join(tmpDir, "anchor-queue.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, cleanup, err
	}
	cleanups = append(cleanups, func() { db.Close() })

	if err := store.InitSchema(db); err != nil {
		return nil, cleanup, err
	}

	// Git store.
	repoPath := filepath.Join(tmpDir, "anchor-data")
	gs, err := gitstore.NewStore(repoPath, "anchor-test", "test@anchor.local")
	if err != nil {
		return nil, cleanup, err
	}

	// Seed sample FHIR files so the Merkle tree has data to work with.
	if err := seedSampleFiles(gs); err != nil {
		return nil, cleanup, fmt.Errorf("seed files: %w", err)
	}

	// Ed25519 keypair.
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, cleanup, err
	}

	// Create stores.
	queue := store.NewAnchorQueue(db)
	anchorStore := store.NewAnchorStore(gs)
	credStore := store.NewCredentialStore(gs)
	didStore := store.NewDIDStore(gs)

	// Create engines.
	anchorEngine := openanchor.NewStubBackend()
	identityEngine := openanchor.NewLocalIdentityEngine()

	svc := service.New(gs, anchorEngine, identityEngine, queue, anchorStore, credStore, didStore, priv)
	if err := svc.Bootstrap(); err != nil {
		return nil, cleanup, fmt.Errorf("bootstrap: %w", err)
	}

	srv := server.NewServer(svc)
	grpcServer := grpc.NewServer()
	anchorv1.RegisterAnchorServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, cleanup, err
	}

	go func() { _ = grpcServer.Serve(lis) }()
	cleanups = append(cleanups, func() { grpcServer.Stop() })

	return &Env{
		Addr: lis.Addr().String(),
		Svc:  svc,
	}, cleanup, nil
}

// seedSampleFiles creates sample FHIR files for Merkle tree computation.
func seedSampleFiles(gs gitstore.Store) error {
	patients := []map[string]any{
		{"resourceType": "Patient", "id": "p1", "name": []map[string]string{{"family": "Doe", "given": "John"}}, "gender": "male"},
		{"resourceType": "Patient", "id": "p2", "name": []map[string]string{{"family": "Smith", "given": "Jane"}}, "gender": "female"},
		{"resourceType": "Patient", "id": "p3", "name": []map[string]string{{"family": "Okafor", "given": "Chidi"}}, "gender": "male"},
	}

	encounters := []map[string]any{
		{"resourceType": "Encounter", "id": "e1", "status": "finished", "class": map[string]string{"code": "AMB"}},
		{"resourceType": "Encounter", "id": "e2", "status": "in-progress", "class": map[string]string{"code": "IMP"}},
	}

	for _, p := range patients {
		data, _ := json.MarshalIndent(p, "", "  ")
		path := fmt.Sprintf("patients/%s/%s.json", p["id"], p["id"])
		_, err := gs.WriteAndCommit(path, data, gitstore.CommitMessage{
			ResourceType: "Patient",
			Operation:    "CREATE",
			ResourceID:   p["id"].(string),
			NodeID:       "test-node",
			Author:       "test",
			SiteID:       "test-site",
		})
		if err != nil {
			return err
		}
	}

	for _, e := range encounters {
		data, _ := json.MarshalIndent(e, "", "  ")
		path := fmt.Sprintf("patients/p1/encounters/%s.json", e["id"])
		_, err := gs.WriteAndCommit(path, data, gitstore.CommitMessage{
			ResourceType: "Encounter",
			Operation:    "CREATE",
			ResourceID:   e["id"].(string),
			NodeID:       "test-node",
			Author:       "test",
			SiteID:       "test-site",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
