package sessioncontrol

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/DashSaman/PV-NaivePanel/internal/sessionkill"
)

type killer interface {
	Kill(sessionkill.Key) (sessionkill.KillResult, error)
}

func NewHandler(k killer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sessions/kill", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		limited := http.MaxBytesReader(w, r.Body, maxRequestBytes)
		defer limited.Close()
		decoder := json.NewDecoder(limited)
		decoder.DisallowUnknownFields()
		var wire KillRequest
		if err := decoder.Decode(&wire); err != nil { http.Error(w, "invalid request", http.StatusBadRequest); return }
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) { http.Error(w, "invalid request", http.StatusBadRequest); return }
		if wire.RuntimeCredentialID == "" || wire.NodeID == "" || wire.BootID == "" || wire.SessionID == "" { http.Error(w, "incomplete session tuple", http.StatusBadRequest); return }
		result, err := k.Kill(sessionkill.Key{RuntimeCredentialID: wire.RuntimeCredentialID, NodeID: wire.NodeID, BootID: wire.BootID, SessionID: wire.SessionID})
		if err != nil { http.Error(w, "kill failed", http.StatusInternalServerError); return }
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(KillResult{Found: result.Found, Killed: result.Killed})
	})
	return mux
}
