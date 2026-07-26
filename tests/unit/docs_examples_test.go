package unit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type docSnippet struct {
	Path string
	Lang string
	Body string
}

func findDocSnippets(t *testing.T, root string) []docSnippet {
	t.Helper()

	var snippets []docSnippet
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(string(body), "\n")
		for i := 0; i < len(lines); i++ {
			line := lines[i]
			if !strings.HasPrefix(line, "```") || len(line) < 4 {
				continue
			}
			lang := strings.TrimSpace(strings.TrimPrefix(line, "```"))
			var chunk []string
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				chunk = append(chunk, lines[i])
			}
			snippets = append(snippets, docSnippet{Path: path, Lang: lang, Body: strings.Join(chunk, "\n")})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snippets
}

func TestFindDocSnippetsIncludesSourcePaths(t *testing.T) {
	snippets := findDocSnippets(t, filepath.Join("..", "..", "docs", "content"))
	if len(snippets) == 0 {
		t.Fatal("expected at least one documentation snippet")
	}
	for _, snippet := range snippets {
		if snippet.Path == "" || snippet.Body == "" {
			t.Fatalf("snippet missing source path or body: %+v", snippet)
		}
	}
}

func TestPublicDocumentationExamplesCompile(t *testing.T) {
	dir := filepath.Join("..", "compile", "docs-examples")
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("documentation examples fixture failed: %v\n%s", err, out)
	}
}
