package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func (s *server) runtimeNaiveStatus(w http.ResponseWriter, r *http.Request) {
	inspection, err := s.config.RuntimeService.InspectRuntime(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, envelope{
			"status":            "unavailable",
			"runtime_available": false,
		})
		return
	}
	writeJSON(w, http.StatusOK, envelope{
		"status":            "ready",
		"runtime_available": true,
		"caddy_sha256":      inspection.CaddySHA256,
	})
}

func (s *server) runtimeNaiveCredentials(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	credentials, err := s.config.RuntimeService.List(r.Context(), authenticated.Bound.Tx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "runtime_state_unavailable", "message": "Runtime credential state is unavailable."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{"credentials": credentials})
}

func (s *server) runtimeNaiveImport(w http.ResponseWriter, r *http.Request) {
	idempotencyKey, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeRuntimeInvalidRequest(w)
		return
	}
	var payload struct{}
	if err := decodeRuntimeJSON(r, &payload); err != nil {
		writeRuntimeInvalidRequest(w)
		return
	}
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	credentials, err := s.config.RuntimeService.ImportCurrent(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, idempotencyKey)
	if err != nil {
		finishRuntimeTransaction(r)
		writeRuntimeServiceError(w, err)
		return
	}
	if err := authenticated.Bound.Tx.Commit(); err != nil {
		authenticated.TransactionFinalized = true
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "runtime_import_commit_failed", "message": "Live runtime was not changed and import was not committed."})
		return
	}
	authenticated.TransactionFinalized = true
	writeJSON(w, http.StatusOK, envelope{
		"status":      "imported",
		"credentials": credentials,
		"notice":      "Current live credentials were encrypted into management state without changing Caddy.",
	})
}

func (s *server) runtimeNaiveCreateCredential(w http.ResponseWriter, r *http.Request) {
	idempotencyKey, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeRuntimeInvalidRequest(w)
		return
	}
	var payload struct {
		Username         string `json:"username"`
		Password         string `json:"password"`
		GeneratePassword bool   `json:"generate_password"`
	}
	if err := decodeRuntimeJSON(r, &payload); err != nil || runtimeCreatePayloadInvalid(payload.Username, payload.Password, payload.GeneratePassword) {
		writeRuntimeInvalidRequest(w)
		return
	}
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	mutation, err := s.config.RuntimeService.Create(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, idempotencyKey, runtimecred.CreateInput{
		Username: payload.Username, Password: payload.Password, GeneratePassword: payload.GeneratePassword,
	})
	if err != nil {
		finishRuntimeTransaction(r)
		writeRuntimeServiceError(w, err)
		return
	}
	if err := commitRuntimeMutation(r, mutation); err != nil {
		writeRuntimeServiceError(w, err)
		return
	}
	response := envelope{"credential": mutation.Credential(), "runtime_revision_id": mutation.RuntimeRevisionID()}
	if generated := mutation.TakeGeneratedPassword(); generated != "" {
		response["generated_password"] = generated
		response["generated_password_notice"] = "This password will only be shown once."
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *server) runtimeNaiveUpdateCredential(w http.ResponseWriter, r *http.Request) {
	idempotencyKey, expectedRevision, ok := runtimeMutationHeaders(w, r)
	if !ok {
		return
	}
	var payload struct {
		Username string                       `json:"username"`
		Status   runtimecred.CredentialStatus `json:"status"`
	}
	if err := decodeRuntimeJSON(r, &payload); err != nil || runtimeUpdatePayloadInvalid(payload.Username, payload.Status) {
		writeRuntimeInvalidRequest(w)
		return
	}
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	mutation, err := s.config.RuntimeService.Update(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, idempotencyKey, runtimecred.UpdateInput{
		ID: r.PathValue("id"), ExpectedRevision: expectedRevision, Username: payload.Username, Status: payload.Status,
	})
	if err != nil {
		finishRuntimeTransaction(r)
		writeRuntimeServiceError(w, err)
		return
	}
	if err := commitRuntimeMutation(r, mutation); err != nil {
		writeRuntimeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"credential": mutation.Credential(), "runtime_revision_id": mutation.RuntimeRevisionID()})
}

func (s *server) runtimeNaiveRotateCredential(w http.ResponseWriter, r *http.Request) {
	idempotencyKey, expectedRevision, ok := runtimeMutationHeaders(w, r)
	if !ok {
		return
	}
	var payload struct {
		Password         string `json:"password"`
		GeneratePassword bool   `json:"generate_password"`
	}
	if err := decodeRuntimeJSON(r, &payload); err != nil || runtimeRotatePayloadInvalid(payload.Password, payload.GeneratePassword) {
		writeRuntimeInvalidRequest(w)
		return
	}
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	mutation, err := s.config.RuntimeService.Rotate(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, idempotencyKey, runtimecred.RotateInput{
		ID: r.PathValue("id"), ExpectedRevision: expectedRevision, Password: payload.Password, GeneratePassword: payload.GeneratePassword,
	})
	if err != nil {
		finishRuntimeTransaction(r)
		writeRuntimeServiceError(w, err)
		return
	}
	if err := commitRuntimeMutation(r, mutation); err != nil {
		writeRuntimeServiceError(w, err)
		return
	}
	response := envelope{"credential": mutation.Credential(), "runtime_revision_id": mutation.RuntimeRevisionID()}
	if generated := mutation.TakeGeneratedPassword(); generated != "" {
		response["generated_password"] = generated
		response["generated_password_notice"] = "This password will only be shown once."
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *server) runtimeNaiveRevokeCredential(w http.ResponseWriter, r *http.Request) {
	idempotencyKey, expectedRevision, ok := runtimeMutationHeaders(w, r)
	if !ok {
		return
	}
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, envelope{"code": "authentication_required", "message": "Authentication is required."})
		return
	}
	mutation, err := s.config.RuntimeService.Revoke(r.Context(), authenticated.Bound.Tx, authenticated.Bound.Principal.ActorID, idempotencyKey, runtimecred.RevokeInput{
		ID: r.PathValue("id"), ExpectedRevision: expectedRevision,
	})
	if err != nil {
		finishRuntimeTransaction(r)
		writeRuntimeServiceError(w, err)
		return
	}
	if err := commitRuntimeMutation(r, mutation); err != nil {
		writeRuntimeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"credential": mutation.Credential(), "runtime_revision_id": mutation.RuntimeRevisionID()})
}

func runtimeMutationHeaders(w http.ResponseWriter, r *http.Request) (string, int64, bool) {
	idempotencyKey, err := runtimeIdempotencyKey(r)
	if err != nil {
		writeRuntimeInvalidRequest(w)
		return "", 0, false
	}
	expectedRevision, err := runtimeExpectedRevision(r)
	if err != nil {
		writeJSON(w, http.StatusPreconditionRequired, envelope{"code": "expected_revision_required", "message": "A valid If-Match revision is required."})
		return "", 0, false
	}
	if r.PathValue("id") == "" {
		writeRuntimeInvalidRequest(w)
		return "", 0, false
	}
	return idempotencyKey, expectedRevision, true
}

func runtimeIdempotencyKey(r *http.Request) (string, error) {
	value := r.Header.Get("Idempotency-Key")
	if len(value) < 8 || len(value) > 160 || strings.TrimSpace(value) != value {
		return "", errors.New("invalid idempotency key")
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 0x21 || value[i] > 0x7e {
			return "", errors.New("invalid idempotency key")
		}
	}
	return value, nil
}

func runtimeExpectedRevision(r *http.Request) (int64, error) {
	value := strings.TrimSpace(r.Header.Get("If-Match"))
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	revision, err := strconv.ParseInt(value, 10, 64)
	if err != nil || revision <= 0 {
		return 0, errors.New("invalid expected revision")
	}
	return revision, nil
}

func decodeRuntimeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request contains trailing JSON")
	}
	return nil
}

func runtimeCreatePayloadInvalid(username, password string, generate bool) bool {
	if runtimecred.ValidateUsername(username) != nil {
		return true
	}
	if generate {
		return password != ""
	}
	return runtimecred.ValidatePassword(password, false) != nil
}

func runtimeUpdatePayloadInvalid(username string, status runtimecred.CredentialStatus) bool {
	if runtimecred.ValidateUsername(username) != nil {
		return true
	}
	return status != runtimecred.CredentialActive && status != runtimecred.CredentialDisabled
}

func runtimeRotatePayloadInvalid(password string, generate bool) bool {
	if generate {
		return password != ""
	}
	return runtimecred.ValidatePassword(password, false) != nil
}

func commitRuntimeMutation(r *http.Request, mutation *runtimecred.Mutation) error {
	authenticated, ok := authenticatedFromRequest(r)
	if !ok {
		return errors.New("authenticated transaction missing")
	}
	err := mutation.CommitAndFinalize(r.Context(), authenticated.Bound.Tx)
	if err != nil {
		return err
	}
	authenticated.TransactionFinalized = true
	return nil
}

func finishRuntimeTransaction(r *http.Request) {
	if authenticated, ok := authenticatedFromRequest(r); ok {
		_ = authenticated.Bound.Tx.Rollback()
		authenticated.TransactionFinalized = true
	}
}

func writeRuntimeInvalidRequest(w http.ResponseWriter) {
	writeJSON(w, http.StatusBadRequest, envelope{"code": "invalid_request", "message": "Invalid runtime credential request."})
}

func writeRuntimeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, runtimecred.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, envelope{"code": "revision_conflict", "message": "Credential revision is stale. Refresh and retry."})
	case errors.Is(err, runtimecred.ErrUsernameConflict):
		writeJSON(w, http.StatusConflict, envelope{"code": "username_conflict", "message": "That username already exists."})
	case errors.Is(err, runtimecred.ErrLastActiveCredential):
		writeJSON(w, http.StatusConflict, envelope{"code": "last_active_credential", "message": "At least one active credential must remain."})
	case errors.Is(err, runtimecred.ErrRuntimeAlreadyOwned):
		writeJSON(w, http.StatusConflict, envelope{"code": "runtime_already_owned", "message": "Live runtime credentials are already imported."})
	case errors.Is(err, runtimecred.ErrImportEquivalence):
		writeJSON(w, http.StatusConflict, envelope{"code": "runtime_import_equivalence_failed", "message": "Import was stopped because reconstructed Caddy state was not byte-equivalent to live state."})
	case errors.Is(err, runtimecred.ErrIdempotentReplay):
		writeJSON(w, http.StatusConflict, envelope{"code": "idempotency_replay", "message": "This mutation key has already been used. Refresh runtime state."})
	case errors.Is(err, runtimecred.ErrReconciliationRequired):
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "runtime_reconciliation_required", "message": "Runtime rollback failed; reconciliation is required before further mutations."})
	case errors.Is(err, runtimecred.ErrConsistency):
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "consistency_error", "message": "Runtime mutation was not committed consistently; rollback was attempted."})
	default:
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "runtime_unavailable", "message": "Runtime mutation is unavailable."})
	}
}
