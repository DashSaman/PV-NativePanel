package customer

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

type renewalStoreStub struct {
	ctx              RenewalContext
	plan             PlanPreset
	created          CreateRenewalTermRecord
	reboundRuntimeID string
	projectedTermID  string
	recorded         bool
	next             *ScheduledNextPlan
}

func (s *renewalStoreStub) DirectTenantID(context.Context, *sql.Tx) (string, error) {
	return s.ctx.TenantID, nil
}

func (s *renewalStoreStub) CreateUserTx(context.Context, *sql.Tx, CreateUserRecord) (User, error) {
	return User{}, errors.New("not used")
}

func (s *renewalStoreStub) CreateServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error) {
	return ServiceTerm{}, errors.New("not used")
}

func (s *renewalStoreStub) BindRuntimeCredentialTx(context.Context, *sql.Tx, string, string, string, string) error {
	return errors.New("not used")
}

func (s *renewalStoreStub) CreateSubscriptionTokenTx(context.Context, *sql.Tx, CreateSubscriptionTokenRecord) error {
	return errors.New("not used")
}

func (s *renewalStoreStub) CurrentRenewalContextTx(context.Context, *sql.Tx, string) (RenewalContext, error) {
	return s.ctx, nil
}

func (s *renewalStoreStub) PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error) {
	return s.plan, nil
}

func (s *renewalStoreStub) CreateRenewalTermTx(_ context.Context, _ *sql.Tx, record CreateRenewalTermRecord) (ServiceTerm, error) {
	s.created = record
	return ServiceTerm{
		ID:                "new-term",
		TenantID:          record.TenantID,
		UserID:            record.UserID,
		PlanID:            record.PlanID,
		QuotaBytes:        record.QuotaBytes,
		DurationSeconds:   record.DurationSeconds,
		NoExpiry:          record.NoExpiry,
		StartPolicy:       record.StartPolicy,
		PurchasedAt:       record.PurchasedAt,
		StartsAt:          record.StartsAt,
		ExpiresAt:         record.ExpiresAt,
		State:             record.State,
		RenewalKind:       record.RenewalKind,
		RenewedFromTermID: record.RenewedFromTermID,
	}, nil
}

func (s *renewalStoreStub) RebindRuntimeCredentialTx(_ context.Context, _ *sql.Tx, _ string, _ string, _ string, newTermID, runtimeID string) error {
	s.reboundRuntimeID = runtimeID
	if newTermID != "new-term" {
		return errors.New("wrong new term")
	}
	return nil
}

func (s *renewalStoreStub) ProjectSubscriptionToTermTx(_ context.Context, _ *sql.Tx, _ string, newTerm ServiceTerm) error {
	s.projectedTermID = newTerm.ID
	return nil
}

func (s *renewalStoreStub) RecordRenewalProfileTx(context.Context, *sql.Tx, string, string, time.Time, bool) error {
	s.recorded = true
	return nil
}

func (s *renewalStoreStub) ScheduledNextPlanTx(context.Context, *sql.Tx, string) (*ScheduledNextPlan, error) {
	return s.next, nil
}

func TestRenewUsingPlanCreatesNewTermAndPreservesRuntimeIdentity(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	quota := int64(100 * 1073741824)
	store := &renewalStoreStub{
		ctx: RenewalContext{
			TenantID:            "tenant",
			UserID:              "user",
			RuntimeCredentialID: "runtime-stable",
			Current:             ServiceTerm{ID: "old-term", UserID: "user", DurationSeconds: 30 * 86400, State: TermActive},
		},
		plan: PlanPreset{
			ID:              "plan-100",
			Name:            "100GB",
			QuotaBytes:      &quota,
			ValiditySeconds: 30 * 86400,
			StartPolicy:     StartOnCreation,
			ResetStrategy:   ResetNone,
			Enabled:         true,
		},
	}
	service := &Service{store: store, now: func() time.Time { return now }}
	result, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalUsingPlan, PlanID: "plan-100"})
	if err != nil {
		t.Fatal(err)
	}
	if result.PreviousTermID != "old-term" || result.ServiceTerm.ID != "new-term" {
		t.Fatalf("renewal did not create a new term: %#v", result)
	}
	if store.reboundRuntimeID != "runtime-stable" || result.RuntimeCredentialID != "runtime-stable" {
		t.Fatalf("runtime UUID changed during renewal: %#v", result)
	}
	if store.projectedTermID != "new-term" || !store.recorded {
		t.Fatalf("subscription/profile projection incomplete: %#v", store)
	}
	if store.created.RenewedFromTermID != "old-term" || store.created.RenewalKind != string(RenewalUsingPlan) {
		t.Fatalf("history lineage missing: %#v", store.created)
	}
}

func TestNextPlanCannotApplyBeforeCurrentServiceEnds(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	store := &renewalStoreStub{
		ctx: RenewalContext{
			TenantID:            "tenant",
			UserID:              "user",
			RuntimeCredentialID: "runtime",
			Current:             ServiceTerm{ID: "old", State: TermActive, ExpiresAt: &expires},
		},
		next: &ScheduledNextPlan{PlanID: "plan-next", SourceTermID: "old"},
		plan: PlanPreset{
			ID:              "plan-next",
			Name:            "next",
			ValiditySeconds: 30 * 86400,
			StartPolicy:     StartOnCreation,
			ResetStrategy:   ResetNone,
			Enabled:         true,
		},
	}
	service := &Service{store: store, now: func() time.Time { return now }}
	_, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalNextPlan})
	if !errors.Is(err, ErrNextPlanNotReady) {
		t.Fatalf("next plan applied early: %v", err)
	}
}

func TestNextPlanAppliesAfterExpiryAndIsConsumed(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expires := now.Add(-time.Minute)
	store := &renewalStoreStub{
		ctx: RenewalContext{
			TenantID:            "tenant",
			UserID:              "user",
			RuntimeCredentialID: "runtime",
			Current:             ServiceTerm{ID: "old", State: TermActive, ExpiresAt: &expires},
		},
		next: &ScheduledNextPlan{PlanID: "plan-next", SourceTermID: "old"},
		plan: PlanPreset{
			ID:              "plan-next",
			Name:            "next",
			ValiditySeconds: 30 * 86400,
			StartPolicy:     StartOnFirstSuccessfulConnection,
			ResetStrategy:   ResetNone,
			Enabled:         true,
		},
	}
	service := &Service{store: store, now: func() time.Time { return now }}
	result, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalNextPlan})
	if err != nil {
		t.Fatal(err)
	}
	if !result.NextPlanConsumed || store.created.State != TermPending || store.created.StartPolicy != StartOnFirstSuccessfulConnection {
		t.Fatalf("next plan semantics lost: result=%#v created=%#v", result, store.created)
	}
}
