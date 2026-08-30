package telemetry

import (
	"errors"
	"math"
	"time"
)

var (
	ErrInvalidProjection   = errors.New("telemetry: invalid projection input")
	ErrServiceTermMismatch = errors.New("telemetry: service term mismatch")
)

type QuotaState string

const (
	QuotaUnlimited QuotaState = "unlimited"
	QuotaActive    QuotaState = "active"
	QuotaDepleted  QuotaState = "depleted"
)

type TermPolicy struct {
	ServiceTermID string
	QuotaBytes    *int64
}

type TermUsage struct {
	ServiceTermID    string
	UploadBytes      int64
	DownloadBytes    int64
	FirstConnectedAt *time.Time
}

type SessionSnapshot struct {
	Key            SessionKey
	LastObservedAt time.Time
	Final          bool
}

type ReadModel struct {
	ServiceTermID      string     `json:"service_term_id"`
	UploadBytes        int64      `json:"upload_bytes"`
	DownloadBytes      int64      `json:"download_bytes"`
	UsedBytes          int64      `json:"used_bytes"`
	QuotaBytes         *int64     `json:"quota_bytes"`
	RemainingBytes     *int64     `json:"remaining_bytes"`
	QuotaState         QuotaState `json:"quota_state"`
	FirstConnectedAt   *time.Time `json:"first_connected_at"`
	LastOnline         *time.Time `json:"last_online"`
	LastResetAt        *time.Time `json:"last_reset_at,omitempty"`
	Online             bool       `json:"online"`
	SessionCount       int        `json:"session_count"`
	AccountingComplete bool       `json:"accounting_complete"`
}

func BuildReadModel(policy TermPolicy, usage TermUsage, sessions []SessionSnapshot, now time.Time, staleAfter time.Duration, telemetryHealthy bool) (ReadModel, error) {
	if !validUUID(policy.ServiceTermID) || !validUUID(usage.ServiceTermID) ||
		now.IsZero() || staleAfter <= 0 || usage.UploadBytes < 0 || usage.DownloadBytes < 0 {
		return ReadModel{}, ErrInvalidProjection
	}
	if policy.ServiceTermID != usage.ServiceTermID {
		return ReadModel{}, ErrServiceTermMismatch
	}
	if policy.QuotaBytes != nil && *policy.QuotaBytes <= 0 {
		return ReadModel{}, ErrInvalidProjection
	}
	if usage.UploadBytes > math.MaxInt64-usage.DownloadBytes {
		return ReadModel{}, ErrInvalidProjection
	}

	used := usage.UploadBytes + usage.DownloadBytes
	model := ReadModel{
		ServiceTermID:      policy.ServiceTermID,
		UploadBytes:        usage.UploadBytes,
		DownloadBytes:      usage.DownloadBytes,
		UsedBytes:          used,
		QuotaBytes:         cloneInt64(policy.QuotaBytes),
		FirstConnectedAt:   cloneTime(usage.FirstConnectedAt),
		AccountingComplete: telemetryHealthy,
	}

	if policy.QuotaBytes == nil {
		model.QuotaState = QuotaUnlimited
	} else {
		remaining := *policy.QuotaBytes - used
		if remaining <= 0 {
			remaining = 0
			model.QuotaState = QuotaDepleted
		} else {
			model.QuotaState = QuotaActive
		}
		model.RemainingBytes = &remaining
	}

	for _, session := range sessions {
		if !validSessionKey(session.Key) || session.LastObservedAt.IsZero() || session.LastObservedAt.After(now) {
			return ReadModel{}, ErrInvalidProjection
		}
		if model.LastOnline == nil || session.LastObservedAt.After(*model.LastOnline) {
			last := session.LastObservedAt
			model.LastOnline = &last
		}
		if session.Final {
			continue
		}
		if now.Sub(session.LastObservedAt) > staleAfter {
			model.AccountingComplete = false
			continue
		}
		model.Online = true
		model.SessionCount++
	}

	return model, nil
}

func validSessionKey(key SessionKey) bool {
	return validUUID(key.RuntimeCredentialID) &&
		validDiagnostic(key.NodeID) &&
		validUUID(key.BootID) &&
		validUUID(key.SessionID)
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
