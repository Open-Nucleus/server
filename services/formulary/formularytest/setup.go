package formularytest

import (
	"database/sql"
	_ "embed"
	"net"
	"testing"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/dosing"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/server"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/service"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "modernc.org/sqlite"
)

//go:embed testdata/medications/medications.json
var medicationsJSON []byte

//go:embed testdata/interactions/interaction-rules.json
var interactionsJSON []byte

// Env holds the test environment for a formulary service.
type Env struct {
	Addr   string
	Client formularyv1.FormularyServiceClient
	Svc    *service.FormularyService
}

// Start boots a formulary service in-process for testing.
func Start(t *testing.T, tmpDir string) *Env {
	t.Helper()

	env, cleanup, err := boot(tmpDir)
	if err != nil {
		t.Fatalf("start formulary: %v", err)
	}
	t.Cleanup(cleanup)

	conn, err := grpc.NewClient(env.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial formulary: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	env.Client = formularyv1.NewFormularyServiceClient(conn)
	return env
}

func boot(tmpDir string) (*Env, func(), error) {
	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	db, err := sql.Open("sqlite", tmpDir+"/formulary.db")
	if err != nil {
		return nil, cleanup, err
	}
	cleanups = append(cleanups, func() { db.Close() })

	if err := store.InitSchema(db); err != nil {
		return nil, cleanup, err
	}

	drugDB := store.NewDrugDB()
	if err := drugDB.LoadFromJSON(medicationsJSON); err != nil {
		return nil, cleanup, err
	}

	interactions := store.NewInteractionIndex()
	if err := interactions.LoadFromJSON(interactionsJSON); err != nil {
		return nil, cleanup, err
	}

	stockStore := store.NewStockStore(db)
	dosingEngine := dosing.NewStubEngine()

	svc := service.New(drugDB, interactions, stockStore, dosingEngine)
	srv := server.NewServer(svc)
	grpcServer := grpc.NewServer()
	formularyv1.RegisterFormularyServiceServer(grpcServer, srv)

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
