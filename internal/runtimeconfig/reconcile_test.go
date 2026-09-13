package runtimeconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestInstallReconciledConfigValidatesBeforeSwap(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "Caddyfile")
	original := []byte("original\n")
	if err := os.WriteFile(configPath, original, 0o644); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	candidate := []byte("reconciled\n")
	validateCalls := 0
	err := installReconciledConfig(configPath, candidate, func(path string) error {
		validateCalls++
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if string(content) != string(candidate) {
			t.Fatalf("validator saw %q, want candidate bytes", content)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("installReconciledConfig: %v", err)
	}
	if validateCalls != 1 {
		t.Fatalf("validator calls = %d, want 1", validateCalls)
	}
	installed, err := os.ReadFile(configPath)
	if err != nil || string(installed) != string(candidate) {
		t.Fatalf("installed config = %q (err=%v), want reconciled candidate", installed, err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("directory has %d entries, want exactly the installed config (temp files must be cleaned)", len(entries))
	}
}

func TestInstallReconciledConfigKeepsOriginalOnValidationFailure(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "Caddyfile")
	original := []byte("original\n")
	if err := os.WriteFile(configPath, original, 0o644); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	err := installReconciledConfig(configPath, []byte("broken\n"), func(path string) error {
		return errors.New("validation rejected")
	})
	if err == nil || !strings.Contains(err.Error(), "validation rejected") {
		t.Fatalf("installReconciledConfig error = %v, want validation failure", err)
	}
	installed, readErr := os.ReadFile(configPath)
	if readErr != nil || string(installed) != string(original) {
		t.Fatalf("config mutated despite validation failure: %q (err=%v)", installed, readErr)
	}
	entries, _ := os.ReadDir(directory)
	if len(entries) != 1 {
		t.Fatalf("temp candidate left behind after failed validation")
	}
}

func TestInstallReconciledConfigWithoutValidatorInstallsDirectly(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "nested", "Caddyfile")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("original\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := installReconciledConfig(configPath, []byte("ok\n"), nil); err != nil {
		t.Fatalf("install without validator: %v", err)
	}
	installed, _ := os.ReadFile(configPath)
	if string(installed) != "ok\n" {
		t.Fatalf("installed = %q", installed)
	}
}

func TestDesiredRenderInputsRejectsInvalidCredential(t *testing.T) {
	desired := []runtimecred.AgentCredential{{
		ID:       "11111111-1111-1111-1111-111111111111",
		Username: "bad user with space",
		Password: "whatever-secret-123",
		Status:   runtimecred.CredentialActive,
	}}
	_, err := desiredRenderInputs(desired)
	if err == nil || !strings.Contains(err.Error(), "bad user with space") {
		t.Fatalf("desiredRenderInputs error = %v, want username rejection", err)
	}
	if desired[0].Password != "" {
		t.Fatal("plaintext password survived a rejected render input conversion")
	}
}
