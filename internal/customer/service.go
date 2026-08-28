package customer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

type Store interface {
	DirectTenantID(context.Context, *sql.Tx) (string, error)
	CreateUserTx(context.Context, *sql.Tx, CreateUserRecord) (User, error)
	CreateServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error)
	BindRuntimeCredentialTx(context.Context, *sql.Tx, string, string, string, string) error
	CreateSubscriptionTokenTx(context.Context, *sql.Tx, CreateSubscriptionTokenRecord) error
}

type RuntimeMutation interface {
	Credential() runtimecred.CredentialView
	RuntimeRevisionID() string
	CommitAndFinalize(context.Context, *sql.Tx) error
	AbortAndRollback(context.Context, *sql.Tx) error
	TakeGeneratedPassword() string
}

type RuntimeCreateFunc func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error)

type Service struct {
	store         Store
	createRuntime RuntimeCreateFunc
	now           func() time.Time
}

type CreateCustomerInput struct {
	Username         string
	Password         string
	GeneratePassword bool
	QuotaGB          *int64
	Validity         ValidityInput
}

type CreateCustomerResult struct {
	User              User                       `json:"user"`
	ServiceTerm       ServiceTerm                `json:"service_term"`
	RuntimeCredential runtimecred.CredentialView `json:"runtime_credential"`
	GeneratedPassword string                     `json:"generated_password,omitempty"`
	SubscriptionToken string                     `json:"subscription_token,omitempty"`
	UsageCapability   UsageCapability            `json:"usage_capability"`
}

func NewService(store Store, createRuntime RuntimeCreateFunc, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, createRuntime: createRuntime, now: now}
}

func (s *Service) CreateCustomer(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input CreateCustomerInput) (CreateCustomerResult, error) {
	if s == nil || s.store == nil || s.createRuntime == nil {
		return CreateCustomerResult{}, errors.New("customer: service dependencies are required")
	}
	if strings.TrimSpace(actorID) == "" {
		return CreateCustomerResult{}, errors.New("customer: actor id is required")
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 160 {
		return CreateCustomerResult{}, errors.New("customer: idempotency key length must be 8-160 bytes")
	}
	username := strings.TrimSpace(input.Username)
	if err := runtimecred.ValidateUsername(username); err != nil {
		return CreateCustomerResult{}, err
	}
	quotaBytes, err := QuotaGBToBytes(input.QuotaGB)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	now := s.now().UTC()
	timing, duration, err := NormalizeValidity(input.Validity, now)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	durationSeconds := durationSecondsCeil(duration)
	if durationSeconds <= 0 {
		return CreateCustomerResult{}, ErrInvalidDuration
	}
	startPolicy, err := validityModeToStartPolicy(input.Validity.Mode)
	if err != nil {
		return CreateCustomerResult{}, err
	}

	tenantID, err := s.store.DirectTenantID(ctx, tx)
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: resolve direct tenant: %w", err)
	}
	user, err := s.store.CreateUserTx(ctx, tx, CreateUserRecord{
		TenantID: tenantID, Username: username, DisplayName: username, ActorID: actorID,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create user: %w", err)
	}
	term, err := s.store.CreateServiceTermTx(ctx, tx, CreateServiceTermRecord{
		TenantID: tenantID, UserID: user.ID, QuotaBytes: quotaBytes, DurationSeconds: durationSeconds,
		StartPolicy: startPolicy, PurchasedAt: now, StartsAt: timing.StartsAt, ExpiresAt: timing.ExpiresAt, State: timing.State,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create service term: %w", err)
	}

	mutation, err := s.createRuntime(ctx, tx, actorID, runtimeIdempotencyKey(idempotencyKey), runtimecred.CreateInput{
		Username: username, Password: input.Password, GeneratePassword: input.GeneratePassword,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create runtime credential: %w", err)
	}
	credential := mutation.Credential()
	if credential.ID == "" {
		_ = mutation.AbortAndRollback(ctx, tx)
		return CreateCustomerResult{}, errors.New("customer: runtime mutation returned empty credential id")
	}
	if err := s.store.BindRuntimeCredentialTx(ctx, tx, tenantID, user.ID, term.ID, credential.ID); err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "bind runtime credential", err)
	}

	rawSubscriptionToken, tokenHash, err := subscription.GenerateToken()
	if err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "generate subscription token", err)
	}
	tokenPrefix := rawSubscriptionToken
	if len(tokenPrefix) > 10 {
		tokenPrefix = tokenPrefix[:10]
	}
	if err := s.store.CreateSubscriptionTokenTx(ctx, tx, CreateSubscriptionTokenRecord{
		TenantID:            tenantID,
		UserID:              user.ID,
		ServiceTermID:       term.ID,
		RuntimeCredentialID: credential.ID,
		TokenHash:           append([]byte(nil), tokenHash[:]...),
		TokenPrefix:         tokenPrefix,
		ExpiresAt:           term.ExpiresAt,
	}); err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "persist subscription token", err)
	}

	if err := mutation.CommitAndFinalize(ctx, tx); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: finalize runtime mutation: %w", err)
	}

	return CreateCustomerResult{
		User: user, ServiceTerm: term, RuntimeCredential: credential,
		GeneratedPassword: mutation.TakeGeneratedPassword(), SubscriptionToken: rawSubscriptionToken,
		UsageCapability: DefaultUsageCapability(),
	}, nil
}

func abortRuntimeMutation(ctx context.Context, tx *sql.Tx, mutation RuntimeMutation, operation string, cause error) error {
	abortErr := mutation.AbortAndRollback(ctx, tx)
	if abortErr != nil {
		return fmt.Errorf("customer: %s: %v; abort: %w", operation, cause, abortErr)
	}
	return fmt.Errorf("customer: %s: %w", operation, cause)
}

func validityModeToStartPolicy(mode ValidityMode) (StartPolicy, error) {
	switch mode {
	case ValidityOnCreation:
		return StartOnCreation, nil
	case ValidityOnFirstSuccessfulConnection:
		return StartOnFirstConnection, nil
	case ValidityFixedExpiry:
		return StartFixedTimestamp, nil
	default:
		return "", ErrInvalidValidityMode
	}
}

func durationSecondsCeil(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	seconds := int64(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	return seconds
}

func runtimeIdempotencyKey(key string) string {
	const suffix = ":runtime"
	if len(key)+len(suffix) <= 160 {
		return key + suffix
	}
	sum := sha256.Sum256([]byte("customer-runtime:" + key))
	return "customer-runtime-" + hex.EncodeToString(sum[:])
}
