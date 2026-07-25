package config

import (
	"fmt"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
)

// BoundedContext is a named grouping of related models sharing one Go
// package and output directory (data-model.md "BoundedContext").
type BoundedContext struct {
	Name      string
	Package   string
	Directory string
	Models    []ModelConfiguration
}

type contextJSON struct {
	Name      string        `json:"name"`
	Package   string        `json:"package"`
	Directory string        `json:"directory"`
	Models    OrderedObject `json:"models"`
}

// decodeContext decodes and validates one bounded context. index identifies
// its position for diagnostics when the context is malformed before its
// name can even be read.
func decodeContext(raw []byte, index int) (BoundedContext, []diagnostics.Diagnostic) {
	genericPath := fmt.Sprintf("contexts[%d]", index)

	var cj contextJSON
	if err := DecodeStrict(raw, &cj); err != nil {
		return BoundedContext{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     genericPath,
			Message:  fmt.Sprintf("invalid bounded context: %v", err),
		}}
	}

	path := fmt.Sprintf("contexts.%s", cj.Name)
	if cj.Name == "" {
		path = genericPath
	}

	var diags []diagnostics.Diagnostic
	if cj.Name == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".name",
			Message:  "bounded context: \"name\" is required",
		})
	}
	if cj.Package == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".package",
			Context:  cj.Name,
			Message:  fmt.Sprintf("context %q: \"package\" is required", cj.Name),
		})
	}
	if cj.Directory == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".directory",
			Context:  cj.Name,
			Message:  fmt.Sprintf("context %q: \"directory\" is required", cj.Name),
		})
	}

	bc := BoundedContext{Name: cj.Name, Package: cj.Package, Directory: cj.Directory}

	seen := make(map[string]bool, len(cj.Models))
	for _, entry := range cj.Models {
		if seen[entry.Key] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fmt.Sprintf("%s.models.%s", path, entry.Key),
				Context:  cj.Name,
				Message:  fmt.Sprintf("context %q: duplicate model %q", cj.Name, entry.Key),
			})
			continue
		}
		seen[entry.Key] = true

		mc, mdiags := decodeModel(entry.Key, entry.Value, path, cj.Name)
		diags = append(diags, mdiags...)
		bc.Models = append(bc.Models, mc)
	}

	return bc, diags
}
