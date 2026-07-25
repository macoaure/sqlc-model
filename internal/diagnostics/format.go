package diagnostics

import (
	"fmt"
	"strings"
)

// Format renders a single Diagnostic in the textual shape documented by
// contracts/diagnostics-contract.md's example:
//
//	error: contexts.content.models.User.operations.insert
//
//	query CreateUser uses :exec, but the default insert lifecycle requires
//	:one so the persisted row can hydrate generated fields.
//
//	Hint: add RETURNING ... and change the annotation to :one, or select an
//	explicit non-hydrating insert policy when that policy becomes supported.
func Format(d Diagnostic) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s\n\n%s", d.Severity, d.Path, d.Message)
	if d.Hint != "" {
		fmt.Fprintf(&b, "\n\nHint: %s", d.Hint)
	}
	return b.String()
}

// FormatAll renders a sorted diagnostic list as the plugin's full textual
// error output, one diagnostic per paragraph.
func FormatAll(diags []Diagnostic) string {
	parts := make([]string, len(diags))
	for i, d := range diags {
		parts[i] = Format(d)
	}
	return strings.Join(parts, "\n\n---\n\n")
}
