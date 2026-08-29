package customer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type lifecycleCustomerStore struct {
	fakeCustomerStore
	target CustomerRuntimeTarget
	state  UserAdminState
}

func (s *lifecycleCustomerStore) CustomerRuntimeTargetTx(context.Context, *sql.Tx, string) (CustomerRuntimeTarget, error) {
	return s.target, nil
}

func (s *lifecycleCustomerStore) SetUserAdminStateTx(_ context.Context, _ *sql.Tx, _ string, state UserAdminState) error {
	s.state = state
	return nil
}

func lifecycleService(t *testing.T) (*Service, *lifecycleCustomerStore, *fakeCustomerRuntimeMutation, *runtimecred.UpdateInput, *runtimecred.RotateInput, *runtimecred.RevokeInput) {
	t.Helper()
	store := &lifecycleCustomerStore{target: CustomerRuntimeTarget{
		UserID: "user-1", Username: "alice", UserState: UserActive,
		RuntimeCredentialID: "runtime-1", RuntimeUsername: "alice",
		RuntimeStatus: runtimecred.CredentialActive, RuntimeRevision: 7,
	}}
	service := NewService(store, nil, func() time.Time { return time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC) })
	mutation := &fakeCustomerRuntimeMutation{view: runtimecred.CredentialView{ID: "runtime-1", Username: "alice", Status: runtimecred.CredentialActive, Revision: 8}, password: "new-generated-password"}
	var update runtimecred.UpdateInput
	var rotate runtimecred.RotateInput
	var revoke runtimecred.RevokeInput
	err := service.ConfigureRuntimeOperations(
		func(_ context.Context, _ *sql.Tx, _, _ string, input runtimecred.UpdateInput) (RuntimeMutation, error) {
			update = input
			return mutation, nil
		},
		func(_ context.Context, _ *sql.Tx, _, _ string, input runtimecred.RotateInput) (RuntimeMutation, error) {
			rotate = input
			return mutation, nil
		},
		func(_ context.Context, _ *sql.Tx, _, _ string, input runtimecred.RevokeInput) (RuntimeMutation, error) {
			revoke = input
			return mutation, nil
		},
	)
	if err != nil {
		t.Fatalf("ConfigureRuntimeOperations() error = %v", err)
	}
	return service, store, mutation, &update, &rotate, &revoke
}

func TestSuspendCustomerDisablesRuntimeAndSetsBusinessState(t *testing.T) {
	service, store, mutation, update, _, _ := lifecycleService(t)
	_, err := service.SuspendCustomer(context.Background(), nil, "owner-1", "idem-suspend-0001", "user-1")
	if err != nil {
		t.Fatalf("SuspendCustomer() error = %v", err)
	}
	if update.ID != "runtime-1" || update.ExpectedRevision != 7 || update.Username != "alice" || update.Status != runtimecred.CredentialDisabled {
		t.Fatalf("runtime update = %#v", *update)
	}
	if store.state != UserSuspended || !mutation.committed {
		t.Fatalf("state=%q committed=%v", store.state, mutation.committed)
	}
}

func TestResumeCustomerEnablesSameRuntimeCredential(t *testing.T) {
	service, store, mutation, update, _, _ := lifecycleService(t)
	store.target.UserState = UserSuspended
	store.target.RuntimeStatus = runtimecred.CredentialDisabled
	_, err := service.ResumeCustomer(context.Background(), nil, "owner-1", "idem-resume-0001", "user-1")
	if err != nil {
		t.Fatalf("ResumeCustomer() error = %v", err)
	}
	if update.ID != "runtime-1" || update.ExpectedRevision != 7 || update.Username != "alice" || update.Status != runtimecred.CredentialActive {
		t.Fatalf("runtime update = %#v", *update)
	}
	if store.state != UserActive || !mutation.committed {
		t.Fatalf("state=%q committed=%v", store.state, mutation.committed)
	}
}

func TestRevokeCustomerRevokesRuntimeAndBusinessUser(t *testing.T) {
	service, store, mutation, _, _, revoke := lifecycleService(t)
	_, err := service.RevokeCustomer(context.Background(), nil, "owner-1", "idem-revoke-0001", "user-1")
	if err != nil {
		t.Fatalf("RevokeCustomer() error = %v", err)
	}
	if revoke.ID != "runtime-1" || revoke.ExpectedRevision != 7 {
		t.Fatalf("runtime revoke = %#v", *revoke)
	}
	if store.state != UserRevoked || !mutation.committed {
		t.Fatalf("state=%q committed=%v", store.state, mutation.committed)
	}
}

func TestRotateCustomerPasswordDoesNotRotateSubscription(t *testing.T) {
	service, _, mutation, _, rotate, _ := lifecycleService(t)
	view, generated, err := service.RotateCustomerPassword(context.Background(), nil, "owner-1", "idem-password-0001", "user-1", "", true)
	if err != nil {
		t.Fatalf("RotateCustomerPassword() error = %v", err)
	}
	if rotate.ID != "runtime-1" || rotate.ExpectedRevision != 7 || !rotate.GeneratePassword {
		t.Fatalf("runtime rotate = %#v", *rotate)
	}
	if view.ID != "runtime-1" || generated != "new-generated-password" || !mutation.committed {
		t.Fatalf("view=%#v generated=%q committed=%v", view, generated, mutation.committed)
	}
}
