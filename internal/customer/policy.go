package customer

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidUserAdminState = errors.New("customer: invalid user administrative state")
	ErrInvalidStartPolicy    = errors.New("customer: invalid service start policy")
	ErrInvalidDuration       = errors.New("customer: invalid service duration")
	ErrInvalidFixedStart     = errors.New("customer: fixed start must be in the future")
)

func ValidateUserAdminState(state UserAdminState) error {
	switch state {
	case UserDraft, UserActive, UserSuspended, UserRevoked:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidUserAdminState, state)
	}
}

func ValidateStartPolicy(policy StartPolicy) error {
	switch policy {
	case StartOnCreation, StartOnFirstSuccessfulConnection, StartAtFixedTimestamp:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidStartPolicy, policy)
	}
}

func DefaultUsageCapability() UsageCapability {
	return UsageCapability{Available: false, Reason: "exact_accounting_not_proven"}
}

func InitialTermTiming(policy StartPolicy, duration time.Duration, now time.Time, fixedStart *time.Time) (TermTiming, error) {
	if err := ValidateStartPolicy(policy); err != nil {
		return TermTiming{}, err
	}
	if duration <= 0 {
		return TermTiming{}, ErrInvalidDuration
	}
	now = now.UTC()

	switch policy {
	case StartOnFirstSuccessfulConnection:
		return TermTiming{State: TermPending}, nil
	case StartOnCreation:
		start := now
		expiry := start.Add(duration)
		return TermTiming{State: TermActive, StartsAt: &start, ExpiresAt: &expiry}, nil
	case StartAtFixedTimestamp:
		if fixedStart == nil {
			return TermTiming{}, ErrInvalidFixedStart
		}
		start := fixedStart.UTC()
		if !start.After(now) {
			return TermTiming{}, ErrInvalidFixedStart
		}
		expiry := start.Add(duration)
		return TermTiming{State: TermPending, StartsAt: &start, ExpiresAt: &expiry}, nil
	default:
		return TermTiming{}, ErrInvalidStartPolicy
	}
}

func EffectiveAccess(input EffectiveAccessInput) AccessState {
	switch input.AdminState {
	case UserDraft:
		return AccessDraft
	case UserSuspended:
		return AccessSuspended
	case UserRevoked:
		return AccessRevoked
	case UserActive:
		// Continue with commercial/service state.
	default:
		return AccessRevoked
	}

	switch input.TermState {
	case TermRevoked:
		return AccessRevoked
	case TermEnded:
		return AccessEnded
	case TermQuotaDepleted:
		return AccessQuotaDepleted
	case TermExpired:
		return AccessExpired
	}

	now := input.Now.UTC()
	if input.ExpiresAt != nil && !input.ExpiresAt.After(now) {
		return AccessExpired
	}
	if input.StartsAt == nil || input.StartsAt.After(now) {
		return AccessPending
	}
	return AccessActive
}
