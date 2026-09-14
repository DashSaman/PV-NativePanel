// Package runtimeconfig reconciles the rendered runtime proxy config file
// with the active credentials persisted in the database. It exists as its
// own package because the reconciliation needs BOTH the encrypted-credential
// domain (internal/runtimecred) and the strict proxy-config renderer
// (internal/naiveruntime), which must not import each other.
package runtimeconfig

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DashSaman/PV-NaivePanel/internal/naiveruntime"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

// ReconcileResult reports the outcome of a boot-time config reconciliation.
type ReconcileResult struct {
	// Changed is true when the config file was rewritten to match the
	// database truth. False means the file already matched (or there was
	// nothing to reconcile) and it was left untouched.
	Changed bool
	// Credentials is the number of active credentials the reconcile
	// ensured are present in the rendered config credential block.
	Credentials int
}

// ReconcileRuntimeConfigFile rewrites the forward_proxy credential block of
// the runtime proxy config so it exactly matches the active credentials
// persisted in the database. This closes the all-in-one boot renderer gap
// (DEPLOY-001): the entrypoint previously rendered a bootstrap-only config
// whenever the persisted file was absent, silently dropping every active
// customer credential on container recreation. The reconcile runs on every
// boot before the proxy starts, treats the database as the single source of
// truth, reuses the exact production renderer so a reconciled file is
// byte-equivalent to one produced by a live runtime apply, and validates the
// candidate with the pinned proxy binary before an atomic swap.
//
// A missing or unparseable config is an error: the entrypoint always renders
// a fresh template first, so reconcile must never see a file it cannot
// parse. Zero active credentials is a no-op (a fresh install keeps the
// bootstrap seed credential until the first runtime apply replaces it).
func ReconcileRuntimeConfigFile(
	ctx context.Context,
	db *sql.DB,
	key []byte,
	configPath string,
	validate func(path string) error,
) (ReconcileResult, error) {
	if db == nil {
		return ReconcileResult{}, errors.New("runtimeconfig: reconcile requires a database connection")
	}
	if len(key) != 32 {
		return ReconcileResult{}, fmt.Errorf("runtimeconfig: runtime key must be 32 bytes, got %d", len(key))
	}
	if filepath.Dir(configPath) == "." {
		return ReconcileResult{}, errors.New("runtimeconfig: config path must include its directory")
	}

	store, err := runtimecred.NewStore(db)
	if err != nil {
		return ReconcileResult{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("runtimeconfig: open reconcile transaction: %w", err)
	}
	credentials, err := store.ListTx(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return ReconcileResult{}, err
	}
	// Read-only snapshot: the boot reconcile never mutates credential rows.
	if err := tx.Rollback(); err != nil {
		return ReconcileResult{}, fmt.Errorf("runtimeconfig: close reconcile transaction: %w", err)
	}

	desired, err := runtimecred.ReconcileDesiredCredentials(key, credentials)
	if err != nil {
		return ReconcileResult{}, err
	}
	if len(desired) == 0 {
		return ReconcileResult{Changed: false, Credentials: 0}, nil
	}

	renderInput, err := desiredRenderInputs(desired)
	if err != nil {
		return ReconcileResult{}, err
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("runtimeconfig: read config for reconcile: %w", err)
	}
	match, err := naiveruntime.CredentialsMatch(content, renderInput)
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("runtimeconfig: inspect config for reconcile: %w", err)
	}
	if match {
		zeroPasswords(desired)
		return ReconcileResult{Changed: false, Credentials: len(desired)}, nil
	}

	candidate, err := naiveruntime.RenderCredentials(content, renderInput)
	zeroPasswords(desired)
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("runtimeconfig: render reconciled config: %w", err)
	}
	if err := installReconciledConfig(configPath, candidate, validate); err != nil {
		return ReconcileResult{}, err
	}
	return ReconcileResult{Changed: true, Credentials: len(desired)}, nil
}

// desiredRenderInputs converts decrypted agent credentials into renderer
// inputs. NewImportedDesiredCredential is the constructor that accepts the
// live credential population without silently rotating legacy passwords.
func desiredRenderInputs(desired []runtimecred.AgentCredential) ([]runtimecred.DesiredCredential, error) {
	inputs := make([]runtimecred.DesiredCredential, 0, len(desired))
	for _, credential := range desired {
		input, err := runtimecred.NewImportedDesiredCredential(
			credential.ID, credential.Username, credential.Password, credential.Status)
		if err != nil {
			zeroPasswords(desired)
			return nil, fmt.Errorf("runtimeconfig: invalid credential %q: %w", credential.Username, err)
		}
		inputs = append(inputs, input)
	}
	return inputs, nil
}

// zeroPasswords clears the exported plaintext password copies. The renderer
// input retains its unexported copy by design of the credential domain
// boundary; the authoritative hygiene contract is covered by the reconciled
// candidate being written, validated and discarded within one call.
func zeroPasswords(desired []runtimecred.AgentCredential) {
	for i := range desired {
		desired[i].Password = ""
	}
}

// installReconciledConfig atomically installs the reconciled candidate: the
// candidate is written next to the live config, handed to the validation
// callback (the pinned proxy binary must accept it) and only then renamed
// over the live path. A validation failure leaves the live config untouched.
func installReconciledConfig(configPath string, candidate []byte, validate func(path string) error) error {
	directory := filepath.Dir(configPath)
	file, err := os.CreateTemp(directory, ".pvnaive-reconcile-*")
	if err != nil {
		return fmt.Errorf("runtimeconfig: create reconcile candidate: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	clean := func(cause error) error {
		_ = file.Close()
		_ = os.Remove(path)
		return cause
	}
	if err := file.Chmod(0o644); err != nil {
		return clean(fmt.Errorf("runtimeconfig: chmod reconcile candidate: %w", err))
	}
	if _, err := file.Write(candidate); err != nil {
		return clean(fmt.Errorf("runtimeconfig: write reconcile candidate: %w", err))
	}
	if err := file.Sync(); err != nil {
		return clean(fmt.Errorf("runtimeconfig: sync reconcile candidate: %w", err))
	}
	if err := file.Close(); err != nil {
		return clean(fmt.Errorf("runtimeconfig: close reconcile candidate: %w", err))
	}
	if validate != nil {
		if err := validate(path); err != nil {
			return fmt.Errorf("runtimeconfig: reconcile candidate rejected by validation: %w", err)
		}
	}
	if err := os.Rename(path, configPath); err != nil {
		return fmt.Errorf("runtimeconfig: install reconciled config: %w", err)
	}
	return nil
}
