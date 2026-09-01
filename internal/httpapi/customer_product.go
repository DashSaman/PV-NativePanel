package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
	"github.com/DashSaman/PV-NaivePanel/internal/customer"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func parseCustomerListQuery(r *http.Request) (customer.CustomerListQuery, error) {
	values := r.URL.Query()
	query := customer.CustomerListQuery{
		Search:     values.Get("q"),
		Status:     values.Get("status"),
		PlanID:     values.Get("plan"),
		GroupID:    values.Get("group"),
		TagID:      values.Get("tag"),
		ResellerID: values.Get("reseller"),
		Sort:       customer.CustomerSort(values.Get("sort")),
		Direction:  customer.SortDirection(values.Get("dir")),
	}
	if raw := values.Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return customer.CustomerListQuery{}, err
		}
		query.Page = value
	}
	if raw := values.Get("page_size"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return customer.CustomerListQuery{}, err
		}
		query.PageSize = value
	}
	parseBool := func(key string) (*bool, error) {
		raw := strings.ToLower(strings.TrimSpace(values.Get(key)))
		switch raw {
		case "":
			return nil, nil
		case "true", "1":
			value := true
			return &value, nil
		case "false", "0":
			value := false
			return &value, nil
		default:
			return nil, errors.New("invalid boolean filter")
		}
	}
	var err error
	if query.UnlimitedVolume, err = parseBool("unlimited_volume"); err != nil {
		return customer.CustomerListQuery{}, err
	}
	if query.UnlimitedExpiry, err = parseBool("unlimited_expiry"); err != nil {
		return customer.CustomerListQuery{}, err
	}
	parseTime := func(key string) (*time.Time, error) {
		raw := strings.TrimSpace(values.Get(key))
		if raw == "" {
			return nil, nil
		}
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	if query.ExpiryFrom, err = parseTime("expiry_from"); err != nil {
		return customer.CustomerListQuery{}, err
	}
	if query.ExpiryTo, err = parseTime("expiry_to"); err != nil {
		return customer.CustomerListQuery{}, err
	}
	return query.Normalize(), nil
}

func (s *server) listProductCustomers(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	query, err := parseCustomerListQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_customer_filter", "message": "Customer filters are invalid."})
		return
	}
	page, err := s.config.CustomerService.SearchCustomers(r.Context(), authenticated.Bound.Tx, query)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "customer_list_failed", "message": "Customers could not be loaded."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{
		"customers": page.Customers,
		"page":      page.Page,
		"page_size": page.PageSize,
		"total":     page.Total,
	})
}

func (s *server) createProductCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	key, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_idempotency_key", "message": "A valid Idempotency-Key is required."})
		return
	}
	var input customer.ProductCreateCustomerInput
	if err := decodeRuntimeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid customer request."})
		return
	}
	if input.PlanID == "" && input.QuotaGB == nil {
		input.UnlimitedQuota = true
	}
	result, err := s.config.CustomerService.CreateProductCustomer(
		r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, input,
	)
	if err != nil {
		finishRuntimeTransaction(r)
		writeJSON(w, http.StatusBadRequest, envelope{"code": "customer_create_failed", "message": "Customer could not be created with these settings."})
		return
	}
	authenticated.TransactionFinalized = true
	response := envelope{
		"user":               result.User,
		"service_term":       result.ServiceTerm,
		"runtime_credential": result.RuntimeCredential,
		"usage_capability":   result.UsageCapability,
		"subscription_path":  "/api/v1/subscriptions/" + url.PathEscape(result.SubscriptionToken),
		"delivery_notice":    "Customer created. Runtime identity is stable unless an explicit security action changes it.",
	}
	if result.GeneratedPassword != "" {
		response["generated_password"] = result.GeneratedPassword
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *server) productUserUpdate(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	userID := r.PathValue("id")
	var payload struct {
		Note         *string  `json:"note,omitempty"`
		GroupID      *string  `json:"group_id,omitempty"`
		OnHold       *bool    `json:"on_hold,omitempty"`
		AddTagIDs    []string `json:"add_tag_ids,omitempty"`
		RemoveTagIDs []string `json:"remove_tag_ids,omitempty"`
		NextPlanID   *string  `json:"next_plan_id,omitempty"`
	}
	if userID == "" || decodeRuntimeJSON(r, &payload) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid customer update."})
		return
	}
	response := envelope{}
	if payload.NextPlanID != nil {
		planID := strings.TrimSpace(*payload.NextPlanID)
		if planID == "" {
			if err := s.config.CustomerService.ClearNextPlan(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, userID); err != nil {
				writeJSON(w, http.StatusBadRequest, envelope{"code": "next_plan_failed", "message": "Next Plan could not be cleared."})
				return
			}
			response["next_plan"] = nil
		} else {
			next, err := s.config.CustomerService.ScheduleNextPlan(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, userID, planID)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, envelope{"code": "next_plan_failed", "message": "Next Plan could not be scheduled."})
				return
			}
			response["next_plan"] = next
		}
	}
	if payload.Note != nil || payload.GroupID != nil || payload.OnHold != nil || len(payload.AddTagIDs) > 0 || len(payload.RemoveTagIDs) > 0 {
		metadata, err := s.config.CustomerService.UpdateCustomerMetadata(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, userID, customer.CustomerMetadataInput{
			Note: payload.Note, GroupID: payload.GroupID, OnHold: payload.OnHold,
			AddTagIDs: payload.AddTagIDs, RemoveTagIDs: payload.RemoveTagIDs,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{"code": "metadata_failed", "message": "Customer metadata could not be updated."})
			return
		}
		response["metadata"] = metadata
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *server) renewProductCustomer(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	var input customer.RenewalInput
	if r.PathValue("id") == "" || decodeRuntimeJSON(r, &input) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_renewal", "message": "Invalid renewal request."})
		return
	}
	result, err := s.config.CustomerService.RenewCustomer(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, r.PathValue("id"), input)
	if err != nil {
		status := http.StatusBadRequest
		code := "renewal_failed"
		if errors.Is(err, customer.ErrNextPlanNotReady) {
			status = http.StatusConflict
			code = "next_plan_not_ready"
		}
		writeJSON(w, status, envelope{"code": code, "message": "Service renewal could not be applied."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"renewal": result})
}

func (s *server) listProductPlans(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	plans, err := s.config.CustomerService.ListPlans(r.Context(), authenticated.Bound.Tx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "plans_failed", "message": "Plans could not be loaded."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"plans": plans, "reset_enforcement_available": false})
}

func (s *server) createProductPlan(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	var payload struct {
		Name             string                 `json:"name"`
		QuotaGB          *int64                 `json:"quota_gb,omitempty"`
		UnlimitedQuota   bool                   `json:"unlimited_quota,omitempty"`
		ValidityDays     int64                  `json:"validity_days,omitempty"`
		NoExpiry         bool                   `json:"no_expiry,omitempty"`
		StartPolicy      customer.StartPolicy   `json:"start_policy"`
		ResetStrategy    customer.ResetStrategy `json:"reset_strategy"`
		ResetCustomDays  int                    `json:"reset_custom_days,omitempty"`
		ConcurrencyLimit *int                   `json:"concurrency_limit"`
		UniqueIPLimit    *int                   `json:"unique_ip_limit"`
		DefaultGroupID   string                 `json:"default_group_id,omitempty"`
		TagIDs           []string               `json:"tag_ids,omitempty"`
		Enabled          *bool                  `json:"enabled,omitempty"`
		SortOrder        int                    `json:"sort_order,omitempty"`
	}
	if decodeRuntimeJSON(r, &payload) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_plan", "message": "Invalid Plan request."})
		return
	}
	if payload.StartPolicy == "" {
		payload.StartPolicy = customer.StartOnCreation
	}
	if payload.ResetStrategy == "" {
		payload.ResetStrategy = customer.ResetNone
	}
	var quotaBytes *int64
	if !payload.UnlimitedQuota {
		if payload.QuotaGB == nil {
			writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_plan", "message": "Plan quota is required."})
			return
		}
		var err error
		quotaBytes, err = customer.QuotaGBToBytes(payload.QuotaGB)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_plan", "message": "Plan quota is invalid."})
			return
		}
	}
	enabled := true
	if payload.Enabled != nil {
		enabled = *payload.Enabled
	}
	validitySeconds := payload.ValidityDays * 86400
	if payload.NoExpiry {
		validitySeconds = 0
	}
	plan, err := s.config.CustomerService.CreatePlan(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, customer.PlanPreset{
		Name: payload.Name, QuotaBytes: quotaBytes, ValiditySeconds: validitySeconds,
		NoExpiry: payload.NoExpiry, StartPolicy: payload.StartPolicy, ResetStrategy: payload.ResetStrategy,
		ResetCustomDays: payload.ResetCustomDays, ConcurrencyLimit: payload.ConcurrencyLimit, UniqueIPLimit: payload.UniqueIPLimit, DefaultGroupID: payload.DefaultGroupID,
		TagIDs: payload.TagIDs, Enabled: enabled, SortOrder: payload.SortOrder,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_plan", "message": "Plan could not be created."})
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"plan": plan})
}

func (s *server) listCustomerGroups(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	groups, err := s.config.CustomerService.ListGroups(r.Context(), authenticated.Bound.Tx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "groups_failed", "message": "Groups could not be loaded."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"groups": groups})
}

func (s *server) createCustomerGroup(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	var payload struct {
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	if decodeRuntimeJSON(r, &payload) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_group", "message": "Invalid Group request."})
		return
	}
	group, err := s.config.CustomerService.CreateGroup(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, payload.Name, payload.SortOrder)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_group", "message": "Group could not be created."})
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"group": group})
}

func (s *server) listCustomerTags(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	tags, err := s.config.CustomerService.ListTags(r.Context(), authenticated.Bound.Tx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "tags_failed", "message": "Tags could not be loaded."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"tags": tags})
}

func (s *server) createCustomerTag(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	var payload struct {
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	if decodeRuntimeJSON(r, &payload) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_tag", "message": "Invalid Tag request."})
		return
	}
	tag, err := s.config.CustomerService.CreateTag(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, payload.Name, payload.SortOrder)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_tag", "message": "Tag could not be created."})
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"tag": tag})
}

func (s *server) previewCustomerBulk(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	key, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_idempotency_key", "message": "A valid Idempotency-Key is required."})
		return
	}
	var request customer.BulkRequest
	if decodeRuntimeJSON(r, &request) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_bulk", "message": "Invalid bulk request."})
		return
	}
	if !bulkActionAllowed(authenticated.Bound.Principal.Role, request.Action) {
		writeJSON(w, http.StatusForbidden, envelope{"code": "forbidden", "message": "This bulk action requires owner access."})
		return
	}
	operation, err := s.config.CustomerService.PreviewBulk(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, request, s.config.AuthStore != nil)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, customer.ErrBulkConflict) {
			status = http.StatusConflict
		}
		writeJSON(w, status, envelope{"code": "bulk_preview_failed", "message": "Bulk preview could not be created."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"bulk": operation})
}

func bulkExcluded(preview customer.BulkPreview) map[string]bool {
	excluded := map[string]bool{}
	for _, list := range [][]customer.BulkItem{preview.Conflicts, preview.Skipped, preview.Invalid} {
		for _, item := range list {
			excluded[item.ID] = true
		}
	}
	return excluded
}

func bulkItemIdempotencyKey(parent string, action customer.BulkAction, userID string) string {
	sum := sha256.Sum256([]byte(parent + "\n" + string(action) + "\n" + userID))
	return "bulk-item-" + hex.EncodeToString(sum[:])
}

func bulkActionAllowed(role string, action customer.BulkAction) bool {
	if action == customer.BulkResetUsage {
		return role == "owner"
	}
	return true
}

func (s *server) executeCustomerBulk(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok || s.config.CustomerService == nil {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	key, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_idempotency_key", "message": "A valid Idempotency-Key is required."})
		return
	}
	var request customer.BulkRequest
	if decodeRuntimeJSON(r, &request) != nil {
		writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_bulk", "message": "Invalid bulk request."})
		return
	}
	if !bulkActionAllowed(authenticated.Bound.Principal.Role, request.Action) {
		writeJSON(w, http.StatusForbidden, envelope{"code": "forbidden", "message": "This bulk action requires owner access."})
		return
	}
	operation, err := s.config.CustomerService.LoadBulk(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, request)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, customer.ErrBulkConflict) {
			status = http.StatusConflict
		}
		writeJSON(w, status, envelope{"code": "bulk_execute_failed", "message": "Bulk request does not match its preview."})
		return
	}
	if operation.Status == "executed" {
		writeJSON(w, http.StatusOK, envelope{"bulk": operation})
		return
	}
	if !customer.IsPerItemBulkAction(operation.Action) {
		operation, err = s.config.CustomerService.ExecuteDatabaseBulk(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, key, operation)
		if err != nil {
			writeJSON(w, http.StatusConflict, envelope{"code": "bulk_execute_failed", "message": "Bulk operation could not be executed atomically."})
			return
		}
		writeJSON(w, http.StatusOK, envelope{"bulk": operation})
		return
	}
	if s.config.AuthStore == nil || len(operation.Preview.Conflicts) > 0 {
		writeJSON(w, http.StatusConflict, envelope{"code": "bulk_item_coordinator_unavailable", "message": "Per-item bulk coordinator is unavailable."})
		return
	}

	if err := authenticated.Bound.Tx.Commit(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "bulk_execute_failed", "message": "Bulk preview state could not be finalized."})
		return
	}
	authenticated.TransactionFinalized = true
	sessionHash := auth.HashOpaqueToken(authenticated.RawSessionToken)
	excluded := bulkExcluded(operation.Preview)
	result := customer.BulkExecutionResult{Items: make([]customer.BulkItemResult, 0, len(operation.Request.CustomerIDs))}
	for _, userID := range operation.Request.CustomerIDs {
		if excluded[userID] {
			result.Skipped++
			result.Items = append(result.Items, customer.BulkItemResult{ID: userID, Status: "skipped"})
			continue
		}
		bound, beginErr := s.config.AuthStore.BeginAuthenticated(r.Context(), sessionHash[:])
		if beginErr != nil {
			result.Failed++
			result.Items = append(result.Items, customer.BulkItemResult{ID: userID, Status: "failed", Reason: "authentication_context_unavailable"})
			continue
		}
		itemKey := bulkItemIdempotencyKey(key, operation.Action, userID)
		var itemErr error
		itemReplay := false
		switch operation.Action {
		case customer.BulkEnable:
			_, itemErr = s.config.CustomerService.ResumeCustomer(r.Context(), bound.Tx, bound.Principal.ActorID, itemKey, userID)
		case customer.BulkSuspend:
			_, itemErr = s.config.CustomerService.SuspendCustomer(r.Context(), bound.Tx, bound.Principal.ActorID, itemKey, userID)
		case customer.BulkRevoke, customer.BulkSafeDelete:
			_, itemErr = s.config.CustomerService.RevokeCustomer(r.Context(), bound.Tx, bound.Principal.ActorID, itemKey, userID)
		case customer.BulkReissueSubscription:
			_, itemErr = s.config.CustomerService.RotateSubscription(r.Context(), bound.Tx, bound.Principal.ActorID, itemKey, userID)
			if itemErr == nil {
				itemErr = bound.Tx.Commit()
			}
		case customer.BulkResetUsage:
			var resetResult customer.UsageResetResult
			resetResult, itemErr = s.config.CustomerService.ResetCustomerUsageForBulk(r.Context(), bound.Tx, bound.Principal.ActorID, itemKey, userID)
			if itemErr == nil {
				itemReplay = resetResult.IdempotentReplay
				itemErr = bound.Tx.Commit()
			}
		default:
			itemErr = customer.ErrInvalidBulkRequest
		}
		if itemErr != nil {
			_ = bound.Tx.Rollback()
			if errors.Is(itemErr, runtimecred.ErrIdempotentReplay) || errors.Is(itemErr, customer.ErrSubscriptionRotationReplay) {
				result.Succeeded++
				result.Items = append(result.Items, customer.BulkItemResult{ID: userID, Status: "succeeded", Reason: "idempotent_replay"})
				continue
			}
			result.Failed++
			result.Items = append(result.Items, customer.BulkItemResult{ID: userID, Status: "failed", Reason: "item_operation_failed"})
			continue
		}
		result.Succeeded++
		if itemReplay {
			result.Items = append(result.Items, customer.BulkItemResult{ID: userID, Status: "succeeded", Reason: "idempotent_replay"})
		} else {
			result.Items = append(result.Items, customer.BulkItemResult{ID: userID, Status: "succeeded"})
		}
	}

	finalBound, err := s.config.AuthStore.BeginAuthenticated(r.Context(), sessionHash[:])
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "bulk_finalize_failed", "message": "Bulk items ran but the execution ledger could not be finalized; retry with the same key."})
		return
	}
	operation, err = s.config.CustomerService.MarkRuntimeBulkExecuted(r.Context(), finalBound.Tx, finalBound.Principal.ActorID, key, operation.Action, result)
	if err != nil {
		_ = finalBound.Tx.Rollback()
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "bulk_finalize_failed", "message": "Bulk items ran but the execution ledger could not be finalized; retry with the same key."})
		return
	}
	if err := finalBound.Tx.Commit(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "bulk_finalize_failed", "message": "Bulk execution ledger commit failed; retry with the same key."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"bulk": operation})
}
