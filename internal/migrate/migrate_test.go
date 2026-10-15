package migrate

import (
	"strings"
	"testing"

	"github.com/okoye-dev/cherryl-media-engine/db"
)

func TestMigrationFilesDiscovered(t *testing.T) {
	files, err := migrationFiles()
	if err != nil {
		t.Fatalf("failed to read migration files: %v", err)
	}

	if len(files) < 4 {
		t.Fatalf("expected at least 4 migration files, got %d: %v", len(files), files)
	}

	// Verify sequential naming
	expectedPrefixes := []string{"0001_", "0002_", "0003_", "0004_"}
	for i, prefix := range expectedPrefixes {
		if !strings.HasPrefix(files[i], prefix) {
			t.Errorf("expected migration [%d] to have prefix %s, got %s", i, prefix, files[i])
		}
	}

	// Verify each file can be read and is non-empty
	for _, name := range files {
		content, err := db.Migrations.ReadFile("migrations/" + name)
		if err != nil {
			t.Errorf("failed to read migration %s: %v", name, err)
		}
		if len(strings.TrimSpace(string(content))) == 0 {
			t.Errorf("migration %s is empty", name)
		}
	}
}
