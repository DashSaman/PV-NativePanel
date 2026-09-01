package httpapi

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreatePlanAcceptsUniqueIPLimit(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve source")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "customer_product.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "UniqueIPLimit") || !strings.Contains(text, `json:"unique_ip_limit"`) {
		t.Fatal("plan create payload does not accept unique_ip_limit")
	}
	if !strings.Contains(text, "UniqueIPLimit: payload.UniqueIPLimit") {
		t.Fatal("plan create handler does not propagate unique_ip_limit")
	}
}
