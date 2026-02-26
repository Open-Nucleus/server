package gitstore

import (
	"os"
	"strings"
	"testing"
	"time"
)

func tempStore(t *testing.T) Store {
	t.Helper()
	dir, err := os.MkdirTemp("", "gitstore-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	store, err := NewStore(dir, "test", "test@test.com")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestWriteAndCommit_CreatesFile(t *testing.T) {
	s := tempStore(t)

	hash, err := s.WriteAndCommit("patients/p1/Patient.json", []byte(`{"id":"p1"}`), CommitMessage{
		ResourceType: "Patient",
		Operation:    "CREATE",
		ResourceID:   "p1",
		NodeID:       "node-1",
		Author:       "dr-test",
		SiteID:       "clinic-1",
		Timestamp:    time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Error("expected non-empty commit hash")
	}
}

func TestRead_ReturnsContent(t *testing.T) {
	s := tempStore(t)

	content := []byte(`{"id":"p1","resourceType":"Patient"}`)
	_, err := s.WriteAndCommit("patients/p1/Patient.json", content, CommitMessage{
		ResourceType: "Patient",
		Operation:    "CREATE",
		ResourceID:   "p1",
		Timestamp:    time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := s.Read("patients/p1/Patient.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Errorf("got %s, want %s", string(data), string(content))
	}
}

func TestLogPath_ReturnsHistory(t *testing.T) {
	s := tempStore(t)

	for i := range 3 {
		_, err := s.WriteAndCommit("patients/p1/Patient.json", []byte(`{"v":`+string(rune('0'+i))+`}`), CommitMessage{
			ResourceType: "Patient",
			Operation:    "UPDATE",
			ResourceID:   "p1",
			Timestamp:    time.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	infos, err := s.LogPath("patients/p1/", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 3 {
		t.Errorf("expected 3 commits, got %d", len(infos))
	}
}

func TestHead_ReturnsHash(t *testing.T) {
	s := tempStore(t)

	// Empty repo
	hash, err := s.Head()
	if err != nil {
		t.Fatal(err)
	}
	if hash != "" {
		t.Errorf("expected empty hash for empty repo, got %s", hash)
	}

	// After commit
	_, err = s.WriteAndCommit("test.txt", []byte("hello"), CommitMessage{
		ResourceType: "Patient",
		Operation:    "CREATE",
		ResourceID:   "p1",
		Timestamp:    time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	hash, err = s.Head()
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Error("expected non-empty hash after commit")
	}
}

func TestTreeWalk_VisitsAllFiles(t *testing.T) {
	s := tempStore(t)

	files := map[string]string{
		"patients/p1/Patient.json":         `{"id":"p1"}`,
		"patients/p1/encounters/e1.json":   `{"id":"e1"}`,
		"patients/p1/observations/o1.json": `{"id":"o1"}`,
	}

	for path, content := range files {
		_, err := s.WriteAndCommit(path, []byte(content), CommitMessage{
			ResourceType: "Patient",
			Operation:    "CREATE",
			ResourceID:   "p1",
			Timestamp:    time.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	visited := make(map[string]bool)
	err := s.TreeWalk(func(path string, data []byte) error {
		visited[path] = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for path := range files {
		if !visited[path] {
			t.Errorf("file %s not visited", path)
		}
	}
}

func TestRollback_RestoresState(t *testing.T) {
	s := tempStore(t)

	_, err := s.WriteAndCommit("test.txt", []byte("v1"), CommitMessage{
		ResourceType: "Patient",
		Operation:    "CREATE",
		ResourceID:   "p1",
		Timestamp:    time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Rollback should not error
	if err := s.Rollback(); err != nil {
		t.Fatal(err)
	}
}

func TestCommitMessage_Format(t *testing.T) {
	cm := CommitMessage{
		ResourceType: "Encounter",
		Operation:    "CREATE",
		ResourceID:   "enc-a1b2c3d4",
		NodeID:       "node-sheffield-01",
		Author:       "dr-osutuk",
		SiteID:       "clinic-maiduguri-03",
		Timestamp:    time.Date(2026, 3, 15, 9, 42, 0, 0, time.UTC),
	}

	msg := cm.Format()
	if !strings.HasPrefix(msg, "[Encounter] CREATE enc-a1b2c3d4") {
		t.Errorf("unexpected header: %s", msg)
	}
	if !strings.Contains(msg, "node: node-sheffield-01") {
		t.Error("missing node")
	}
	if !strings.Contains(msg, "fhir_version: R4") {
		t.Error("missing fhir_version")
	}
}

func TestCommitMessage_Parse(t *testing.T) {
	original := CommitMessage{
		ResourceType: "Patient",
		Operation:    "UPDATE",
		ResourceID:   "p-123",
		NodeID:       "node-1",
		Author:       "dr-test",
		SiteID:       "clinic-1",
		Timestamp:    time.Date(2026, 3, 15, 9, 42, 0, 0, time.UTC),
	}

	msg := original.Format()
	parsed, err := ParseCommitMessage(msg)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.ResourceType != original.ResourceType {
		t.Errorf("ResourceType: got %s, want %s", parsed.ResourceType, original.ResourceType)
	}
	if parsed.Operation != original.Operation {
		t.Errorf("Operation: got %s, want %s", parsed.Operation, original.Operation)
	}
	if parsed.ResourceID != original.ResourceID {
		t.Errorf("ResourceID: got %s, want %s", parsed.ResourceID, original.ResourceID)
	}
	if parsed.NodeID != original.NodeID {
		t.Errorf("NodeID: got %s, want %s", parsed.NodeID, original.NodeID)
	}
}
