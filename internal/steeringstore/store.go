// Package steeringstore persists R2 steering decisions (the steeringsched
// DecisionSink hook, docs/STEERING_SPEC_FA.md §2). The audit side is
// append-only; the current-state side is the future R5 renderer ordering
// input. All details payloads carry only node ids and scores — no
// credentials, no headers (redaction rule).
package steeringstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/steering"
)

// maxDecisionCandidates caps the details payload: the engine may score many
// nodes once the fleet lands, but the audit row keeps the best subset.
const maxDecisionCandidates = 50

// Store applies steering decisions through the 0032 SECURITY DEFINER
// boundary (pvnaive.steering_decision_apply).
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("steeringstore: PostgreSQL handle is required")
	}
	return &Store{db: db}, nil
}

// Current is the last applied decision for a user. Zero value = no decision
// yet (honest Unknown; callers must not invent a node).
type Current struct {
	UserID       string
	NodeID       string
	PreviousNode string
	WindowIndex  int64
	Reason       string
	DecidedAt    time.Time
	Found        bool
}

// ApplyDecision implements steeringsched.DecisionSink. Sink errors are
// counted by the scheduler per user and never abort the tick.
func (s *Store) ApplyDecision(ctx context.Context, decision steering.Decision) error {
	if s == nil || s.db == nil {
		return errors.New("steeringstore: PostgreSQL handle is required")
	}
	if err := ValidateDecision(decision); err != nil {
		return err
	}
	details, err := buildDetails(decision)
	if err != nil {
		return err
	}
	var tracked, accepted, duplicate bool
	err = s.db.QueryRowContext(ctx, `
SELECT tracked, accepted, duplicate
FROM pvnaive.steering_decision_apply($1::uuid, $2, $3, $4, $5, $6, $7)`,
		decision.UserID, decision.Primary, nullableText(decision.Previous),
		decision.WindowIdx, string(decision.Reason), time.Now().UTC(), details,
	).Scan(&tracked, &accepted, &duplicate)
	if err != nil {
		return fmt.Errorf("steeringstore: apply decision: %w", err)
	}
	if !tracked {
		// Honest fail-closed: the decision references a user the panel does
		// not know (deleted mid-flight). Nothing is fabricated.
		return fmt.Errorf("steeringstore: user %s is untracked", decision.UserID)
	}
	if !accepted && !duplicate {
		return fmt.Errorf("steeringstore: decision for user %s was neither applied nor duplicated", decision.UserID)
	}
	return nil
}

// ReadCurrent returns the applied decision for one user.
func (s *Store) ReadCurrent(ctx context.Context, userID string) (Current, error) {
	if s == nil || s.db == nil {
		return Current{}, errors.New("steeringstore: PostgreSQL handle is required")
	}
	if _, err := ParseUserID(userID); err != nil {
		return Current{}, err
	}
	var out Current
	var decidedAt sql.NullTime
	var previous sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT node_id, previous_node, window_index, reason, decided_at
FROM pvnaive.steering_state_read($1::uuid)`, userID,
	).Scan(&out.NodeID, &previous, &out.WindowIndex, &out.Reason, &decidedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Current{UserID: userID, Found: false}, nil
	}
	if err != nil {
		return Current{}, fmt.Errorf("steeringstore: read current: %w", err)
	}
	out.UserID = userID
	out.Found = true
	out.PreviousNode = previous.String
	if decidedAt.Valid {
		out.DecidedAt = decidedAt.Time.UTC()
	}
	return out, nil
}

// ValidateDecision mirrors the SQL contract so bad decisions fail at the
// caller with a precise error instead of a generic DB exception.
func ValidateDecision(decision steering.Decision) error {
	if _, err := ParseUserID(decision.UserID); err != nil {
		return err
	}
	if err := validateNode(decision.Primary); err != nil {
		return err
	}
	if decision.Previous != "" {
		if err := validateNode(decision.Previous); err != nil {
			return err
		}
	}
	if decision.WindowIdx < 0 {
		return fmt.Errorf("steeringstore: window index %d must be >= 0", decision.WindowIdx)
	}
	switch decision.Reason {
	case steering.ReasonInitial, steering.ReasonHysteresis, steering.ReasonKillSwitch:
	default:
		return fmt.Errorf("steeringstore: reason %q is not an actionable decision", string(decision.Reason))
	}
	return nil
}

func validateNode(node string) error {
	node = strings.TrimSpace(node)
	if node == "" || len(node) > 160 {
		return errors.New("steeringstore: node id must be 1..160 chars")
	}
	if strings.ContainsAny(node, " \t\r\n") {
		return errors.New("steeringstore: node id must not contain whitespace")
	}
	return nil
}

func ParseUserID(userID string) (string, error) {
	trimmed := strings.TrimSpace(userID)
	if len(trimmed) != 36 || strings.Count(trimmed, "-") != 4 {
		return "", fmt.Errorf("steeringstore: user id %q is not a uuid", userID)
	}
	return trimmed, nil
}

// decisionDetails is the redacted audit payload: node ids and scores only.
type decisionDetails struct {
	Candidates []decisionCandidate `json:"candidates,omitempty"`
}

type decisionCandidate struct {
	Node  string  `json:"node"`
	Score float64 `json:"score"`
	Kind  string  `json:"kind,omitempty"`
}

func buildDetails(decision steering.Decision) (json.RawMessage, error) {
	if len(decision.Candidates) == 0 {
		return nil, nil
	}
	limit := len(decision.Candidates)
	if limit > maxDecisionCandidates {
		limit = maxDecisionCandidates
	}
	payload := decisionDetails{Candidates: make([]decisionCandidate, 0, limit)}
	for _, candidate := range decision.Candidates[:limit] {
		if strings.TrimSpace(candidate.NodeID) == "" {
			return nil, errors.New("steeringstore: candidate with empty node id")
		}
		payload.Candidates = append(payload.Candidates, decisionCandidate{
			Node:  candidate.NodeID,
			Score: candidate.Score,
			Kind:  candidate.Kind,
		})
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("steeringstore: encode details: %w", err)
	}
	return raw, nil
}

func nullableText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
