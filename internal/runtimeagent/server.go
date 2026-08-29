package runtimeagent

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

const maxCredentialCount = 256

func ListenUnix(socketPath string) (net.Listener, error) {
	if socketPath == "" {
		return nil, errors.New("runtimeagent: socket path is required")
	}
	if info, err := os.Lstat(socketPath); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, errors.New("runtimeagent: refusing to replace non-socket path")
		}
		if err := os.Remove(socketPath); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	if unixListener, ok := listener.(*net.UnixListener); ok {
		unixListener.SetUnlinkOnClose(true)
	}
	return listener, nil
}

func NewHandler(op Operator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", method(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		response, err := op.Health(r.Context())
		if err != nil {
			writeOperationFailure(w)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}))
	mux.HandleFunc("/v1/inspect", method(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		response, err := op.Inspect(r.Context())
		if err != nil {
			writeOperationFailure(w)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}))
	mux.HandleFunc("/v1/validate", method(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		var request ValidateRequest
		if !decodeStrictJSON(w, r, &request) {
			return
		}
		if !validMutationInput(request.ExpectedCaddySHA256, request.Desired) {
			writeClientError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		response, err := op.Validate(r.Context(), request)
		if err != nil {
			writeOperationFailure(w)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}))
	mux.HandleFunc("/v1/apply", method(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		var request ApplyRequest
		if !decodeStrictJSON(w, r, &request) {
			return
		}
		if !validMutationInput(request.ExpectedCaddySHA256, request.Desired) {
			writeClientError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		response, err := op.Apply(r.Context(), request)
		if err != nil {
			writeOperationFailure(w)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}))
	mux.HandleFunc("/v1/rollback", method(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		var request RollbackRequest
		if !decodeStrictJSON(w, r, &request) {
			return
		}
		if !validBackupID(request.BackupID) {
			writeClientError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		response, err := op.Rollback(r.Context(), request)
		if err != nil {
			writeOperationFailure(w)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}))
	return mux
}

func method(want string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != want {
			w.Header().Set("Allow", want)
			writeClientError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		next(w, r)
	}
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeClientError(w, http.StatusRequestEntityTooLarge, "request_too_large")
		} else {
			writeClientError(w, http.StatusBadRequest, "invalid_json")
		}
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeClientError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func validMutationInput(expectedSHA string, desired DesiredStateInput) bool {
	if !validSHA256(expectedSHA) || len(desired.Revision) == 0 || len(desired.Revision) > 160 {
		return false
	}
	if len(desired.Credentials) == 0 || len(desired.Credentials) > maxCredentialCount {
		return false
	}
	seen := make(map[string]struct{}, len(desired.Credentials))
	active := 0
	for _, credential := range desired.Credentials {
		if len(credential.ID) == 0 || len(credential.ID) > 160 {
			return false
		}
		if err := runtimecred.ValidateUsername(credential.Username); err != nil {
			return false
		}
		if err := runtimecred.ValidatePassword(credential.Password, true); err != nil {
			return false
		}
		switch credential.Status {
		case runtimecred.CredentialActive:
			active++
		case runtimecred.CredentialDisabled, runtimecred.CredentialRevoked:
		default:
			return false
		}
		if _, exists := seen[credential.Username]; exists {
			return false
		}
		seen[credential.Username] = struct{}{}
	}
	return active > 0
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}

func validBackupID(value string) bool {
	if len(value) == 0 || len(value) > 160 || strings.Contains(value, "..") {
		return false
	}
	for i := 0; i < len(value); i++ {
		b := value[i]
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' {
			continue
		}
		return false
	}
	return true
}

func writeOperationFailure(w http.ResponseWriter) {
	writeClientError(w, http.StatusInternalServerError, "operation_failed")
}

func writeClientError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
