package runtimecred

import (
	"database/sql"
	"testing"
)

func TestNewStoreRejectsNilDB(t *testing.T) {
	if _, err := NewStore(nil); err == nil {
		t.Fatal("NewStore(nil) succeeded")
	}
}

func TestStoreRequiresBoundTransaction(t *testing.T) {
	store, err := NewStore(&sql.DB{})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if _, err := store.ListTx(t.Context(), nil); err == nil {
		t.Fatal("ListTx accepted nil transaction")
	}
	if _, err := store.CreateTx(t.Context(), nil, Credential{}); err == nil {
		t.Fatal("CreateTx accepted nil transaction")
	}
	if _, err := store.UpdateTx(t.Context(), nil, "id", 1, "user", CredentialActive, "actor"); err == nil {
		t.Fatal("UpdateTx accepted nil transaction")
	}
	if _, err := store.RotateTx(t.Context(), nil, "id", 1, [32]byte{}, nil, nil, "key", "actor"); err == nil {
		t.Fatal("RotateTx accepted nil transaction")
	}
	if _, err := store.RevokeTx(t.Context(), nil, "id", 1, "actor"); err == nil {
		t.Fatal("RevokeTx accepted nil transaction")
	}
}

func TestStoreCreateRequiresAuditActor(t *testing.T) {
	store, err := NewStore(&sql.DB{})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	credential := Credential{Username: "safe.user", Status: CredentialActive, Origin: CredentialPanel, Revision: 1}
	if _, err := store.CreateTx(t.Context(), newDriverTx(t, nil), credential); err == nil {
		t.Fatal("CreateTx accepted credential without created/updated actor metadata")
	}
}
