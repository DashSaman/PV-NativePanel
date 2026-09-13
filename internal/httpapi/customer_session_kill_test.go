package httpapi

import (
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

func TestFindCustomerSessionByIDReturnsTrustedReadModelTuple(t *testing.T) {
	sessions := []customer.CustomerSession{
		{RuntimeCredentialID: "runtime-1", NodeID: "node-1", BootID: "boot-1", SessionID: "session-1"},
		{RuntimeCredentialID: "runtime-2", NodeID: "node-2", BootID: "boot-2", SessionID: "session-2"},
	}
	got, ok := findCustomerSessionByID(sessions, "session-2")
	if !ok {
		t.Fatal("trusted active session was not found")
	}
	if got.RuntimeCredentialID != "runtime-2" || got.NodeID != "node-2" || got.BootID != "boot-2" || got.SessionID != "session-2" {
		t.Fatalf("unexpected trusted tuple: %+v", got)
	}
	if _, ok := findCustomerSessionByID(sessions, "forged-session"); ok {
		t.Fatal("forged/unowned session id matched trusted active session")
	}
}

func TestSessionKillRouteIsReadyAndResellerScoped(t *testing.T) {
	for _, route := range Routes {
		if route.Name != "users.sessions.delete" {
			continue
		}
		if route.Access != Reseller {
			t.Fatalf("session kill access=%s want=%s", route.Access, Reseller)
		}
		if !route.Ready {
			t.Fatal("session kill route must be ready once API wiring is implemented")
		}
		return
	}
	t.Fatal("session kill route missing")
}
