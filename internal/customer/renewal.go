package customer

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type RenewalMode string

const (
	RenewalCurrent   RenewalMode = "renew_current"
	RenewalUsingPlan RenewalMode = "renew_plan"
	RenewalCustom    RenewalMode = "custom"
	RenewalNextPlan  RenewalMode = "next_plan"
)

var (
	ErrInvalidRenewal   = errors.New("customer: invalid renewal request")
	ErrNextPlanNotReady = errors.New("customer: next plan is scheduled but current service is still active")
	ErrNextPlanMissing  = errors.New("customer: no next plan is scheduled")
)

type RenewalInput struct {
	Mode           RenewalMode   `json:"mode"`
	PlanID         string        `json:"plan_id,omitempty"`
	QuotaGB        *int64        `json:"quota_gb,omitempty"`
	UnlimitedQuota bool          `json:"unlimited_quota,omitempty"`
	NoExpiry       bool          `json:"no_expiry,omitempty"`
	Validity       ValidityInput `json:"validity,omitempty"`
}

type RenewalContext struct {
	TenantID            string
	UserID              string
	RuntimeCredentialID string
	Current             ServiceTerm
}

type ScheduledNextPlan struct {
	PlanID       string     `json:"plan_id"`
	PlanName     string     `json:"plan_name,omitempty"`
	SourceTermID string     `json:"source_term_id"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
}

type CreateRenewalTermRecord struct {
	TenantID           string
	UserID             string
	PlanID             string
	QuotaBytes         *int64
	DurationSeconds    int64
	NoExpiry           bool
	StartPolicy        StartPolicy
	PurchasedAt        time.Time
	StartsAt           *time.Time
	ExpiresAt          *time.Time
	State              TermState
	RenewalKind        string
	RenewedFromTermID  string
	AccountingBaseline AccountingBaseline
}

type RenewalResult struct {
	PreviousTermID       string      `json:"previous_term_id"`
	ServiceTerm          ServiceTerm `json:"service_term"`
	RuntimeCredentialID  string      `json:"runtime_credential_id"`
	AppliedPlanID        string      `json:"applied_plan_id,omitempty"`
	NextPlanConsumed     bool        `json:"next_plan_consumed"`
	CredentialRotated    bool        `json:"credential_rotated"`
	SubscriptionReissued bool        `json:"subscription_reissued"`
}

type renewalStore interface {
	CurrentRenewalContextTx(context.Context, *sql.Tx, string) (RenewalContext, error)
	PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error)
	CreateRenewalTermTx(context.Context, *sql.Tx, CreateRenewalTermRecord) (ServiceTerm, error)
	RebindRuntimeCredentialTx(context.Context, *sql.Tx, string, string, string, string, string) error
	ProjectSubscriptionToTermTx(context.Context, *sql.Tx, string, ServiceTerm) error
	RecordRenewalProfileTx(context.Context, *sql.Tx, string, string, time.Time, bool) error
	ScheduledNextPlanTx(context.Context, *sql.Tx, string) (*ScheduledNextPlan, error)
}

type nextPlanStore interface {
	CurrentRenewalContextTx(context.Context, *sql.Tx, string) (RenewalContext, error)
	PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error)
	ScheduleNextPlanTx(context.Context, *sql.Tx, string, string, string, string, time.Time) (ScheduledNextPlan, error)
	ClearNextPlanTx(context.Context, *sql.Tx, string, string) error
}

func (s *Service) RenewCustomer(ctx context.Context, tx *sql.Tx, actorID, userID string, input RenewalInput) (RenewalResult, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(actorID) == "" || strings.TrimSpace(userID) == "" {
		return RenewalResult{}, ErrInvalidRenewal
	}
	store, ok := s.store.(renewalStore)
	if !ok {
		return RenewalResult{}, errors.New("customer: renewal capability is unavailable")
	}
	current, err := store.CurrentRenewalContextTx(ctx, tx, userID)
	if err != nil {
		return RenewalResult{}, err
	}
	if current.Current.State == TermRevoked || (current.Current.State == TermEnded && current.RuntimeCredentialID == "") {
		return RenewalResult{}, ErrInvalidRenewal
	}
	now := s.now().UTC()
	record, consumed, err := s.resolveRenewalRecord(ctx, tx, store, current, input, now)
	if err != nil {
		return RenewalResult{}, err
	}
	term, err := store.CreateRenewalTermTx(ctx, tx, record)
	if err != nil {
		return RenewalResult{}, err
	}
	if err := store.RebindRuntimeCredentialTx(ctx, tx, current.TenantID, current.UserID, current.Current.ID, term.ID, current.RuntimeCredentialID); err != nil {
		return RenewalResult{}, err
	}
	if err := store.ProjectSubscriptionToTermTx(ctx, tx, current.UserID, term); err != nil {
		return RenewalResult{}, err
	}
	if err := store.RecordRenewalProfileTx(ctx, tx, current.TenantID, current.UserID, now, consumed); err != nil {
		return RenewalResult{}, err
	}
	return RenewalResult{
		PreviousTermID:       current.Current.ID,
		ServiceTerm:          term,
		RuntimeCredentialID:  current.RuntimeCredentialID,
		AppliedPlanID:        record.PlanID,
		NextPlanConsumed:     consumed,
		CredentialRotated:    false,
		SubscriptionReissued: false,
	}, nil
}

func (s *Service) resolveRenewalRecord(ctx context.Context, tx *sql.Tx, store renewalStore, current RenewalContext, input RenewalInput, now time.Time) (CreateRenewalTermRecord, bool, error) {
	base := CreateRenewalTermRecord{
		TenantID:          current.TenantID,
		UserID:            current.UserID,
		PurchasedAt:       now,
		RenewedFromTermID: current.Current.ID,
	}
	switch input.Mode {
	case RenewalCurrent:
		base.PlanID = current.Current.PlanID
		base.QuotaBytes = cloneInt64(current.Current.QuotaBytes)
		base.DurationSeconds = current.Current.DurationSeconds
		base.NoExpiry = current.Current.NoExpiry
		base.StartPolicy = current.Current.StartPolicy
		base.RenewalKind = string(RenewalCurrent)
		applyRenewalTiming(&base, now)
		return base, false, nil
	case RenewalUsingPlan:
		if strings.TrimSpace(input.PlanID) == "" {
			return CreateRenewalTermRecord{}, false, ErrInvalidRenewal
		}
		plan, err := store.PlanByIDTx(ctx, tx, input.PlanID)
		if err != nil {
			return CreateRenewalTermRecord{}, false, err
		}
		if !plan.Enabled {
			return CreateRenewalTermRecord{}, false, ErrInvalidRenewal
		}
		applyPlanToRenewal(&base, plan)
		base.RenewalKind = string(RenewalUsingPlan)
		applyRenewalTiming(&base, now)
		return base, false, nil
	case RenewalNextPlan:
		next, err := store.ScheduledNextPlanTx(ctx, tx, current.UserID)
		if err != nil {
			return CreateRenewalTermRecord{}, false, err
		}
		if next == nil || next.PlanID == "" {
			return CreateRenewalTermRecord{}, false, ErrNextPlanMissing
		}
		if !nextPlanEligible(current.Current, now) {
			return CreateRenewalTermRecord{}, false, ErrNextPlanNotReady
		}
		plan, err := store.PlanByIDTx(ctx, tx, next.PlanID)
		if err != nil {
			return CreateRenewalTermRecord{}, false, err
		}
		if !plan.Enabled {
			return CreateRenewalTermRecord{}, false, ErrInvalidRenewal
		}
		applyPlanToRenewal(&base, plan)
		base.RenewalKind = string(RenewalNextPlan)
		applyRenewalTiming(&base, now)
		return base, true, nil
	case RenewalCustom:
		quota, err := customRenewalQuota(input)
		if err != nil {
			return CreateRenewalTermRecord{}, false, err
		}
		base.QuotaBytes = quota
		base.RenewalKind = string(RenewalCustom)
		if input.NoExpiry {
			base.NoExpiry = true
			base.StartPolicy = StartOnCreation
			base.DurationSeconds = 86400
			base.State = TermActive
			return base, false, nil
		}
		timing, duration, err := NormalizeValidity(input.Validity, now)
		if err != nil {
			return CreateRenewalTermRecord{}, false, err
		}
		startPolicy, err := validityModeToStartPolicy(input.Validity.Mode)
		if err != nil {
			return CreateRenewalTermRecord{}, false, err
		}
		base.DurationSeconds = int64(duration / time.Second)
		base.StartPolicy = startPolicy
		base.StartsAt = timing.StartsAt
		base.ExpiresAt = timing.ExpiresAt
		base.State = timing.State
		return base, false, nil
	default:
		return CreateRenewalTermRecord{}, false, ErrInvalidRenewal
	}
}

func customRenewalQuota(input RenewalInput) (*int64, error) {
	if input.UnlimitedQuota {
		if input.QuotaGB != nil {
			return nil, ErrInvalidRenewal
		}
		return nil, nil
	}
	if input.QuotaGB == nil {
		return nil, ErrInvalidRenewal
	}
	return QuotaGBToBytes(input.QuotaGB)
}

func applyPlanToRenewal(record *CreateRenewalTermRecord, plan PlanPreset) {
	record.PlanID = plan.ID
	record.QuotaBytes = cloneInt64(plan.QuotaBytes)
	record.NoExpiry = plan.NoExpiry
	record.StartPolicy = plan.StartPolicy
	if plan.NoExpiry {
		record.DurationSeconds = 86400
	} else {
		record.DurationSeconds = plan.ValiditySeconds
	}
}

func applyRenewalTiming(record *CreateRenewalTermRecord, now time.Time) {
	if record.NoExpiry {
		record.State = TermActive
		record.StartsAt = nil
		record.ExpiresAt = nil
		return
	}
	if record.StartPolicy == StartOnFirstSuccessfulConnection {
		record.State = TermPending
		record.StartsAt = nil
		record.ExpiresAt = nil
		return
	}
	start := now.UTC()
	expires := start.Add(time.Duration(record.DurationSeconds) * time.Second)
	record.State = TermActive
	record.StartsAt = &start
	record.ExpiresAt = &expires
}

func nextPlanEligible(term ServiceTerm, now time.Time) bool {
	if term.State == TermExpired || term.State == TermQuotaDepleted || term.State == TermEnded {
		return true
	}
	return term.ExpiresAt != nil && !term.ExpiresAt.After(now)
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func (s *Service) ScheduleNextPlan(ctx context.Context, tx *sql.Tx, actorID, userID, planID string) (ScheduledNextPlan, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(actorID) == "" || strings.TrimSpace(userID) == "" || strings.TrimSpace(planID) == "" {
		return ScheduledNextPlan{}, ErrInvalidRenewal
	}
	store, ok := s.store.(nextPlanStore)
	if !ok {
		return ScheduledNextPlan{}, errors.New("customer: next-plan capability is unavailable")
	}
	current, err := store.CurrentRenewalContextTx(ctx, tx, userID)
	if err != nil {
		return ScheduledNextPlan{}, err
	}
	plan, err := store.PlanByIDTx(ctx, tx, planID)
	if err != nil {
		return ScheduledNextPlan{}, err
	}
	if !plan.Enabled {
		return ScheduledNextPlan{}, ErrInvalidRenewal
	}
	return store.ScheduleNextPlanTx(ctx, tx, current.TenantID, current.UserID, current.Current.ID, plan.ID, s.now().UTC())
}

func (s *Service) ClearNextPlan(ctx context.Context, tx *sql.Tx, actorID, userID string) error {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(actorID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidRenewal
	}
	store, ok := s.store.(nextPlanStore)
	if !ok {
		return errors.New("customer: next-plan capability is unavailable")
	}
	return store.ClearNextPlanTx(ctx, tx, actorID, userID)
}
