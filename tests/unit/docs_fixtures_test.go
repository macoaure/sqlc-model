package unit

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentationSQLAndConfigSnippetsAreNonEmpty(t *testing.T) {
	snippets := findDocSnippets(t, filepath.Join("..", "..", "docs", "content"))
	found := map[string]bool{}
	for _, snippet := range snippets {
		switch snippet.Lang {
		case "sql":
			found["sql"] = true
			upper := strings.ToUpper(snippet.Body)
			if !strings.Contains(upper, "CREATE") && !strings.Contains(upper, "SELECT") && !strings.Contains(upper, "INSERT") && !strings.Contains(upper, "UPDATE") && !strings.Contains(upper, "DELETE") {
				t.Fatalf("SQL snippet in %s has no recognizable statement:\n%s", snippet.Path, snippet.Body)
			}
		case "yaml", "yml":
			found["yaml"] = true
			if !strings.Contains(snippet.Body, ":") {
				t.Fatalf("YAML snippet in %s has no key/value separator:\n%s", snippet.Path, snippet.Body)
			}
		}
	}
	for _, kind := range []string{"sql", "yaml"} {
		if !found[kind] {
			t.Fatalf("expected at least one %s documentation snippet", kind)
		}
	}
}
