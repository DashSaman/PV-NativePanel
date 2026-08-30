package customer

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type resetTestStore struct {
	target      UsageResetTarget
	claimed     bool
	mutationID  string
	resetResult AccountingResetResult
	event       UsageResetEvent
	resetCalls  int
	eventCalls  int
	auditCalls  int
	lastHash    []byte
}

func (s *resetTestStore) DirectTenantID(context.Context, *sql.Tx) (string, error) { return "", nil }
func (s *resetTestStore) CreateUserTx(context.Context, *sql.Tx, CreateUserRecord) (User, error) {
	return User{}, nil
}
func (s *resetTestStore) CreateServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error) {
	return ServiceTerm{}, nil
}
func (s *resetTestStore) BindRuntimeCredentialTx(context.Context, *sql.Tx, string, string, string, string) error {
	return nil
}
func (s *resetTestStore) CreateSubscriptionTokenTx(context.Context, *sql.Tx, CreateSubscriptionTokenRecord) error {
	return nil
}
func (s *resetTestStore) UsageResetTargetTx(context.Context, *sql.Tx, string) (UsageResetTarget, error) {
	return s.target, nil
}
func (s *resetTestStore) ClaimUsageResetTx(_ context.Context, _ *sql.Tx, _ UsageResetTarget, _ string, _ string, hash []byte) (string, bool, error) {
	s.lastHash = append([]byte(nil), hash...)
	return s.mutationID, s.claimed, nil
}
func (s *resetTestStore) UsageResetEventByMutationKeyTx(context.Context, *sql.Tx, string) (UsageResetEvent, error) {
	return s.event, nil
}
func (s *resetTestStore) ResetDirectAccountingTx(context.Context, *sql.Tx, string, time.Time, time.Duration) (AccountingResetResult, error) {
	s.resetCalls++
	return s.resetResult, nil
}
func (s *resetTestStore) AppendUsageResetEventTx(context.Context, *sql.Tx, UsageResetTarget, string, string, AccountingResetResult) (UsageResetEvent, error) {
	s.eventCalls++
	return s.event, nil
}
func (s *resetTestStore) AppendUsageResetAuditTx(context.Context, *sql.Tx, UsageResetEvent) error {
	s.auditCalls++
	return nil
}

func TestResetCustomerUsageIsIdempotentAndDoesNotNeedRuntimeMutation(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	store := &resetTestStore{
		target:  UsageResetTarget{TenantID: "tenant-1", UserID: "user-1", ServiceTermID: "term-1"},
		claimed: true, mutationID: "mutation-1",
		resetResult: AccountingResetResult{Resettable: true, Reason: "reset", ServiceTermID: "term-1", TenantID: "tenant-1", UserID: "user-1", PreviousUploadBytes: 100, PreviousDownloadBytes: 200, PreviousUsedBytes: 300, ResetAt: now, ServiceState: TermActive},
		event:       UsageResetEvent{ID: "reset-1", TenantID: "tenant-1", UserID: "user-1", ServiceTermID: "term-1", ActorID: "owner-1", MutationKeyID: "mutation-1", Reason: UsageResetManual, ResetAt: now, PreviousUploadBytes: 100, PreviousDownloadBytes: 200, PreviousUsedBytes: 300},
	}
	service := NewService(store, nil, func() time.Time { return now })
	result, err := service.ResetCustomerUsage(context.Background(), nil, "owner-1", "usage-reset-0001", "user-1")
	if err != nil {
		t.Fatalf("ResetCustomerUsage() error = %v", err)
	}
	if result.IdempotentReplay || result.Event.ID != "reset-1" || store.resetCalls != 1 || store.eventCalls != 1 || store.auditCalls != 1 {
		t.Fatalf("reset result=%#v calls reset/event/audit=%d/%d/%d", result, store.resetCalls, store.eventCalls, store.auditCalls)
	}
	if len(store.lastHash) != 32 {
		t.Fatalf("request hash length=%d", len(store.lastHash))
	}
}

func TestResetCustomerUsageReplayReturnsStoredEventWithoutSecondReset(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	store := &resetTestStore{
		target:  UsageResetTarget{TenantID: "tenant-1", UserID: "user-1", ServiceTermID: "term-1"},
		claimed: false, mutationID: "mutation-1",
		event: UsageResetEvent{ID: "reset-1", MutationKeyID: "mutation-1", ResetAt: now},
	}
	service := NewService(store, nil, func() time.Time { return now.Add(time.Minute) })
	result, err := service.ResetCustomerUsage(context.Background(), nil, "owner-1", "usage-reset-0001", "user-1")
	if err != nil {
		t.Fatalf("ResetCustomerUsage() replay error = %v", err)
	}
	if !result.IdempotentReplay || result.Event.ID != "reset-1" || store.resetCalls != 0 || store.eventCalls != 0 || store.auditCalls != 0 {
		t.Fatalf("replay result=%#v calls=%d/%d/%d", result, store.resetCalls, store.eventCalls, store.auditCalls)
	}
}

func TestResetCustomerUsageRefusesUnsafeAccountingState(t *testing.T) {
	store := &resetTestStore{
		target:  UsageResetTarget{TenantID: "tenant-1", UserID: "user-1", ServiceTermID: "term-1"},
		claimed: true, mutationID: "mutation-1",
		resetResult: AccountingResetResult{Resettable: false, Reason: "reservation_pending"},
	}
	service := NewService(store, nil, time.Now)
	_, err := service.ResetCustomerUsage(context.Background(), nil, "owner-1", "usage-reset-0002", "user-1")
	if err != ErrUsageResetReservationPending {
		t.Fatalf("error=%v want ErrUsageResetReservationPending", err)
	}
	if store.eventCalls != 0 || store.auditCalls != 0 {
		t.Fatal("unsafe reset appended history/audit")
	}
}
