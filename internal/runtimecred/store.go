package runtimecred

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const credentialColumns = `
id::text,
username,
secret_hash,
secret_ciphertext,
secret_nonce,
encryption_key_id,
status,
origin,
created_by_actor_id::text,
updated_by_actor_id::text,
revision,
created_at,
updated_at,
rotated_at,
revoked_at`

var ErrUsernameConflict = errors.New("runtimecred: username already exists")

type Store struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("runtimecred: nil database")
	}
	return &Store{db: db}, nil
}

func (s *Store) ListTx(ctx context.Context, tx *sql.Tx) ([]Credential, error) {
	if err := requireTx(tx); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+credentialColumns+` FROM pvnaive.naive_runtime_credentials ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("runtimecred: list credentials: %w", err)
	}
	defer rows.Close()

	credentials := make([]Credential, 0)
	for rows.Next() {
		credential, err := scanCredential(rows)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("runtimecred: iterate credentials: %w", err)
	}
	return credentials, nil
}

func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, credential Credential) (Credential, error) {
	if err := requireTx(tx); err != nil {
		return Credential{}, err
	}
	if err := validateCredentialForCreate(credential); err != nil {
		return Credential{}, err
	}

	created, err := scanCredential(tx.QueryRowContext(ctx, `
INSERT INTO pvnaive.naive_runtime_credentials (
    username, secret_hash, secret_ciphertext, secret_nonce, encryption_key_id,
    status, origin, created_by_actor_id, updated_by_actor_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    pvnaive.current_actor_id(), pvnaive.current_actor_id()
)
RETURNING `+credentialColumns,
		credential.Username,
		credential.secretHash[:],
		credential.secretCiphertext,
		credential.secretNonce,
		credential.EncryptionKeyID,
		string(credential.Status),
		string(credential.Origin),
	))
	if err != nil {
		return Credential{}, mapMutationError("create credential", err)
	}
	return created, nil
}

func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, id string, expectedRevision int64, username string, status CredentialStatus, actorID string) (Credential, error) {
	if err := requireMutationArgs(tx, id, expectedRevision, actorID); err != nil {
		return Credential{}, err
	}
	if err := ValidateUsername(username); err != nil {
		return Credential{}, err
	}
	if !validCredentialStatus(status) {
		return Credential{}, ErrInvalidCredentialStatus
	}

	updated, err := scanCredential(tx.QueryRowContext(ctx, `
UPDATE pvnaive.naive_runtime_credentials
   SET username = $3,
       status = $4,
       updated_by_actor_id = pvnaive.current_actor_id(),
       revision = revision + 1,
       updated_at = clock_timestamp(),
       revoked_at = CASE
           WHEN $4 = 'revoked' AND status <> 'revoked' THEN clock_timestamp()
           WHEN $4 <> 'revoked' THEN NULL
           ELSE revoked_at
       END
 WHERE id = $1::uuid
   AND revision = $2
   AND status <> 'revoked'
RETURNING `+credentialColumns, id, expectedRevision, username, string(status)))
	if err != nil {
		return Credential{}, mapRevisionMutationError("update credential", err)
	}
	return updated, nil
}

func (s *Store) RotateTx(ctx context.Context, tx *sql.Tx, id string, expectedRevision int64, hash [32]byte, ciphertext, nonce []byte, keyID, actorID string) (Credential, error) {
	if err := requireMutationArgs(tx, id, expectedRevision, actorID); err != nil {
		return Credential{}, err
	}
	if len(ciphertext) < 16 || len(nonce) != 12 || strings.TrimSpace(keyID) == "" || len(keyID) > 160 {
		return Credential{}, errors.New("runtimecred: invalid rotated secret envelope")
	}

	rotated, err := scanCredential(tx.QueryRowContext(ctx, `
UPDATE pvnaive.naive_runtime_credentials
   SET secret_hash = $3,
       secret_ciphertext = $4,
       secret_nonce = $5,
       encryption_key_id = $6,
       updated_by_actor_id = pvnaive.current_actor_id(),
       revision = revision + 1,
       rotated_at = clock_timestamp(),
       updated_at = clock_timestamp()
 WHERE id = $1::uuid
   AND revision = $2
   AND status <> 'revoked'
RETURNING `+credentialColumns, id, expectedRevision, hash[:], ciphertext, nonce, keyID))
	if err != nil {
		return Credential{}, mapRevisionMutationError("rotate credential", err)
	}
	return rotated, nil
}

func (s *Store) RevokeTx(ctx context.Context, tx *sql.Tx, id string, expectedRevision int64, actorID string) (Credential, error) {
	if err := requireMutationArgs(tx, id, expectedRevision, actorID); err != nil {
		return Credential{}, err
	}

	revoked, err := scanCredential(tx.QueryRowContext(ctx, `
UPDATE pvnaive.naive_runtime_credentials
   SET status = 'revoked',
       updated_by_actor_id = pvnaive.current_actor_id(),
       revision = revision + 1,
       revoked_at = clock_timestamp(),
       updated_at = clock_timestamp()
 WHERE id = $1::uuid
   AND revision = $2
   AND status <> 'revoked'
RETURNING `+credentialColumns, id, expectedRevision))
	if err != nil {
		return Credential{}, mapRevisionMutationError("revoke credential", err)
	}
	return revoked, nil
}

func (s *Store) FindRevisionByIdempotencyTx(ctx context.Context, tx *sql.Tx, idempotencyKey string) (*RuntimeRevision, error) {
	if err := requireTx(tx); err != nil {
		return nil, err
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 160 {
		return nil, errors.New("runtimecred: invalid idempotency key")
	}

	var revision RuntimeRevision
	var manifestJSON []byte
	var failure sql.NullString
	var appliedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT id::text, revision_no, state, idempotency_key,
       config_checksum_sha256, config_ciphertext, encryption_key_id,
       manifest, created_by_actor_id::text, failure_code, created_at, applied_at
  FROM pvnaive.runtime_revisions
 WHERE tenant_id IS NULL
   AND protocol_id = 'naive'
   AND idempotency_key = $1
 LIMIT 1`, idempotencyKey).Scan(
		&revision.ID,
		&revision.RevisionNo,
		&revision.State,
		&revision.IdempotencyKey,
		&revision.ConfigChecksumSHA256,
		&revision.ConfigCiphertext,
		&revision.EncryptionKeyID,
		&manifestJSON,
		&revision.CreatedByActorID,
		&failure,
		&revision.CreatedAt,
		&appliedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("runtimecred: find idempotency revision: %w", err)
	}
	if failure.Valid {
		revision.FailureCode = failure.String
	}
	if appliedAt.Valid {
		v := appliedAt.Time
		revision.AppliedAt = &v
	}
	if len(manifestJSON) > 0 {
		if err := json.Unmarshal(manifestJSON, &revision.Manifest); err != nil {
			return nil, fmt.Errorf("runtimecred: decode revision manifest: %w", err)
		}
		if raw, ok := revision.Manifest["runtime_apply"]; ok {
			encoded, err := json.Marshal(raw)
			if err == nil {
				_ = json.Unmarshal(encoded, &revision.Applied)
			}
		}
	}
	return &revision, nil
}

func (s *Store) CreateRuntimeRevisionTx(ctx context.Context, tx *sql.Tx, revision RuntimeRevision) (RuntimeRevision, error) {
	if err := requireTx(tx); err != nil {
		return RuntimeRevision{}, err
	}
	if revision.State != "staged" {
		return RuntimeRevision{}, errors.New("runtimecred: new runtime revision must be staged")
	}
	if len(revision.IdempotencyKey) < 8 || len(revision.IdempotencyKey) > 160 {
		return RuntimeRevision{}, errors.New("runtimecred: invalid idempotency key")
	}
	if len(revision.ConfigChecksumSHA256) != 64 || len(revision.ConfigCiphertext) == 0 || revision.EncryptionKeyID == "" || revision.CreatedByActorID == "" {
		return RuntimeRevision{}, errors.New("runtimecred: incomplete runtime revision")
	}
	manifestJSON, err := json.Marshal(revision.Manifest)
	if err != nil {
		return RuntimeRevision{}, fmt.Errorf("runtimecred: encode revision manifest: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('pvnaive:runtime:naive', 0))`); err != nil {
		return RuntimeRevision{}, fmt.Errorf("runtimecred: lock runtime revision sequence: %w", err)
	}

	var created RuntimeRevision
	var returnedManifest []byte
	var previous sql.NullString
	err = tx.QueryRowContext(ctx, `
WITH next_revision AS (
    SELECT COALESCE(MAX(revision_no), 0) + 1 AS revision_no
      FROM pvnaive.runtime_revisions
     WHERE tenant_id IS NULL AND protocol_id = 'naive'
), previous_revision AS (
    SELECT id
      FROM pvnaive.runtime_revisions
     WHERE tenant_id IS NULL AND protocol_id = 'naive' AND state = 'applied'
     ORDER BY revision_no DESC
     LIMIT 1
)
INSERT INTO pvnaive.runtime_revisions (
    tenant_id, protocol_id, revision_no, state, config_checksum_sha256,
    config_ciphertext, encryption_key_id, manifest, previous_revision_id,
    created_by_actor_id, idempotency_key
)
SELECT NULL, 'naive', n.revision_no, 'staged', $1, $2, $3, $4::jsonb,
       p.id, $5::uuid, $6
  FROM next_revision n
  LEFT JOIN previous_revision p ON true
RETURNING id::text, revision_no, state, config_checksum_sha256,
          config_ciphertext, encryption_key_id, manifest,
          previous_revision_id::text, created_by_actor_id::text,
          idempotency_key, created_at`,
		revision.ConfigChecksumSHA256,
		revision.ConfigCiphertext,
		revision.EncryptionKeyID,
		manifestJSON,
		revision.CreatedByActorID,
		revision.IdempotencyKey,
	).Scan(
		&created.ID,
		&created.RevisionNo,
		&created.State,
		&created.ConfigChecksumSHA256,
		&created.ConfigCiphertext,
		&created.EncryptionKeyID,
		&returnedManifest,
		&previous,
		&created.CreatedByActorID,
		&created.IdempotencyKey,
		&created.CreatedAt,
	)
	if err != nil {
		return RuntimeRevision{}, mapMutationError("create runtime revision", err)
	}
	if err := json.Unmarshal(returnedManifest, &created.Manifest); err != nil {
		return RuntimeRevision{}, fmt.Errorf("runtimecred: decode created revision manifest: %w", err)
	}
	return created, nil
}

func (s *Store) MarkRevisionAppliedTx(ctx context.Context, tx *sql.Tx, id string, metadata AppliedRuntimeMetadata) error {
	if err := requireTx(tx); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" || metadata.BackupID == "" || len(metadata.PreviousSHA256) != 64 || len(metadata.AppliedSHA256) != 64 {
		return errors.New("runtimecred: invalid applied revision metadata")
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("runtimecred: encode applied metadata: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
UPDATE pvnaive.runtime_revisions
   SET state = 'applied',
       applied_at = clock_timestamp(),
       manifest = manifest || jsonb_build_object('runtime_apply', $2::jsonb)
 WHERE id = $1::uuid
   AND tenant_id IS NULL
   AND protocol_id = 'naive'
   AND state IN ('staged', 'validated')`, id, encoded)
	if err != nil {
		return fmt.Errorf("runtimecred: mark revision applied: %w", err)
	}
	return requireOneRow(result, "mark revision applied")
}

func (s *Store) MarkRevisionFailedTx(ctx context.Context, tx *sql.Tx, id, code string) error {
	if err := requireTx(tx); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" || strings.TrimSpace(code) == "" || len(code) > 160 {
		return errors.New("runtimecred: invalid failed revision metadata")
	}
	result, err := tx.ExecContext(ctx, `
UPDATE pvnaive.runtime_revisions
   SET state = 'failed', failure_code = $2
 WHERE id = $1::uuid
   AND tenant_id IS NULL
   AND protocol_id = 'naive'
   AND state IN ('staged', 'validated')`, id, code)
	if err != nil {
		return fmt.Errorf("runtimecred: mark revision failed: %w", err)
	}
	return requireOneRow(result, "mark revision failed")
}

func scanCredential(scanner rowScanner) (Credential, error) {
	var credential Credential
	var hash []byte
	var status, origin string
	var createdBy, updatedBy sql.NullString
	var rotatedAt, revokedAt sql.NullTime
	if err := scanner.Scan(
		&credential.ID,
		&credential.Username,
		&hash,
		&credential.secretCiphertext,
		&credential.secretNonce,
		&credential.EncryptionKeyID,
		&status,
		&origin,
		&createdBy,
		&updatedBy,
		&credential.Revision,
		&credential.CreatedAt,
		&credential.UpdatedAt,
		&rotatedAt,
		&revokedAt,
	); err != nil {
		return Credential{}, err
	}
	if len(hash) != 32 || len(credential.secretNonce) != 12 || len(credential.secretCiphertext) < 16 {
		return Credential{}, errors.New("runtimecred: invalid persisted secret envelope")
	}
	copy(credential.secretHash[:], hash)
	credential.secretCiphertext = append([]byte(nil), credential.secretCiphertext...)
	credential.secretNonce = append([]byte(nil), credential.secretNonce...)
	credential.Status = CredentialStatus(status)
	credential.Origin = CredentialOrigin(origin)
	if !validCredentialStatus(credential.Status) || (credential.Origin != CredentialImported && credential.Origin != CredentialPanel) {
		return Credential{}, errors.New("runtimecred: invalid persisted credential state")
	}
	if createdBy.Valid {
		credential.CreatedByActorID = createdBy.String
	}
	if updatedBy.Valid {
		credential.UpdatedByActorID = updatedBy.String
	}
	if rotatedAt.Valid {
		v := rotatedAt.Time
		credential.RotatedAt = &v
	}
	if revokedAt.Valid {
		v := revokedAt.Time
		credential.RevokedAt = &v
	}
	return credential, nil
}

func validateCredentialForCreate(credential Credential) error {
	if err := ValidateUsername(credential.Username); err != nil {
		return err
	}
	if len(credential.secretNonce) != 12 || len(credential.secretCiphertext) < 16 || strings.TrimSpace(credential.EncryptionKeyID) == "" || len(credential.EncryptionKeyID) > 160 {
		return errors.New("runtimecred: invalid credential secret envelope")
	}
	if !validCredentialStatus(credential.Status) {
		return ErrInvalidCredentialStatus
	}
	if credential.Origin != CredentialImported && credential.Origin != CredentialPanel {
		return errors.New("runtimecred: invalid credential origin")
	}
	return nil
}

func requireTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("runtimecred: bound transaction is required")
	}
	return nil
}

func requireMutationArgs(tx *sql.Tx, id string, expectedRevision int64, actorID string) error {
	if err := requireTx(tx); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" || expectedRevision <= 0 || strings.TrimSpace(actorID) == "" {
		return errors.New("runtimecred: credential id, expected revision and actor id are required")
	}
	return nil
}

func requireOneRow(result sql.Result, action string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("runtimecred: %s rows affected: %w", action, err)
	}
	if rows != 1 {
		return fmt.Errorf("runtimecred: %s affected %d rows", action, rows)
	}
	return nil
}

func mapRevisionMutationError(action string, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRevisionConflict
	}
	return mapMutationError(action, err)
}

func mapMutationError(action string, err error) error {
	if sqlState(err) == "23505" {
		if action == "create runtime revision" {
			return ErrIdempotentReplay
		}
		return ErrUsernameConflict
	}
	return fmt.Errorf("runtimecred: %s: %w", action, err)
}

func sqlState(err error) string {
	type stateError interface {
		SQLState() string
	}
	var state stateError
	if errors.As(err, &state) {
		return state.SQLState()
	}
	return ""
}
