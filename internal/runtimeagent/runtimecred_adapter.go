package runtimeagent

import (
	"context"
	"errors"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type runtimeCredRPC interface {
	Inspect(context.Context) (InspectResponse, error)
	Apply(context.Context, ApplyRequest) (ApplyResponse, error)
	Rollback(context.Context, RollbackRequest) (RollbackResponse, error)
}

type RuntimeCredAdapter struct {
	rpc runtimeCredRPC
}

func NewRuntimeCredAdapter(rpc runtimeCredRPC) *RuntimeCredAdapter {
	return &RuntimeCredAdapter{rpc: rpc}
}

func (a *RuntimeCredAdapter) Inspect(ctx context.Context) (runtimecred.AgentInspection, error) {
	if a == nil || a.rpc == nil {
		return runtimecred.AgentInspection{}, errors.New("runtimeagent: runtime credential adapter is not initialized")
	}
	response, err := a.rpc.Inspect(ctx)
	if err != nil {
		return runtimecred.AgentInspection{}, err
	}
	return runtimecred.AgentInspection{CaddySHA256: response.CaddySHA256}, nil
}

func (a *RuntimeCredAdapter) Apply(ctx context.Context, request runtimecred.AgentApplyRequest) (runtimecred.AgentApplyResult, error) {
	if a == nil || a.rpc == nil {
		return runtimecred.AgentApplyResult{}, errors.New("runtimeagent: runtime credential adapter is not initialized")
	}
	credentials := make([]CredentialInput, 0, len(request.Credentials))
	for _, credential := range request.Credentials {
		credentials = append(credentials, CredentialInput{
			ID:       credential.ID,
			Username: credential.Username,
			Password: credential.Password,
			Status:   credential.Status,
		})
	}
	response, err := a.rpc.Apply(ctx, ApplyRequest{
		ExpectedCaddySHA256: request.ExpectedCaddySHA256,
		Desired: DesiredStateInput{
			Revision:    request.Revision,
			Credentials: credentials,
		},
	})
	for i := range credentials {
		credentials[i].Password = ""
	}
	if err != nil {
		return runtimecred.AgentApplyResult{}, err
	}
	return runtimecred.AgentApplyResult{
		PreviousSHA256: response.PreviousSHA256,
		AppliedSHA256:  response.AppliedSHA256,
		BackupID:       response.BackupID,
		MainPID:        response.MainPID,
		NRestarts:      response.NRestarts,
	}, nil
}

func (a *RuntimeCredAdapter) Rollback(ctx context.Context, backupID string) error {
	if a == nil || a.rpc == nil {
		return errors.New("runtimeagent: runtime credential adapter is not initialized")
	}
	_, err := a.rpc.Rollback(ctx, RollbackRequest{BackupID: backupID})
	return err
}
