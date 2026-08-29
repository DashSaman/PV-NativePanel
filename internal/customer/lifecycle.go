package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

var ErrCustomerLifecycleUnavailable = errors.New("customer: lifecycle operation is unavailable")

type RuntimeUpdateFunc func(context.Context, *sql.Tx, string, string, runtimecred.UpdateInput) (RuntimeMutation, error)
type RuntimeRotateFunc func(context.Context, *sql.Tx, string, string, runtimecred.RotateInput) (RuntimeMutation, error)
type RuntimeRevokeFunc func(context.Context, *sql.Tx, string, string, runtimecred.RevokeInput) (RuntimeMutation, error)

type CustomerRuntimeTarget struct {
	UserID              string
	Username            string
	UserState           UserAdminState
	RuntimeCredentialID string
	RuntimeUsername     string
	RuntimeStatus       runtimecred.CredentialStatus
	RuntimeRevision     int64
}

type lifecycleStore interface {
	CustomerRuntimeTargetTx(context.Context, *sql.Tx, string) (CustomerRuntimeTarget, error)
	SetUserAdminStateTx(context.Context, *sql.Tx, string, UserAdminState) error
}

func (s *Service) ConfigureRuntimeOperations(update RuntimeUpdateFunc, rotate RuntimeRotateFunc, revoke RuntimeRevokeFunc) error {
	if s == nil || update == nil || rotate == nil || revoke == nil {
		return errors.New("customer: all runtime operation hooks are required")
	}
	s.updateRuntime = update
	s.rotateRuntime = rotate
	s.revokeRuntime = revoke
	return nil
}

func validateLifecycleIdentity(actorID, idempotencyKey, userID string) error {
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(userID) == "" {
		return ErrCustomerLifecycleUnavailable
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 160 || strings.TrimSpace(idempotencyKey) != idempotencyKey {
		return errors.New("customer: invalid lifecycle idempotency key")
	}
	return nil
}

func (s *Service) lifecycleTarget(ctx context.Context, tx *sql.Tx, userID string) (lifecycleStore, CustomerRuntimeTarget, error) {
	if s == nil || s.store == nil {
		return nil, CustomerRuntimeTarget{}, ErrCustomerLifecycleUnavailable
	}
	store, ok := s.store.(lifecycleStore)
	if !ok {
		return nil, CustomerRuntimeTarget{}, ErrCustomerLifecycleUnavailable
	}
	target, err := store.CustomerRuntimeTargetTx(ctx, tx, userID)
	if err != nil {
		return nil, CustomerRuntimeTarget{}, err
	}
	if target.UserID == "" || target.RuntimeCredentialID == "" || target.RuntimeRevision <= 0 || target.RuntimeUsername == "" {
		return nil, CustomerRuntimeTarget{}, ErrCustomerLifecycleUnavailable
	}
	return store, target, nil
}

func finalizeLifecycleMutation(ctx context.Context, tx *sql.Tx, mutation RuntimeMutation, store lifecycleStore, userID string, state UserAdminState) (runtimecred.CredentialView, error) {
	if err := store.SetUserAdminStateTx(ctx, tx, userID, state); err != nil {
		return runtimecred.CredentialView{}, abortRuntimeMutation(ctx, tx, mutation, "update customer lifecycle state", err)
	}
	if err := mutation.CommitAndFinalize(ctx, tx); err != nil {
		return runtimecred.CredentialView{}, fmt.Errorf("customer: finalize lifecycle mutation: %w", err)
	}
	return mutation.Credential(), nil
}

func (s *Service) SuspendCustomer(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey, userID string) (runtimecred.CredentialView, error) {
	if err := validateLifecycleIdentity(actorID, idempotencyKey, userID); err != nil {
		return runtimecred.CredentialView{}, err
	}
	if s.updateRuntime == nil {
		return runtimecred.CredentialView{}, ErrCustomerLifecycleUnavailable
	}
	store, target, err := s.lifecycleTarget(ctx, tx, userID)
	if err != nil {
		return runtimecred.CredentialView{}, err
	}
	if target.UserState == UserRevoked || target.RuntimeStatus == runtimecred.CredentialRevoked {
		return runtimecred.CredentialView{}, errors.New("customer: revoked customer cannot be suspended")
	}
	mutation, err := s.updateRuntime(ctx, tx, actorID, runtimeIdempotencyKey(idempotencyKey), runtimecred.UpdateInput{
		ID: target.RuntimeCredentialID, ExpectedRevision: target.RuntimeRevision,
		Username: target.RuntimeUsername, Status: runtimecred.CredentialDisabled,
	})
	if err != nil {
		return runtimecred.CredentialView{}, err
	}
	return finalizeLifecycleMutation(ctx, tx, mutation, store, target.UserID, UserSuspended)
}

func (s *Service) ResumeCustomer(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey, userID string) (runtimecred.CredentialView, error) {
	if err := validateLifecycleIdentity(actorID, idempotencyKey, userID); err != nil {
		return runtimecred.CredentialView{}, err
	}
	if s.updateRuntime == nil {
		return runtimecred.CredentialView{}, ErrCustomerLifecycleUnavailable
	}
	store, target, err := s.lifecycleTarget(ctx, tx, userID)
	if err != nil {
		return runtimecred.CredentialView{}, err
	}
	if target.UserState == UserRevoked || target.RuntimeStatus == runtimecred.CredentialRevoked {
		return runtimecred.CredentialView{}, errors.New("customer: revoked customer cannot be resumed")
	}
	mutation, err := s.updateRuntime(ctx, tx, actorID, runtimeIdempotencyKey(idempotencyKey), runtimecred.UpdateInput{
		ID: target.RuntimeCredentialID, ExpectedRevision: target.RuntimeRevision,
		Username: target.RuntimeUsername, Status: runtimecred.CredentialActive,
	})
	if err != nil {
		return runtimecred.CredentialView{}, err
	}
	return finalizeLifecycleMutation(ctx, tx, mutation, store, target.UserID, UserActive)
}

func (s *Service) RevokeCustomer(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey, userID string) (runtimecred.CredentialView, error) {
	if err := validateLifecycleIdentity(actorID, idempotencyKey, userID); err != nil {
		return runtimecred.CredentialView{}, err
	}
	if s.revokeRuntime == nil {
		return runtimecred.CredentialView{}, ErrCustomerLifecycleUnavailable
	}
	store, target, err := s.lifecycleTarget(ctx, tx, userID)
	if err != nil {
		return runtimecred.CredentialView{}, err
	}
	if target.UserState == UserRevoked || target.RuntimeStatus == runtimecred.CredentialRevoked {
		return runtimecred.CredentialView{ID: target.RuntimeCredentialID, Username: target.RuntimeUsername, Status: runtimecred.CredentialRevoked, Revision: target.RuntimeRevision}, nil
	}
	mutation, err := s.revokeRuntime(ctx, tx, actorID, runtimeIdempotencyKey(idempotencyKey), runtimecred.RevokeInput{
		ID: target.RuntimeCredentialID, ExpectedRevision: target.RuntimeRevision,
	})
	if err != nil {
		return runtimecred.CredentialView{}, err
	}
	return finalizeLifecycleMutation(ctx, tx, mutation, store, target.UserID, UserRevoked)
}

func (s *Service) RotateCustomerPassword(
	ctx context.Context,
	tx *sql.Tx,
	actorID, idempotencyKey, userID, password string,
	generate bool,
) (runtimecred.CredentialView, string, error) {
	if err := validateLifecycleIdentity(actorID, idempotencyKey, userID); err != nil {
		return runtimecred.CredentialView{}, "", err
	}
	if s.rotateRuntime == nil {
		return runtimecred.CredentialView{}, "", ErrCustomerLifecycleUnavailable
	}
	_, target, err := s.lifecycleTarget(ctx, tx, userID)
	if err != nil {
		return runtimecred.CredentialView{}, "", err
	}
	if target.UserState == UserRevoked || target.RuntimeStatus == runtimecred.CredentialRevoked {
		return runtimecred.CredentialView{}, "", errors.New("customer: revoked customer password cannot be rotated")
	}
	mutation, err := s.rotateRuntime(ctx, tx, actorID, runtimeIdempotencyKey(idempotencyKey), runtimecred.RotateInput{
		ID: target.RuntimeCredentialID, ExpectedRevision: target.RuntimeRevision,
		Password: password, GeneratePassword: generate,
	})
	if err != nil {
		return runtimecred.CredentialView{}, "", err
	}
	if err := mutation.CommitAndFinalize(ctx, tx); err != nil {
		return runtimecred.CredentialView{}, "", fmt.Errorf("customer: finalize password rotation: %w", err)
	}
	return mutation.Credential(), mutation.TakeGeneratedPassword(), nil
}
