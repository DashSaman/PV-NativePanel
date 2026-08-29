package customer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProductListQueryUsesRealResellerKey(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	sourcePath := filepath.Join(filepath.Dir(filename), "product_list_store.go")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read product list store: %v", err)
	}
	text := string(source)
	if strings.Contains(text, "r.id") {
		t.Fatal("product list query references pvnaive.resellers.id, but resellers is keyed by tenant_id/primary_actor_id")
	}
	if !strings.Contains(text, "r.primary_actor_id") {
		t.Fatal("product list query must expose/filter reseller identity through primary_actor_id")
	}
}
