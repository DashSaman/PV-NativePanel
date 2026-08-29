package telemetry

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountingMigrationManifestHashes(t *testing.T) {
	root := filepath.Join("..", "..")
	manifestPath := filepath.Join(root, "db", "migrations", "SHA256SUMS")
	manifestFile, err := os.Open(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	defer manifestFile.Close()

	manifest := map[string]string{}
	scanner := bufio.NewScanner(manifestFile)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 {
			manifest[fields[1]] = fields[0]
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"0009_direct_naive_exact_accounting.up.sql",
		"0009_direct_naive_exact_accounting.down.sql",
		"0010_pending_reservation_completeness.up.sql",
		"0010_pending_reservation_completeness.down.sql",
	} {
		data, err := os.ReadFile(filepath.Join(root, "db", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		actual := fmt.Sprintf("%x", sha256.Sum256(data))
		expected, ok := manifest[name]
		if !ok {
			t.Fatalf("migration manifest missing %s sha256=%s", name, actual)
		}
		if expected != actual {
			t.Fatalf("migration manifest mismatch %s expected=%s actual=%s", name, expected, actual)
		}
	}
}
