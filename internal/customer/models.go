package customer

import "time"

type User struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"-"`
	Username    string         `json:"username"`
	DisplayName string         `json:"display_name"`
	Status      UserAdminState `json:"status"`
	Revision    int64          `json:"revision"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty"`
}

type ServiceTerm struct {
	ID               string      `json:"id"`
	TenantID         string      `json:"-"`
	UserID           string      `json:"user_id"`
	QuotaBytes       *int64      `json:"quota_bytes"`
	DurationSeconds  int64       `json:"duration_seconds"`
	StartPolicy      StartPolicy `json:"start_policy"`
	PurchasedAt      time.Time   `json:"purchased_at"`
	StartsAt         *time.Time  `json:"starts_at,omitempty"`
	FirstConnectedAt *time.Time  `json:"first_connected_at,omitempty"`
	ExpiresAt        *time.Time  `json:"expires_at,omitempty"`
	State            TermState   `json:"state"`
	Revision         int64       `json:"revision"`
}

type CreateUserRecord struct {
	TenantID    string
	Username    string
	DisplayName string
	ActorID     string
}

type CreateServiceTermRecord struct {
	TenantID        string
	UserID          string
	QuotaBytes      *int64
	DurationSeconds int64
	StartPolicy     StartPolicy
	PurchasedAt     time.Time
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	State           TermState
}
