package steeringstore

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/steering"
)

func validDecision() steering.Decision {
	return steering.Decision{
		UserID:    "3f6d2a5e-91c7-4b0f-8f1a-2b3c4d5e6f70",
		WindowIdx: 42,
		Primary:   "node-a",
		Previous:  "node-b",
		Reason:    steering.ReasonHysteresis,
		Changed:   true,
		Candidates: []steering.Scored{
			{NodeID: "node-a", Score: 1.25, Kind: "eligible"},
			{NodeID: "node-b", Score: 1.00, Kind: "eligible"},
		},
	}
}

func TestValidateDecisionAcceptsActionableReasons(t *testing.T) {
	for _, reason := range []steering.Reason{steering.ReasonInitial, steering.ReasonHysteresis, steering.ReasonKillSwitch} {
		decision := validDecision()
		decision.Reason = reason
		if err := ValidateDecision(decision); err != nil {
			t.Fatalf("reason %q must validate: %v", reason, err)
		}
	}
}

func TestValidateDecisionRejectsNonActionableReasons(t *testing.T) {
	for _, reason := range []steering.Reason{steering.ReasonNoChange, steering.ReasonNoCandidates, ""} {
		decision := validDecision()
		decision.Reason = reason
		if err := ValidateDecision(decision); err == nil {
			t.Fatalf("reason %q must be rejected (scheduler never dispatches it)", reason)
		}
	}
}

func TestValidateDecisionRejectsBadIdentityAndNodes(t *testing.T) {
	bad := validDecision()
	bad.UserID = "not-a-uuid"
	if err := ValidateDecision(bad); err == nil {
		t.Fatal("malformed user id must be rejected")
	}
	empty := validDecision()
	empty.Primary = " "
	if err := ValidateDecision(empty); err == nil {
		t.Fatal("empty primary node must be rejected")
	}
	longNode := validDecision()
	longNode.Primary = strings.Repeat("n", 161)
	if err := ValidateDecision(longNode); err == nil {
		t.Fatal("oversized node id must be rejected")
	}
	whitespaceNode := validDecision()
	whitespaceNode.Primary = "node a"
	if err := ValidateDecision(whitespaceNode); err == nil {
		t.Fatal("node id with whitespace must be rejected")
	}
	negativeWindow := validDecision()
	negativeWindow.WindowIdx = -1
	if err := ValidateDecision(negativeWindow); err == nil {
		t.Fatal("negative window index must be rejected")
	}
}

func TestBuildDetailsIsRedactedScoresOnly(t *testing.T) {
	raw, err := buildDetails(validDecision())
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Candidates []struct {
			Node  string  `json:"node"`
			Score float64 `json:"score"`
			Kind  string  `json:"kind"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Candidates) != 2 || payload.Candidates[0].Node != "node-a" {
		t.Fatalf("details candidates wrong: %+v", payload)
	}
	for _, banned := range []string{"password", "secret", "credential", "token"} {
		if strings.Contains(strings.ToLower(string(raw)), banned) {
			t.Fatalf("details payload must never carry %q", banned)
		}
	}
}

func TestBuildDetailsEmptyAndCap(t *testing.T) {
	empty := validDecision()
	empty.Candidates = nil
	raw, err := buildDetails(empty)
	if err != nil || raw != nil {
		t.Fatalf("no candidates must encode to nil, got %v %v", raw, err)
	}
	many := validDecision()
	many.Candidates = make([]steering.Scored, 0, maxDecisionCandidates+5)
	for i := 0; i < maxDecisionCandidates+5; i++ {
		many.Candidates = append(many.Candidates, steering.Scored{NodeID: "n", Score: float64(i)})
	}
	raw, err = buildDetails(many)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Candidates []json.RawMessage `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Candidates) != maxDecisionCandidates {
		t.Fatalf("details must cap at %d candidates, got %d", maxDecisionCandidates, len(payload.Candidates))
	}
}

func TestApplyDecisionRequiresStore(t *testing.T) {
	var nilStore *Store
	if err := nilStore.ApplyDecision(t.Context(), validDecision()); err == nil {
		t.Fatal("nil store must error")
	}
	if _, err := NewStore(nil); err == nil {
		t.Fatal("nil db must error at construction")
	}
}

func TestReadCurrentValidatesUserID(t *testing.T) {
	store := &Store{}
	if _, err := store.ReadCurrent(t.Context(), "junk"); err == nil {
		t.Fatal("malformed user id must fail before touching the db")
	}
}
