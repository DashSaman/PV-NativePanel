package runtimeagent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// caddyAdminRunner accepts the validate + admin-API reload command sequence
// and refuses every systemctl call, exactly like the all-in-one container
// where no PID 1 systemd exists.
type caddyAdminRunner struct {
	calls      []string
	validateOK bool
	reloadPlan []bool // one entry per reload call; empty means always succeed
}

func (r *caddyAdminRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	r.calls = append(r.calls, name+" "+joined)
	switch {
	case name == "/opt/pvnaive/bin/caddy" && strings.HasPrefix(joined, "validate "):
		if r.validateOK {
			return []byte("valid\n"), nil
		}
		return nil, fmt.Errorf("validate failed")
	case name == "/opt/pvnaive/bin/caddy" && strings.HasPrefix(joined, "reload "):
		if len(r.reloadPlan) > 0 {
			ok := r.reloadPlan[0]
			r.reloadPlan = r.reloadPlan[1:]
			if !ok {
				return nil, fmt.Errorf("reload failed")
			}
			return nil, nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected command in caddy-admin mode: %s %s", name, joined)
	}
}

func newCaddyAdminTestOperator(t *testing.T) (*FixedOperator, string, *caddyAdminRunner) {
	t.Helper()
	root := t.TempDir()
	caddyfile := filepath.Join(root, "etc", "caddy", "Caddyfile")
	backupRoot := filepath.Join(root, "backups")
	runner := &caddyAdminRunner{validateOK: true}
	op, err := newOperator(operatorConfig{
		caddyfilePath: caddyfile,
		caddyBinary:   "/opt/pvnaive/bin/caddy",
		serviceName:   "caddy-naive.service",
		backupRoot:    backupRoot,
		reloadMode:    ReloadModeCaddyAdmin,
	}, runner, func() time.Time {
		return time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	}, bytes.NewReader(bytes.Repeat([]byte{0xa1, 0xb2, 0xc3, 0xd4}, 32)))
	if err != nil {
		t.Fatalf("newOperator(caddy-admin): %v", err)
	}
	return op, caddyfile, runner
}

func TestOperatorCaddyAdminModeApplyReloadsWithoutSystemd(t *testing.T) {
	op, caddyfile, runner := newCaddyAdminTestOperator(t)
	oldBytes := liveLikeCaddy("legacy.user", "legacy password 123")
	mustWriteFile(t, caddyfile, oldBytes, 0640)

	response, err := op.Apply(context.Background(), ApplyRequest{
		ExpectedCaddySHA256: shaHex(oldBytes),
		Desired:             desiredInput(t, "rev-docker-1", "docker.user", "docker safe password 123"),
	})
	if err != nil {
		t.Fatalf("Apply(caddy-admin) error = %v", err)
	}
	newBytes := mustReadFile(t, caddyfile)
	if bytes.Equal(newBytes, oldBytes) {
		t.Fatal("Apply(caddy-admin) did not install candidate bytes")
	}
	if response.AppliedSHA256 != shaHex(newBytes) {
		t.Fatalf("applied SHA mismatch: %q", response.AppliedSHA256)
	}
	if response.MainPID != 1 || response.NRestarts != 0 {
		t.Fatalf("caddy-admin snapshot = pid:%d restarts:%d, want synthetic 1/0", response.MainPID, response.NRestarts)
	}

	validateIndex, reloadIndex := -1, -1
	for i, call := range runner.calls {
		if strings.HasPrefix(call, "/opt/pvnaive/bin/caddy validate") && validateIndex < 0 {
			validateIndex = i
		}
		if strings.Contains(call, " reload --config ") && strings.Contains(call, "--adapter caddyfile") && reloadIndex < 0 {
			reloadIndex = i
		}
		if strings.Contains(call, "systemctl") {
			t.Fatalf("caddy-admin mode invoked systemctl: %s", call)
		}
	}
	if validateIndex < 0 || reloadIndex < 0 || validateIndex > reloadIndex {
		t.Fatalf("expected validate before admin-API reload, calls: %#v", runner.calls)
	}
}

func TestOperatorCaddyAdminModeFailedReloadRollsBackExactBytes(t *testing.T) {
	op, caddyfile, runner := newCaddyAdminTestOperator(t)
	oldBytes := liveLikeCaddy("legacy.user", "legacy password 123")
	mustWriteFile(t, caddyfile, oldBytes, 0640)

	// First reload (after install) fails; the rollback path reloads again (succeeds).
	runner.reloadPlan = []bool{false, true}

	_, err := op.Apply(context.Background(), ApplyRequest{
		ExpectedCaddySHA256: shaHex(oldBytes),
		Desired:             desiredInput(t, "rev-docker-2", "docker.user", "docker safe password 456"),
	})
	if err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("Apply(caddy-admin) error = %v, want rolled-back failure", err)
	}
	restored := mustReadFile(t, caddyfile)
	if !bytes.Equal(restored, oldBytes) {
		t.Fatal("failed caddy-admin reload did not restore exact pre-apply bytes")
	}
}

func TestNewOperatorRejectsUnknownReloadMode(t *testing.T) {
	t.Setenv("PVNAIVE_RUNTIME_RELOAD_MODE", "rc-control")
	if _, err := NewOperator(); err == nil || !strings.Contains(err.Error(), "unsupported PVNAIVE_RUNTIME_RELOAD_MODE") {
		t.Fatalf("NewOperator() error = %v, want unsupported reload mode", err)
	}
}

func caddyAdminSHAHex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
