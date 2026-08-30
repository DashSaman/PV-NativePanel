package httpapi

import "testing"

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
