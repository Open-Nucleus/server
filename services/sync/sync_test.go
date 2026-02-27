package sync_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"testing"
	"time"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/config"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/server"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/store"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport/localnet"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "modernc.org/sqlite"
)

type testEnv struct {
	syncClient    syncv1.SyncServiceClient
	conflictClient syncv1.ConflictServiceClient
	engine        *service.SyncEngine
	conflicts     *store.ConflictStore
	history       *store.HistoryStore
	peers         *store.PeerStore
	eventBus      *service.EventBus
}

func setupTestServer(t *testing.T) *testEnv {
	t.Helper()
	tmpDir := t.TempDir()

	cfg := &config.Config{
		GRPCPort: 0,
		Git: config.GitConfig{
			RepoPath:    tmpDir + "/data",
			AuthorName:  "test",
			AuthorEmail: "test@test.local",
		},
		Sync: config.SyncConfig{
			MaxConcurrent:    1,
			Cooldown:         time.Second,
			HandshakeTimeout: 5 * time.Second,
			TransferTimeout:  30 * time.Second,
			ChunkSize:        65536,
		},
		History: config.HistoryConfig{
			DBPath:     tmpDir + "/sync.db",
			MaxEntries: 1000,
		},
		Events: config.EventsConfig{BufferSize: 100},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	require.NoError(t, err)

	db, err := sql.Open("sqlite", cfg.History.DBPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	require.NoError(t, store.InitSchema(db))

	conflictStore := store.NewConflictStore(db)
	historyStore := store.NewHistoryStore(db, cfg.History.MaxEntries)
	peerStore := store.NewPeerStore(db)
	mergeDriver := merge.NewDriver(nil)
	eventBus := service.NewEventBus(cfg.Events.BufferSize)

	engine := service.NewSyncEngine(cfg, git, conflictStore, historyStore, peerStore, mergeDriver, eventBus, "test-node", "test-site")

	// Register local network transport
	ln := localnet.New("test-node", "test-site", "", 0)
	engine.RegisterTransport(ln)
	engine.RegisterTransport(&transport.StubAdapter{TransportName: "bluetooth"})

	srv := server.NewServer(cfg, engine, conflictStore, historyStore, peerStore, eventBus)
	grpcServer := grpc.NewServer()
	syncv1.RegisterSyncServiceServer(grpcServer, srv)
	syncv1.RegisterConflictServiceServer(grpcServer, srv)
	syncv1.RegisterNodeSyncServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return &testEnv{
		syncClient:    syncv1.NewSyncServiceClient(conn),
		conflictClient: syncv1.NewConflictServiceClient(conn),
		engine:        engine,
		conflicts:     conflictStore,
		history:       historyStore,
		peers:         peerStore,
		eventBus:      eventBus,
	}
}

func TestGetStatus_Idle(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	resp, err := env.syncClient.GetStatus(ctx, &syncv1.GetStatusRequest{})
	require.NoError(t, err)
	assert.Equal(t, "idle", resp.State)
	assert.Equal(t, "test-node", resp.NodeId)
	assert.Equal(t, "test-site", resp.SiteId)
}

func TestConflictCRUD(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	// Create conflict directly in store
	conflictID := uuid.New().String()
	localVersion := json.RawMessage(`{"resourceType":"Patient","id":"p1","name":[{"family":"Smith"}]}`)
	remoteVersion := json.RawMessage(`{"resourceType":"Patient","id":"p1","name":[{"family":"Jones"}]}`)

	err := env.conflicts.Create(&store.ConflictRecord{
		ID:            conflictID,
		ResourceType:  "Patient",
		ResourceID:    "p1",
		PatientID:     "p1",
		Level:         "block",
		Status:        "pending",
		LocalVersion:  localVersion,
		RemoteVersion: remoteVersion,
		ChangedFields: []string{"name"},
		Reason:        "conflicting patient identity fields",
		LocalNode:     "node-a",
		RemoteNode:    "node-b",
	})
	require.NoError(t, err)

	// List conflicts
	listResp, err := env.conflictClient.ListConflicts(ctx, &syncv1.ListConflictsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Conflicts, 1)
	assert.Equal(t, "block", listResp.Conflicts[0].Level)

	// Get conflict
	getResp, err := env.conflictClient.GetConflict(ctx, &syncv1.GetConflictRequest{ConflictId: conflictID})
	require.NoError(t, err)
	assert.Equal(t, "Patient", getResp.Conflict.ResourceType)
	assert.Equal(t, "p1", getResp.Conflict.PatientId)
}

func TestConflictResolution_AcceptLocal(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	conflictID := uuid.New().String()
	err := env.conflicts.Create(&store.ConflictRecord{
		ID: conflictID, ResourceType: "Patient", ResourceID: "p1",
		Level: "review", Status: "pending",
	})
	require.NoError(t, err)

	_, err = env.conflictClient.ResolveConflict(ctx, &syncv1.ResolveConflictRequest{
		ConflictId: conflictID,
		Resolution: "accept_local",
		Author:     "dr-jones",
	})
	require.NoError(t, err)

	// Verify resolved
	getResp, err := env.conflictClient.GetConflict(ctx, &syncv1.GetConflictRequest{ConflictId: conflictID})
	require.NoError(t, err)
	assert.Equal(t, "resolved", getResp.Conflict.Status)
}

func TestConflictDeferral(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	conflictID := uuid.New().String()
	err := env.conflicts.Create(&store.ConflictRecord{
		ID: conflictID, ResourceType: "Observation", ResourceID: "o1",
		Level: "review", Status: "pending",
	})
	require.NoError(t, err)

	_, err = env.conflictClient.DeferConflict(ctx, &syncv1.DeferConflictRequest{
		ConflictId: conflictID,
		Reason:     "need more context",
	})
	require.NoError(t, err)

	getResp, err := env.conflictClient.GetConflict(ctx, &syncv1.GetConflictRequest{ConflictId: conflictID})
	require.NoError(t, err)
	assert.Equal(t, "deferred", getResp.Conflict.Status)
}

func TestSyncHistory(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	// Trigger a sync (this records history)
	_, err := env.syncClient.TriggerSync(ctx, &syncv1.TriggerSyncRequest{TargetNode: "peer-1"})
	require.NoError(t, err)

	// Check history
	histResp, err := env.syncClient.GetHistory(ctx, &syncv1.GetSyncHistoryRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(histResp.Entries), 1)
	assert.Equal(t, "peer-1", histResp.Entries[0].PeerNode)
}

func TestPeerState(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	// Add peer
	err := env.peers.Upsert(&store.PeerRecord{
		NodeID:    "peer-1",
		SiteID:    "site-2",
		TheirHead: "abc123",
		Transport: "local_network",
	})
	require.NoError(t, err)

	// List peers
	resp, err := env.syncClient.ListPeers(ctx, &syncv1.ListPeersRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Peers, 1)
	assert.Equal(t, "peer-1", resp.Peers[0].NodeId)

	// Trust peer
	trustResp, err := env.syncClient.TrustPeer(ctx, &syncv1.TrustPeerRequest{NodeId: "peer-1"})
	require.NoError(t, err)
	assert.True(t, trustResp.Peer.Trusted)

	// Untrust peer
	untrustResp, err := env.syncClient.UntrustPeer(ctx, &syncv1.UntrustPeerRequest{NodeId: "peer-1"})
	require.NoError(t, err)
	assert.False(t, untrustResp.Peer.Trusted)
}

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := service.NewEventBus(10)

	sub := bus.Subscribe(nil) // all events
	defer bus.Unsubscribe(sub)

	bus.Publish(service.Event{
		Type:    service.EventSyncStarted,
		Payload: map[string]string{"sync_id": "s1"},
	})

	select {
	case event := <-sub.Ch:
		assert.Equal(t, service.EventSyncStarted, event.Type)
		assert.Equal(t, "s1", event.Payload["sync_id"])
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_Filter(t *testing.T) {
	bus := service.NewEventBus(10)

	sub := bus.Subscribe([]string{service.EventConflictNew})
	defer bus.Unsubscribe(sub)

	// Publish event that doesn't match filter
	bus.Publish(service.Event{Type: service.EventSyncStarted})

	// Publish event that matches
	bus.Publish(service.Event{Type: service.EventConflictNew, Payload: map[string]string{"id": "c1"}})

	select {
	case event := <-sub.Ch:
		assert.Equal(t, service.EventConflictNew, event.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestBundleExportImport(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	// Export (empty repo)
	exportResp, err := env.syncClient.ExportBundle(ctx, &syncv1.ExportBundleRequest{})
	require.NoError(t, err)
	assert.Equal(t, "nucleus-bundle-v1", exportResp.Format)

	// Import the exported bundle
	importResp, err := env.syncClient.ImportBundle(ctx, &syncv1.ImportBundleRequest{
		BundleData: exportResp.BundleData,
		Format:     "nucleus-bundle-v1",
		Author:     "test",
		NodeId:     "test-node",
	})
	require.NoError(t, err)
	assert.NotNil(t, importResp)
}

func TestTransportList(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	resp, err := env.syncClient.ListTransports(ctx, &syncv1.ListTransportsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Transports, 2)

	// Check local_network has capabilities
	var found bool
	for _, tr := range resp.Transports {
		if tr.Name == "local_network" {
			found = true
			assert.True(t, tr.Available)
			assert.Contains(t, tr.Capabilities, "discovery")
			assert.Contains(t, tr.Capabilities, "streaming")
		}
	}
	assert.True(t, found)
}

func TestMergeDriver_Integration(t *testing.T) {
	driver := merge.NewDriver(nil)

	// AutoMerge case — non-overlapping changes (no meta diff)
	base := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}]}`)
	local := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}],"telecom":[{"value":"555"}]}`)
	remote := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}],"address":[{"city":"Lagos"}]}`)

	result := driver.MergeFile("Patient", "p1", "p1", base, local, remote)
	assert.Equal(t, merge.AutoMerge, result.Level)
	assert.NotNil(t, result.MergedDoc)

	// Block case
	base = json.RawMessage(`{"resourceType":"AllergyIntolerance","criticality":"low"}`)
	local = json.RawMessage(`{"resourceType":"AllergyIntolerance","criticality":"high"}`)
	remote = json.RawMessage(`{"resourceType":"AllergyIntolerance","criticality":"unable-to-assess"}`)

	result = driver.MergeFile("AllergyIntolerance", "a1", "p1", base, local, remote)
	assert.Equal(t, merge.Block, result.Level)
}

func TestHealthCheck(t *testing.T) {
	env := setupTestServer(t)
	ctx := context.Background()

	resp, err := env.syncClient.Health(ctx, &syncv1.HealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "0.4.0", resp.Version)
}
