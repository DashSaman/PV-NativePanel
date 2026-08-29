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

type CustomerView struct {
	UserID                  string          `json:"id"`
	Username                string          `json:"username"`
	Status                  UserAdminState  `json:"status"`
	ServiceTermID           string          `json:"service_term_id"`
	ServiceState            TermState       `json:"service_state"`
	QuotaBytes              *int64          `json:"quota_bytes"`
	DurationSeconds         int64           `json:"duration_seconds"`
	StartPolicy             StartPolicy     `json:"start_policy"`
	StartsAt                *time.Time      `json:"starts_at,omitempty"`
	FirstConnectedAt        *time.Time      `json:"first_connected_at,omitempty"`
	ExpiresAt               *time.Time      `json:"expires_at,omitempty"`
	RuntimeCredentialID     string          `json:"runtime_credential_id"`
	SubscriptionAvailable   bool            `json:"subscription_available"`
	SubscriptionRetrievable bool            `json:"subscription_retrievable"`
	UsageCapability         UsageCapability `json:"usage_capability"`
}

type SubscriptionTarget struct {
	TenantID            string
	UserID              string
	ServiceTermID       string
	RuntimeCredentialID string
	ExpiresAt           *time.Time
}

type EncryptedSubscriptionToken struct {
	UserID          string
	Ciphertext      []byte
	Nonce           []byte
	EncryptionKeyID string
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

type CreateSubscriptionTokenRecord struct {
	TenantID            string
	UserID              string
	ServiceTermID       string
	RuntimeCredentialID string
	TokenHash           []byte
	TokenPrefix         string
	TokenCiphertext     []byte
	TokenNonce          []byte
	TokenEncryptionKeyID string
	ExpiresAt           *time.Time
}
