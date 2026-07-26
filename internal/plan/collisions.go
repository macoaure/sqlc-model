package plan

import (
	"fmt"

	"github.com/macoaure/sqlc-model/internal/diagnostics"
)

// checkContextCollisions detects two generated declarations within the same
// context that would collide: currently, two models sharing the same `row`
// (generated struct name), which would emit `type Row struct` twice into
// the same Go package.
func checkContextCollisions(contextName string, rc ResolvedContext) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	seenRow := make(map[string]string, len(rc.Models)) // row name -> first model that claimed it
	for _, m := range rc.Models {
		if m.Row == "" {
			continue
		}
		if first, ok := seenRow[m.Row]; ok {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fmt.Sprintf("contexts.%s.models.%s.row", contextName, m.Name),
				Context:  contextName,
				Model:    m.Name,
				Message:  fmt.Sprintf("models %q and %q both declare row %q in context %q — the generated struct name would collide", first, m.Name, m.Row, contextName),
			})
			continue
		}
		seenRow[m.Row] = m.Name
	}
	return diags
}
