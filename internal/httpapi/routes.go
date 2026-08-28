package httpapi

type Access string

const (
	Public        Access = "public"
	Authenticated Access = "authenticated"
	Reseller      Access = "reseller"
	Admin         Access = "admin"
	Owner         Access = "owner"
	Auditor       Access = "auditor"
	Operator      Access = "operator"
)

type Route struct {
	Method string
	Path   string
	Name   string
	Access Access
	Ready  bool
}

var Routes = []Route{
	{"GET", "/api/v1/health/live", "health.live", Public, true},
	{"GET", "/api/v1/health/ready", "health.ready", Public, true},
	{"POST", "/api/v1/auth/login", "auth.login", Public, false},
	{"POST", "/api/v1/auth/logout", "auth.logout", Authenticated, false},
	{"POST", "/api/v1/auth/refresh", "auth.refresh", Public, false},
	{"GET", "/api/v1/subscriptions/{token}", "subscriptions.show", Public, true},
	{"GET", "/api/v1/subscriptions/{token}/info", "subscriptions.info", Public, false},
	{"GET", "/api/v1/subscriptions/{token}/usage", "subscriptions.usage", Public, false},
	{"GET", "/api/v1/me", "me.show", Authenticated, false},
	{"GET", "/api/v1/me/notifications", "me.notifications.index", Authenticated, false},
	{"PATCH", "/api/v1/me/notification-preferences", "me.notifications.preferences", Authenticated, false},
	{"GET", "/api/v1/me/sessions", "me.sessions.index", Authenticated, false},
	{"DELETE", "/api/v1/me/sessions/{id}", "me.sessions.delete", Authenticated, false},
	{"POST", "/api/v1/me/mfa/totp/enroll", "me.mfa.enroll", Authenticated, false},
	{"POST", "/api/v1/me/mfa/totp/confirm", "me.mfa.confirm", Authenticated, false},
	{"DELETE", "/api/v1/me/mfa/totp", "me.mfa.delete", Authenticated, false},
	{"GET", "/api/v1/customers", "customers.index", Owner, true},
	{"POST", "/api/v1/customers", "customers.create", Owner, true},
	{"POST", "/api/v1/customers/{id}/subscription/rotate", "customers.subscription.rotate", Owner, true},
	{"GET", "/api/v1/users", "users.index", Reseller, false},
	{"POST", "/api/v1/users", "users.create", Reseller, false},
	{"GET", "/api/v1/users/{id}", "users.show", Reseller, false},
	{"PATCH", "/api/v1/users/{id}", "users.update", Reseller, false},
	{"POST", "/api/v1/users/{id}/suspend", "users.suspend", Reseller, false},
	{"POST", "/api/v1/users/{id}/resume", "users.resume", Reseller, false},
	{"POST", "/api/v1/users/{id}/revoke", "users.revoke", Reseller, false},
	{"POST", "/api/v1/users/{id}/renew", "users.renew", Reseller, false},
	{"POST", "/api/v1/users/{id}/reset-usage", "users.usage.reset", Reseller, false},
	{"GET", "/api/v1/users/{id}/credentials", "credentials.index", Reseller, false},
	{"POST", "/api/v1/users/{id}/credentials", "credentials.create", Reseller, false},
	{"POST", "/api/v1/users/{id}/credentials/{credentialId}/rotate", "credentials.rotate", Reseller, false},
	{"DELETE", "/api/v1/users/{id}/credentials/{credentialId}", "credentials.delete", Reseller, false},
	{"GET", "/api/v1/users/{id}/sessions", "users.sessions.index", Reseller, false},
	{"DELETE", "/api/v1/users/{id}/sessions/{sessionId}", "users.sessions.delete", Reseller, false},
	{"GET", "/api/v1/resellers", "resellers.index", Admin, false},
	{"POST", "/api/v1/resellers", "resellers.create", Admin, false},
	{"GET", "/api/v1/resellers/{id}", "resellers.show", Admin, false},
	{"PATCH", "/api/v1/resellers/{id}", "resellers.update", Admin, false},
	{"POST", "/api/v1/resellers/{id}/credit-adjustments", "resellers.credit.adjust", Owner, false},
	{"GET", "/api/v1/resellers/{id}/ledger", "resellers.ledger", Auditor, false},
	{"GET", "/api/v1/plans", "plans.index", Reseller, false},
	{"POST", "/api/v1/plans", "plans.create", Admin, false},
	{"GET", "/api/v1/notifications", "notifications.index", Admin, false},
	{"GET", "/api/v1/notification-rules", "notification-rules.index", Admin, false},
	{"POST", "/api/v1/notification-rules", "notification-rules.create", Admin, false},
	{"PATCH", "/api/v1/notification-rules/{id}", "notification-rules.update", Admin, false},
	{"GET", "/api/v1/runtime/status", "runtime.status", Operator, false},
	{"GET", "/api/v1/runtime/naive", "runtime.naive.show", Owner, true},
	{"GET", "/api/v1/runtime/naive/credentials", "runtime.naive.credentials.index", Owner, true},
	{"POST", "/api/v1/runtime/naive/import", "runtime.naive.import", Owner, true},
	{"POST", "/api/v1/runtime/naive/credentials", "runtime.naive.credentials.create", Owner, true},
	{"PATCH", "/api/v1/runtime/naive/credentials/{id}", "runtime.naive.credentials.update", Owner, true},
	{"POST", "/api/v1/runtime/naive/credentials/{id}/rotate-password", "runtime.naive.credentials.rotate", Owner, true},
	{"DELETE", "/api/v1/runtime/naive/credentials/{id}", "runtime.naive.credentials.revoke", Owner, true},
	{"GET", "/api/v1/runtime/revisions", "runtime.revisions.index", Admin, false},
	{"POST", "/api/v1/runtime/revisions/validate", "runtime.revisions.validate", Admin, false},
	{"POST", "/api/v1/runtime/revisions/{id}/apply", "runtime.revisions.apply", Owner, false},
	{"POST", "/api/v1/runtime/revisions/{id}/rollback", "runtime.revisions.rollback", Owner, false},
	{"GET", "/api/v1/usage/summary", "usage.summary", Auditor, false},
	{"GET", "/api/v1/usage/users/{id}", "usage.user", Auditor, false},
	{"GET", "/api/v1/usage/reconciliation", "usage.reconciliation", Auditor, false},
	{"GET", "/api/v1/system/status", "system.status", Operator, false},
	{"GET", "/api/v1/audit-events", "audit.index", Auditor, false},
	{"GET", "/api/v1/logs/application", "logs.application", Operator, false},
	{"GET", "/api/v1/logs/runtime", "logs.runtime", Operator, false},
	{"GET", "/api/v1/logs/security", "logs.security", Auditor, false},
	{"GET", "/api/v1/diagnostics/requests/{requestId}", "diagnostics.request", Operator, false},
	{"GET", "/api/v1/diagnostics/domain-activity", "diagnostics.domains", Owner, false},
	{"POST", "/api/v1/diagnostics/domain-activity/enable", "diagnostics.domains.enable", Owner, false},
	{"POST", "/api/v1/diagnostics/domain-activity/disable", "diagnostics.domains.disable", Owner, false},
	{"DELETE", "/api/v1/diagnostics/domain-activity", "diagnostics.domains.purge", Owner, false},
	{"POST", "/api/v1/diagnostics/bundles", "diagnostics.bundles.create", Owner, false},
	{"POST", "/api/v1/backups", "backups.create", Owner, false},
	{"GET", "/api/v1/backups", "backups.index", Owner, false},
	{"POST", "/api/v1/backups/{id}/verify", "backups.verify", Owner, false},
}
