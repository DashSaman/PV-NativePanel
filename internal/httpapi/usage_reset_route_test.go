package httpapi

import (
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/customer"
)

func TestManualUsageResetRoutesAreOwnerReady(t *testing.T) {
	want := map[string]string{
		"users.usage.reset":     "/api/v1/users/{id}/reset-usage",
		"customers.usage.reset": "/api/v1/customers/{id}/reset-usage",
	}
	seen := map[string]bool{}
	for _, route := range Routes {
		path, ok := want[route.Name]
		if !ok {
			continue
		}
		seen[route.Name] = true
		if route.Method != "POST" || route.Path != path || route.Access != Owner || !route.Ready {
			t.Fatalf("route %s = %#v, want POST %s owner ready", route.Name, route, path)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("missing usage reset route %s", name)
		}
	}
}

func TestBulkResetUsagePreservesOwnerOnlyAuthorization(t *testing.T) {
	if !bulkActionAllowed("owner", customer.BulkResetUsage) {
		t.Fatal("owner must be allowed to bulk reset usage")
	}
	for _, role := range []string{"admin", "reseller", "auditor"} {
		if bulkActionAllowed(role, customer.BulkResetUsage) {
			t.Fatalf("role %s unexpectedly allowed to bulk reset usage", role)
		}
	}
	if !bulkActionAllowed("reseller", customer.BulkSuspend) {
		t.Fatal("existing reseller bulk actions must remain unchanged")
	}
}

func TestBulkResetChildIdempotencyKeyIsStableAndActionScoped(t *testing.T) {
	first := bulkItemIdempotencyKey("parent-key-0001", customer.BulkResetUsage, "user-1")
	second := bulkItemIdempotencyKey("parent-key-0001", customer.BulkResetUsage, "user-1")
	other := bulkItemIdempotencyKey("parent-key-0002", customer.BulkResetUsage, "user-1")
	if first != second || first == other || len(first) < 16 {
		t.Fatalf("bulk reset child keys first=%q second=%q other=%q", first, second, other)
	}
}
