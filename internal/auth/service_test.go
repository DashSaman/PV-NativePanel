package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

type fakeLoginStore struct {
	actor             ActorRecord
	lookupErr         error
	createTokenHash   []byte
	createCSRFHash    []byte
	createdSessionID  string
	failureCalls      int
	successCalls      int
	totpFactor        TOTPFactorRecord
	consumeTOTPOK     bool
	createSessionCall int
}

func (f *fakeLoginStore) LookupActor(context.Context, string) (ActorRecord, error) {
	if f.lookupErr != nil {
		return ActorRecord{}, f.lookupErr
	}
	return f.actor, nil
}
func (f *fakeLoginStore) RecordLoginFailure(context.Context, string) (*time.Time, error) {
	f.failureCalls++
	return nil, nil
}
func (f *fakeLoginStore) RecordLoginSuccess(context.Context, string) error {
	f.successCalls++
	return nil
}
func (f *fakeLoginStore) CreateSession(_ context.Context, _ string, tokenHash, csrfHash []byte, _ string, _ []byte, _, _ time.Time) (string, error) {
	f.createSessionCall++
	f.createTokenHash = append([]byte(nil), tokenHash...)
	f.createCSRFHash = append([]byte(nil), csrfHash...)
	if f.createdSessionID == "" {
		f.createdSessionID = "00000000-0000-0000-0000-000000000099"
	}
	return f.createdSessionID, nil
}
func (f *fakeLoginStore) GetTOTPFactorPreAuth(context.Context, string) (TOTPFactorRecord, error) {
	return f.totpFactor, nil
}
func (f *fakeLoginStore) ConsumeTOTPStepPreAuth(context.Context, string, int64) (bool, error) {
	return f.consumeTOTPOK, nil
}
func (f *fakeLoginStore) AppendAudit(context.Context, *string, string, string, string) error {
	return nil
}

func TestLoginCreatesOnlyHashedSessionMaterialInStore(t *testing.T) {
	passwordHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeLoginStore{actor: ActorRecord{
		ID:           "00000000-0000-0000-0000-000000000001",
		Role:         "owner",
		PasswordHash: passwordHash,
		Status:       "active",
	}}
	service, err := NewService(store, make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Login(t.Context(), LoginInput{
		Email: "owner@example.invalid", Password: "correct horse battery staple", UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if result.SessionToken == "" || result.CSRFToken == "" || result.SessionID == "" {
		t.Fatalf("missing login material: %#v", result)
	}
	if len(store.createTokenHash) != 32 || len(store.createCSRFHash) != 32 {
		t.Fatalf("store did not receive SHA-256 material: token=%d csrf=%d", len(store.createTokenHash), len(store.createCSRFHash))
	}
	if store.createSessionCall != 1 || store.successCalls != 1 {
		t.Fatalf("unexpected calls: create=%d success=%d", store.createSessionCall, store.successCalls)
	}
}

func TestLoginMFARequiredDoesNotCreateSession(t *testing.T) {
	passwordHash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeLoginStore{actor: ActorRecord{
		ID: "00000000-0000-0000-0000-000000000001", Role: "owner", PasswordHash: passwordHash,
		Status: "active", MFARequired: true, TOTPConfirmed: true,
	}}
	service, err := NewService(store, make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Login(t.Context(), LoginInput{Email: "owner@example.invalid", Password: "secret-password"})
	if !errors.Is(err, ErrMFARequired) {
		t.Fatalf("error=%v want ErrMFARequired", err)
	}
	if store.createSessionCall != 0 {
		t.Fatal("session created before MFA")
	}
}

func TestLoginUnknownAndWrongPasswordShareGenericError(t *testing.T) {
	unknown, err := NewService(&fakeLoginStore{lookupErr: sql.ErrNoRows}, make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	_, unknownErr := unknown.Login(t.Context(), LoginInput{Email: "missing@example.invalid", Password: "guess"})
	if !errors.Is(unknownErr, ErrInvalidCredentials) {
		t.Fatalf("unknown actor error=%v", unknownErr)
	}

	passwordHash, err := HashPassword("real-password")
	if err != nil {
		t.Fatal(err)
	}
	wrong, err := NewService(&fakeLoginStore{actor: ActorRecord{ID: "00000000-0000-0000-0000-000000000001", Role: "owner", PasswordHash: passwordHash, Status: "active"}}, make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	_, wrongErr := wrong.Login(t.Context(), LoginInput{Email: "owner@example.invalid", Password: "guess"})
	if !errors.Is(wrongErr, ErrInvalidCredentials) {
		t.Fatalf("wrong-password error=%v", wrongErr)
	}
}
