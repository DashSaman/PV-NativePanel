package auth

import (
	"database/sql"
	"testing"
	"time"
)

func TestNewStoreRejectsNilDB(t *testing.T) {
	if _, err := NewStore(nil); err == nil {
		t.Fatal("NewStore(nil) succeeded")
	}
}

func TestStoreRejectsNonSHA256SessionMaterialBeforeSQL(t *testing.T) {
	db := &sql.DB{}
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if _, err := store.CreateSession(t.Context(), "00000000-0000-0000-0000-000000000001", []byte("short"), make([]byte, 32), "00000000-0000-0000-0000-000000000002", nil, time.Now().Add(time.Hour), time.Now().Add(12*time.Hour)); err == nil {
		t.Fatal("CreateSession accepted a non-SHA256 token hash")
	}
	if _, err := store.BeginAuthenticated(t.Context(), []byte("short")); err == nil {
		t.Fatal("BeginAuthenticated accepted a non-SHA256 token hash")
	}
}

func TestStoreRejectsInvalidActorAndSessionIdentifiersBeforeSQL(t *testing.T) {
	db := &sql.DB{}
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if _, err := store.RecordLoginFailure(t.Context(), ""); err == nil {
		t.Fatal("RecordLoginFailure accepted empty actor ID")
	}
	if _, err := store.RevokeSession(t.Context(), make([]byte, 31)); err == nil {
		t.Fatal("RevokeSession accepted a 31-byte hash")
	}
}
