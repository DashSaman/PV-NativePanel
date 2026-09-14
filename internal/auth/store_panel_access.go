package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// R7 panel access settings (ACCESS-001..002). Storage lives behind the 0030
// SECURITY DEFINER functions: pvnaive_app never receives direct table
// privileges, and password hashes are audited only as fixed redaction markers.

type PanelSettings struct {
	AdminUsername string
	PasswordHash  string // never serialized to API responses
	BasePath      string
	ListenPort    int
	GraceMinutes  int
	ExposureMode  string // "reverse_proxy" | "direct"
	SessionTTL    time.Duration
	UpdatedAt     *time.Time
	UpdatedBy     *string
}

type PanelAccessInput struct {
	AdminUsername string
	// PasswordHash: pass an argon2id hash to rotate the password, or "" to
	// keep the current one. Plaintext never reaches this layer.
	PasswordHash string
	BasePath     string
	ListenPort   int
	GraceMinutes int
	ExposureMode string
}

// ReadPanelSettings returns the stored settings row, or nil when the operator
// has never applied an access change (the caller then falls back to the
// actors-table identity and deployment defaults).
func (s *Store) ReadPanelSettings(ctx context.Context) (*PanelSettings, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("auth: store unavailable")
	}
	var (
		adminUsername string
		passwordHash  string
		basePath      string
		listenPort    int
		graceMinutes  int
		exposureMode  string
		sessionTTLSec int64
		updatedAt     sql.NullTime
		updatedBy     sql.NullString
	)
	err := s.db.QueryRowContext(ctx, `
SELECT admin_username, password_hash, base_path, listen_port, grace_minutes,
       exposure_mode, EXTRACT(epoch FROM session_ttl)::bigint, updated_at, updated_by
FROM pvnaive.panel_settings_read()`).Scan(
		&adminUsername, &passwordHash, &basePath, &listenPort, &graceMinutes,
		&exposureMode, &sessionTTLSec, &updatedAt, &updatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := &PanelSettings{
		AdminUsername: adminUsername,
		PasswordHash:  passwordHash,
		BasePath:      basePath,
		ListenPort:    listenPort,
		GraceMinutes:  graceMinutes,
		ExposureMode:  exposureMode,
		SessionTTL:    time.Duration(sessionTTLSec) * time.Second,
	}
	if updatedAt.Valid {
		t := updatedAt.Time.UTC()
		out.UpdatedAt = &t
	}
	if updatedBy.Valid && updatedBy.String != "" {
		id := updatedBy.String
		out.UpdatedBy = &id
	}
	return out, nil
}

// UpsertPanelSettings applies a validated change through the SECURITY DEFINER
// function and records the per-field audit trail (secrets redacted in SQL).
func (s *Store) UpsertPanelSettings(ctx context.Context, input PanelAccessInput, actorID *string) error {
	if s == nil || s.db == nil {
		return errors.New("auth: store unavailable")
	}
	if input.AdminUsername == "" || input.BasePath == "" || input.ListenPort <= 0 {
		return errors.New("auth: panel access input incomplete")
	}
	if input.GraceMinutes <= 0 {
		input.GraceMinutes = 10
	}
	if input.ExposureMode == "" {
		input.ExposureMode = "reverse_proxy"
	}
	var actorArg interface{}
	if actorID != nil && *actorID != "" {
		actorArg = *actorID
	}
	if _, err := s.db.ExecContext(ctx, `
SELECT pvnaive.panel_settings_upsert($1, $2, $3, $4, $5, $6, $7::uuid)`,
		input.AdminUsername, input.PasswordHash, input.BasePath, input.ListenPort,
		input.GraceMinutes, input.ExposureMode, actorArg); err != nil {
		return err
	}
	return nil
}

// EffectiveAdminUsername resolves the login identity used by the panel:
// the settings row when present, otherwise the owner actor's login email.
func (s *Store) EffectiveAdminUsername(ctx context.Context) (string, bool, error) {
	settings, err := s.ReadPanelSettings(ctx)
	if err != nil {
		return "", false, err
	}
	if settings != nil {
		return settings.AdminUsername, true, nil
	}
	var email string
	err = s.db.QueryRowContext(ctx, `
SELECT email FROM pvnaive.actors WHERE actor_role = 'owner' AND status = 'active'
ORDER BY created_at LIMIT 1`).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return email, false, nil
}
