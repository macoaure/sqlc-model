package config

import (
	"fmt"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
)

// Operations maps each lifecycle operation kind to the sqlc query name that
// implements it. An empty string means that operation was not configured
// (no corresponding generated method); Insert is the only operation
// required to be non-empty.
type Operations struct {
	Find    string
	Insert  string
	Update  string
	Delete  string
	Refresh string
}

// ModelConfiguration is the mapping for one database entity: its canonical
// sqlc row/result type, its lifecycle-operation-to-query mapping, and its
// field policies (data-model.md "ModelConfiguration").
type ModelConfiguration struct {
	Name       string
	Row        string
	Operations Operations
	Fields     []FieldPolicy
}

type operationsJSON struct {
	Find    string `json:"find"`
	Insert  string `json:"insert"`
	Update  string `json:"update"`
	Delete  string `json:"delete"`
	Refresh string `json:"refresh"`
}

type modelConfigJSON struct {
	Row        string         `json:"row"`
	Operations operationsJSON `json:"operations"`
	Fields     OrderedObject  `json:"fields"`
}

// decodeModel decodes and validates one model entry. contextPath is the
// path prefix for this model's context, e.g. "contexts.content".
func decodeModel(name string, raw []byte, contextPath, contextName string) (ModelConfiguration, []diagnostics.Diagnostic) {
	path := fmt.Sprintf("%s.models.%s", contextPath, name)

	var mj modelConfigJSON
	if err := DecodeStrict(raw, &mj); err != nil {
		return ModelConfiguration{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  contextName,
			Model:    name,
			Message:  fmt.Sprintf("model %q: invalid model configuration: %v", name, err),
		}}
	}

	var diags []diagnostics.Diagnostic

	if mj.Row == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".row",
			Context:  contextName,
			Model:    name,
			Message:  fmt.Sprintf("model %q: \"row\" is required", name),
		})
	}
	if mj.Operations.Insert == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".operations.insert",
			Context:  contextName,
			Model:    name,
			Message:  fmt.Sprintf("model %q: \"operations.insert\" is required — a model that can never be created is not meaningful", name),
		})
	}

	mc := ModelConfiguration{
		Name: name,
		Row:  mj.Row,
		Operations: Operations{
			Find:    mj.Operations.Find,
			Insert:  mj.Operations.Insert,
			Update:  mj.Operations.Update,
			Delete:  mj.Operations.Delete,
			Refresh: mj.Operations.Refresh,
		},
	}

	seen := make(map[string]bool, len(mj.Fields))
	for _, entry := range mj.Fields {
		if seen[entry.Key] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fmt.Sprintf("%s.fields.%s", path, entry.Key),
				Context:  contextName,
				Model:    name,
				Message:  fmt.Sprintf("model %q: duplicate field %q", name, entry.Key),
			})
			continue
		}
		seen[entry.Key] = true

		fieldPath := fmt.Sprintf("%s.fields.%s", path, entry.Key)
		fp, fdiags := decodeField(entry.Key, entry.Value, fieldPath, contextName, name)
		diags = append(diags, fdiags...)
		mc.Fields = append(mc.Fields, fp)
	}

	diags = append(diags, validateModelFields(mc, path, contextName)...)

	return mc, diags
}

// validateModelFields applies cross-field rules that need the whole
// model's field list: at most one version field per model (research.md
// "Field-policy combination validity").
func validateModelFields(mc ModelConfiguration, path, contextName string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	var versionFields []string
	for _, f := range mc.Fields {
		if f.Version {
			versionFields = append(versionFields, f.Name)
		}
	}
	if len(versionFields) > 1 {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".fields",
			Context:  contextName,
			Model:    mc.Name,
			Message:  fmt.Sprintf("model %q: at most one field may be marked version, found %d: %v", mc.Name, len(versionFields), versionFields),
		})
	}
	return diags
}
