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

type ProductCreateCustomerInput struct {
	Username         string        `json:"username"`
	Password         string        `json:"password,omitempty"`
	GeneratePassword bool          `json:"generate_password"`
	PlanID           string        `json:"plan_id,omitempty"`
	QuotaGB          *int64        `json:"quota_gb,omitempty"`
	UnlimitedQuota   bool          `json:"unlimited_quota,omitempty"`
	NoExpiry         bool          `json:"no_expiry,omitempty"`
	Validity         ValidityInput `json:"validity,omitempty"`
	GroupID          string        `json:"group_id,omitempty"`
	TagIDs           []string      `json:"tag_ids,omitempty"`
	OnHold           bool          `json:"on_hold,omitempty"`
}

func (s *Service) CreateProductCustomer(
	ctx context.Context,
	tx *sql.Tx,
	actorID, idempotencyKey string,
	input ProductCreateCustomerInput,
) (CreateCustomerResult, error) {
	if s == nil || s.store == nil || s.createRuntime == nil || tx == nil {
		return CreateCustomerResult{}, errors.New("customer: product create dependencies are required")
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
	store, ok := s.store.(productCreateStore)
	if !ok {
		return CreateCustomerResult{}, errors.New("customer: product create capability is unavailable")
	}
	tenantID, err := store.OperationTenantIDTx(ctx, tx)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	now := s.now().UTC()

	termRecord, defaultGroupID, defaultTagIDs, err := s.productCreateTerm(ctx, tx, store, tenantID, username, input, now)
	if err != nil {
		return CreateCustomerResult{}, err
	}
	user, err := s.store.CreateUserTx(ctx, tx, CreateUserRecord{
		TenantID:    tenantID,
		Username:    username,
		DisplayName: username,
		ActorID:     actorID,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create product user: %w", err)
	}
	termRecord.UserID = user.ID
	term, err := store.CreateProductServiceTermTx(ctx, tx, termRecord)
	if err != nil {
		return CreateCustomerResult{}, err
	}

	mutation, err := s.createRuntime(ctx, tx, actorID, runtimeIdempotencyKey(idempotencyKey), runtimecred.CreateInput{
		Username:         username,
		Password:         input.Password,
		GeneratePassword: input.GeneratePassword,
	})
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: create product runtime credential: %w", err)
	}
	credential := mutation.Credential()
	if credential.ID == "" {
		_ = mutation.AbortAndRollback(ctx, tx)
		return CreateCustomerResult{}, errors.New("customer: runtime mutation returned empty credential id")
	}
	if err := s.store.BindRuntimeCredentialTx(ctx, tx, tenantID, user.ID, term.ID, credential.ID); err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "bind product runtime credential", err)
	}

	rawToken, tokenHash, err := subscription.GenerateToken()
	if err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "generate product subscription token", err)
	}
	tokenPrefix := rawToken
	if len(tokenPrefix) > 10 {
		tokenPrefix = tokenPrefix[:10]
	}
	ciphertext, nonce, keyID, err := s.encryptedSubscriptionToken(rawToken)
	if err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "encrypt product subscription token", err)
	}
	if err := s.store.CreateSubscriptionTokenTx(ctx, tx, CreateSubscriptionTokenRecord{
		TenantID:             tenantID,
		UserID:               user.ID,
		ServiceTermID:        term.ID,
		RuntimeCredentialID:  credential.ID,
		TokenHash:            append([]byte(nil), tokenHash[:]...),
		TokenPrefix:          tokenPrefix,
		TokenCiphertext:      ciphertext,
		TokenNonce:           nonce,
		TokenEncryptionKeyID: keyID,
		ExpiresAt:            term.ExpiresAt,
	}); err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "persist product subscription token", err)
	}
	if err := s.persistSubscriptionRecovery(ctx, tx, tokenHash[:], ciphertext, nonce, keyID); err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "persist product subscription recovery", err)
	}

	groupID := strings.TrimSpace(input.GroupID)
	if groupID == "" {
		groupID = defaultGroupID
	}
	tagIDs := append([]string(nil), defaultTagIDs...)
	tagIDs = append(tagIDs, input.TagIDs...)
	if err := store.ApplyCustomerMetadataTx(ctx, tx, tenantID, user.ID, actorID, groupID, tagIDs, input.OnHold); err != nil {
		return CreateCustomerResult{}, abortRuntimeMutation(ctx, tx, mutation, "apply product customer metadata", err)
	}
	if err := mutation.CommitAndFinalize(ctx, tx); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("customer: finalize product runtime mutation: %w", err)
	}
	return CreateCustomerResult{
		User:              user,
		ServiceTerm:       term,
		RuntimeCredential: credential,
		GeneratedPassword: mutation.TakeGeneratedPassword(),
		SubscriptionToken: rawToken,
		UsageCapability:   DefaultUsageCapability(),
	}, nil
}

func (s *Service) productCreateTerm(
	ctx context.Context,
	tx *sql.Tx,
	store productCreateStore,
	tenantID, username string,
	input ProductCreateCustomerInput,
	now time.Time,
) (CreateServiceTermRecord, string, []string, error) {
	record := CreateServiceTermRecord{
		TenantID:    tenantID,
		PurchasedAt: now,
		RenewalKind: "initial",
	}
	if strings.TrimSpace(input.PlanID) != "" {
		plan, err := store.PlanByIDTx(ctx, tx, input.PlanID)
		if err != nil {
			return CreateServiceTermRecord{}, "", nil, err
		}
		if !plan.Enabled {
			return CreateServiceTermRecord{}, "", nil, ErrInvalidRenewal
		}
		record.PlanID = plan.ID
		record.QuotaBytes = cloneInt64(plan.QuotaBytes)
		record.NoExpiry = plan.NoExpiry
		record.StartPolicy = plan.StartPolicy
		if plan.NoExpiry {
			record.DurationSeconds = 86400
		} else {
			record.DurationSeconds = plan.ValiditySeconds
		}
		applyRenewalTiming((*CreateRenewalTermRecord)(&record), now)
		return record, plan.DefaultGroupID, append([]string(nil), plan.TagIDs...), nil
	}

	quota, err := customRenewalQuota(RenewalInput{
		QuotaGB:        input.QuotaGB,
		UnlimitedQuota: input.UnlimitedQuota,
	})
	if err != nil {
		return CreateServiceTermRecord{}, "", nil, err
	}
	record.QuotaBytes = quota
	if input.NoExpiry {
		record.NoExpiry = true
		record.DurationSeconds = 86400
		record.StartPolicy = StartOnCreation
		record.State = TermActive
		return record, "", nil, nil
	}
	timing, duration, err := NormalizeValidity(input.Validity, now)
	if err != nil {
		return CreateServiceTermRecord{}, "", nil, err
	}
	startPolicy, err := validityModeToStartPolicy(input.Validity.Mode)
	if err != nil {
		return CreateServiceTermRecord{}, "", nil, err
	}
	record.DurationSeconds = durationSecondsCeil(duration)
	record.StartPolicy = startPolicy
	record.StartsAt = timing.StartsAt
	record.ExpiresAt = timing.ExpiresAt
	record.State = timing.State
	_ = username
	return record, "", nil, nil
}
