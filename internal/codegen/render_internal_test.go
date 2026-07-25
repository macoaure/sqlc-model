package codegen

import (
	"path/filepath"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"
)

// These exercise render's four failure branches directly: none of them are
// reachable through the package's exported Render* entrypoints, since every
// one of those always supplies a fixed, valid template and well-typed data.

func TestRender_TemplateParseError(t *testing.T) {
	_, diags := render("path.go", "broken", "{{", nil, "ctx", "Model")
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected a template parse error diagnostic, got %+v", diags)
	}
}

func TestRender_TemplateExecuteError(t *testing.T) {
	_, diags := render("path.go", "missingfield", "{{.Nonexistent}}", struct{}{}, "ctx", "Model")
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected a template execute error diagnostic, got %+v", diags)
	}
}

func TestRender_FormatSourceError(t *testing.T) {
	_, diags := render("path.go", "badgo", "package {{.}}\nfunc (", "foo", "ctx", "Model")
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected a format.Source error diagnostic, got %+v", diags)
	}
}

// TestRender_ImportsProcessError covers imports.Process's error branch: an
// empty template produces a nil formatted byte slice (bytes.Buffer.Bytes()
// on an untouched buffer is nil, and format.Source(nil) succeeds with a nil
// result too), and imports.Process falls back to reading filePath from disk
// when handed a nil source — which fails for a path that doesn't exist.
func TestRender_ImportsProcessError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does", "not", "exist.go")
	_, diags := render(missing, "empty", "", nil, "ctx", "Model")
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an imports.Process error diagnostic, got %+v", diags)
	}
}

func TestLowerFirst(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"ID":   "id",
		"Name": "name",
	}
	for in, want := range cases {
		if got := lowerFirst(in); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSortedParams(t *testing.T) {
	in := []plan.ParamBinding{
		{Number: 3},
		{Number: 1},
		{Number: 2},
	}
	out := sortedParams(in)
	for i, want := range []int32{1, 2, 3} {
		if out[i].Number != want {
			t.Fatalf("sortedParams order = %+v, want ascending by Number", out)
		}
	}
}
