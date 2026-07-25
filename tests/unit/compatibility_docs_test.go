package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompatibilityDocsCoverGuaranteedTypesAndDeferredPlatforms(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "content", "reference", "compatibility.md")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(body)
	for _, want := range []string{
		"Go 1.25.0",
		"sqlc configuration version 2",
		"PostgreSQL",
		"pgx/v5",
		"database/sql",
		"MySQL",
		"SQLite",
		"UUIDs",
		"integer identifiers",
		"timestamps",
		"booleans",
		"text",
		"numeric values",
		"JSON",
		"JSONB",
		"byte arrays",
		"arrays",
		"enums",
		"nullable values",
		"custom overrides",
		"renamed fields",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("compatibility docs missing %q:\n%s", want, doc)
		}
	}
}
