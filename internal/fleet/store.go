package fleet

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// PoolNode is one inventory row of the R5 pull-model registry
// (migration 0033). DesiredRevision/health/last-seen feed the owner UI
// (drift visibility); Maintenance carries the drain state machine.
type PoolNode struct {
	ID              string
	DisplayName     string
	Region          string
	CapacityWeight  int
	Health          string
	Maintenance     string
	DesiredRevision int64
	AppliedRevision int64
	LastSeenAt      *time.Time
}

// Registry lifecycle states reuse the model.go vocabulary (NodeState /
// MaintenanceState); the 0033 CHECK constraints accept the same literals.

const (
	// enrollmentTokenBytes: 32 random bytes -> 64 hex chars. Only the
	// SHA-256 hash is persisted; the raw token is shown to the operator
	// exactly once.
	enrollmentTokenBytes = 32
	// enrollmentTokenTTL bounds: 5 minutes .. 24 hours (match 0033).
	enrollmentTokenMinTTL = 5 * time.Minute
	enrollmentTokenMaxTTL = 24 * time.Hour
	// DefaultEnrollmentTTL is the operator default (one working session).
	DefaultEnrollmentTTL = 2 * time.Hour
)

var (
	validHealth = map[string]struct{}{
		string(NodeUnknown): {}, string(NodeHealthy): {}, string(NodeDegraded): {}, string(NodeOffline): {},
	}
	validMaintenance = map[string]struct{}{
		string(MaintenanceActive): {}, string(MaintenanceDraining): {}, string(MaintenanceDisabled): {},
	}
)

// ValidateNodeName / ValidateRegion / ValidateWeight mirror the 0032-style
// fail-closed input contract: the store refuses to send malformed state to
// the trusted boundary instead of letting the database reject it later.
func ValidateNodeName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 1 || len(trimmed) > 120 {
		return errors.New("fleet: node name must be 1..120 characters")
	}
	return nil
}

func ValidateRegion(region string) error {
	trimmed := strings.TrimSpace(region)
	if region == "" {
		return nil
	}
	if len(trimmed) < 1 || len(trimmed) > 60 {
		return errors.New("fleet: region must be empty or 1..60 characters")
	}
	return nil
}

func ValidateWeight(weight int) error {
	if weight < 1 || weight > 10000 {
		return errors.New("fleet: capacity weight must be 1..10000")
	}
	return nil
}

func ValidateHealth(health string) error {
	if _, ok := validHealth[health]; !ok {
		return fmt.Errorf("fleet: invalid health state %q", health)
	}
	return nil
}

func ValidateMaintenance(state string) error {
	if _, ok := validMaintenance[state]; !ok {
		return fmt.Errorf("fleet: invalid maintenance state %q", state)
	}
	return nil
}

// PrepareEnrollmentToken generates the raw one-time token and its SHA-256
// hash. The raw token never touches the database.
func PrepareEnrollmentToken() (rawToken, tokenHash string, err error) {
	buf := make([]byte, enrollmentTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("fleet: generate enrollment token: %w", err)
	}
	rawToken = hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

// HashEnrollmentToken derives the stored hash for a raw token (enrollment
// path). Constant-time not required (hash input, not a key comparison) but
// subtle is kept for parity with the credential boundary.
func HashEnrollmentToken(rawToken string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}

// ValidTokenHash reports whether hash is a 64-char lowercase hex digest
// (the 0033 CHECK contract).
func ValidTokenHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for _, c := range hash {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// ValidateEnrollmentTTL enforces the 0033 token TTL window.
func ValidateEnrollmentTTL(ttl time.Duration) error {
	if ttl < enrollmentTokenMinTTL || ttl > enrollmentTokenMaxTTL {
		return fmt.Errorf("fleet: enrollment TTL must be %s..%s", enrollmentTokenMinTTL, enrollmentTokenMaxTTL)
	}
	return nil
}

// PrepareRevision validates the manifest and produces the signed canonical
// envelope body the store persists. Pure: no database access, fully unit
// testable. The manifest validity window is generous (publish-time issuance,
// trust-time expiry enforcement per the R4 contract).
func PrepareRevision(m Manifest, privHex string, now time.Time) (ManifestEnvelope, error) {
	if !m.ValidFrom.IsZero() && !m.ValidUntil.IsZero() {
		if err := m.Validate(now); err != nil {
			return ManifestEnvelope{}, fmt.Errorf("fleet: manifest not publishable: %w", err)
		}
	}
	sig, err := SignManifest(m, privHex)
	if err != nil {
		return ManifestEnvelope{}, err
	}
	return ManifestEnvelope{Manifest: m, Signature: sig}, nil
}

// VerifyEnvelope is the pull-side trust gate (thin wrapper kept here so the
// registry and the agent share one verification path).
func VerifyEnvelope(env ManifestEnvelope, now time.Time) error {
	return VerifyManifestEnvelope(env, now)
}

// Store is the durable R5 registry accessor. It talks to migration 0033's
// SECURITY DEFINER functions only — never to the tables directly — and every
// method fails closed on store-unavailable (nil receiver), mirroring the
// steeringstore contract.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

// IssueEnrollmentToken persists a fresh token hash and returns the raw token
// exactly once (with its expiry) for the operator UI.
func (s *Store) IssueEnrollmentToken(ctx context.Context, nodeName string, ttl time.Duration) (rawToken string, expiresAt time.Time, err error) {
	if err := s.available(); err != nil {
		return "", time.Time{}, err
	}
	if err := ValidateNodeName(nodeName); err != nil {
		return "", time.Time{}, err
	}
	if err := ValidateEnrollmentTTL(ttl); err != nil {
		return "", time.Time{}, err
	}
	rawToken, tokenHash, err := PrepareEnrollmentToken()
	if err != nil {
		return "", time.Time{}, err
	}
	err = s.queryRowContext(ctx,
		`SELECT expires_at FROM pvnaive.pool_enrollment_token_record($1, $2, $3)`,
		tokenHash, strings.TrimSpace(nodeName), int(ttl.Seconds()),
	).Scan(&expiresAt)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("fleet: enrollment token record: %w", err)
	}
	return rawToken, expiresAt, nil
}

// Enroll consumes a raw token and creates the inventory node. Invalid,
// expired or replayed tokens report enrolled=false (honest, never fabricated).
func (s *Store) Enroll(ctx context.Context, rawToken, displayName, region string, weight int, actorID string) (nodeID string, enrolled bool, reason string, err error) {
	if err := s.available(); err != nil {
		return "", false, "", err
	}
	if err := ValidateNodeName(displayName); err != nil {
		return "", false, "", err
	}
	if err := ValidateRegion(region); err != nil {
		return "", false, "", err
	}
	if err := ValidateWeight(weight); err != nil {
		return "", false, "", err
	}
	if strings.TrimSpace(actorID) == "" {
		return "", false, "", errors.New("fleet: enrollment requires the acting owner/admin id")
	}
	tokenHash := HashEnrollmentToken(rawToken)
	if !ValidTokenHash(tokenHash) {
		return "", false, "", errors.New("fleet: enrollment token malformed")
	}
	err = s.queryRowContext(ctx,
		`SELECT node_id, enrolled, reason FROM pvnaive.pool_node_enroll($1, $2, NULLIF(btrim($3), ''), $4, $5)`,
		tokenHash, strings.TrimSpace(displayName), region, weight, actorID,
	).Scan(&nodeID, &enrolled, &reason)
	if err != nil {
		return "", false, "", fmt.Errorf("fleet: enrollment: %w", err)
	}
	return nodeID, enrolled, reason, nil
}

// PublishRevision signs and appends the next monotonic desired-state revision
// for a node. accepted=false means the node is unknown or disabled — the
// caller reports it instead of inventing a revision.
func (s *Store) PublishRevision(ctx context.Context, nodeID string, m Manifest, privHex string, now time.Time) (revision int64, accepted bool, err error) {
	if err := s.available(); err != nil {
		return 0, false, err
	}
	if strings.TrimSpace(nodeID) == "" {
		return 0, false, errors.New("fleet: node id is required")
	}
	env, err := PrepareRevision(m, privHex, now)
	if err != nil {
		return 0, false, err
	}
	manifestJSON, err := json.Marshal(env.Manifest)
	if err != nil {
		return 0, false, fmt.Errorf("fleet: manifest encode: %w", err)
	}
	err = s.queryRowContext(ctx,
		`SELECT revision, accepted FROM pvnaive.pool_revision_publish($1, $2, $3)`,
		nodeID, string(manifestJSON), env.Signature,
	).Scan(&revision, &accepted)
	if err != nil {
		return 0, false, fmt.Errorf("fleet: revision publish: %w", err)
	}
	return revision, accepted, nil
}

// LatestRevision is the pull-side read for sibling agents.
func (s *Store) LatestRevision(ctx context.Context, nodeID string) (revision int64, env ManifestEnvelope, found bool, err error) {
	if err := s.available(); err != nil {
		return 0, ManifestEnvelope{}, false, err
	}
	if strings.TrimSpace(nodeID) == "" {
		return 0, ManifestEnvelope{}, false, errors.New("fleet: node id is required")
	}
	var manifestJSON string
	var signature string
	err = s.queryRowContext(ctx,
		`SELECT revision, manifest::text, signature FROM pvnaive.pool_revision_latest($1)`,
		nodeID,
	).Scan(&revision, &manifestJSON, &signature)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, ManifestEnvelope{}, false, nil
	case err != nil:
		return 0, ManifestEnvelope{}, false, fmt.Errorf("fleet: revision latest: %w", err)
	}
	if err := json.Unmarshal([]byte(manifestJSON), &env.Manifest); err != nil {
		return 0, ManifestEnvelope{}, false, fmt.Errorf("fleet: stored manifest decode: %w", err)
	}
	env.Signature = signature
	return revision, env, true, nil
}

// Heartbeat records an agent's liveness + applied revision progress.
func (s *Store) Heartbeat(ctx context.Context, nodeID, health string, appliedRevision int64) (tracked bool, err error) {
	if err := s.available(); err != nil {
		return false, err
	}
	if err := ValidateHealth(health); err != nil {
		return false, err
	}
	if appliedRevision < 0 {
		return false, errors.New("fleet: applied revision must be non-negative")
	}
	err = s.queryRowContext(ctx,
		`SELECT tracked FROM pvnaive.pool_node_heartbeat($1, $2, $3)`,
		nodeID, health, appliedRevision,
	).Scan(&tracked)
	if err != nil {
		return false, fmt.Errorf("fleet: heartbeat: %w", err)
	}
	return tracked, nil
}

// ListNodes is the owner inventory view (drift included).
func (s *Store) ListNodes(ctx context.Context) ([]PoolNode, error) {
	if err := s.available(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT node_id, display_name, region, capacity_weight, health, maintenance, desired_revision, applied_revision, last_seen_at FROM pvnaive.pool_nodes_list()`)
	if err != nil {
		return nil, fmt.Errorf("fleet: node list: %w", err)
	}
	defer rows.Close()
	nodes := make([]PoolNode, 0, 8)
	for rows.Next() {
		var (
			n      PoolNode
			region sql.NullString
			seen   sql.NullTime
		)
		if err := rows.Scan(&n.ID, &n.DisplayName, &region, &n.CapacityWeight, &n.Health, &n.Maintenance, &n.DesiredRevision, &n.AppliedRevision, &seen); err != nil {
			return nil, fmt.Errorf("fleet: node list scan: %w", err)
		}
		n.Region = region.String
		if seen.Valid {
			t := seen.Time
			n.LastSeenAt = &t
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fleet: node list rows: %w", err)
	}
	return nodes, nil
}

// SetMaintenance drives the drain state machine; the database refuses an
// active -> disabled jump (23514) and the store surfaces it as a typed error.
var ErrDrainRequired = errors.New("fleet: drain required before disabling a live node")

func (s *Store) SetMaintenance(ctx context.Context, nodeID, state string) (tracked bool, err error) {
	if err := s.available(); err != nil {
		return false, err
	}
	if err := ValidateMaintenance(state); err != nil {
		return false, err
	}
	err = s.queryRowContext(ctx,
		`SELECT tracked FROM pvnaive.pool_node_maintenance_set($1, $2)`,
		nodeID, state,
	).Scan(&tracked)
	if err != nil {
		if strings.Contains(err.Error(), "23514") || strings.Contains(err.Error(), "drain required") {
			return false, ErrDrainRequired
		}
		return false, fmt.Errorf("fleet: maintenance set: %w", err)
	}
	return tracked, nil
}

func (s *Store) available() error {
	if s == nil || s.db == nil {
		return errors.New("fleet: registry store unavailable")
	}
	return nil
}
