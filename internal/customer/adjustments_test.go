package customer

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type adjustmentStore struct {
	fakeCustomerStore
	quotaDelta int64
	extendBy   int64
}

func (s *adjustmentStore) AddCurrentServiceQuotaTx(_ context.Context, _ *sql.Tx, _ string, deltaBytes int64) (ServiceTerm, error) {
	s.quotaDelta = deltaBytes
	quota := int64(70 * 1073741824)
	return ServiceTerm{ID: "term-1", UserID: "user-1", QuotaBytes: &quota, State: TermActive}, nil
}

func (s *adjustmentStore) ExtendCurrentServiceTx(_ context.Context, _ *sql.Tx, _ string, seconds int64, _ time.Time) (ServiceTerm, error) {
	s.extendBy = seconds
	return ServiceTerm{ID: "term-1", UserID: "user-1", DurationSeconds: 60 * 86400, State: TermActive}, nil
}

func TestAddCustomerVolumeConvertsBinaryGBAndKeepsRuntimeOutOfScope(t *testing.T) {
	store := &adjustmentStore{}
	service := NewService(store, nil, time.Now)

	term, err := service.AddCustomerVolume(context.Background(), nil, "user-1", 20)
	if err != nil {
		t.Fatalf("AddCustomerVolume() error = %v", err)
	}
	if store.quotaDelta != 20*1073741824 {
		t.Fatalf("quota delta = %d, want %d", store.quotaDelta, int64(20*1073741824))
	}
	if term.QuotaBytes == nil || *term.QuotaBytes != 70*1073741824 {
		t.Fatalf("quota = %#v", term.QuotaBytes)
	}
}

func TestExtendCustomerTimeUsesWholeDays(t *testing.T) {
	store := &adjustmentStore{}
	service := NewService(store, nil, func() time.Time { return time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC) })

	_, err := service.ExtendCustomerTime(context.Background(), nil, "user-1", 30)
	if err != nil {
		t.Fatalf("ExtendCustomerTime() error = %v", err)
	}
	if store.extendBy != 30*86400 {
		t.Fatalf("extend seconds = %d, want %d", store.extendBy, int64(30*86400))
	}
}

func TestAdjustmentsRejectNonPositiveValues(t *testing.T) {
	service := NewService(&adjustmentStore{}, nil, time.Now)
	if _, err := service.AddCustomerVolume(context.Background(), nil, "user-1", 0); err == nil {
		t.Fatal("AddCustomerVolume() accepted zero")
	}
	if _, err := service.ExtendCustomerTime(context.Background(), nil, "user-1", -1); err == nil {
		t.Fatal("ExtendCustomerTime() accepted negative days")
	}
}
