package customer

import (
	"testing"
	"time"
)

func TestEffectiveAccessRequiresActiveAdminAndStartedTerm(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)

	got := EffectiveAccess(EffectiveAccessInput{
		AdminState: UserActive,
		TermState:  TermPending,
		StartsAt:   nil,
		ExpiresAt:  nil,
		Now:        now,
	})
	if got != AccessPending {
		t.Fatalf("pending unstarted term: got %q want %q", got, AccessPending)
	}

	started := now.Add(-time.Hour)
	got = EffectiveAccess(EffectiveAccessInput{
		AdminState: UserActive,
		TermState:  TermActive,
		StartsAt:   &started,
		ExpiresAt:  nil,
		Now:        now,
	})
	if got != AccessActive {
		t.Fatalf("started active term: got %q want %q", got, AccessActive)
	}
}

func TestEffectiveAccessReportsSuspendedBeforeCommercialState(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	started := now.Add(-48 * time.Hour)
	expired := now.Add(-time.Hour)

	got := EffectiveAccess(EffectiveAccessInput{
		AdminState: UserSuspended,
		TermState:  TermExpired,
		StartsAt:   &started,
		ExpiresAt:  &expired,
		Now:        now,
	})
	if got != AccessSuspended {
		t.Fatalf("administrative suspension must win: got %q want %q", got, AccessSuspended)
	}
}

func TestEffectiveAccessReportsExpiredWhenNowAtOrAfterExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	started := now.Add(-30 * 24 * time.Hour)

	for _, expiry := range []time.Time{now, now.Add(-time.Nanosecond)} {
		got := EffectiveAccess(EffectiveAccessInput{
			AdminState: UserActive,
			TermState:  TermActive,
			StartsAt:   &started,
			ExpiresAt:  &expiry,
			Now:        now,
		})
		if got != AccessExpired {
			t.Fatalf("expiry %s: got %q want %q", expiry, got, AccessExpired)
		}
	}
}

func TestEffectiveAccessReportsQuotaDepletedFromCommercialState(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	started := now.Add(-time.Hour)

	got := EffectiveAccess(EffectiveAccessInput{
		AdminState: UserActive,
		TermState:  TermQuotaDepleted,
		StartsAt:   &started,
		Now:        now,
	})
	if got != AccessQuotaDepleted {
		t.Fatalf("depleted term: got %q want %q", got, AccessQuotaDepleted)
	}
}

func TestUserAdminStateRejectsExpiredAndDepleted(t *testing.T) {
	for _, state := range []UserAdminState{"expired", "depleted", "quota_depleted", "ended", "pending"} {
		if ValidateUserAdminState(state) == nil {
			t.Fatalf("admin state %q must be rejected", state)
		}
	}
	for _, state := range []UserAdminState{UserDraft, UserActive, UserSuspended, UserRevoked} {
		if err := ValidateUserAdminState(state); err != nil {
			t.Fatalf("admin state %q unexpectedly rejected: %v", state, err)
		}
	}
}

func TestStartPolicyValidationSupportsApprovedPolicies(t *testing.T) {
	for _, policy := range []StartPolicy{StartOnCreation, StartOnFirstSuccessfulConnection, StartAtFixedTimestamp} {
		if err := ValidateStartPolicy(policy); err != nil {
			t.Fatalf("approved start policy %q rejected: %v", policy, err)
		}
	}
	if ValidateStartPolicy("first_seen") == nil {
		t.Fatal("unsupported start policy was accepted")
	}
}

func TestOnFirstConnectionTermStartsPendingWithoutExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	term, err := InitialTermTiming(StartOnFirstSuccessfulConnection, 30*24*time.Hour, now, nil)
	if err != nil {
		t.Fatalf("InitialTermTiming: %v", err)
	}
	if term.State != TermPending {
		t.Fatalf("state: got %q want %q", term.State, TermPending)
	}
	if term.StartsAt != nil || term.ExpiresAt != nil {
		t.Fatalf("first-connection term must have nil start/expiry before proven connection: start=%v expiry=%v", term.StartsAt, term.ExpiresAt)
	}
}

func TestOnCreationTermSnapshotsStartAndExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	term, err := InitialTermTiming(StartOnCreation, 30*24*time.Hour, now, nil)
	if err != nil {
		t.Fatalf("InitialTermTiming: %v", err)
	}
	if term.State != TermActive || term.StartsAt == nil || term.ExpiresAt == nil {
		t.Fatalf("unexpected creation timing: %+v", term)
	}
	if !term.StartsAt.Equal(now) || !term.ExpiresAt.Equal(now.Add(30*24*time.Hour)) {
		t.Fatalf("creation timing mismatch: start=%v expiry=%v", term.StartsAt, term.ExpiresAt)
	}
}

func TestFixedTimestampRequiresFutureStartAndDerivesExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	fixed := now.Add(2 * time.Hour)
	term, err := InitialTermTiming(StartAtFixedTimestamp, 60*24*time.Hour, now, &fixed)
	if err != nil {
		t.Fatalf("InitialTermTiming: %v", err)
	}
	if term.State != TermPending || term.StartsAt == nil || term.ExpiresAt == nil {
		t.Fatalf("unexpected fixed timing: %+v", term)
	}
	if !term.StartsAt.Equal(fixed) || !term.ExpiresAt.Equal(fixed.Add(60*24*time.Hour)) {
		t.Fatalf("fixed timing mismatch: start=%v expiry=%v", term.StartsAt, term.ExpiresAt)
	}

	past := now.Add(-time.Second)
	if _, err := InitialTermTiming(StartAtFixedTimestamp, 24*time.Hour, now, &past); err == nil {
		t.Fatal("past fixed start was accepted")
	}
}

func TestUsageCapabilityIsUnavailableUntilExactAccountingProof(t *testing.T) {
	capability := DefaultUsageCapability()
	if capability.Available {
		t.Fatal("exact usage must remain unavailable before accounting PoC")
	}
	if capability.Reason != "exact_accounting_not_proven" {
		t.Fatalf("reason: got %q", capability.Reason)
	}
}
