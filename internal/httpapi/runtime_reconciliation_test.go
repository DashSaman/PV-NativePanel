package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func TestRuntimeReconciliationRequiredHasDedicatedAPIError(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeRuntimeServiceError(recorder, runtimecred.ErrReconciliationRequired)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"code":"runtime_reconciliation_required"`) {
		t.Fatalf("body = %s, want dedicated reconciliation code", body)
	}
}
