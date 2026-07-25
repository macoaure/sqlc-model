package config

import (
	"fmt"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
)

// GeneratedKind describes when a field's value is supplied by the database
// rather than the caller.
type GeneratedKind int

const (
	// GeneratedNone means the field is not database-generated.
	GeneratedNone GeneratedKind = iota
	// GeneratedInsert means the field is populated only from the insert
	// operation's RETURNING row.
	GeneratedInsert
	// GeneratedSave means the field may be repopulated on both insert and
	// update.
	GeneratedSave
)

// FieldPolicy is the per-column configuration described by FR-005: what API
// surface a field gets, how it's populated, and how it's treated in
// diagnostics.
type FieldPolicy struct {
	Name                 string
	Readable             bool
	Fillable             bool
	Mutable              bool
	Generated            GeneratedKind
	ImmutableAfterInsert bool
	Sensitive            bool
	Version              bool
	// Column and RowField are explicit overrides; empty means "resolve
	// automatically" (internal/mapping's responsibility).
	Column   string
	RowField string
}

type fieldPolicyJSON struct {
	Readable             bool   `json:"readable"`
	Fillable             bool   `json:"fillable"`
	Mutable              bool   `json:"mutable"`
	Generated            string `json:"generated"`
	ImmutableAfterInsert bool   `json:"immutable_after_insert"`
	Sensitive            bool   `json:"sensitive"`
	Version              bool   `json:"version"`
	Column               string `json:"column"`
	RowField             string `json:"row_field"`
}

// decodeField decodes and validates one field entry. path identifies this
// field's location for diagnostic reporting, e.g.
// "contexts.content.models.User.fields.name".
func decodeField(name string, raw []byte, path, context, model string) (FieldPolicy, []diagnostics.Diagnostic) {
	var fj fieldPolicyJSON
	if err := DecodeStrict(raw, &fj); err != nil {
		return FieldPolicy{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: invalid field policy: %v", name, err),
		}}
	}

	fp := FieldPolicy{
		Name:                 name,
		Readable:             fj.Readable,
		Fillable:             fj.Fillable,
		Mutable:              fj.Mutable,
		ImmutableAfterInsert: fj.ImmutableAfterInsert,
		Sensitive:            fj.Sensitive,
		Version:              fj.Version,
		Column:               fj.Column,
		RowField:             fj.RowField,
	}

	var diags []diagnostics.Diagnostic
	switch fj.Generated {
	case "":
		fp.Generated = GeneratedNone
	case "insert":
		fp.Generated = GeneratedInsert
	case "save":
		fp.Generated = GeneratedSave
	default:
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: unsupported generated value %q, expected \"insert\" or \"save\"", name, fj.Generated),
		})
	}

	// version fields are implicitly readable (research.md "Field-policy
	// combination validity") regardless of whether readable was set
	// explicitly — matching the same "database/framework-managed value"
	// treatment as generated.
	if fp.Version {
		fp.Readable = true
	}

	diags = append(diags, validateFieldPolicy(fp, path, context, model)...)
	return fp, diags
}

// validateFieldPolicy applies the field-policy combination rules from
// research.md "Field-policy combination validity". Cross-field rules that
// need the whole model's field list (e.g. "at most one version field") are
// checked by validateModelFields in model.go.
func validateFieldPolicy(fp FieldPolicy, path, context, model string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic

	if fp.Generated != GeneratedNone && fp.Fillable {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: cannot be both fillable and generated — a caller-supplied value the database would immediately overwrite", fp.Name),
			Hint:     "remove fillable, or remove generated if the caller is expected to supply this value",
		})
	}

	if fp.Version && (fp.Fillable || fp.Mutable) {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: cannot be both version and fillable/mutable — an optimistic-concurrency value is database/framework-managed, not caller-supplied", fp.Name),
			Hint:     "remove fillable/mutable from the version field",
		})
	}

	if fp.ImmutableAfterInsert && !fp.Fillable && !fp.Mutable {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityWarning,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: immutable_after_insert has no effect on a field that is neither fillable nor mutable", fp.Name),
			Hint:     "remove immutable_after_insert, or add fillable/mutable if mutation should be allowed before insert",
		})
	}

	return diags
}
