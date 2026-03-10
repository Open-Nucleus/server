package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/pkg/consent"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
)

// setup creates a ConsentManager backed by real SQLite for testing.
func setup(t *testing.T) *consent.Manager {
	t.Helper()
	dir := t.TempDir()
	idx, err := sqliteindex.NewIndex(dir + "/test.db")
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}
	t.Cleanup(func() { idx.Close() })
	git := &memGit{files: make(map[string][]byte)}
	return consent.NewManager(idx, git, slog.Default())
}

type memGit struct {
	files map[string][]byte
}

func (g *memGit) WriteAndCommit(path string, data []byte, msg gitstore.CommitMessage) (string, error) {
	g.files[path] = append([]byte(nil), data...)
	return "hash", nil
}
func (g *memGit) Read(path string) ([]byte, error) {
	data, ok := g.files[path]
	if !ok {
		return nil, nil
	}
	return data, nil
}
func (g *memGit) LogPath(string, int) ([]gitstore.CommitInfo, error) { return nil, nil }
func (g *memGit) Head() (string, error)                              { return "head", nil }
func (g *memGit) TreeWalk(func(string, []byte) error) error          { return nil }
func (g *memGit) Rollback() error                                    { return nil }

// serveWithConsent creates a chi router with consent middleware and a test handler.
func serveWithConsent(mgr *consent.Manager, claims *model.NucleusClaims) *chi.Mux {
	r := chi.NewRouter()

	// Inject claims into context
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if claims != nil {
				ctx := context.WithValue(r.Context(), model.CtxClaims, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})

	r.Route("/api/v1/patients/{id}", func(r chi.Router) {
		r.Use(ConsentCheck(mgr, slog.Default()))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
	})

	return r
}

func TestConsentCheck_NilManager(t *testing.T) {
	r := serveWithConsent(nil, &model.NucleusClaims{DeviceID: "d1", Role: "physician"})

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestConsentCheck_AdminBypass(t *testing.T) {
	mgr := setup(t)
	claims := &model.NucleusClaims{DeviceID: "admin-device", Role: "site_administrator"}
	r := serveWithConsent(mgr, claims)

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConsentCheck_Denied(t *testing.T) {
	mgr := setup(t)
	claims := &model.NucleusClaims{DeviceID: "device-1", Role: "physician"}
	r := serveWithConsent(mgr, claims)

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConsentCheck_Allowed(t *testing.T) {
	mgr := setup(t)

	// Grant consent
	_, _, err := mgr.GrantConsent("p1", "device-1", fhir.ConsentScopePatientPrivacy, nil, "")
	if err != nil {
		t.Fatalf("GrantConsent: %v", err)
	}

	claims := &model.NucleusClaims{DeviceID: "device-1", Role: "physician"}
	r := serveWithConsent(mgr, claims)

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 after consent, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConsentCheck_BreakGlass(t *testing.T) {
	mgr := setup(t)
	claims := &model.NucleusClaims{DeviceID: "device-bg", Role: "physician"}
	r := serveWithConsent(mgr, claims)

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/", nil)
	req.Header.Set("X-Break-Glass", "true")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for break-glass, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConsentCheck_NoClaims(t *testing.T) {
	mgr := setup(t)
	r := serveWithConsent(mgr, nil)

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
