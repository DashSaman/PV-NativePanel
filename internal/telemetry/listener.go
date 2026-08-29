package telemetry

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func ListenUnix(path string) (net.Listener, error) {
	if path != DefaultTelemetrySocketPath {
		return nil, fmt.Errorf("telemetry: socket path must be %s", DefaultTelemetrySocketPath)
	}
	if filepath.Clean(path) != path || filepath.Dir(path) != "/run/pvnaive" {
		return nil, errors.New("telemetry: socket path escaped fixed runtime directory")
	}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, errors.New("telemetry: refusing to replace non-socket path")
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("telemetry: remove stale socket: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("telemetry: inspect socket: %w", err)
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0660); err != nil {
		_ = listener.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("telemetry: chmod socket: %w", err)
	}
	return listener, nil
}
