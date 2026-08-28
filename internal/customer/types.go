package customer

import "time"

type UserAdminState string

const (
	UserDraft     UserAdminState = "draft"
	UserActive    UserAdminState = "active"
	UserSuspended UserAdminState = "suspended"
	UserRevoked   UserAdminState = "revoked"
)

type TermState string

const (
	TermPending       TermState = "pending"
	TermActive        TermState = "active"
	TermExpired       TermState = "expired"
	TermQuotaDepleted TermState = "quota_depleted"
	TermEnded         TermState = "ended"
	TermRevoked       TermState = "revoked"
)

type StartPolicy string

const (
	StartOnCreation                  StartPolicy = "on_creation"
	StartOnFirstSuccessfulConnection StartPolicy = "on_first_successful_connection"
	StartAtFixedTimestamp            StartPolicy = "fixed_timestamp"
)

type AccessState string

const (
	AccessDraft         AccessState = "draft"
	AccessPending       AccessState = "pending"
	AccessActive        AccessState = "active"
	AccessSuspended     AccessState = "suspended"
	AccessExpired       AccessState = "expired"
	AccessQuotaDepleted AccessState = "quota_depleted"
	AccessEnded         AccessState = "ended"
	AccessRevoked       AccessState = "revoked"
)

type UsageCapability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type EffectiveAccessInput struct {
	AdminState UserAdminState
	TermState  TermState
	StartsAt   *time.Time
	ExpiresAt  *time.Time
	Now        time.Time
}

type TermTiming struct {
	State     TermState
	StartsAt  *time.Time
	ExpiresAt *time.Time
}
