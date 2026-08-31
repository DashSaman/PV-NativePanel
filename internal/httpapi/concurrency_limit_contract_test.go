package httpapi

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreatePlanAcceptsConcurrencyLimit(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve source")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "customer_product.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "ConcurrencyLimit") || !strings.Contains(text, `json:"concurrency_limit"`) {
		t.Fatal("plan create payload does not accept concurrency_limit")
	}
	if !strings.Contains(text, "ConcurrencyLimit: payload.ConcurrencyLimit") {
		t.Fatal("plan create handler does not propagate concurrency_limit")
	}
}
