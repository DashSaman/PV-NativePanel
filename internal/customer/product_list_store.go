package customer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *PostgresStore) SearchCustomersTx(ctx context.Context, tx *sql.Tx, query CustomerListQuery) (CustomerPage, error) {
	if tx == nil {
		return CustomerPage{}, errors.New("customer: transaction is required")
	}
	query = query.Normalize()
	sortColumn := map[CustomerSort]string{
		SortUsername:    "lower(u.username)",
		SortCreated:     "u.created_at",
		SortUpdated:     "u.updated_at",
		SortExpiry:      "st.expires_at",
		SortLastRenewal: "cp.last_renewal_at",
		SortUsage:       "COALESCE(acct.used_bytes,0)",
		SortRemaining:   "acct.remaining_bytes",
		SortLastOnline:  "acct.last_online",
	}[query.Sort]
	if sortColumn == "" {
		sortColumn = "u.updated_at"
	}
	direction := "DESC"
	if query.Direction == SortAscending {
		direction = "ASC"
	}

	statement := fmt.Sprintf(`
WITH current_terms AS (
    SELECT DISTINCT ON (st.user_id) st.*
    FROM pvnaive.service_terms st
    ORDER BY st.user_id, st.purchased_at DESC, st.created_at DESC
), tag_rollup AS (
    SELECT cta.user_id,
           jsonb_agg(jsonb_build_object(
               'id', ct.id::text,
               'name', ct.name,
               'enabled', ct.enabled,
               'sort_order', ct.sort_order
           ) ORDER BY ct.sort_order, ct.name) AS tags
    FROM pvnaive.customer_tag_assignments cta
    JOIN pvnaive.customer_tags ct ON ct.id = cta.tag_id
    GROUP BY cta.user_id
)
SELECT
    u.id::text,
    u.username,
    u.display_name,
    u.status,
    st.id::text,
    st.state,
    st.quota_bytes,
    st.duration_seconds,
    st.no_expiry,
    st.start_policy,
    st.starts_at,
    st.first_connected_at,
    st.expires_at,
    COALESCE(st.plan_id::text,''),
    COALESCE(p.name,''),
    COALESCE(urc.runtime_credential_id::text,''),
    EXISTS (
        SELECT 1 FROM pvnaive.direct_subscription_tokens dst
        WHERE dst.user_id=u.id AND dst.service_term_id=st.id
          AND dst.status='active' AND dst.revoked_at IS NULL
    ),
    EXISTS (
        SELECT 1 FROM pvnaive.direct_subscription_tokens dst
        WHERE dst.user_id=u.id AND dst.service_term_id=st.id
          AND dst.status='active' AND dst.revoked_at IS NULL
          AND octet_length(dst.token_ciphertext) >= 16
          AND octet_length(dst.token_nonce) = 12
          AND dst.encryption_key_id <> ''
    ),
    COALESCE(acct.upload_bytes,0),
    COALESCE(acct.download_bytes,0),
    COALESCE(acct.used_bytes,0),
    acct.remaining_bytes,
    COALESCE(acct.online,false),
    COALESCE(acct.session_count,0),
    acct.last_online,
    COALESCE(acct.accounting_complete,false),
    (acct.service_term_id IS NOT NULL),
    COALESCE(cp.note,''),
    COALESCE(cg.id::text,''),
    COALESCE(cg.name,''),
    COALESCE(cg.enabled,false),
    COALESCE(cg.sort_order,0),
    COALESCE(tr.tags,'[]'::jsonb),
    COALESCE(cp.assigned_actor_id::text,''),
    COALESCE(u.created_by_actor_id::text,''),
    COALESCE(r.id::text,''),
    u.created_at,
    u.updated_at,
    cp.last_renewal_at,
    COALESCE(cp.on_hold,false),
    COALESCE(cp.next_plan_id::text,''),
    COALESCE(np.name,''),
    count(*) OVER()
FROM pvnaive.users u
JOIN current_terms st ON st.user_id=u.id AND st.tenant_id=u.tenant_id
LEFT JOIN pvnaive.plans p ON p.id=st.plan_id
LEFT JOIN pvnaive.customer_profiles cp ON cp.user_id=u.id AND cp.tenant_id=u.tenant_id
LEFT JOIN pvnaive.customer_groups cg ON cg.id=cp.group_id
LEFT JOIN tag_rollup tr ON tr.user_id=u.id
LEFT JOIN pvnaive.plans np ON np.id=cp.next_plan_id
LEFT JOIN pvnaive.resellers r ON r.tenant_id=u.tenant_id
LEFT JOIN LATERAL (
    SELECT binding.runtime_credential_id
    FROM pvnaive.user_runtime_credentials binding
    WHERE binding.user_id=u.id AND binding.service_term_id=st.id AND binding.unbound_at IS NULL
    ORDER BY binding.bound_at DESC
    LIMIT 1
) urc ON TRUE
LEFT JOIN LATERAL pvnaive.direct_naive_accounting_read(
    st.id, clock_timestamp(), 90
) acct ON TRUE
WHERE
    ($1 = '' OR
        u.username ILIKE '%%' || $1 || '%%' OR
        u.id::text ILIKE '%%' || $1 || '%%' OR
        COALESCE(cp.note,'') ILIKE '%%' || $1 || '%%' OR
        EXISTS (
            SELECT 1 FROM pvnaive.direct_subscription_tokens search_token
            WHERE search_token.user_id=u.id
              AND search_token.token_prefix ILIKE $1 || '%%'
        )
    )
    AND ($2 = '' OR
        ($2='active' AND u.status='active' AND NOT COALESCE(cp.on_hold,false)
            AND st.state='active' AND (st.expires_at IS NULL OR st.expires_at > clock_timestamp())) OR
        ($2='disabled' AND u.status='draft') OR
        ($2='suspended' AND u.status='suspended') OR
        ($2='revoked' AND u.status='revoked') OR
        ($2='expired' AND (st.state='expired' OR (st.expires_at IS NOT NULL AND st.expires_at <= clock_timestamp()))) OR
        ($2='depleted' AND (st.state='quota_depleted' OR u.status='depleted')) OR
        ($2='pending' AND st.state='pending' AND st.start_policy='on_first_successful_connection') OR
        ($2='on_hold' AND COALESCE(cp.on_hold,false))
    )
    AND ($3 = '' OR st.plan_id = $3::uuid)
    AND ($4 = '' OR cp.group_id = $4::uuid)
    AND ($5 = '' OR EXISTS (
        SELECT 1 FROM pvnaive.customer_tag_assignments filter_tag
        WHERE filter_tag.user_id=u.id AND filter_tag.tag_id=$5::uuid
    ))
    AND ($6 = '' OR r.id = $6::uuid)
    AND ($7::boolean IS NULL OR (st.quota_bytes IS NULL) = $7::boolean)
    AND ($8::boolean IS NULL OR st.no_expiry = $8::boolean)
    AND ($9::timestamptz IS NULL OR st.expires_at >= $9::timestamptz)
    AND ($10::timestamptz IS NULL OR st.expires_at <= $10::timestamptz)
ORDER BY %s %s NULLS LAST, u.id ASC
LIMIT $11 OFFSET $12`, sortColumn, direction)

	offset := (query.Page - 1) * query.PageSize
	rows, err := tx.QueryContext(ctx, statement,
		query.Search, query.Status, query.PlanID, query.GroupID, query.TagID, query.ResellerID,
		nullableBool(query.UnlimitedVolume), nullableBool(query.UnlimitedExpiry),
		query.ExpiryFrom, query.ExpiryTo, query.PageSize, offset,
	)
	if err != nil {
		return CustomerPage{}, fmt.Errorf("customer: search customers: %w", err)
	}
	defer rows.Close()

	page := CustomerPage{Customers: make([]CustomerView, 0), Page: query.Page, PageSize: query.PageSize}
	for rows.Next() {
		var view CustomerView
		var userState, termState, startPolicy string
		var groupID, groupName string
		var groupEnabled bool
		var groupSort int
		var tagsJSON []byte
		var total int64
		var accountingPresent bool
		if err := rows.Scan(
			&view.UserID, &view.Username, &view.DisplayName, &userState,
			&view.ServiceTermID, &termState, &view.QuotaBytes, &view.DurationSeconds,
			&view.NoExpiry, &startPolicy, &view.StartsAt, &view.FirstConnectedAt,
			&view.ExpiresAt, &view.PlanID, &view.PlanName, &view.RuntimeCredentialID,
			&view.SubscriptionAvailable, &view.SubscriptionRetrievable,
			&view.UploadBytes, &view.DownloadBytes, &view.UsedBytes, &view.RemainingBytes,
			&view.Online, &view.OnlineSessions, &view.LastOnline, &view.AccountingComplete,
			&accountingPresent, &view.Note,
			&groupID, &groupName, &groupEnabled, &groupSort, &tagsJSON,
			&view.AssignedActorID, &view.CreatedByActorID, &view.ResellerID,
			&view.CreatedAt, &view.UpdatedAt, &view.LastRenewalAt, &view.OnHold,
			&view.NextPlanID, &view.NextPlanName, &total,
		); err != nil {
			return CustomerPage{}, fmt.Errorf("customer: scan product customer: %w", err)
		}
		view.Status = UserAdminState(userState)
		view.ServiceState = TermState(termState)
		view.StartPolicy = StartPolicy(startPolicy)
		ApplyAccountingSnapshot(&view, AccountingSnapshot{
			Present:            accountingPresent,
			AccountingComplete: view.AccountingComplete,
			UploadBytes:        view.UploadBytes,
			DownloadBytes:      view.DownloadBytes,
			UsedBytes:          view.UsedBytes,
			RemainingBytes:     view.RemainingBytes,
			Online:             view.Online,
			OnlineSessions:     view.OnlineSessions,
			LastOnline:         view.LastOnline,
		})
		if groupID != "" {
			view.Group = &CustomerGroup{ID: groupID, Name: groupName, Enabled: groupEnabled, SortOrder: groupSort}
		}
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &view.Tags); err != nil {
				return CustomerPage{}, fmt.Errorf("customer: decode product tags: %w", err)
			}
		}
		page.Total = total
		page.Customers = append(page.Customers, view)
	}
	if err := rows.Err(); err != nil {
		return CustomerPage{}, fmt.Errorf("customer: search customer rows: %w", err)
	}
	return page, nil
}

func nullableBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func parseOptionalBool(value string) (*bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return nil, nil
	case "true", "1", "yes":
		v := true
		return &v, nil
	case "false", "0", "no":
		v := false
		return &v, nil
	default:
		return nil, errors.New("customer: invalid boolean filter")
	}
}
