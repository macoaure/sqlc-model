package unit

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-model/internal/diagnostics"
)

func TestSeverity_String(t *testing.T) {
	if got := diagnostics.SeverityError.String(); got != "error" {
		t.Fatalf("expected %q, got %q", "error", got)
	}
	if got := diagnostics.SeverityWarning.String(); got != "warning" {
		t.Fatalf("expected %q, got %q", "warning", got)
	}
}

func TestSort(t *testing.T) {
	diags := []diagnostics.Diagnostic{
		{Path: "b", Context: "z", Model: "z", Relation: "z", Query: "z"},
		{Path: "a", Context: "z", Model: "z", Relation: "b", Query: "z"},
		{Path: "a", Context: "z", Model: "z", Relation: "a", Query: "z"},
		{Path: "a", Context: "y", Model: "z", Relation: "z", Query: "z"},
		{Path: "a", Context: "z", Model: "y", Relation: "z", Query: "z"},
	}

	sorted := diagnostics.Sort(diags)

	want := []string{"a:y:z:z:z", "a:z:y:z:z", "a:z:z:a:z", "a:z:z:b:z", "b:z:z:z:z"}
	got := make([]string, len(sorted))
	for i, d := range sorted {
		got[i] = strings.Join([]string{d.Path, d.Context, d.Model, d.Relation, d.Query}, ":")
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q (full: %v)", i, want[i], got[i], got)
		}
	}
}

func TestHasError(t *testing.T) {
	if diagnostics.HasError(nil) {
		t.Fatal("expected no error for nil slice")
	}
	if diagnostics.HasError([]diagnostics.Diagnostic{{Severity: diagnostics.SeverityWarning}}) {
		t.Fatal("expected no error for warning-only slice")
	}
	if !diagnostics.HasError([]diagnostics.Diagnostic{
		{Severity: diagnostics.SeverityWarning},
		{Severity: diagnostics.SeverityError},
	}) {
		t.Fatal("expected error to be detected")
	}
}

func TestFormat(t *testing.T) {
	d := diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Path:     "contexts.content.models.User.operations.insert",
		Message:  "query CreateUser uses :exec, but the default insert lifecycle requires :one",
	}
	got := diagnostics.Format(d)
	want := "error: contexts.content.models.User.operations.insert\n\nquery CreateUser uses :exec, but the default insert lifecycle requires :one"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	dWithHint := diagnostics.Diagnostic{
		Severity: diagnostics.SeverityWarning,
		Path:     "some.path",
		Message:  "some message",
		Hint:     "some hint",
	}
	got = diagnostics.Format(dWithHint)
	want = "warning: some.path\n\nsome message\n\nHint: some hint"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFormatAll(t *testing.T) {
	if got := diagnostics.FormatAll(nil); got != "" {
		t.Fatalf("expected empty string for nil slice, got %q", got)
	}

	diags := []diagnostics.Diagnostic{
		{Severity: diagnostics.SeverityError, Path: "a", Message: "first"},
		{Severity: diagnostics.SeverityWarning, Path: "b", Message: "second", Hint: "fix it"},
	}
	got := diagnostics.FormatAll(diags)
	want := "error: a\n\nfirst\n\n---\n\nwarning: b\n\nsecond\n\nHint: fix it"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
