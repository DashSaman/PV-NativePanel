package customer

import (
	"errors"
	"math"
	"time"
)

const BytesPerCustomerGB int64 = 1073741824

type ValidityMode string

const (
	ValidityOnCreation                  ValidityMode = "on_creation"
	ValidityOnFirstSuccessfulConnection ValidityMode = "on_first_successful_connection"
	ValidityFixedExpiry                 ValidityMode = "fixed_expiry"
)

type ValidityInput struct {
	Mode         ValidityMode
	DurationDays int
	ExpiresAt    *time.Time
}

var (
	ErrInvalidQuotaGB      = errors.New("customer: quota GB must be a positive integer or unlimited")
	ErrQuotaOverflow       = errors.New("customer: quota GB overflows byte storage")
	ErrInvalidValidityMode = errors.New("customer: invalid validity mode")
	ErrStaleValidityField  = errors.New("customer: validity contains fields from another mode")
	ErrInvalidFixedExpiry  = errors.New("customer: fixed expiry must be in the future")
)

func QuotaGBToBytes(gb *int64) (*int64, error) {
	if gb == nil {
		return nil, nil
	}
	if *gb <= 0 {
		return nil, ErrInvalidQuotaGB
	}
	if *gb > math.MaxInt64/BytesPerCustomerGB {
		return nil, ErrQuotaOverflow
	}
	bytes := *gb * BytesPerCustomerGB
	return &bytes, nil
}

func NormalizeValidity(input ValidityInput, now time.Time) (TermTiming, time.Duration, error) {
	now = now.UTC()

	switch input.Mode {
	case ValidityOnCreation:
		if input.ExpiresAt != nil {
			return TermTiming{}, 0, ErrStaleValidityField
		}
		duration, err := durationFromDays(input.DurationDays)
		if err != nil {
			return TermTiming{}, 0, err
		}
		expires := now.Add(duration)
		start := now
		return TermTiming{State: TermActive, StartsAt: &start, ExpiresAt: &expires}, duration, nil

	case ValidityOnFirstSuccessfulConnection:
		if input.ExpiresAt != nil {
			return TermTiming{}, 0, ErrStaleValidityField
		}
		duration, err := durationFromDays(input.DurationDays)
		if err != nil {
			return TermTiming{}, 0, err
		}
		return TermTiming{State: TermPending}, duration, nil

	case ValidityFixedExpiry:
		if input.DurationDays != 0 {
			return TermTiming{}, 0, ErrStaleValidityField
		}
		if input.ExpiresAt == nil {
			return TermTiming{}, 0, ErrInvalidFixedExpiry
		}
		expires := input.ExpiresAt.UTC()
		if !expires.After(now) {
			return TermTiming{}, 0, ErrInvalidFixedExpiry
		}
		duration := expires.Sub(now)
		if duration <= 0 {
			return TermTiming{}, 0, ErrInvalidFixedExpiry
		}
		start := now
		return TermTiming{State: TermActive, StartsAt: &start, ExpiresAt: &expires}, duration, nil

	default:
		return TermTiming{}, 0, ErrInvalidValidityMode
	}
}

func ActivateOnFirstUse(current TermTiming, duration time.Duration, observedAt time.Time) (TermTiming, bool, error) {
	if duration <= 0 {
		return current, false, ErrInvalidDuration
	}
	if current.StartsAt != nil || current.ExpiresAt != nil || current.State == TermActive {
		return current, false, nil
	}
	if current.State != TermPending {
		return current, false, ErrInvalidStartPolicy
	}
	start := observedAt.UTC()
	expires := start.Add(duration)
	return TermTiming{State: TermActive, StartsAt: &start, ExpiresAt: &expires}, true, nil
}

func durationFromDays(days int) (time.Duration, error) {
	if days <= 0 {
		return 0, ErrInvalidDuration
	}
	const day = 24 * time.Hour
	if int64(days) > int64(math.MaxInt64)/int64(day) {
		return 0, ErrInvalidDuration
	}
	return time.Duration(days) * day, nil
}
