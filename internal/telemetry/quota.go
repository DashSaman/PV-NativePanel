package telemetry

import (
	"errors"
	"sync/atomic"
)

var ErrInvalidQuotaClaim = errors.New("telemetry: invalid quota claim")

type SharedQuotaBudget struct {
	remaining atomic.Int64
}

func NewSharedQuotaBudget(quotaBytes int64) (*SharedQuotaBudget, error) {
	if quotaBytes < 0 {
		return nil, ErrInvalidQuotaClaim
	}
	budget := &SharedQuotaBudget{}
	budget.remaining.Store(quotaBytes)
	return budget, nil
}

func (b *SharedQuotaBudget) Claim(requested int64) (int64, error) {
	if b == nil || requested <= 0 {
		return 0, ErrInvalidQuotaClaim
	}
	for {
		remaining := b.remaining.Load()
		if remaining <= 0 {
			return 0, nil
		}
		granted := requested
		if granted > remaining {
			granted = remaining
		}
		if b.remaining.CompareAndSwap(remaining, remaining-granted) {
			return granted, nil
		}
	}
}

func (b *SharedQuotaBudget) Remaining() int64 {
	if b == nil {
		return 0
	}
	return b.remaining.Load()
}
