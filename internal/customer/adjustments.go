package customer

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrInvalidAdjustment      = errors.New("customer: adjustment must be positive")
	ErrUnlimitedQuotaAddition = errors.New("customer: cannot add volume to an unlimited service; set a total quota first")
)

type serviceAdjustmentStore interface {
	AddCurrentServiceQuotaTx(context.Context, *sql.Tx, string, int64) (ServiceTerm, error)
	ExtendCurrentServiceTx(context.Context, *sql.Tx, string, int64, time.Time) (ServiceTerm, error)
}

func (s *Service) AddCustomerVolume(ctx context.Context, tx *sql.Tx, userID string, deltaGB int64) (ServiceTerm, error) {
	if s == nil || s.store == nil || strings.TrimSpace(userID) == "" || deltaGB <= 0 {
		return ServiceTerm{}, ErrInvalidAdjustment
	}
	if deltaGB > math.MaxInt64/1073741824 {
		return ServiceTerm{}, ErrQuotaOverflow
	}
	store, ok := s.store.(serviceAdjustmentStore)
	if !ok {
		return ServiceTerm{}, ErrCustomerServiceNotFound
	}
	return store.AddCurrentServiceQuotaTx(ctx, tx, userID, deltaGB*1073741824)
}

func (s *Service) ExtendCustomerTime(ctx context.Context, tx *sql.Tx, userID string, days int64) (ServiceTerm, error) {
	if s == nil || s.store == nil || strings.TrimSpace(userID) == "" || days <= 0 {
		return ServiceTerm{}, ErrInvalidAdjustment
	}
	if days > math.MaxInt64/86400 {
		return ServiceTerm{}, ErrInvalidAdjustment
	}
	store, ok := s.store.(serviceAdjustmentStore)
	if !ok {
		return ServiceTerm{}, ErrCustomerServiceNotFound
	}
	return store.ExtendCurrentServiceTx(ctx, tx, userID, days*86400, s.now().UTC())
}
