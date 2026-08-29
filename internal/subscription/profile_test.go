package subscription

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestServiceResolveProfileReturnsCommercialMetadataWithoutFakingUsage(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	ciphertext, nonce, err := runtimecred.EncryptSecret(key, []byte("customer-secret-123"))
	if err != nil {
		t.Fatal(err)
	}
	rawBytes := make([]byte, 32)
	for i := range rawBytes {
		rawBytes[i] = byte(0x40 + i)
	}
	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	hash := sha256.Sum256(rawBytes)
	quota := int64(50 * 1024 * 1024 * 1024)
	starts := time.Date(2026, 8, 29, 14, 0, 0, 0, time.UTC)
	expires := starts.Add(30 * 24 * time.Hour)

	store := &fakeStore{wantHash: hash, record: Record{
		RuntimeCredentialID: "runtime-1",
		Username:            "Amir22",
		SecretCiphertext:    ciphertext,
		SecretNonce:         nonce,
		EncryptionKeyID:     "runtime-v1",
		UserState:           "active",
		TermState:           "active",
		QuotaBytes:          &quota,
		DurationSeconds:     30 * 24 * 60 * 60,
		StartPolicy:         "on_creation",
		StartsAt:            &starts,
		ExpiresAt:           &expires,
	}}
	service, err := NewService(store, key, "runtime-v1")
	if err != nil {
		t.Fatal(err)
	}

	profile, err := service.ResolveProfile(context.Background(), raw, "namir.softarg.ir:443")
	if err != nil {
		t.Fatalf("ResolveProfile() error = %v", err)
	}
	if profile.Username != "Amir22" || profile.DirectURI == "" {
		t.Fatalf("profile identity/delivery = %#v", profile)
	}
	if profile.QuotaBytes == nil || *profile.QuotaBytes != quota {
		t.Fatalf("profile quota = %v", profile.QuotaBytes)
	}
	if profile.StartPolicy != "on_creation" || profile.ExpiresAt == nil || !profile.ExpiresAt.Equal(expires) {
		t.Fatalf("profile validity = %#v", profile)
	}
	if profile.UsageAvailable {
		t.Fatal("exact usage must remain unavailable until accounting proof is complete")
	}
	if profile.UsedBytes != nil || profile.RemainingBytes != nil {
		t.Fatal("profile must not fabricate used/remaining byte counters")
	}
}
