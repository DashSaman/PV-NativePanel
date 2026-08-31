package customer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidBulkRequest = errors.New("customer: invalid bulk request")
	ErrBulkConflict       = errors.New("customer: bulk idempotency conflict")
	ErrBulkNotPreviewed   = errors.New("customer: bulk operation must be previewed first")
)

type BulkRequest struct {
	Action          BulkAction `json:"action"`
	CustomerIDs     []string   `json:"customer_ids"`
	Days            int64      `json:"days,omitempty"`
	VolumeGB        int64      `json:"volume_gb,omitempty"`
	TotalVolumeGB   *int64     `json:"total_volume_gb,omitempty"`
	UnlimitedVolume bool       `json:"unlimited_volume,omitempty"`
	PlanID          string     `json:"plan_id,omitempty"`
	GroupID         string     `json:"group_id,omitempty"`
	TagID           string     `json:"tag_id,omitempty"`
}

type BulkItemResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type BulkExecutionResult struct {
	Succeeded int              `json:"succeeded"`
	Failed    int              `json:"failed"`
	Skipped   int              `json:"skipped"`
	Items     []BulkItemResult `json:"items"`
}

type BulkOperation struct {
	ID             string               `json:"id"`
	Action         BulkAction           `json:"action"`
	Status         string               `json:"status"`
	Preview        BulkPreview          `json:"preview"`
	Result         *BulkExecutionResult `json:"result,omitempty"`
	Request        BulkRequest          `json:"request"`
	IdempotencyKey string               `json:"-"`
}

type bulkStore interface {
	OperationTenantIDTx(context.Context, *sql.Tx) (string, error)
	BulkCustomersTx(context.Context, *sql.Tx, []string) ([]BulkCustomer, error)
	ClaimBulkPreviewTx(context.Context, *sql.Tx, string, string, string, []byte, BulkRequest, BulkPreview) (BulkOperation, []byte, error)
	BulkOperationTx(context.Context, *sql.Tx, string, string, BulkAction) (BulkOperation, []byte, error)
	MarkBulkExecutedTx(context.Context, *sql.Tx, string, string, BulkAction, BulkExecutionResult) (BulkOperation, error)
}

func (r BulkRequest) Normalize() (BulkRequest, error) {
	r.CustomerIDs = uniqueStrings(r.CustomerIDs)
	r.PlanID = strings.TrimSpace(r.PlanID)
	r.GroupID = strings.TrimSpace(r.GroupID)
	r.TagID = strings.TrimSpace(r.TagID)
	if len(r.CustomerIDs) == 0 || len(r.CustomerIDs) > 500 {
		return BulkRequest{}, ErrInvalidBulkRequest
	}
	switch r.Action {
	case BulkEnable, BulkSuspend, BulkRevoke, BulkSafeDelete, BulkReissueSubscription, BulkResetUsage:
	case BulkExtendDays:
		if r.Days <= 0 {
			return BulkRequest{}, ErrInvalidBulkRequest
		}
	case BulkAddVolume:
		if r.VolumeGB <= 0 {
			return BulkRequest{}, ErrInvalidBulkRequest
		}
	case BulkSetVolume:
		if r.UnlimitedVolume {
			if r.TotalVolumeGB != nil {
				return BulkRequest{}, ErrInvalidBulkRequest
			}
		} else if r.TotalVolumeGB == nil || *r.TotalVolumeGB <= 0 {
			return BulkRequest{}, ErrInvalidBulkRequest
		}
	case BulkApplyPlan:
		if r.PlanID == "" {
			return BulkRequest{}, ErrInvalidBulkRequest
		}
	case BulkAssignGroup:
		// Empty group id intentionally clears the group.
	case BulkAddTag, BulkRemoveTag:
		if r.TagID == "" {
			return BulkRequest{}, ErrInvalidBulkRequest
		}
	default:
		return BulkRequest{}, ErrInvalidBulkRequest
	}
	return r, nil
}

func IsRuntimeBulkAction(action BulkAction) bool {
	switch action {
	case BulkEnable, BulkSuspend, BulkRevoke, BulkSafeDelete, BulkReissueSubscription:
		return true
	default:
		return false
	}
}

func IsPerItemBulkAction(action BulkAction) bool {
	return IsRuntimeBulkAction(action) || action == BulkResetUsage
}

func bulkRequestHash(request BulkRequest) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(payload)
	return sum[:], nil
}

func (s *Service) PreviewBulk(
	ctx context.Context,
	tx *sql.Tx,
	actorID, idempotencyKey string,
	request BulkRequest,
	runtimeCoordinatorAvailable bool,
) (BulkOperation, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(actorID) == "" || len(idempotencyKey) < 8 || len(idempotencyKey) > 160 {
		return BulkOperation{}, ErrInvalidBulkRequest
	}
	request, err := request.Normalize()
	if err != nil {
		return BulkOperation{}, err
	}
	store, ok := s.store.(bulkStore)
	if !ok {
		return BulkOperation{}, errors.New("customer: bulk capability is unavailable")
	}
	tenantID, err := store.OperationTenantIDTx(ctx, tx)
	if err != nil {
		return BulkOperation{}, err
	}
	customers, err := store.BulkCustomersTx(ctx, tx, request.CustomerIDs)
	if err != nil {
		return BulkOperation{}, err
	}
	preview := BuildBulkPreview(BulkPreviewInput{
		Action:                      request.Action,
		RequestedIDs:                request.CustomerIDs,
		Customers:                   customers,
		RuntimeCoordinatorAvailable: runtimeCoordinatorAvailable,
	})
	hash, err := bulkRequestHash(request)
	if err != nil {
		return BulkOperation{}, err
	}
	operation, storedHash, err := store.ClaimBulkPreviewTx(ctx, tx, tenantID, actorID, idempotencyKey, hash, request, preview)
	if err != nil {
		return BulkOperation{}, err
	}
	if !bytes.Equal(hash, storedHash) {
		return BulkOperation{}, ErrBulkConflict
	}
	return operation, nil
}

func (s *Service) LoadBulk(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, request BulkRequest) (BulkOperation, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(actorID) == "" {
		return BulkOperation{}, ErrInvalidBulkRequest
	}
	request, err := request.Normalize()
	if err != nil {
		return BulkOperation{}, err
	}
	store, ok := s.store.(bulkStore)
	if !ok {
		return BulkOperation{}, errors.New("customer: bulk capability is unavailable")
	}
	hash, err := bulkRequestHash(request)
	if err != nil {
		return BulkOperation{}, err
	}
	operation, storedHash, err := store.BulkOperationTx(ctx, tx, actorID, idempotencyKey, request.Action)
	if err != nil {
		return BulkOperation{}, ErrBulkNotPreviewed
	}
	if !bytes.Equal(hash, storedHash) {
		return BulkOperation{}, ErrBulkConflict
	}
	return operation, nil
}

func excludedBulkIDs(preview BulkPreview) map[string]bool {
	excluded := make(map[string]bool, len(preview.Conflicts)+len(preview.Skipped)+len(preview.Invalid))
	for _, list := range [][]BulkItem{preview.Conflicts, preview.Skipped, preview.Invalid} {
		for _, item := range list {
			excluded[item.ID] = true
		}
	}
	return excluded
}

func (s *Service) ExecuteDatabaseBulk(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, operation BulkOperation) (BulkOperation, error) {
	if IsPerItemBulkAction(operation.Action) {
		return BulkOperation{}, ErrInvalidBulkRequest
	}
	store, ok := s.store.(bulkStore)
	if !ok || tx == nil {
		return BulkOperation{}, errors.New("customer: bulk capability is unavailable")
	}
	if operation.Status == "executed" {
		return operation, nil
	}
	if len(operation.Preview.Conflicts) > 0 {
		return BulkOperation{}, ErrInvalidBulkRequest
	}
	excluded := excludedBulkIDs(operation.Preview)
	result := BulkExecutionResult{Items: make([]BulkItemResult, 0, len(operation.Request.CustomerIDs))}
	for _, userID := range operation.Request.CustomerIDs {
		if excluded[userID] {
			result.Skipped++
			result.Items = append(result.Items, BulkItemResult{ID: userID, Status: "skipped"})
			continue
		}
		var err error
		switch operation.Action {
		case BulkExtendDays:
			_, err = s.ExtendCustomerTime(ctx, tx, userID, operation.Request.Days)
		case BulkAddVolume:
			_, err = s.AddCustomerVolume(ctx, tx, userID, operation.Request.VolumeGB)
		case BulkSetVolume:
			quota := operation.Request.TotalVolumeGB
			if operation.Request.UnlimitedVolume {
				quota = nil
			}
			_, err = s.SetCustomerVolume(ctx, tx, userID, quota)
		case BulkApplyPlan:
			_, err = s.RenewCustomer(ctx, tx, actorID, userID, RenewalInput{Mode: RenewalUsingPlan, PlanID: operation.Request.PlanID})
		case BulkAssignGroup:
			groupID := operation.Request.GroupID
			_, err = s.UpdateCustomerMetadata(ctx, tx, actorID, userID, CustomerMetadataInput{GroupID: &groupID})
		case BulkAddTag:
			_, err = s.UpdateCustomerMetadata(ctx, tx, actorID, userID, CustomerMetadataInput{AddTagIDs: []string{operation.Request.TagID}})
		case BulkRemoveTag:
			_, err = s.UpdateCustomerMetadata(ctx, tx, actorID, userID, CustomerMetadataInput{RemoveTagIDs: []string{operation.Request.TagID}})
		default:
			err = ErrInvalidBulkRequest
		}
		if err != nil {
			return BulkOperation{}, fmt.Errorf("customer: bulk item %s: %w", userID, err)
		}
		result.Succeeded++
		result.Items = append(result.Items, BulkItemResult{ID: userID, Status: "succeeded"})
	}
	return store.MarkBulkExecutedTx(ctx, tx, actorID, idempotencyKey, operation.Action, result)
}

func (s *Service) MarkRuntimeBulkExecuted(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, action BulkAction, result BulkExecutionResult) (BulkOperation, error) {
	store, ok := s.store.(bulkStore)
	if !ok || tx == nil {
		return BulkOperation{}, errors.New("customer: bulk capability is unavailable")
	}
	return store.MarkBulkExecutedTx(ctx, tx, actorID, idempotencyKey, action, result)
}
