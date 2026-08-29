package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
	"github.com/DashSaman/PV-NaivePanel/internal/subscription"
)

var (
	ErrRuntimeCredentialNotAdoptable = errors.New("customer: runtime credential is not available for adoption")
	ErrCustomerServiceNotFound       = errors.New("customer: managed service was not found")
)

type AdoptRuntimeInput struct {
	RuntimeCredentialID string
	QuotaGB             *int64
	Validity            ValidityInput
}

type UpdateServiceInput struct {
	QuotaGB  *int64
	Validity ValidityInput
}

type UpdateServiceTermRecord struct {
	QuotaBytes      *int64
	DurationSeconds int64
	StartPolicy     StartPolicy
	EffectiveAt     time.Time
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	State           TermState
}

type runtimeAdoptionStore interface {
	AdoptableRuntimeCredentialTx(context.Context, *sql.Tx, string) (runtimecred.CredentialView, error)
}

type serviceUpdateStore interface {
	UpdateCurrentServiceTermTx(context.Context, *sql.Tx, string, UpdateServiceTermRecord) (ServiceTerm, error)
}

func normalizeServiceSettings(quotaGB *int64, validity ValidityInput, now time.Time) (*int64, TermTiming, int64, StartPolicy, error) {
	quotaBytes, err := QuotaGBToBytes(quotaGB)
	if err != nil {
		return nil, TermTiming{}, 0, "", err
	}
	timing, duration, err := NormalizeValidity(validity, now)
	if err != nil {
		return nil, TermTiming{}, 0, "", err
	}
	durationSeconds := durationSecondsCeil(duration)
	if durationSeconds <= 0 {
		return nil, TermTiming{}, 0, "", ErrInvalidDuration
	}
	startPolicy, err := validityModeToStartPolicy(validity.Mode)
	if err != nil {
		return nil, TermTiming{}, 0, "", err
	}
	return quotaBytes, timing, durationSeconds, startPolicy, nil
}

// AdoptRuntimeCredential adds commercial/customer metadata around an existing
// active Runtime credential. It deliberately does not invoke createRuntime,
// rotate a password, rename the credential, or apply/reload Caddy.
func (s *Service) AdoptRuntimeCredential(ctx context.Context, tx *sql.Tx, actorID string, input AdoptRuntimeInput) (CreateCustomerResult, error) {
	if s == nil || s.store == nil {
		return CreateCustomerResult{}, errors.New("customer: service dependencies are required")
	}
	if strings.TrimSpace(actorID) == "" {
		return CreateCustomerResult{}, errors.New("customer: actor id is required")
	}
	runtimeID := strings.TrimSpace(input.RuntimeCredentialID)
	if runtimeID == "" {
		return CreateCustomerResult{}, ErrRuntimeCredentialNotAdoptable
	}
	adoptionStore, ok := s.store.(runtimeAdoptionStore)
	if !ok {
		return CreateCustomerResult{}, errors.New("customer: runtime adoption capability is unavailable")
	}

	now := s.now().UTC()
	quotaBytes, timing, durationSeconds, startPolicy, err := normalizeServiceSettings(input.QuotaGB, input.Validity, now)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	credential, err := adoptionStore.AdoptableRuntimeCredentialTx(ctx, tx, runtimeID)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	if credential.ID != runtimeID || credential.Status != runtimecred.CredentialActive {
		return CreateCustomerResult{}, ErrRuntimeCredentialNotAdoptable
	}
	if err := runtimecred.ValidateUsername(credential.Username); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: legacy runtime username: %w", err)
	}

	tenantID, err := s.store.DirectTenantID(ctx, tx)
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: resolve direct tenant: %w", err)
	}
	user, err := s.store.CreateUserTx(ctx, tx, CreateUserRecord{
		TenantID: tenantID, Username: credential.Username, DisplayName: credential.Username, ActorID: actorID,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create adopted user: %w", err)
	}
	term, err := s.store.CreateServiceTermTx(ctx, tx, CreateServiceTermRecord{
		TenantID: tenantID, UserID: user.ID, QuotaBytes: quotaBytes, DurationSeconds: durationSeconds,
		StartPolicy: startPolicy, PurchasedAt: now, StartsAt: timing.StartsAt, ExpiresAt: timing.ExpiresAt, State: timing.State,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create adopted service term: %w", err)
	}
	if err := s.store.BindRuntimeCredentialTx(ctx, tx, tenantID, user.ID, term.ID, credential.ID); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: bind existing runtime credential: %w", err)
	}

	rawToken, tokenHash, err := subscription.GenerateToken()
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: generate adopted subscription token: %w", err)
	}
	tokenPrefix := rawToken
	if len(tokenPrefix) > 10 {
		tokenPrefix = tokenPrefix[:10]
	}
	tokenCiphertext, tokenNonce, tokenEncryptionKeyID, err := s.encryptedSubscriptionToken(rawToken)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	if err := s.store.CreateSubscriptionTokenTx(ctx, tx, CreateSubscriptionTokenRecord{
		TenantID: tenantID, UserID: user.ID, ServiceTermID: term.ID, RuntimeCredentialID: credential.ID,
		TokenHash: append([]byte(nil), tokenHash[:]...), TokenPrefix: tokenPrefix,
		TokenCiphertext: tokenCiphertext, TokenNonce: tokenNonce, TokenEncryptionKeyID: tokenEncryptionKeyID,
		ExpiresAt: term.ExpiresAt,
	}); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: persist adopted subscription token: %w", err)
	}
	if err := s.persistSubscriptionRecovery(ctx, tx, tokenHash[:], tokenCiphertext, tokenNonce, tokenEncryptionKeyID); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: persist adopted subscription recovery: %w", err)
	}

	return CreateCustomerResult{
		User: user, ServiceTerm: term, RuntimeCredential: credential,
		SubscriptionToken: rawToken, UsageCapability: DefaultUsageCapability(),
	}, nil
}

// UpdateCustomerService edits only commercial service metadata. Runtime
// credential identity and secret are outside this operation by design.
func (s *Service) UpdateCustomerService(ctx context.Context, tx *sql.Tx, userID string, input UpdateServiceInput) (ServiceTerm, error) {
	if s == nil || s.store == nil {
		return ServiceTerm{}, errors.New("customer: service dependencies are required")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ServiceTerm{}, ErrCustomerServiceNotFound
	}
	updateStore, ok := s.store.(serviceUpdateStore)
	if !ok {
		return ServiceTerm{}, errors.New("customer: service update capability is unavailable")
	}

	now := s.now().UTC()
	quotaBytes, timing, durationSeconds, startPolicy, err := normalizeServiceSettings(input.QuotaGB, input.Validity, now)
	if err != nil {
		return ServiceTerm{}, err
	}
	term, err := updateStore.UpdateCurrentServiceTermTx(ctx, tx, userID, UpdateServiceTermRecord{
		QuotaBytes: quotaBytes, DurationSeconds: durationSeconds, StartPolicy: startPolicy,
		EffectiveAt: now, StartsAt: timing.StartsAt, ExpiresAt: timing.ExpiresAt, State: timing.State,
	})
	if err != nil {
		return ServiceTerm{}, err
	}
	return term, nil
}
