package smart

import (
	"testing"
	"time"
)

func TestLaunchStore_CreateAndConsume(t *testing.T) {
	store := NewLaunchStore(5 * time.Second)
	defer store.Close()

	token, err := store.Create("client-1", "patient-123", "encounter-456", "device-1")
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if token == "" {
		t.Fatal("Create() returned empty token")
	}

	lt, err := store.Consume(token)
	if err != nil {
		t.Fatalf("Consume() error: %v", err)
	}
	if lt.ClientID != "client-1" {
		t.Errorf("Consume().ClientID = %q, want %q", lt.ClientID, "client-1")
	}
	if lt.PatientID != "patient-123" {
		t.Errorf("Consume().PatientID = %q, want %q", lt.PatientID, "patient-123")
	}
	if lt.EncounterID != "encounter-456" {
		t.Errorf("Consume().EncounterID = %q, want %q", lt.EncounterID, "encounter-456")
	}
}

func TestLaunchStore_Expired(t *testing.T) {
	store := NewLaunchStore(1 * time.Millisecond)
	defer store.Close()

	token, _ := store.Create("client-1", "patient-123", "", "device-1")
	time.Sleep(10 * time.Millisecond)

	_, err := store.Consume(token)
	if err == nil {
		t.Fatal("Consume() expected error for expired token")
	}
}

func TestLaunchStore_AlreadyConsumed(t *testing.T) {
	store := NewLaunchStore(5 * time.Second)
	defer store.Close()

	token, _ := store.Create("client-1", "patient-123", "", "device-1")

	_, err := store.Consume(token)
	if err != nil {
		t.Fatalf("first Consume() error: %v", err)
	}

	_, err = store.Consume(token)
	if err == nil {
		t.Fatal("second Consume() expected error for already consumed token")
	}
}

func TestLaunchStore_MissingClientID(t *testing.T) {
	store := NewLaunchStore(5 * time.Second)
	defer store.Close()

	_, err := store.Create("", "patient-123", "", "device-1")
	if err == nil {
		t.Fatal("Create() expected error for missing client_id")
	}
}
