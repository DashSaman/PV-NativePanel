package sessioncontrol

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/sessionkill"
)

func TestHandlerKillsOnlyExactTuple(t *testing.T) {
	registry := sessionkill.New()
	target := testKey()
	sibling := target
	sibling.SessionID = "44444444-4444-4444-8444-444444444444"
	var targetKills, siblingKills atomic.Int32
	registry.Register(target, func(sessionkill.Key) { targetKills.Add(1) })
	registry.Register(sibling, func(sessionkill.Key) { siblingKills.Add(1) })

	body, err := json.Marshal(KillRequest{RuntimeCredentialID: target.RuntimeCredentialID, NodeID: target.NodeID, BootID: target.BootID, SessionID: target.SessionID})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/kill", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	NewHandler(registry).ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	var got KillResult
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Found || !got.Killed {
		t.Fatalf("result=%+v", got)
	}
	if targetKills.Load() != 1 || siblingKills.Load() != 0 {
		t.Fatalf("kill counts target=%d sibling=%d", targetKills.Load(), siblingKills.Load())
	}
	if registry.IsLive(target) || !registry.IsLive(sibling) {
		t.Fatalf("live target=%v sibling=%v", registry.IsLive(target), registry.IsLive(sibling))
	}
}

func TestHandlerRejectsPartialTupleWithoutTouchingLiveSession(t *testing.T) {
	registry := sessionkill.New()
	target := testKey()
	var kills atomic.Int32
	registry.Register(target, func(sessionkill.Key) { kills.Add(1) })

	body, err := json.Marshal(KillRequest{RuntimeCredentialID: target.RuntimeCredentialID, NodeID: target.NodeID, BootID: target.BootID})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/kill", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	NewHandler(registry).ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	if kills.Load() != 0 || !registry.IsLive(target) {
		t.Fatalf("partial tuple touched live session: kills=%d live=%v", kills.Load(), registry.IsLive(target))
	}
}
