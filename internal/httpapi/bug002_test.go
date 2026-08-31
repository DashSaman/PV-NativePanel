package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeCommitTx implements transactionCommit for testing finalize.
type fakeCommitTx struct {
	commitErr  error
	committed  bool
	rolledBack bool
}

func (f *fakeCommitTx) Commit() error {
	f.committed = true
	if f.commitErr != nil {
		return f.commitErr
	}
	return nil
}

func (f *fakeCommitTx) Rollback() error {
	f.rolledBack = true
	return nil
}

// --- RED: responseBuffer must NOT satisfy http.Flusher ---

func TestResponseBufferDoesNotSatisfyFlusher(t *testing.T) {
	var buf *responseBuffer
	if _, ok := interface{}(buf).(http.Flusher); ok {
		t.Fatal("BUG-002: *responseBuffer must NOT satisfy http.Flusher; Flush method must be renamed to commitToClient")
	}
}

// --- RED: finalize with commit error: buffered 200 never reaches client ---

func TestFinalizeCommitErrorDiscardsBufferedSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	tx := &fakeCommitTx{commitErr: errors.New("commit failed")}

	writeJSON(buf, http.StatusOK, envelope{"status": "ok", "service_term": "term-1"})

	err := finalize(tx, buf, rec)
	if err == nil {
		t.Fatal("expected error from finalize on commit failure")
	}

	if rec.Code == http.StatusOK {
		t.Fatalf("BUG-002: 200 status leaked to client on commit failure: got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] == "ok" {
		t.Fatalf("BUG-002: success body leaked to client on commit failure: %v", body)
	}
	if body["code"] != "transaction_commit_failed" {
		t.Fatalf("expected transaction_commit_failed error, got: %v", body)
	}
}

// --- RED: finalize with commit error: 500 commit-failed response DOES reach client ---

func TestFinalizeCommitErrorWrites500ToClient(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	tx := &fakeCommitTx{commitErr: errors.New("commit failed")}

	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})

	finalize(tx, buf, rec)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on commit failure, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "transaction_commit_failed" {
		t.Fatalf("expected transaction_commit_failed error, got: %v", body)
	}
	if body["status"] == "ok" {
		t.Fatalf("BUG-002: success body still present in error response")
	}
}

// --- RED: finalize success path: buffered 200 DOES reach client ---

func TestFinalizeSuccessFlushesBufferedResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	tx := &fakeCommitTx{}

	writeJSON(buf, http.StatusOK, envelope{"status": "ok", "service_term": "term-1"})

	err := finalize(tx, buf, rec)
	if err != nil {
		t.Fatalf("unexpected error from finalize on commit success: %v", err)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected ok status, got: %v", body)
	}
}

// --- Existing tests (adapted: Flush → commitToClient) ---

func TestResponseBufferCapturesWritesWithoutLeakingToUnderlyingWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})
	if rec.Body.Len() != 0 {
		t.Fatalf("body leaked to underlying writer: %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "" {
		t.Fatalf("headers leaked to underlying writer: %v", rec.Header())
	}
}

func TestResponseBufferCapturesStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	writeJSON(buf, http.StatusCreated, envelope{"created": true})
	if buf.statusCode != http.StatusCreated {
		t.Fatalf("statusCode=%d want %d", buf.statusCode, http.StatusCreated)
	}
}

func TestResponseBufferCapturesHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	buf.Header().Set("X-Custom", "value")
	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})
	if buf.header.Get("X-Custom") != "value" {
		t.Fatalf("header not captured: %v", buf.header)
	}
}

func TestResponseBufferCommitToClientWritesToUnderlying(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})
	buf.commitToClient()
	if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("headers not flushed: %v", rec.Header())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body not flushed: %v", body)
	}
}

func TestResponseBufferDiscardDoesNotWriteToUnderlying(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})
	buf.Discard()
	if rec.Body.Len() != 0 {
		t.Fatalf("body leaked after discard: %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "" {
		t.Fatalf("headers leaked after discard: %v", rec.Header())
	}
}

func TestResponseBufferCommitToClientPreservesHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	buf.Header().Set("X-Request-ID", "test-123")
	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})
	buf.commitToClient()
	if rec.Header().Get("X-Request-ID") != "test-123" {
		t.Fatalf("header not flushed: %v", rec.Header())
	}
}

func TestCommitFailureAfterMutationHandlerSuccessLeaksNoBody(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}

	writeJSON(buf, http.StatusOK, envelope{"status": "ok", "runtime_mutated": false})

	buf.Discard()

	if rec.Body.Len() != 0 {
		t.Fatalf("BUG-002: success body leaked on commit failure: %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "" {
		t.Fatalf("BUG-002: headers leaked on commit failure: %v", rec.Header())
	}
}

func TestCommitSuccessAfterMutationHandlerFlushesBody(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}

	writeJSON(buf, http.StatusOK, envelope{"status": "ok", "service_term": "term-1"})

	buf.commitToClient()

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected ok status, got: %v", body)
	}
}

func TestTransactionFinalizedBypassesMiddlewareCommitAndFlushes(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}

	writeJSON(buf, http.StatusOK, envelope{"status": "suspended"})

	authenticated := &authenticatedRequest{TransactionFinalized: true}
	if authenticated.TransactionFinalized {
		buf.commitToClient()
	}

	if rec.Body.Len() == 0 {
		t.Fatal("TransactionFinalized path should flush body")
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "suspended" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestMutationHandlerErrorPathWritesErrorNotSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}

	writeJSON(buf, http.StatusServiceUnavailable, envelope{"code": "operation_failed", "message": "Operation could not be completed."})
	buf.commitToClient()

	if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("error headers not flushed")
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "operation_failed" {
		t.Fatalf("expected error code, got: %v", body)
	}
}

func TestResponseBufferWriteBeforeWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	buf.Write([]byte(`{"status":"ok"}`))
	if buf.statusCode != http.StatusOK {
		t.Fatalf("implicit WriteHeader(200) not set, got statusCode=%d", buf.statusCode)
	}
	buf.commitToClient()
	if rec.Header().Get("Content-Type") != "" {
		// Content-Type is empty because no header was set before commitToClient
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body not flushed: %v", body)
	}
}

func TestResponseBufferDoubleWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	buf.WriteHeader(http.StatusCreated)
	buf.WriteHeader(http.StatusConflict)
	if buf.statusCode != http.StatusCreated {
		t.Fatalf("second WriteHeader overwrote first: %d", buf.statusCode)
	}
}

func TestMiddlewareCommitFailureAfterSuccessWrites500Not200(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}

	writeJSON(buf, http.StatusOK, envelope{"status": "ok"})

	tx := &fakeCommitTx{commitErr: errors.New("commit failed")}
	finalize(tx, buf, rec)

	if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("error headers missing")
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "transaction_commit_failed" {
		t.Fatalf("expected commit_failed error, got: %v", body)
	}
	if body["status"] == "ok" {
		t.Fatalf("BUG-002: success body still present in error response")
	}
}

func TestResponseBufferEmptyBodyCommitToClient(t *testing.T) {
	rec := httptest.NewRecorder()
	buf := &responseBuffer{w: rec}
	buf.WriteHeader(http.StatusNoContent)
	buf.commitToClient()
	if rec.Body.Len() != 0 {
		t.Fatalf("empty body commitToClient wrote bytes: %s", rec.Body.String())
	}
}
