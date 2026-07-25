package unit

import (
	"regexp"
	"strings"
	"testing"
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
