package customer

import (
	"testing"
	"time"
)

func TestQuotaGBToBytesUsesBinaryGB(t *testing.T) {
	gb := int64(50)
	got, err := QuotaGBToBytes(&gb)
	if err != nil {
		t.Fatalf("QuotaGBToBytes returned error: %v", err)
	}
	if got == nil {
		t.Fatal("QuotaGBToBytes returned nil for finite quota")
	}
	want := int64(50 * 1073741824)
	if *got != want {
		t.Fatalf("QuotaGBToBytes = %d, want %d", *got, want)
	}
}

func TestQuotaGBToBytesNilMeansUnlimited(t *testing.T) {
	got, err := QuotaGBToBytes(nil)
	if err != nil {
		t.Fatalf("QuotaGBToBytes(nil) returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("QuotaGBToBytes(nil) = %d, want nil", *got)
	}
}

func TestQuotaGBToBytesRejectsZeroAndOverflow(t *testing.T) {
	zero := int64(0)
	if _, err := QuotaGBToBytes(&zero); err == nil {
		t.Fatal("QuotaGBToBytes accepted zero quota")
	}
	overflow := int64(8589934592)
	if _, err := QuotaGBToBytes(&overflow); err == nil {
		t.Fatal("QuotaGBToBytes accepted a value that overflows int64 bytes")
	}
}

func TestNormalizeValidityFirstUseStartsPending(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	got, duration, err := NormalizeValidity(ValidityInput{
		Mode:         ValidityOnFirstSuccessfulConnection,
		DurationDays: 30,
	}, now)
	if err != nil {
		t.Fatalf("NormalizeValidity returned error: %v", err)
	}
	if duration != 30*24*time.Hour {
		t.Fatalf("duration = %s, want 720h", duration)
	}
	if got.State != TermPending || got.StartsAt != nil || got.ExpiresAt != nil {
		t.Fatalf("first-use timing = %#v, want pending with nil start/expiry", got)
	}
}

func TestNormalizeValidityCreationStartsNow(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.FixedZone("IRST", 3*3600+30*60))
	got, duration, err := NormalizeValidity(ValidityInput{
		Mode:         ValidityOnCreation,
		DurationDays: 30,
	}, now)
	if err != nil {
		t.Fatalf("NormalizeValidity returned error: %v", err)
	}
	if duration != 30*24*time.Hour {
		t.Fatalf("duration = %s, want 720h", duration)
	}
	wantStart := now.UTC()
	wantExpiry := wantStart.Add(30 * 24 * time.Hour)
	if got.State != TermActive || got.StartsAt == nil || !got.StartsAt.Equal(wantStart) {
		t.Fatalf("starts_at = %v, want %v", got.StartsAt, wantStart)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expires_at = %v, want %v", got.ExpiresAt, wantExpiry)
	}
}

func TestNormalizeValidityFixedExpiryPreservesInstant(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expiry := time.Date(2026, 10, 15, 23, 59, 0, 0, time.FixedZone("IRST", 3*3600+30*60))
	got, duration, err := NormalizeValidity(ValidityInput{
		Mode:      ValidityFixedExpiry,
		ExpiresAt: &expiry,
	}, now)
	if err != nil {
		t.Fatalf("NormalizeValidity returned error: %v", err)
	}
	if got.State != TermActive || got.StartsAt == nil || !got.StartsAt.Equal(now.UTC()) {
		t.Fatalf("fixed-expiry starts_at = %v, want %v", got.StartsAt, now.UTC())
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expiry) {
		t.Fatalf("fixed-expiry expires_at = %v, want same instant %v", got.ExpiresAt, expiry)
	}
	if duration != expiry.UTC().Sub(now.UTC()) {
		t.Fatalf("duration = %s, want %s", duration, expiry.UTC().Sub(now.UTC()))
	}
}

func TestNormalizeValidityRejectsPastFixedExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expiry := now.Add(-time.Second)
	if _, _, err := NormalizeValidity(ValidityInput{Mode: ValidityFixedExpiry, ExpiresAt: &expiry}, now); err == nil {
		t.Fatal("NormalizeValidity accepted a past fixed expiry")
	}
}

func TestNormalizeValidityRejectsZeroDays(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	for _, mode := range []ValidityMode{ValidityOnCreation, ValidityOnFirstSuccessfulConnection} {
		if _, _, err := NormalizeValidity(ValidityInput{Mode: mode, DurationDays: 0}, now); err == nil {
			t.Fatalf("NormalizeValidity accepted zero days for %s", mode)
		}
	}
}

func TestNormalizeValidityRejectsStaleFieldsFromOtherModes(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	expiry := now.Add(48 * time.Hour)
	if _, _, err := NormalizeValidity(ValidityInput{
		Mode:         ValidityFixedExpiry,
		DurationDays: 30,
		ExpiresAt:    &expiry,
	}, now); err == nil {
		t.Fatal("fixed expiry accepted stale duration_days")
	}
	if _, _, err := NormalizeValidity(ValidityInput{
		Mode:         ValidityOnFirstSuccessfulConnection,
		DurationDays: 30,
		ExpiresAt:    &expiry,
	}, now); err == nil {
		t.Fatal("duration mode accepted stale expires_at")
	}
}

func TestActivateOnFirstUseIsIdempotent(t *testing.T) {
	observed := time.Date(2026, 8, 30, 9, 10, 11, 0, time.UTC)
	pending := TermTiming{State: TermPending}
	first, changed, err := ActivateOnFirstUse(pending, 30*24*time.Hour, observed)
	if err != nil {
		t.Fatalf("ActivateOnFirstUse returned error: %v", err)
	}
	if !changed || first.State != TermActive || first.StartsAt == nil || first.ExpiresAt == nil {
		t.Fatalf("first activation = %#v changed=%v", first, changed)
	}
	if !first.StartsAt.Equal(observed) || !first.ExpiresAt.Equal(observed.Add(30*24*time.Hour)) {
		t.Fatalf("activation timestamps = start %v expiry %v", first.StartsAt, first.ExpiresAt)
	}
	second, changed, err := ActivateOnFirstUse(first, 30*24*time.Hour, observed.Add(time.Hour))
	if err != nil {
		t.Fatalf("second ActivateOnFirstUse returned error: %v", err)
	}
	if changed {
		t.Fatal("second activation changed an already-started term")
	}
	if second.StartsAt == nil || !second.StartsAt.Equal(observed) || second.ExpiresAt == nil || !second.ExpiresAt.Equal(observed.Add(30*24*time.Hour)) {
		t.Fatalf("second activation moved timestamps: %#v", second)
	}
}
