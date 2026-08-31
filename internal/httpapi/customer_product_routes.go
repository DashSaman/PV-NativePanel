package httpapi

func init() {
	ready := map[string]bool{
		"users.index":          true,
		"users.create":         true,
		"users.update":         true,
		"users.suspend":        true,
		"users.resume":         true,
		"users.revoke":         true,
		"users.renew":          true,
		"users.sessions.index": true,
		"plans.index":          true,
		"plans.create":         true,
	}
	for i := range Routes {
		if ready[Routes[i].Name] {
			Routes[i].Ready = true
		}
	}
	Routes = append(Routes,
		Route{"GET", "/api/v1/users/{id}/subscription", "users.subscription.current", Reseller, true},
		Route{"POST", "/api/v1/users/{id}/subscription/rotate", "users.subscription.rotate", Reseller, true},
		Route{"POST", "/api/v1/users/{id}/rotate-password", "users.password.rotate", Reseller, true},
		Route{"PATCH", "/api/v1/users/{id}/service", "users.service.update", Reseller, true},
		Route{"POST", "/api/v1/users/{id}/volume/add", "users.volume.add", Reseller, true},
		Route{"POST", "/api/v1/users/{id}/validity/extend", "users.validity.extend", Reseller, true},
		Route{"GET", "/api/v1/customer-groups", "customer-groups.index", Reseller, true},
		Route{"POST", "/api/v1/customer-groups", "customer-groups.create", Reseller, true},
		Route{"GET", "/api/v1/customer-tags", "customer-tags.index", Reseller, true},
		Route{"POST", "/api/v1/customer-tags", "customer-tags.create", Reseller, true},
		Route{"POST", "/api/v1/users/bulk/preview", "users.bulk.preview", Reseller, true},
		Route{"POST", "/api/v1/users/bulk/execute", "users.bulk.execute", Reseller, true},
	)
}
