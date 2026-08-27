package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pvdb "github.com/DashSaman/PV-NaivePanel/internal/database"
)

type Store struct {
	db *sql.DB
}

type ActorRecord struct {
	ID            string
	TenantID      *string
	Role          string
	PasswordHash  string
	MFARequired   bool
	Status        string
	LockedUntil   *time.Time
	TOTPConfirmed bool
}

type Principal struct {
	ActorID     string  `json:"actor_id"`
	TenantID    *string `json:"tenant_id,omitempty"`
	Role        string  `json:"role"`
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
}

type SessionRecord struct {
	ID                string     `json:"id"`
	CSRFTokenHash     []byte     `json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
	LastSeenAt        *time.Time `json:"last_seen_at,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	AbsoluteExpiresAt time.Time  `json:"absolute_expires_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}

type AuthenticatedTx struct {
	Tx        *sql.Tx
	Principal Principal
	Session   SessionRecord
}

type RotatedSession struct {
	SessionID         string
	ActorID           string
	TenantID          *string
	RefreshFamilyID   string
	AbsoluteExpiresAt time.Time
	ReuseDetected     bool
}

type TOTPFactorRecord struct {
	Ciphertext   []byte
	Nonce        []byte
	KeyID        string
	LastUsedStep *int64
	ConfirmedAt  *time.Time
}

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("auth: nil database")
	}
	return &Store{db: db}, nil
}

func (s *Store) LookupActor(ctx context.Context, email string) (ActorRecord, error) {
	if s == nil || s.db == nil {
		return ActorRecord{}, errors.New("auth: store is not initialized")
	}
	if email == "" {
		return ActorRecord{}, errors.New("auth: email must not be empty")
	}
	var out ActorRecord
	var tenant sql.NullString
	var password sql.NullString
	var locked sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT actor_id::text, tenant_id::text, actor_role, password_hash, mfa_required, status, locked_until, totp_confirmed FROM pvnaive.auth_lookup_actor($1)`, email).
		Scan(&out.ID, &tenant, &out.Role, &password, &out.MFARequired, &out.Status, &locked, &out.TOTPConfirmed)
	if err != nil {
		return ActorRecord{}, fmt.Errorf("auth: lookup actor: %w", err)
	}
	if tenant.Valid {
		v := tenant.String
		out.TenantID = &v
	}
	if password.Valid {
		out.PasswordHash = password.String
	}
	if locked.Valid {
		v := locked.Time
		out.LockedUntil = &v
	}
	return out, nil
}

func (s *Store) RecordLoginFailure(ctx context.Context, actorID string) (*time.Time, error) {
	if actorID == "" {
		return nil, errors.New("auth: actor ID must not be empty")
	}
	var locked sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT pvnaive.auth_record_login_failure($1::uuid)`, actorID).Scan(&locked); err != nil {
		return nil, fmt.Errorf("auth: record login failure: %w", err)
	}
	if !locked.Valid {
		return nil, nil
	}
	v := locked.Time
	return &v, nil
}

func (s *Store) RecordLoginSuccess(ctx context.Context, actorID string) error {
	if actorID == "" {
		return errors.New("auth: actor ID must not be empty")
	}
	if _, err := s.db.ExecContext(ctx, `SELECT pvnaive.auth_record_login_success($1::uuid)`, actorID); err != nil {
		return fmt.Errorf("auth: record login success: %w", err)
	}
	return nil
}

func (s *Store) CreateSession(ctx context.Context, actorID string, tokenHash, csrfHash []byte, refreshFamilyID string, userAgentHash []byte, expiresAt, absoluteExpiresAt time.Time) (string, error) {
	if actorID == "" || refreshFamilyID == "" {
		return "", errors.New("auth: actor and refresh-family IDs are required")
	}
	if err := requireSHA256("session token", tokenHash); err != nil {
		return "", err
	}
	if err := requireSHA256("CSRF token", csrfHash); err != nil {
		return "", err
	}
	if len(userAgentHash) != 0 && len(userAgentHash) != 32 {
		return "", errors.New("auth: user-agent hash must be 32 bytes")
	}
	var id string
	if err := s.db.QueryRowContext(ctx, `SELECT pvnaive.auth_create_session($1::uuid,$2,$3,$4::uuid,$5,$6,$7)`, actorID, tokenHash, csrfHash, refreshFamilyID, nullableBytes(userAgentHash), expiresAt, absoluteExpiresAt).Scan(&id); err != nil {
		return "", fmt.Errorf("auth: create session: %w", err)
	}
	return id, nil
}

func (s *Store) RotateSession(ctx context.Context, oldHash, newHash, newCSRFHash, newUserAgentHash []byte, newExpiresAt time.Time) (RotatedSession, error) {
	for label, value := range map[string][]byte{"old session token": oldHash, "new session token": newHash, "new CSRF token": newCSRFHash} {
		if err := requireSHA256(label, value); err != nil {
			return RotatedSession{}, err
		}
	}
	if len(newUserAgentHash) != 0 && len(newUserAgentHash) != 32 {
		return RotatedSession{}, errors.New("auth: user-agent hash must be 32 bytes")
	}
	var out RotatedSession
	var tenant sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT session_id::text, actor_id::text, tenant_id::text, refresh_family_id::text, absolute_expires_at, reuse_detected FROM pvnaive.auth_rotate_session($1,$2,$3,$4,$5)`, oldHash, newHash, newCSRFHash, nullableBytes(newUserAgentHash), newExpiresAt).
		Scan(&out.SessionID, &out.ActorID, &tenant, &out.RefreshFamilyID, &out.AbsoluteExpiresAt, &out.ReuseDetected)
	if err != nil {
		return RotatedSession{}, fmt.Errorf("auth: rotate session: %w", err)
	}
	if tenant.Valid {
		v := tenant.String
		out.TenantID = &v
	}
	return out, nil
}

func (s *Store) RevokeSession(ctx context.Context, tokenHash []byte) (bool, error) {
	if err := requireSHA256("session token", tokenHash); err != nil {
		return false, err
	}
	var revoked bool
	if err := s.db.QueryRowContext(ctx, `SELECT pvnaive.auth_revoke_session($1)`, tokenHash).Scan(&revoked); err != nil {
		return false, fmt.Errorf("auth: revoke session: %w", err)
	}
	return revoked, nil
}

func (s *Store) RevokeActorSessions(ctx context.Context, actorID string) (int, error) {
	if actorID == "" {
		return 0, errors.New("auth: actor ID must not be empty")
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT pvnaive.auth_revoke_actor_sessions($1::uuid)`, actorID).Scan(&count); err != nil {
		return 0, fmt.Errorf("auth: revoke actor sessions: %w", err)
	}
	return count, nil
}

func (s *Store) BeginAuthenticated(ctx context.Context, tokenHash []byte) (*AuthenticatedTx, error) {
	if err := requireSHA256("session token", tokenHash); err != nil {
		return nil, err
	}
	if s == nil || s.db == nil {
		return nil, errors.New("auth: store is not initialized")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		return nil, fmt.Errorf("auth: begin request transaction: %w", err)
	}
	fail := func(cause error) (*AuthenticatedTx, error) {
		_ = tx.Rollback()
		return nil, cause
	}
	if err := pvdb.BindRequestContext(ctx, tx, tokenHash); err != nil {
		return fail(fmt.Errorf("auth: bind request context: %w", err))
	}

	var principal Principal
	var session SessionRecord
	var tenant sql.NullString
	var lastSeen, revoked sql.NullTime
	err = tx.QueryRowContext(ctx, `
SELECT a.id::text, a.tenant_id::text, a.actor_role, a.email, a.display_name,
       s.id::text, s.csrf_token_hash, s.created_at, s.last_seen_at,
       s.expires_at, s.absolute_expires_at, s.revoked_at
  FROM pvnaive.actors a
  JOIN pvnaive.auth_sessions s ON s.actor_id = a.id
 WHERE a.id = pvnaive.current_actor_id()
   AND s.token_hash = $1
   AND s.revoked_at IS NULL
   AND s.expires_at > clock_timestamp()
   AND s.absolute_expires_at > clock_timestamp()
 LIMIT 1`, tokenHash).Scan(
		&principal.ActorID, &tenant, &principal.Role, &principal.Email, &principal.DisplayName,
		&session.ID, &session.CSRFTokenHash, &session.CreatedAt, &lastSeen,
		&session.ExpiresAt, &session.AbsoluteExpiresAt, &revoked,
	)
	if err != nil {
		return fail(fmt.Errorf("auth: load authenticated principal: %w", err))
	}
	if tenant.Valid {
		v := tenant.String
		principal.TenantID = &v
	}
	if lastSeen.Valid {
		v := lastSeen.Time
		session.LastSeenAt = &v
	}
	if revoked.Valid {
		v := revoked.Time
		session.RevokedAt = &v
	}
	return &AuthenticatedTx{Tx: tx, Principal: principal, Session: session}, nil
}

func (s *Store) ListSessions(ctx context.Context, tx *sql.Tx, actorID string) ([]SessionRecord, error) {
	if tx == nil || actorID == "" {
		return nil, errors.New("auth: bound transaction and actor ID are required")
	}
	rows, err := tx.QueryContext(ctx, `SELECT id::text, created_at, last_seen_at, expires_at, absolute_expires_at, revoked_at FROM pvnaive.auth_sessions WHERE actor_id=$1::uuid ORDER BY created_at DESC`, actorID)
	if err != nil {
		return nil, fmt.Errorf("auth: list sessions: %w", err)
	}
	defer rows.Close()
	var out []SessionRecord
	for rows.Next() {
		var item SessionRecord
		var lastSeen, revoked sql.NullTime
		if err := rows.Scan(&item.ID, &item.CreatedAt, &lastSeen, &item.ExpiresAt, &item.AbsoluteExpiresAt, &revoked); err != nil {
			return nil, fmt.Errorf("auth: scan session: %w", err)
		}
		if lastSeen.Valid {
			v := lastSeen.Time
			item.LastSeenAt = &v
		}
		if revoked.Valid {
			v := revoked.Time
			item.RevokedAt = &v
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("auth: list sessions: %w", err)
	}
	return out, nil
}

func (s *Store) RevokeSessionByID(ctx context.Context, tx *sql.Tx, actorID, sessionID string) (bool, error) {
	if tx == nil || actorID == "" || sessionID == "" {
		return false, errors.New("auth: bound transaction, actor ID and session ID are required")
	}
	result, err := tx.ExecContext(ctx, `UPDATE pvnaive.auth_sessions SET revoked_at=COALESCE(revoked_at,clock_timestamp()) WHERE id=$1::uuid AND actor_id=$2::uuid AND revoked_at IS NULL`, sessionID, actorID)
	if err != nil {
		return false, fmt.Errorf("auth: revoke session by ID: %w", err)
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (s *Store) GetTOTPFactor(ctx context.Context, tx *sql.Tx, actorID string) (TOTPFactorRecord, error) {
	if tx == nil || actorID == "" {
		return TOTPFactorRecord{}, errors.New("auth: bound transaction and actor ID are required")
	}
	var out TOTPFactorRecord
	var step sql.NullInt64
	var confirmed sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT secret_ciphertext,secret_nonce,encryption_key_id,last_used_step,confirmed_at FROM pvnaive.auth_get_totp_factor($1::uuid)`, actorID).Scan(&out.Ciphertext, &out.Nonce, &out.KeyID, &step, &confirmed); err != nil {
		return TOTPFactorRecord{}, fmt.Errorf("auth: get TOTP factor: %w", err)
	}
	if step.Valid {
		v := step.Int64
		out.LastUsedStep = &v
	}
	if confirmed.Valid {
		v := confirmed.Time
		out.ConfirmedAt = &v
	}
	return out, nil
}

func (s *Store) UpsertTOTPFactor(ctx context.Context, tx *sql.Tx, actorID string, ciphertext, nonce []byte, keyID string) error {
	if tx == nil || actorID == "" || len(ciphertext) < 16 || len(nonce) != 12 || keyID == "" {
		return errors.New("auth: invalid TOTP persistence material")
	}
	_, err := tx.ExecContext(ctx, `SELECT pvnaive.auth_upsert_totp_factor($1::uuid,$2,$3,$4)`, actorID, ciphertext, nonce, keyID)
	if err != nil {
		return fmt.Errorf("auth: upsert TOTP factor: %w", err)
	}
	return nil
}

func (s *Store) ConfirmTOTPFactor(ctx context.Context, tx *sql.Tx, actorID string, step int64, recoveryHashes [][32]byte) error {
	if tx == nil || actorID == "" || step < 0 || len(recoveryHashes) != 10 {
		return errors.New("auth: invalid MFA confirmation material")
	}
	values := make([][]byte, len(recoveryHashes))
	for i := range recoveryHashes {
		values[i] = recoveryHashes[i][:]
	}
	_, err := tx.ExecContext(ctx, `SELECT pvnaive.auth_confirm_totp_factor($1::uuid,$2,$3::bytea[])`, actorID, step, values)
	if err != nil {
		return fmt.Errorf("auth: confirm TOTP factor: %w", err)
	}
	return nil
}

func (s *Store) ConsumeTOTPStep(ctx context.Context, tx *sql.Tx, actorID string, step int64) (bool, error) {
	if tx == nil || actorID == "" || step < 0 {
		return false, errors.New("auth: invalid TOTP consume request")
	}
	var ok bool
	if err := tx.QueryRowContext(ctx, `SELECT pvnaive.auth_consume_totp_step($1::uuid,$2)`, actorID, step).Scan(&ok); err != nil {
		return false, fmt.Errorf("auth: consume TOTP step: %w", err)
	}
	return ok, nil
}

func (s *Store) ConsumeRecoveryCode(ctx context.Context, tx *sql.Tx, actorID string, hash [32]byte) (bool, error) {
	if tx == nil || actorID == "" {
		return false, errors.New("auth: invalid recovery-code consume request")
	}
	var ok bool
	if err := tx.QueryRowContext(ctx, `SELECT pvnaive.auth_consume_recovery_code($1::uuid,$2)`, actorID, hash[:]).Scan(&ok); err != nil {
		return false, fmt.Errorf("auth: consume recovery code: %w", err)
	}
	return ok, nil
}

func (s *Store) RemoveMFA(ctx context.Context, tx *sql.Tx, actorID string) error {
	if tx == nil || actorID == "" {
		return errors.New("auth: invalid MFA removal request")
	}
	if _, err := tx.ExecContext(ctx, `SELECT pvnaive.auth_remove_mfa($1::uuid)`, actorID); err != nil {
		return fmt.Errorf("auth: remove MFA: %w", err)
	}
	return nil
}

func (s *Store) AppendAudit(ctx context.Context, actorID *string, action, outcome, reason string) error {
	if action == "" || outcome == "" {
		return errors.New("auth: audit action and outcome are required")
	}
	var id any
	if actorID != nil {
		id = *actorID
	}
	if _, err := s.db.ExecContext(ctx, `SELECT pvnaive.auth_append_audit($1::uuid,$2,$3,$4,NULL,NULL)`, id, action, outcome, nullableString(reason)); err != nil {
		return fmt.Errorf("auth: append audit: %w", err)
	}
	return nil
}

func requireSHA256(label string, value []byte) error {
	if len(value) != 32 {
		return fmt.Errorf("auth: %s hash must be 32 bytes", label)
	}
	return nil
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
