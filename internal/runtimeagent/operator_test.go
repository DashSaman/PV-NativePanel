package runtimeagent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type commandCall struct {
	name string
	args []string
}

type scriptedRunner struct {
	mu          sync.Mutex
	calls       []commandCall
	mainPIDs    []string
	nRestarts   []string
	active      []string
	validateErr error
	reloadErr   error
}

func (r *scriptedRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, commandCall{name: name, args: append([]string(nil), args...)})

	joined := strings.Join(args, " ")
	switch {
	case strings.Contains(joined, " validate ") || (len(args) > 0 && args[0] == "validate"):
		if r.validateErr != nil {
			return nil, r.validateErr
		}
		return []byte("valid\n"), nil
	case name == "systemctl" && joined == "reload caddy-naive.service":
		if r.reloadErr != nil {
			return nil, r.reloadErr
		}
		return nil, nil
	case name == "systemctl" && joined == "is-active caddy-naive.service":
		return popScript(&r.active, "active\n"), nil
	case name == "systemctl" && joined == "show --property=MainPID --value caddy-naive.service":
		return popScript(&r.mainPIDs, "1234\n"), nil
	case name == "systemctl" && joined == "show --property=NRestarts --value caddy-naive.service":
		return popScript(&r.nRestarts, "0\n"), nil
	default:
		return nil, fmt.Errorf("unexpected command: %s %s", name, joined)
	}
}

func popScript(values *[]string, fallback string) []byte {
	if len(*values) == 0 {
		return []byte(fallback)
	}
	value := (*values)[0]
	*values = (*values)[1:]
	return []byte(value)
}

func (r *scriptedRunner) snapshotCalls() []commandCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]commandCall, len(r.calls))
	copy(out, r.calls)
	return out
}

func TestOperatorApplyValidatesBacksUpReloadsAndPreservesServiceProcess(t *testing.T) {
	op, caddyfile, backupRoot, runner := newTestOperator(t)
	oldBytes := liveLikeCaddy("legacy.user", "legacy password 123")
	mustWriteFile(t, caddyfile, oldBytes, 0640)
	oldSHA := shaHex(oldBytes)

	response, err := op.Apply(context.Background(), ApplyRequest{
		ExpectedCaddySHA256: oldSHA,
		Desired:             desiredInput(t, "rev-2", "new.user", "new safe password 123"),
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if response.PreviousSHA256 != oldSHA {
		t.Fatalf("previous SHA = %q, want %q", response.PreviousSHA256, oldSHA)
	}
	newBytes := mustReadFile(t, caddyfile)
	if bytes.Equal(newBytes, oldBytes) {
		t.Fatal("Apply() did not install candidate bytes")
	}
	if response.AppliedSHA256 != shaHex(newBytes) {
		t.Fatalf("applied SHA = %q, want %q", response.AppliedSHA256, shaHex(newBytes))
	}
	if response.MainPID != 1234 || response.NRestarts != 0 {
		t.Fatalf("service process metadata = pid:%d restarts:%d", response.MainPID, response.NRestarts)
	}
	if response.BackupID == "" {
		t.Fatal("Apply() returned empty backup id")
	}
	backupBytes := mustReadFile(t, filepath.Join(backupRoot, response.BackupID, "Caddyfile"))
	if !bytes.Equal(backupBytes, oldBytes) {
		t.Fatal("backup bytes differ from exact pre-apply Caddyfile")
	}

	calls := runner.snapshotCalls()
	validateIndex := commandIndex(calls, "/usr/local/bin/caddy", "validate")
	reloadIndex := commandIndex(calls, "systemctl", "reload caddy-naive.service")
	if validateIndex < 0 || reloadIndex < 0 || validateIndex > reloadIndex {
		t.Fatalf("command order did not validate before reload: %#v", calls)
	}
	for _, call := range calls {
		if call.name == "systemctl" && len(call.args) > 0 && call.args[0] == "restart" {
			t.Fatalf("Apply() invoked forbidden restart: %#v", call)
		}
	}
}

func TestOperatorApplyRejectsExpectedSHAMismatchBeforeAnyCommandOrBackup(t *testing.T) {
	op, caddyfile, backupRoot, runner := newTestOperator(t)
	oldBytes := liveLikeCaddy("legacy.user", "legacy password 123")
	mustWriteFile(t, caddyfile, oldBytes, 0640)

	_, err := op.Apply(context.Background(), ApplyRequest{
		ExpectedCaddySHA256: strings.Repeat("0", 64),
		Desired:             desiredInput(t, "rev-2", "new.user", "new safe password 123"),
	})
	if err == nil {
		t.Fatal("Apply() unexpectedly accepted stale expected SHA")
	}
	if got := runner.snapshotCalls(); len(got) != 0 {
		t.Fatalf("commands ran before SHA mismatch rejection: %#v", got)
	}
	if !bytes.Equal(mustReadFile(t, caddyfile), oldBytes) {
		t.Fatal("Caddyfile changed on SHA mismatch")
	}
	entries, err := os.ReadDir(backupRoot)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("backup created on SHA mismatch: %d entries", len(entries))
	}
}

func TestOperatorApplyValidationFailureNeverInstallsOrReloads(t *testing.T) {
	op, caddyfile, backupRoot, runner := newTestOperator(t)
	runner.validateErr = errors.New("candidate rejected")
	oldBytes := liveLikeCaddy("legacy.user", "legacy password 123")
	mustWriteFile(t, caddyfile, oldBytes, 0640)

	_, err := op.Apply(context.Background(), ApplyRequest{
		ExpectedCaddySHA256: shaHex(oldBytes),
		Desired:             desiredInput(t, "rev-2", "new.user", "new safe password 123"),
	})
	if err == nil {
		t.Fatal("Apply() unexpectedly succeeded after caddy validate failure")
	}
	if !bytes.Equal(mustReadFile(t, caddyfile), oldBytes) {
		t.Fatal("Caddyfile changed after validation failure")
	}
	for _, call := range runner.snapshotCalls() {
		if call.name == "systemctl" && len(call.args) > 0 && call.args[0] == "reload" {
			t.Fatalf("reload ran after validation failure: %#v", call)
		}
	}
	entries, err := os.ReadDir(backupRoot)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("backup created before successful validation: %d entries", len(entries))
	}
}

func TestOperatorApplyPostflightPIDChangeRestoresExactOldBytes(t *testing.T) {
	op, caddyfile, _, runner := newTestOperator(t)
	runner.mainPIDs = []string{"1234\n", "9999\n", "1234\n"}
	oldBytes := liveLikeCaddy("legacy.user", "legacy password 123")
	mustWriteFile(t, caddyfile, oldBytes, 0640)

	_, err := op.Apply(context.Background(), ApplyRequest{
		ExpectedCaddySHA256: shaHex(oldBytes),
		Desired:             desiredInput(t, "rev-2", "new.user", "new safe password 123"),
	})
	if err == nil {
		t.Fatal("Apply() unexpectedly accepted MainPID change")
	}
	if !bytes.Equal(mustReadFile(t, caddyfile), oldBytes) {
		t.Fatal("failed postflight did not restore exact old Caddyfile bytes")
	}
	calls := runner.snapshotCalls()
	if countCommand(calls, "systemctl", "reload caddy-naive.service") < 2 {
		t.Fatalf("expected apply reload plus compensation reload, calls=%#v", calls)
	}
	for _, call := range calls {
		if call.name == "systemctl" && len(call.args) > 0 && call.args[0] == "restart" {
			t.Fatalf("compensation invoked forbidden restart: %#v", call)
		}
	}
}

func TestOperatorRollbackRejectsTraversalAndRestoresAgentBackup(t *testing.T) {
	op, caddyfile, backupRoot, runner := newTestOperator(t)
	current := liveLikeCaddy("current.user", "current password 123")
	old := liveLikeCaddy("old.user", "old safe password 123")
	mustWriteFile(t, caddyfile, current, 0640)

	if _, err := op.Rollback(context.Background(), RollbackRequest{BackupID: "../outside"}); err == nil {
		t.Fatal("Rollback() accepted traversal backup id")
	}
	if got := runner.snapshotCalls(); len(got) != 0 {
		t.Fatalf("commands ran for traversal id: %#v", got)
	}

	backupID := "backup-20260828T120000Z-a1b2c3d4"
	mustWriteFile(t, filepath.Join(backupRoot, backupID, "Caddyfile"), old, 0600)
	response, err := op.Rollback(context.Background(), RollbackRequest{BackupID: backupID})
	if err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	if !bytes.Equal(mustReadFile(t, caddyfile), old) {
		t.Fatal("Rollback() did not restore backup bytes")
	}
	if response.RestoredSHA256 != shaHex(old) || response.MainPID != 1234 || response.NRestarts != 0 {
		t.Fatalf("Rollback() response = %#v", response)
	}
}

func TestOperatorInspectReturnsExactCredentialSecretOnlyOnInternalType(t *testing.T) {
	op, caddyfile, _, _ := newTestOperator(t)
	input := liveLikeCaddy("import.user", "fixture-secret")
	mustWriteFile(t, caddyfile, input, 0640)
	response, err := op.Inspect(context.Background())
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if response.CaddySHA256 != shaHex(input) {
		t.Fatalf("inspect SHA = %q", response.CaddySHA256)
	}
	if len(response.Credentials) != 1 || response.Credentials[0].Username != "import.user" || response.Credentials[0].Password != "fixture-secret" {
		t.Fatalf("inspect credentials = %#v", response.Credentials)
	}
}

func newTestOperator(t *testing.T) (*FixedOperator, string, string, *scriptedRunner) {
	t.Helper()
	root := t.TempDir()
	caddyfile := filepath.Join(root, "etc", "caddy", "Caddyfile")
	backupRoot := filepath.Join(root, "backups")
	runner := &scriptedRunner{}
	op, err := newOperator(operatorConfig{
		caddyfilePath: caddyfile,
		caddyBinary:   "/usr/local/bin/caddy",
		serviceName:   "caddy-naive.service",
		backupRoot:    backupRoot,
	}, runner, func() time.Time {
		return time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	}, bytes.NewReader(bytes.Repeat([]byte{0xa1, 0xb2, 0xc3, 0xd4}, 32)))
	if err != nil {
		t.Fatalf("newOperator() error = %v", err)
	}
	return op, caddyfile, backupRoot, runner
}

func desiredInput(t *testing.T, revision, username, password string) DesiredStateInput {
	t.Helper()
	if err := runtimecred.ValidatePassword(password, true); err != nil {
		t.Fatal(err)
	}
	return DesiredStateInput{
		Revision: revision,
		Credentials: []CredentialInput{{
			ID:       "cred-1",
			Username: username,
			Password: password,
			Status:   runtimecred.CredentialActive,
		}},
	}
}

func liveLikeCaddy(username, password string) []byte {
	return []byte("{\n    order forward_proxy before file_server\n}\n\nexample.test {\n    forward_proxy {\n        basic_auth " + username + " \"" + password + "\"\n        hide_ip\n        hide_via\n        probe_resistance \"probe-safe-value\"\n    }\n    root * /var/www/pvnaive-camouflage\n    file_server\n}\n")
}

func mustWriteFile(t *testing.T, path string, content []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		t.Fatal(err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func shaHex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func commandIndex(calls []commandCall, name, firstArg string) int {
	for i, call := range calls {
		if call.name == name && len(call.args) > 0 && call.args[0] == firstArg {
			return i
		}
	}
	return -1
}

func countCommand(calls []commandCall, name, joinedArgs string) int {
	count := 0
	for _, call := range calls {
		if call.name == name && strings.Join(call.args, " ") == joinedArgs {
			count++
		}
	}
	return count
}

var _ io.Reader = (*bytes.Reader)(nil)
