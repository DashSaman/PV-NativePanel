package telemetry

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidEvent       = errors.New("telemetry: invalid event")
	ErrSequenceConflict   = errors.New("telemetry: sequence conflict")
	ErrSequenceGap        = errors.New("telemetry: sequence gap")
	ErrSequenceOutOfOrder = errors.New("telemetry: sequence out of order")
	ErrCounterRegression  = errors.New("telemetry: counter regression")
	ErrSessionIdentity    = errors.New("telemetry: session identity changed")
	ErrSessionClosed      = errors.New("telemetry: session already closed")
)

type Event struct {
	RuntimeCredentialID  string    `json:"runtime_credential_id"`
	Username             string    `json:"username"`
	NodeID               string    `json:"node_id"`
	BootID               string    `json:"boot_id"`
	SessionID            string    `json:"session_id"`
	Sequence             int64     `json:"sequence"`
	ObservedAt           time.Time `json:"timestamp"`
	AuthenticatedConnect bool      `json:"authenticated_connect"`
	UploadBytes          int64     `json:"upload_bytes"`
	DownloadBytes        int64     `json:"download_bytes"`
	Final                bool      `json:"final,omitempty"`
	ClientIP             string    `json:"client_ip,omitempty"`
}

type SessionKey struct {
	RuntimeCredentialID string
	NodeID              string
	BootID              string
	SessionID           string
}

func (e Event) Key() SessionKey {
	return SessionKey{
		RuntimeCredentialID: e.RuntimeCredentialID,
		NodeID:              e.NodeID,
		BootID:              e.BootID,
		SessionID:           e.SessionID,
	}
}

type ApplyResult struct {
	UploadDelta   int64
	DownloadDelta int64
	Duplicate     bool
	Final         bool
}

type SessionState struct {
	initialized bool
	key         SessionKey
	last        Event
	complete    bool
}

func ValidateEvent(event Event) error {
	if !validUUID(event.RuntimeCredentialID) ||
		!validUUID(event.BootID) ||
		!validUUID(event.SessionID) ||
		!validDiagnostic(event.Username) ||
		!validDiagnostic(event.NodeID) ||
		event.Sequence < 1 ||
		event.ObservedAt.IsZero() ||
		!event.AuthenticatedConnect ||
		event.UploadBytes < 0 ||
		event.DownloadBytes < 0 {
		return ErrInvalidEvent
	}
	return nil
}

func (s *SessionState) Apply(event Event) (ApplyResult, error) {
	if err := ValidateEvent(event); err != nil {
		return ApplyResult{}, err
	}

	key := event.Key()
	if !s.initialized {
		if event.Sequence != 1 {
			return ApplyResult{}, fmt.Errorf("%w: first sequence is %d", ErrSequenceGap, event.Sequence)
		}
		s.initialized = true
		s.key = key
		s.last = event
		s.complete = event.Final
		return ApplyResult{
			UploadDelta:   event.UploadBytes,
			DownloadDelta: event.DownloadBytes,
			Final:         event.Final,
		}, nil
	}

	if s.key != key {
		return ApplyResult{}, ErrSessionIdentity
	}

	if event.Sequence == s.last.Sequence {
		if sameEvent(s.last, event) {
			return ApplyResult{Duplicate: true, Final: event.Final}, nil
		}
		return ApplyResult{}, ErrSequenceConflict
	}
	if s.complete {
		return ApplyResult{}, ErrSessionClosed
	}
	if event.Sequence < s.last.Sequence {
		return ApplyResult{}, ErrSequenceOutOfOrder
	}
	if event.Sequence != s.last.Sequence+1 {
		return ApplyResult{}, ErrSequenceGap
	}
	if event.ObservedAt.Before(s.last.ObservedAt) {
		return ApplyResult{}, ErrSequenceOutOfOrder
	}
	if event.UploadBytes < s.last.UploadBytes || event.DownloadBytes < s.last.DownloadBytes {
		return ApplyResult{}, ErrCounterRegression
	}

	result := ApplyResult{
		UploadDelta:   event.UploadBytes - s.last.UploadBytes,
		DownloadDelta: event.DownloadBytes - s.last.DownloadBytes,
		Final:         event.Final,
	}
	s.last = event
	s.complete = event.Final
	return result, nil
}

func (s SessionState) Complete() bool {
	return s.initialized && s.complete
}

func sameEvent(a, b Event) bool {
	return a.RuntimeCredentialID == b.RuntimeCredentialID &&
		a.Username == b.Username &&
		a.NodeID == b.NodeID &&
		a.BootID == b.BootID &&
		a.SessionID == b.SessionID &&
		a.Sequence == b.Sequence &&
		a.ObservedAt.Equal(b.ObservedAt) &&
		a.AuthenticatedConnect == b.AuthenticatedConnect &&
		a.UploadBytes == b.UploadBytes &&
		a.DownloadBytes == b.DownloadBytes &&
		a.Final == b.Final &&
		a.ClientIP == b.ClientIP
}

func validDiagnostic(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && trimmed == value && len(value) <= 160
}

func validUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i := 0; i < len(value); i++ {
		switch i {
		case 8, 13, 18, 23:
			if value[i] != '-' {
				return false
			}
		default:
			if !isHex(value[i]) {
				return false
			}
		}
	}
	return true
}

func isHex(value byte) bool {
	return (value >= '0' && value <= '9') ||
		(value >= 'a' && value <= 'f') ||
		(value >= 'A' && value <= 'F')
}
