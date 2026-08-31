package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestCustomerProductRoutePermissions(t *testing.T) {
	want := map[string]Access{
		"users.index":            Reseller,
		"users.create":           Reseller,
		"users.update":           Reseller,
		"users.suspend":          Reseller,
		"users.resume":           Reseller,
		"users.revoke":           Reseller,
		"users.renew":            Reseller,
		"users.sessions.index":   Reseller,
		"plans.index":            Reseller,
		"plans.create":           Admin,
		"customer-groups.index":  Reseller,
		"customer-groups.create": Reseller,
		"customer-tags.index":    Reseller,
		"customer-tags.create":   Reseller,
		"users.bulk.preview":     Reseller,
		"users.bulk.execute":     Reseller,
		"customers.delete":       Owner,
	}
	seen := map[string]bool{}
	for _, route := range Routes {
		access, ok := want[route.Name]
		if !ok {
			continue
		}
		seen[route.Name] = true
		if route.Access != access {
			t.Fatalf("route %s access=%s want=%s", route.Name, route.Access, access)
		}
		if route.Name != "customers.delete" && !route.Ready {
			t.Fatalf("product route %s is not marked ready", route.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("missing product route %s", name)
		}
	}
	if roleAllowed(Admin, "reseller") {
		t.Fatal("reseller may not create plans")
	}
	if roleAllowed(Owner, "admin") || roleAllowed(Owner, "reseller") {
		t.Fatal("admin/reseller may not use owner-only safe delete")
	}
}

func TestCustomerListQueryParsesURLState(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/users?q=ali&status=on_hold&page=2&page_size=25&sort=expiry&dir=asc&unlimited_volume=true&unlimited_expiry=false", nil)
	query, err := parseCustomerListQuery(req)
	if err != nil {
		t.Fatal(err)
	}
	if query.Search != "ali" || query.Status != "on_hold" || query.Page != 2 || query.PageSize != 25 || query.Sort != "expiry" || query.Direction != "asc" {
		t.Fatalf("unexpected parsed query: %#v", query)
	}
	if query.UnlimitedVolume == nil || !*query.UnlimitedVolume {
		t.Fatal("unlimited volume filter was lost")
	}
	if query.UnlimitedExpiry == nil || *query.UnlimitedExpiry {
		t.Fatal("unlimited expiry filter was lost")
	}
}

func TestCustomerListQueryRejectsInvalidBoolean(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/users?unlimited_volume=maybe", nil)
	if _, err := parseCustomerListQuery(req); err == nil {
		t.Fatal("invalid boolean filter accepted")
	}
}
