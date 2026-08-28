package customer

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

var fakeFirstUseResult bool
var fakeFirstUseError error
var fakeFirstUseCredential string
var fakeFirstUseObserved time.Time

func (f *fakeCustomerStore) ActivateFirstUseTx(_ context.Context, _ *sql.Tx, runtimeCredentialID string, observedAt time.Time) (bool, error) {
	fakeFirstUseCredential = runtimeCredentialID
	fakeFirstUseObserved = observedAt
	return fakeFirstUseResult, fakeFirstUseError
}

func TestServiceActivateFirstUseDelegatesToAtomicStore(t *testing.T) {
	fakeFirstUseResult = true
	fakeFirstUseError = nil
	fakeFirstUseCredential = ""
	fakeFirstUseObserved = time.Time{}
	observed := time.Date(2026, 8, 29, 5, 6, 7, 0, time.FixedZone("test", 3*60*60))
	service := NewService(&fakeCustomerStore{}, nil, time.Now)
	activated, err := service.ActivateFirstUse(context.Background(), nil, "runtime-1", observed)
	if err != nil {
		t.Fatalf("ActivateFirstUse() error = %v", err)
	}
	if !activated || fakeFirstUseCredential != "runtime-1" || !fakeFirstUseObserved.Equal(observed.UTC()) {
		t.Fatalf("activated=%v credential=%q observed=%v", activated, fakeFirstUseCredential, fakeFirstUseObserved)
	}
}

func TestServiceActivateFirstUsePropagatesStoreError(t *testing.T) {
	fakeFirstUseResult = false
	fakeFirstUseError = errors.New("boom")
	service := NewService(&fakeCustomerStore{}, nil, time.Now)
	if _, err := service.ActivateFirstUse(context.Background(), nil, "runtime-1", time.Now()); err == nil {
		t.Fatal("ActivateFirstUse() swallowed store error")
	}
	fakeFirstUseError = nil
}
