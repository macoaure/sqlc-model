package unit

import (
	"regexp"
	"strings"
	"testing"

	"github.com/macoaure/sqlc-model/internal/codegen"
	"github.com/macoaure/sqlc-model/internal/diagnostics"
	"github.com/macoaure/sqlc-model/internal/plan"
)

// errsAssignmentPattern matches any assignment/mutation of the model's
// unexported error map (u.errs[...] = ..., delete(u.errs, ...), u.errs =
// ...). Every match must originate inside setFieldError/clearFieldError/
// ClearErrors — this is what makes those the *sole* sanctioned mutation
// path for handwritten extension code (FR-019), not merely a convention.
var errsAssignmentPattern = regexp.MustCompile(`u\.errs(\[[^\]]*\]\s*=|\s*=|,)`)

func TestGeneratedModel_ErrorMutationConfinedToSanctionedHelpers(t *testing.T) {
	src := generateSecretModel(t)

	allowedFuncs := []string{"func (u *Thing) setFieldError(", "func (u *Thing) clearFieldError(", "func (u *Thing) ClearErrors("}

	funcs := splitFuncs(src)
	for _, fn := range funcs {
		if !errsAssignmentPattern.MatchString(fn) {
			continue
		}
		allowed := false
		for _, prefix := range allowedFuncs {
			if strings.HasPrefix(fn, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			t.Fatalf("found u.errs mutation outside setFieldError/clearFieldError/ClearErrors:\n%s", fn)
		}
	}
}

// splitFuncs splits a Go source file into per-top-level-func chunks, purely
// by scanning for lines starting with "func ".
func splitFuncs(src string) []string {
	lines := strings.Split(src, "\n")
	var funcs []string
	var current []string
	for _, line := range lines {
		if strings.HasPrefix(line, "func ") {
			if len(current) > 0 {
				funcs = append(funcs, strings.Join(current, "\n"))
			}
			current = []string{line}
			continue
		}
		if len(current) > 0 {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		funcs = append(funcs, strings.Join(current, "\n"))
	}
	return funcs
}

func TestGeneratedSession_DatabaseErrorContract(t *testing.T) {
	src := generateSessionSource(t, false)
	for _, want := range []string{
		"type DatabaseErrorKind uint8",
		"DatabaseErrorUniqueViolation",
		"DatabaseErrorForeignKeyViolation",
		"DatabaseErrorNotNullViolation",
		"DatabaseErrorCheckViolation",
		"DatabaseErrorSerializationFailure",
		"DatabaseErrorDeadlock",
		"type DatabaseError struct",
		"func (e *DatabaseError) Unwrap() error",
		"func classifyDatabaseError(err error) error",
		`case "23505":`,
		`case "23503":`,
		`case "23502":`,
		`case "23514":`,
		`case "40001":`,
		`case "40P01":`,
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated session to contain %q:\n%s", want, src)
		}
	}

}

func generateSessionSource(t *testing.T, hasRelations bool) string {
	t.Helper()
	ctx := plan.ResolvedContext{
		Name:      "content",
		Package:   "content",
		Directory: "content",
		Models:    []plan.ResolvedModel{{Name: "User", Row: "User"}},
	}
	if hasRelations {
		ctx.Models[0].Relations = []plan.ResolvedRelation{{Name: "Posts"}}
	}
	out, diags := codegen.RenderSession(ctx)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	return string(out)
}
