package ops

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

type CheckFunc struct {
	Name string
	Func func(context.Context) CheckResult
}

func (c CheckFunc) Run(ctx context.Context) CheckResult {
	if c.Func == nil {
		return CheckResult{Name: c.Name, Status: Fail, Detail: "check is not configured"}
	}
	result := c.Func(ctx)
	if result.Name == "" {
		result.Name = c.Name
	}
	return result
}

func RequiredEnvironmentCheck(names []string, getenv func(string) string) Check {
	return CheckFunc{Name: "required-env", Func: func(context.Context) CheckResult {
		var missing []string
		for _, name := range names {
			if strings.TrimSpace(getenv(name)) == "" {
				missing = append(missing, name)
			}
		}
		if len(missing) == 0 {
			return CheckResult{Name: "required-env", Status: Pass, Detail: "required variable names are present"}
		}
		sort.Strings(missing)
		return CheckResult{Name: "required-env", Status: Warn, Detail: "not present in current process: " + strings.Join(missing, ",")}
	}}
}

func ServiceCheck(unit string) Check {
	return CheckFunc{Name: unit, Func: func(ctx context.Context) CheckResult {
		command := exec.CommandContext(ctx, "systemctl", "is-active", unit)
		output, err := command.Output()
		if err != nil || strings.TrimSpace(string(output)) != "active" {
			return CheckResult{Name: unit, Status: Fail, Detail: "not active"}
		}
		return CheckResult{Name: unit, Status: Pass, Detail: "active"}
	}}
}

func CommandCheck(name string, command string, args ...string) Check {
	return CheckFunc{Name: name, Func: func(ctx context.Context) CheckResult {
		cmd := exec.CommandContext(ctx, command, args...)
		output, err := cmd.CombinedOutput()
		detail := strings.TrimSpace(string(output))
		if len(detail) > 300 {
			detail = detail[:300]
		}
		if err != nil {
			if detail == "" {
				detail = "command failed"
			}
			return CheckResult{Name: name, Status: Fail, Detail: detail}
		}
		if detail == "" {
			detail = "ok"
		}
		return CheckResult{Name: name, Status: Pass, Detail: detail}
	}}
}

func UnixSocketCheck(name, path string) Check {
	return CheckFunc{Name: name, Func: func(context.Context) CheckResult {
		info, err := os.Stat(path)
		if err != nil {
			return CheckResult{Name: name, Status: Fail, Detail: "socket missing"}
		}
		if info.Mode()&os.ModeSocket == 0 {
			return CheckResult{Name: name, Status: Fail, Detail: "path is not a Unix socket"}
		}
		if info.Mode().Perm()&0o007 != 0 {
			return CheckResult{Name: name, Status: Warn, Detail: fmt.Sprintf("socket mode %04o is accessible to other users", info.Mode().Perm())}
		}
		return CheckResult{Name: name, Status: Pass, Detail: fmt.Sprintf("socket mode %04o", info.Mode().Perm())}
	}}
}

func UnixHTTPCheck(name, socketPath, requestPath string) Check {
	return CheckFunc{Name: name, Func: func(ctx context.Context) CheckResult {
		transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		}}
		defer transport.CloseIdleConnections()
		client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix"+requestPath, nil)
		if err != nil {
			return CheckResult{Name: name, Status: Fail, Detail: "invalid health request"}
		}
		res, err := client.Do(req)
		if err != nil {
			return CheckResult{Name: name, Status: Fail, Detail: "health socket unreachable"}
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return CheckResult{Name: name, Status: Fail, Detail: fmt.Sprintf("HTTP %d", res.StatusCode)}
		}
		return CheckResult{Name: name, Status: Pass, Detail: fmt.Sprintf("HTTP %d", res.StatusCode)}
	}}
}

func DiskCheck(path string, warnPercent, failPercent float64) Check {
	return CheckFunc{Name: "disk", Func: func(context.Context) CheckResult {
		var stat unix.Statfs_t
		if err := unix.Statfs(path, &stat); err != nil {
			return CheckResult{Name: "disk", Status: Fail, Detail: "filesystem statistics unavailable"}
		}
		if stat.Blocks == 0 {
			return CheckResult{Name: "disk", Status: Fail, Detail: "filesystem size is zero"}
		}
		used := 100 * (1 - float64(stat.Bavail)/float64(stat.Blocks))
		status := Pass
		if used >= failPercent {
			status = Fail
		} else if used >= warnPercent {
			status = Warn
		}
		return CheckResult{Name: "disk", Status: status, Detail: fmt.Sprintf("%.1f%% used", used)}
	}}
}

func BackupCheck(root string, maxAge time.Duration) Check {
	return CheckFunc{Name: "backup", Func: func(context.Context) CheckResult {
		entries, err := os.ReadDir(root)
		if err != nil {
			return CheckResult{Name: "backup", Status: Fail, Detail: "backup directory unavailable"}
		}
		var newest time.Time
		var encrypted bool
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			info, infoErr := entry.Info()
			if infoErr != nil {
				continue
			}
			if info.ModTime().After(newest) {
				newest = info.ModTime()
				_, dumpErr := os.Stat(filepath.Join(root, entry.Name(), "pvnaive.dump.age"))
				_, sumErr := os.Stat(filepath.Join(root, entry.Name(), "SHA256SUMS"))
				encrypted = dumpErr == nil && sumErr == nil
			}
		}
		if newest.IsZero() {
			return CheckResult{Name: "backup", Status: Fail, Detail: "no completed backup found"}
		}
		if !encrypted {
			return CheckResult{Name: "backup", Status: Fail, Detail: "latest backup lacks encrypted dump or checksums"}
		}
		age := time.Since(newest)
		if age > maxAge {
			return CheckResult{Name: "backup", Status: Warn, Detail: fmt.Sprintf("latest backup age %s", age.Round(time.Minute))}
		}
		return CheckResult{Name: "backup", Status: Pass, Detail: fmt.Sprintf("latest backup age %s", age.Round(time.Minute))}
	}}
}

func HTTPCheck(name, url string) Check {
	return CheckFunc{Name: name, Func: func(ctx context.Context) CheckResult {
		client := &http.Client{Timeout: 3 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return CheckResult{Name: name, Status: Fail, Detail: "invalid health URL"}
		}
		res, err := client.Do(req)
		if err != nil {
			return CheckResult{Name: name, Status: Fail, Detail: "health endpoint unreachable"}
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return CheckResult{Name: name, Status: Fail, Detail: fmt.Sprintf("HTTP %d", res.StatusCode)}
		}
		return CheckResult{Name: name, Status: Pass, Detail: fmt.Sprintf("HTTP %d", res.StatusCode)}
	}}
}

func PortCheck(name, address string, expectListening bool) Check {
	return CheckFunc{Name: name, Func: func(ctx context.Context) CheckResult {
		dialer := &net.Dialer{Timeout: time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = conn.Close()
		}
		listening := err == nil
		if listening == expectListening {
			return CheckResult{Name: name, Status: Pass, Detail: "expected listen state"}
		}
		if expectListening {
			return CheckResult{Name: name, Status: Fail, Detail: "expected listener is unavailable"}
		}
		return CheckResult{Name: name, Status: Fail, Detail: "unexpected listener detected"}
	}}
}

func FileModeCheck(name, path string, maxPerm os.FileMode) Check {
	return CheckFunc{Name: name, Func: func(context.Context) CheckResult {
		info, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			return CheckResult{Name: name, Status: Fail, Detail: "file missing"}
		}
		if err != nil {
			return CheckResult{Name: name, Status: Fail, Detail: "file metadata unavailable"}
		}
		if info.Mode().Perm()&^maxPerm != 0 {
			return CheckResult{Name: name, Status: Fail, Detail: fmt.Sprintf("unsafe mode %04o", info.Mode().Perm())}
		}
		return CheckResult{Name: name, Status: Pass, Detail: fmt.Sprintf("mode %04o", info.Mode().Perm())}
	}}
}
