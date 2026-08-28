package runtimeagent

import (
	"context"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

const (
	DefaultSocketPath = "/run/pvnaive/runtime-agent.sock"
	maxRequestBytes   = 64 << 10
)

type HealthResponse struct {
	Status string `json:"status"`
}

type InspectCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type InspectResponse struct {
	CaddySHA256 string              `json:"caddy_sha256"`
	Credentials []InspectCredential `json:"credentials"`
}

type CredentialInput struct {
	ID       string                       `json:"id"`
	Username string                       `json:"username"`
	Password string                       `json:"password"`
	Status   runtimecred.CredentialStatus `json:"status"`
}

type DesiredStateInput struct {
	Revision    string            `json:"revision"`
	Credentials []CredentialInput `json:"credentials"`
}

type ValidateRequest struct {
	ExpectedCaddySHA256 string            `json:"expected_caddy_sha256"`
	Desired             DesiredStateInput `json:"desired"`
}

type ValidateResponse struct {
	CandidateSHA256 string `json:"candidate_sha256"`
}

type ApplyRequest struct {
	ExpectedCaddySHA256 string            `json:"expected_caddy_sha256"`
	Desired             DesiredStateInput `json:"desired"`
}

type ApplyResponse struct {
	PreviousSHA256 string `json:"previous_sha256"`
	AppliedSHA256  string `json:"applied_sha256"`
	BackupID       string `json:"backup_id"`
	MainPID        int    `json:"main_pid"`
	NRestarts      int    `json:"n_restarts"`
}

type RollbackRequest struct {
	BackupID string `json:"backup_id"`
}

type RollbackResponse struct {
	RestoredSHA256 string `json:"restored_sha256"`
	MainPID        int    `json:"main_pid"`
	NRestarts      int    `json:"n_restarts"`
}

type Operator interface {
	Health(context.Context) (HealthResponse, error)
	Inspect(context.Context) (InspectResponse, error)
	Validate(context.Context, ValidateRequest) (ValidateResponse, error)
	Apply(context.Context, ApplyRequest) (ApplyResponse, error)
	Rollback(context.Context, RollbackRequest) (RollbackResponse, error)
}
