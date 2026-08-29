package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrMFARequired        = errors.New("auth: MFA required")
)

const (
	sessionIdleLifetime     = time.Hour
	sessionAbsoluteLifetime = 12 * time.Hour
)

type loginStore interface {
	LookupActor(context.Context, string) (ActorRecord, error)
	RecordLoginFailure(context.Context, string) (*time.Time, error)
	RecordLoginSuccess(context.Context, string) error
	CreateSession(context.Context, string, []byte, []byte, string, []byte, time.Time, time.Time) (string, error)
	GetTOTPFactorPreAuth(context.Context, string) (TOTPFactorRecord, error)
	ConsumeTOTPStepPreAuth(context.Context, string, int64) (bool, error)
	AppendAudit(context.Context, *string, string, string, string) error
}

type Service struct {
	store     loginStore
	mfaKey    [32]byte
	dummyHash string
	now       func() time.Time
}

type LoginInput struct {
	Email     string
	Password  string
	TOTPCode  string
	UserAgent string
}

type LoginResult struct {
	SessionToken      string
	CSRFToken         string
	SessionID         string
	ActorID           string
	Role              string
	ExpiresAt         time.Time
	AbsoluteExpiresAt time.Time
}

func NewService(store loginStore, mfaKey []byte) (*Service, error) {
	if store == nil {
		return nil, errors.New("auth: nil login store")
	}
	if len(mfaKey) != 32 {
		return nil, errors.New("auth: MFA encryption key must be 32 bytes")
	}
	dummyHash, err := HashPassword("PVNaive-dummy-password-not-an-account")
	if err != nil {
		return nil, fmt.Errorf("auth: initialize dummy password path: %w", err)
	}
	var key [32]byte
	copy(key[:], mfaKey)
	return &Service{store: store, mfaKey: key, dummyHash: dummyHash, now: time.Now}, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	if s == nil || s.store == nil || in.Email == "" || in.Password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	actor, err := s.store.LookupActor(ctx, in.Email)
	if err != nil {
		// Always execute the expensive password path for an unknown identity.
		_, _ = VerifyPassword(in.Password, s.dummyHash)
		if errors.Is(err, sql.ErrNoRows) {
			_ = s.store.AppendAudit(ctx, nil, "auth.login", "denied", "invalid_credentials")
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("auth: login lookup: %w", err)
	}
	actorID := actor.ID
	now := s.now().UTC()
	if actor.Status != "active" || actor.PasswordHash == "" || (actor.LockedUntil != nil && actor.LockedUntil.After(now)) {
		_, _ = VerifyPassword(in.Password, s.dummyHash)
		_ = s.store.AppendAudit(ctx, &actorID, "auth.login", "denied", "account_unavailable")
		return LoginResult{}, ErrInvalidCredentials
	}

	passwordOK, err := VerifyPassword(in.Password, actor.PasswordHash)
	if err != nil || !passwordOK {
		_, _ = s.store.RecordLoginFailure(ctx, actor.ID)
		_ = s.store.AppendAudit(ctx, &actorID, "auth.login", "denied", "invalid_credentials")
		return LoginResult{}, ErrInvalidCredentials
	}

	if actor.MFARequired && actor.TOTPConfirmed {
		if in.TOTPCode == "" {
			return LoginResult{}, ErrMFARequired
		}
		factor, err := s.store.GetTOTPFactorPreAuth(ctx, actor.ID)
		if err != nil || factor.ConfirmedAt == nil {
			_ = s.store.AppendAudit(ctx, &actorID, "auth.login", "denied", "mfa_invalid")
			return LoginResult{}, ErrInvalidCredentials
		}
		secret, err := DecryptSecret(s.mfaKey[:], factor.Nonce, factor.Ciphertext)
		if err != nil {
			return LoginResult{}, fmt.Errorf("auth: decrypt TOTP factor: %w", err)
		}
		step, ok, err := ValidateTOTP(string(secret), in.TOTPCode, now, factor.LastUsedStep)
		for i := range secret {
			secret[i] = 0
		}
		if err != nil || !ok {
			_, _ = s.store.RecordLoginFailure(ctx, actor.ID)
			_ = s.store.AppendAudit(ctx, &actorID, "auth.login", "denied", "mfa_invalid")
			return LoginResult{}, ErrInvalidCredentials
		}
		consumed, err := s.store.ConsumeTOTPStepPreAuth(ctx, actor.ID, step)
		if err != nil {
			return LoginResult{}, fmt.Errorf("auth: consume TOTP step: %w", err)
		}
		if !consumed {
			_ = s.store.AppendAudit(ctx, &actorID, "auth.login", "denied", "mfa_replay")
			return LoginResult{}, ErrInvalidCredentials
		}
	}

	sessionRaw, sessionHash, err := NewOpaqueToken()
	if err != nil {
		return LoginResult{}, err
	}
	csrfRaw, csrfHash, err := NewOpaqueToken()
	if err != nil {
		return LoginResult{}, err
	}
	refreshFamilyID, err := newUUIDv4()
	if err != nil {
		return LoginResult{}, err
	}
	var uaHash []byte
	if in.UserAgent != "" {
		sum := sha256.Sum256([]byte(in.UserAgent))
		uaHash = sum[:]
	}
	expiresAt := now.Add(sessionIdleLifetime)
	absoluteExpiresAt := now.Add(sessionAbsoluteLifetime)
	sessionID, err := s.store.CreateSession(ctx, actor.ID, sessionHash[:], csrfHash[:], refreshFamilyID, uaHash, expiresAt, absoluteExpiresAt)
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.store.RecordLoginSuccess(ctx, actor.ID); err != nil {
		return LoginResult{}, err
	}
	_ = s.store.AppendAudit(ctx, &actorID, "auth.login", "success", "")
	return LoginResult{
		SessionToken: sessionRaw, CSRFToken: csrfRaw, SessionID: sessionID,
		ActorID: actor.ID, Role: actor.Role, ExpiresAt: expiresAt, AbsoluteExpiresAt: absoluteExpiresAt,
	}, nil
}

func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("auth: generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
