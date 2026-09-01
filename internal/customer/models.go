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
	ID                 string             `json:"id"`
	TenantID           string             `json:"-"`
	UserID             string             `json:"user_id"`
	PlanID             string             `json:"plan_id,omitempty"`
	QuotaBytes         *int64             `json:"quota_bytes"`
	DurationSeconds    int64              `json:"duration_seconds"`
	NoExpiry           bool               `json:"no_expiry"`
	ConcurrencyLimit   *int               `json:"concurrency_limit"`
	UniqueIPLimit      *int               `json:"unique_ip_limit"`
	StartPolicy        StartPolicy        `json:"start_policy"`
	PurchasedAt        time.Time          `json:"purchased_at"`
	StartsAt           *time.Time         `json:"starts_at,omitempty"`
	FirstConnectedAt   *time.Time         `json:"first_connected_at,omitempty"`
	ExpiresAt          *time.Time         `json:"expires_at,omitempty"`
	State              TermState          `json:"state"`
	RenewalKind        string             `json:"renewal_kind,omitempty"`
	RenewedFromTermID  string             `json:"renewed_from_term_id,omitempty"`
	AccountingBaseline AccountingBaseline `json:"accounting_baseline"`
	Revision           int64              `json:"revision"`
}

type CustomerGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

type CustomerTag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

type CustomerUsage struct {
	Available            bool               `json:"available"`
	AccountingComplete   bool               `json:"accounting_complete"`
	Baseline             AccountingBaseline `json:"baseline"`
	DirectUploadBytes    int64              `json:"direct_upload_bytes"`
	DirectDownloadBytes  int64              `json:"direct_download_bytes"`
	DirectUsedBytes      int64              `json:"direct_used_bytes"`
	UploadBytes          *int64             `json:"upload_bytes"`
	DownloadBytes        *int64             `json:"download_bytes"`
	UsedBytes            *int64             `json:"used_bytes"`
	RemainingBytes       *int64             `json:"remaining_bytes"`
	LastOnline           *time.Time         `json:"last_online,omitempty"`
	Online               bool               `json:"online"`
	SessionCount         int                `json:"session_count"`
	UsagePeriodStartedAt *time.Time         `json:"usage_period_started_at,omitempty"`
	UsageResetApplied    bool               `json:"usage_reset_applied"`
}

type CustomerView struct {
	UserID                  string             `json:"id"`
	Username                string             `json:"username"`
	DisplayName             string             `json:"display_name,omitempty"`
	Status                  UserAdminState     `json:"status"`
	StatusDimensions        StatusDimensions   `json:"status_dimensions"`
	ServiceTermID           string             `json:"service_term_id"`
	ServiceState            TermState          `json:"service_state"`
	PlanID                  string             `json:"plan_id,omitempty"`
	PlanName                string             `json:"plan_name,omitempty"`
	QuotaBytes              *int64             `json:"quota_bytes"`
	DurationSeconds         int64              `json:"duration_seconds"`
	NoExpiry                bool               `json:"no_expiry"`
	StartPolicy             StartPolicy        `json:"start_policy"`
	StartsAt                *time.Time         `json:"starts_at,omitempty"`
	FirstConnectedAt        *time.Time         `json:"first_connected_at,omitempty"`
	ExpiresAt               *time.Time         `json:"expires_at,omitempty"`
	RuntimeCredentialID     string             `json:"runtime_credential_id"`
	SubscriptionAvailable   bool               `json:"subscription_available"`
	SubscriptionRetrievable bool               `json:"subscription_retrievable"`
	AccountingBaseline      AccountingBaseline `json:"accounting_baseline"`
	UsageCapability         UsageCapability    `json:"usage_capability"`
	Usage                   *CustomerUsage     `json:"usage,omitempty"`
	Note                    string             `json:"note,omitempty"`
	Group                   *CustomerGroup     `json:"group,omitempty"`
	Tags                    []CustomerTag      `json:"tags,omitempty"`
	AssignedActorID         string             `json:"assigned_actor_id,omitempty"`
	CreatedByActorID        string             `json:"created_by_actor_id,omitempty"`
	ResellerID              string             `json:"reseller_id,omitempty"`
	CreatedAt               time.Time          `json:"created_at,omitempty"`
	UpdatedAt               time.Time          `json:"updated_at,omitempty"`
	LastRenewalAt           *time.Time         `json:"last_renewal_at,omitempty"`
	OnHold                  bool               `json:"on_hold"`
	NextPlanID              string             `json:"next_plan_id,omitempty"`
	NextPlanName            string             `json:"next_plan_name,omitempty"`
}

type CustomerPage struct {
	Customers []CustomerView `json:"customers"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
	Total     int64          `json:"total"`
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
	TenantID           string
	UserID             string
	PlanID             string
	QuotaBytes         *int64
	DurationSeconds    int64
	NoExpiry           bool
	ConcurrencyLimit   *int
	UniqueIPLimit      *int
	StartPolicy        StartPolicy
	PurchasedAt        time.Time
	StartsAt           *time.Time
	ExpiresAt          *time.Time
	State              TermState
	RenewalKind        string
	RenewedFromTermID  string
	AccountingBaseline AccountingBaseline
}

type CreateSubscriptionTokenRecord struct {
	TenantID             string
	UserID               string
	ServiceTermID        string
	RuntimeCredentialID  string
	TokenHash            []byte
	TokenPrefix          string
	TokenCiphertext      []byte
	TokenNonce           []byte
	TokenEncryptionKeyID string
	ExpiresAt            *time.Time
}
