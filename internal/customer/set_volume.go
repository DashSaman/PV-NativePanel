package customer

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
)

type setVolumeStore interface {
	SetCurrentServiceQuotaTx(context.Context, *sql.Tx, string, *int64) (ServiceTerm, error)
}

func (s *Service) SetCustomerVolume(ctx context.Context, tx *sql.Tx, userID string, quotaGB *int64) (ServiceTerm, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(userID) == "" {
		return ServiceTerm{}, ErrInvalidAdjustment
	}
	var quotaBytes *int64
	if quotaGB != nil {
		if *quotaGB <= 0 || *quotaGB > math.MaxInt64/gibibyte {
			return ServiceTerm{}, ErrInvalidAdjustment
		}
		value := *quotaGB * gibibyte
		quotaBytes = &value
	}
	store, ok := s.store.(setVolumeStore)
	if !ok {
		return ServiceTerm{}, errors.New("customer: set-volume capability is unavailable")
	}
	return store.SetCurrentServiceQuotaTx(ctx, tx, userID, quotaBytes)
}
