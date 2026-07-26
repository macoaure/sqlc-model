package golden

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/macoaure/sqlc-model/internal/generate"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// TestExtensionFile_CreateOnceNeverOverwritten validates User Story 3's
// independent test end to end: a model is generated, a handwritten method
// is added to its extension file, an unrelated field-policy change is made,
// and regeneration leaves the extension file untouched while generated
// files reflect the change (FR-015, FR-016).
func TestExtensionFile_CreateOnceNeverOverwritten(t *testing.T) {
	outDir := t.TempDir()
	req := userBasicRequest(t)
	req.Settings.Codegen.Out = outDir // absolute path: outputRoot() joins cwd+Out, so make Out absolute-as-is

	// First run: content/user.go doesn't exist yet, so it must be emitted.
	resp1, diags1 := generate.Generate(req)
	if resp1 == nil {
		t.Fatalf("first generation failed: %v", diags1)
	}
	extPath := "content/user.go"
	if !containsFile(resp1.Files, extPath) {
		t.Fatalf("expected %s to be emitted on first run, got %v", extPath, fileNames(resp1.Files))
	}
	writeFiles(t, outDir, resp1.Files)

	// Simulate the developer adding a handwritten method.
	fullExtPath := filepath.Join(outDir, filepath.FromSlash(extPath))
	original, err := os.ReadFile(fullExtPath)
	if err != nil {
		t.Fatal(err)
	}
	handwritten := string(original) + "\nfunc (u *User) Greeting() string { return \"hi \" + u.Name() }\n"
	if err := os.WriteFile(fullExtPath, []byte(handwritten), 0o644); err != nil {
		t.Fatal(err)
	}

	// Unrelated field-policy change: mark email sensitive.
	req.PluginOptions = []byte(strings.Replace(string(req.PluginOptions),
		`"email": {"readable": true, "fillable": true, "mutable": true}`,
		`"email": {"readable": true, "fillable": true, "mutable": true, "sensitive": true}`, 1))

	resp2, diags2 := generate.Generate(req)
	if resp2 == nil {
		t.Fatalf("second generation failed: %v", diags2)
	}

	if containsFile(resp2.Files, extPath) {
		t.Fatalf("expected %s to be omitted from the response once it already exists on disk (FR-015), got %v", extPath, fileNames(resp2.Files))
	}

	afterRegen, err := os.ReadFile(fullExtPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterRegen) != handwritten {
		t.Fatalf("extension file was modified by regeneration; handwritten content must survive untouched.\nbefore:\n%s\nafter:\n%s", handwritten, afterRegen)
	}

	var modelSrc string
	for _, f := range resp2.Files {
		if f.Name == "content/user_gen.go" {
			modelSrc = string(f.Contents)
		}
	}
	if !strings.Contains(modelSrc, `"Email=<redacted>"`) {
		t.Fatalf("expected regenerated model to reflect the sensitive field-policy change, got:\n%s", modelSrc)
	}
}

func TestCompileMatrixFixturesRegistered(t *testing.T) {
	for _, dir := range []string{
		"identifier-styles",
		"type-matrix",
		"query-matrix",
		"config-matrix",
		"relation-session-matrix",
	} {
		if _, err := os.Stat(filepath.Join("..", "compile", dir, "go.mod")); err != nil {
			t.Fatalf("compile matrix fixture %s is not registered: %v", dir, err)
		}
	}
}

func containsFile(files []*pb.File, name string) bool {
	for _, f := range files {
		if f.Name == name {
			return true
		}
	}
	return false
}

func fileNames(files []*pb.File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Name)
	}
	return out
}

func writeFiles(t *testing.T, dir string, files []*pb.File) {
	t.Helper()
	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, f.Contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
