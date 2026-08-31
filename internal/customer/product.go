package customer

import (
	"errors"
	"sort"
	"strings"
	"time"
)

const gibibyte int64 = 1073741824

type ResetStrategy string

const PeriodicResetEnforcementAvailable = true

const (
	ResetNone    ResetStrategy = "none"
	ResetDaily   ResetStrategy = "daily"
	ResetWeekly  ResetStrategy = "weekly"
	ResetMonthly ResetStrategy = "monthly"
	ResetYearly  ResetStrategy = "yearly"
	ResetCustom  ResetStrategy = "custom"
)

type PlanPreset struct {
	ID               string        `json:"id,omitempty"`
	Name             string        `json:"name"`
	QuotaBytes       *int64        `json:"quota_bytes"`
	ValiditySeconds  int64         `json:"validity_seconds,omitempty"`
	NoExpiry         bool          `json:"no_expiry"`
	StartPolicy      StartPolicy   `json:"start_policy"`
	ResetStrategy    ResetStrategy `json:"reset_strategy"`
	ResetCustomDays  int           `json:"reset_custom_days,omitempty"`
	ConcurrencyLimit *int          `json:"concurrency_limit"`
	DefaultGroupID   string        `json:"default_group_id,omitempty"`
	TagIDs           []string      `json:"tag_ids,omitempty"`
	Enabled          bool          `json:"enabled"`
	SortOrder        int           `json:"sort_order"`
	ResetEnforcement bool          `json:"reset_enforcement_available"`
}

type ServiceSnapshot struct {
	PlanID           string
	QuotaBytes       *int64
	DurationSeconds  int64
	NoExpiry         bool
	StartPolicy      StartPolicy
	PurchasedAt      time.Time
	ResetStrategy    ResetStrategy
	ResetCustomDays  int
	ConcurrencyLimit *int
}

func (p PlanPreset) Validate() error {
	if strings.TrimSpace(p.Name) == "" || len(strings.TrimSpace(p.Name)) > 120 {
		return errors.New("customer: plan name is required")
	}
	if p.QuotaBytes != nil && *p.QuotaBytes <= 0 {
		return ErrInvalidQuotaGB
	}
	if p.NoExpiry {
		if p.ValiditySeconds != 0 {
			return errors.New("customer: no-expiry plan must not carry validity seconds")
		}
		if p.StartPolicy != StartOnCreation {
			return errors.New("customer: no-expiry plan must start on creation")
		}
	} else if p.ValiditySeconds <= 0 {
		return errors.New("customer: finite plan requires validity")
	}
	switch p.StartPolicy {
	case StartOnCreation, StartOnFirstSuccessfulConnection:
	default:
		return ErrInvalidValidityMode
	}
	if p.ConcurrencyLimit != nil && *p.ConcurrencyLimit <= 0 {
		return errors.New("customer: concurrency limit must be positive or null for Unlimited")
	}
	switch p.ResetStrategy {
	case ResetNone, ResetDaily, ResetWeekly, ResetMonthly, ResetYearly:
		if p.ResetCustomDays != 0 {
			return errors.New("customer: custom reset days only apply to custom reset")
		}
	case ResetCustom:
		if p.ResetCustomDays <= 0 || p.ResetCustomDays > 3660 {
			return errors.New("customer: custom reset days are invalid")
		}
	default:
		return errors.New("customer: reset strategy is invalid")
	}
	return nil
}

func (p PlanPreset) ServiceSnapshot(now time.Time) ServiceSnapshot {
	quota := p.QuotaBytes
	if quota != nil {
		value := *quota
		quota = &value
	}
	var concurrencyLimit *int
	if p.ConcurrencyLimit != nil {
		value := *p.ConcurrencyLimit
		concurrencyLimit = &value
	}
	return ServiceSnapshot{
		PlanID:           p.ID,
		QuotaBytes:       quota,
		DurationSeconds:  p.ValiditySeconds,
		NoExpiry:         p.NoExpiry,
		StartPolicy:      p.StartPolicy,
		PurchasedAt:      now.UTC(),
		ResetStrategy:    p.ResetStrategy,
		ResetCustomDays:  p.ResetCustomDays,
		ConcurrencyLimit: concurrencyLimit,
	}
}

type LifecycleStatus string
type CommercialStatus string
type PresenceStatus string
type QuotaStatus string
type RuntimeStatus string

const (
	LifecycleActive    LifecycleStatus = "active"
	LifecycleDisabled  LifecycleStatus = "disabled"
	LifecycleSuspended LifecycleStatus = "suspended"
	LifecycleRevoked   LifecycleStatus = "revoked"

	CommercialPendingFirstUse CommercialStatus = "pending_first_use"
	CommercialActive          CommercialStatus = "active"
	CommercialExpired         CommercialStatus = "expired"
	CommercialDepleted        CommercialStatus = "depleted"
	CommercialOnHold          CommercialStatus = "on_hold"

	PresenceOnline  PresenceStatus = "online"
	PresenceOffline PresenceStatus = "offline"
	PresenceUnknown PresenceStatus = "unknown"

	QuotaUnlimited   QuotaStatus = "unlimited"
	QuotaHealthy     QuotaStatus = "healthy"
	QuotaWarning     QuotaStatus = "warning"
	QuotaDepleted    QuotaStatus = "depleted"
	QuotaUnavailable QuotaStatus = "unavailable"

	RuntimeHealthy  RuntimeStatus = "healthy"
	RuntimeDegraded RuntimeStatus = "degraded"
	RuntimeDown     RuntimeStatus = "down"
	RuntimeUnknown  RuntimeStatus = "unknown"
)

type StatusDimensions struct {
	Lifecycle  LifecycleStatus  `json:"lifecycle"`
	Commercial CommercialStatus `json:"commercial"`
	Presence   PresenceStatus   `json:"presence"`
	Quota      QuotaStatus      `json:"quota"`
	Runtime    RuntimeStatus    `json:"runtime"`
}

type StatusInput struct {
	UserState              UserAdminState
	TermState              TermState
	StartPolicy            StartPolicy
	QuotaBytes             *int64
	ExpiresAt              *time.Time
	OnHold                 bool
	AccountingAvailable    bool
	UsedBytes              *int64
	Presence               *PresenceStatus
	RuntimeHealthAvailable bool
	RuntimeHealth          RuntimeStatus
	Now                    time.Time
}

func DeriveStatusDimensions(in StatusInput) StatusDimensions {
	out := StatusDimensions{
		Presence: PresenceUnknown,
		Quota:    QuotaUnavailable,
		Runtime:  RuntimeUnknown,
	}
	switch in.UserState {
	case UserRevoked:
		out.Lifecycle = LifecycleRevoked
	case UserSuspended:
		out.Lifecycle = LifecycleSuspended
	case UserDisabled, UserDraft:
		out.Lifecycle = LifecycleDisabled
	default:
		out.Lifecycle = LifecycleActive
	}

	if in.OnHold {
		out.Commercial = CommercialOnHold
	} else {
		now := in.Now
		if now.IsZero() {
			now = time.Now().UTC()
		}
		switch {
		case in.TermState == TermQuotaDepleted:
			out.Commercial = CommercialDepleted
		case in.TermState == TermExpired || (in.ExpiresAt != nil && !in.ExpiresAt.After(now)):
			out.Commercial = CommercialExpired
		case in.TermState == TermPending && in.StartPolicy == StartOnFirstSuccessfulConnection:
			out.Commercial = CommercialPendingFirstUse
		default:
			out.Commercial = CommercialActive
		}
	}

	if in.QuotaBytes == nil {
		out.Quota = QuotaUnlimited
	} else if in.TermState == TermQuotaDepleted {
		out.Quota = QuotaDepleted
	} else if in.AccountingAvailable && in.UsedBytes != nil && *in.QuotaBytes > 0 {
		ratio := float64(*in.UsedBytes) / float64(*in.QuotaBytes)
		switch {
		case ratio >= 1:
			out.Quota = QuotaDepleted
		case ratio >= 0.8:
			out.Quota = QuotaWarning
		default:
			out.Quota = QuotaHealthy
		}
	}
	if in.Presence != nil {
		out.Presence = *in.Presence
	}
	if in.RuntimeHealthAvailable {
		switch in.RuntimeHealth {
		case RuntimeHealthy, RuntimeDegraded, RuntimeDown:
			out.Runtime = in.RuntimeHealth
		}
	}
	return out
}

type CustomerAction string

const (
	ActionViewCustomers          CustomerAction = "view_customers"
	ActionCreateCustomer         CustomerAction = "create_customer"
	ActionEditCustomer           CustomerAction = "edit_customer"
	ActionRenewCustomer          CustomerAction = "renew_customer"
	ActionSuspendCustomer        CustomerAction = "suspend_customer"
	ActionResumeCustomer         CustomerAction = "resume_customer"
	ActionRevokeCustomer         CustomerAction = "revoke_customer"
	ActionDeleteCustomer         CustomerAction = "delete_customer"
	ActionManagePlans            CustomerAction = "manage_plans"
	ActionManageGroupsTags       CustomerAction = "manage_groups_tags"
	ActionSubscriptionOperations CustomerAction = "subscription_operations"
	ActionViewAudit              CustomerAction = "view_audit"
)

func CustomerActionAllowed(role string, action CustomerAction) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "owner" {
		return true
	}
	switch role {
	case "admin":
		switch action {
		case ActionViewCustomers, ActionCreateCustomer, ActionEditCustomer, ActionRenewCustomer,
			ActionSuspendCustomer, ActionResumeCustomer, ActionRevokeCustomer, ActionManagePlans,
			ActionManageGroupsTags, ActionSubscriptionOperations, ActionViewAudit:
			return true
		}
	case "reseller":
		switch action {
		case ActionViewCustomers, ActionCreateCustomer, ActionEditCustomer, ActionRenewCustomer,
			ActionSuspendCustomer, ActionResumeCustomer, ActionRevokeCustomer,
			ActionManageGroupsTags, ActionSubscriptionOperations:
			return true
		}
	case "auditor":
		return action == ActionViewAudit
	}
	return false
}

type CustomerSort string
type SortDirection string

const (
	SortUsername    CustomerSort = "username"
	SortCreated     CustomerSort = "created"
	SortUpdated     CustomerSort = "updated"
	SortExpiry      CustomerSort = "expiry"
	SortLastRenewal CustomerSort = "last_renewal"
	SortUsage       CustomerSort = "usage"
	SortRemaining   CustomerSort = "remaining"
	SortLastOnline  CustomerSort = "last_online"

	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type CustomerListQuery struct {
	Search          string        `json:"q,omitempty"`
	Status          string        `json:"status,omitempty"`
	ExpiryFrom      *time.Time    `json:"expiry_from,omitempty"`
	ExpiryTo        *time.Time    `json:"expiry_to,omitempty"`
	PlanID          string        `json:"plan_id,omitempty"`
	GroupID         string        `json:"group_id,omitempty"`
	TagID           string        `json:"tag_id,omitempty"`
	ResellerID      string        `json:"reseller_id,omitempty"`
	UnlimitedVolume *bool         `json:"unlimited_volume,omitempty"`
	UnlimitedExpiry *bool         `json:"unlimited_expiry,omitempty"`
	Page            int           `json:"page"`
	PageSize        int           `json:"page_size"`
	Sort            CustomerSort  `json:"sort"`
	Direction       SortDirection `json:"direction"`
}

func (q CustomerListQuery) Normalize() CustomerListQuery {
	q.Search = strings.TrimSpace(q.Search)
	if q.Page < 1 {
		q.Page = 1
	}
	switch q.PageSize {
	case 10, 20, 25, 50, 100:
	default:
		q.PageSize = 50
	}
	switch q.Sort {
	case SortUsername, SortCreated, SortUpdated, SortExpiry, SortLastRenewal:
	default:
		q.Sort = SortUpdated
	}
	if q.Direction != SortAscending && q.Direction != SortDescending {
		q.Direction = SortDescending
	}
	return q
}

type BulkAction string

const (
	BulkEnable              BulkAction = "enable"
	BulkSuspend             BulkAction = "suspend"
	BulkRevoke              BulkAction = "revoke"
	BulkSafeDelete          BulkAction = "safe_delete"
	BulkExtendDays          BulkAction = "extend_days"
	BulkAddVolume           BulkAction = "add_volume"
	BulkSetVolume           BulkAction = "set_volume"
	BulkApplyPlan           BulkAction = "apply_plan"
	BulkAssignGroup         BulkAction = "assign_group"
	BulkAddTag              BulkAction = "add_tag"
	BulkRemoveTag           BulkAction = "remove_tag"
	BulkReissueSubscription BulkAction = "reissue_subscription"
	BulkResetUsage          BulkAction = "reset_usage"
)

type BulkCustomer struct {
	ID        string
	Lifecycle LifecycleStatus
}

type BulkItem struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

type BulkPreview struct {
	Requested int        `json:"requested"`
	Affected  int        `json:"affected"`
	Changes   []string   `json:"changes"`
	Conflicts []BulkItem `json:"conflicts"`
	Skipped   []BulkItem `json:"skipped"`
	Invalid   []BulkItem `json:"invalid"`
}

type BulkPreviewInput struct {
	Action                      BulkAction
	RequestedIDs                []string
	Customers                   []BulkCustomer
	RuntimeCoordinatorAvailable bool
}

func BuildBulkPreview(in BulkPreviewInput) BulkPreview {
	preview := BulkPreview{
		Requested: len(in.RequestedIDs),
		Conflicts: []BulkItem{},
		Skipped:   []BulkItem{},
		Invalid:   []BulkItem{},
	}
	known := make(map[string]BulkCustomer, len(in.Customers))
	for _, item := range in.Customers {
		known[item.ID] = item
	}
	runtimeSensitive := in.Action == BulkEnable || in.Action == BulkSuspend || in.Action == BulkRevoke || in.Action == BulkSafeDelete
	for _, id := range uniqueStrings(in.RequestedIDs) {
		item, ok := known[id]
		if !ok {
			preview.Invalid = append(preview.Invalid, BulkItem{ID: id, Reason: "customer_not_found_or_out_of_scope"})
			continue
		}
		if item.Lifecycle == LifecycleRevoked && in.Action != BulkRevoke && in.Action != BulkSafeDelete {
			preview.Skipped = append(preview.Skipped, BulkItem{ID: id, Reason: "customer_revoked"})
			continue
		}
		if runtimeSensitive && !in.RuntimeCoordinatorAvailable {
			preview.Conflicts = append(preview.Conflicts, BulkItem{ID: id, Reason: "runtime_coordinator_unavailable"})
			continue
		}
		preview.Affected++
	}
	preview.Changes = []string{string(in.Action)}
	return preview
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
