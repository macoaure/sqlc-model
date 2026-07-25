package mapping

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// ResolvedField is a field's mapping to a concrete query result column,
// fully resolved to real sqlc metadata (data-model.md "ResolvedField").
type ResolvedField struct {
	Name            string // the field's declared config key
	GoField         string // generated Go struct field name, PascalCase(Name)
	ColumnName      string // the resolved query result column's exposed name
	GoType          GoType // exposed generated model type
	PersistedGoType GoType // sqlc-compatible scan/parameter type
	ValueObject     *config.ValueObjectMapping
	NotNull         bool
}

// Resolve determines which of columns a field policy identifies, per
// research.md "Column <-> row-field mapping resolution":
//
//   - column and row_field both explicit: the column whose underlying
//     database identity (OriginalName, falling back to Name) matches
//     `column` AND whose query-result identity (Name) matches `row_field`.
//   - column only: matches the column's underlying database identity.
//   - row_field only: matches the column's query-result identity verbatim.
//   - neither (automatic): the field's declared key must match exactly one
//     column's query-result identity, case-insensitively.
//
// Any outcome other than exactly one match is an FR-007 ambiguity/absence
// diagnostic — never a guess.
func Resolve(fp config.FieldPolicy, columns []*pb.Column, path, context, model string) (ResolvedField, []diagnostics.Diagnostic) {
	var candidates []*pb.Column

	switch {
	case fp.Column != "" && fp.RowField != "":
		for _, c := range columns {
			if columnIdentity(c) == fp.Column && c.Name == fp.RowField {
				candidates = append(candidates, c)
			}
		}
	case fp.Column != "":
		for _, c := range columns {
			if columnIdentity(c) == fp.Column {
				candidates = append(candidates, c)
			}
		}
	case fp.RowField != "":
		for _, c := range columns {
			if c.Name == fp.RowField {
				candidates = append(candidates, c)
			}
		}
	default:
		for _, c := range columns {
			if strings.EqualFold(c.Name, fp.Name) {
				candidates = append(candidates, c)
			}
		}
	}

	switch len(candidates) {
	case 0:
		return ResolvedField{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: no matching column found (%s)", fp.Name, describeMatch(fp)),
			Hint:     "set an explicit column and/or row_field override, or check the field name matches the query's result column",
		}}
	case 1:
		col := candidates[0]
		gt := ResolveGoType(col)
		exposed := gt
		if fp.ValueObject != nil {
			exposed = GoType{Expr: fp.ValueObject.Type}
		}
		var diags []diagnostics.Diagnostic
		if gt.Unmapped {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path,
				Context:  context,
				Model:    model,
				Message:  fmt.Sprintf("field %q: column type %q has no supported Go type mapping", fp.Name, typeName(col)),
				Hint:     "use a supported PostgreSQL type or configure a custom type override",
			})
		}
		return ResolvedField{
			Name:            fp.Name,
			GoField:         PascalCase(fp.Name),
			ColumnName:      col.Name,
			GoType:          exposed,
			PersistedGoType: gt,
			ValueObject:     fp.ValueObject,
			NotNull:         col.NotNull,
		}, diags
	default:
		return ResolvedField{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: ambiguous match, %d columns match (%s)", fp.Name, len(candidates), describeMatch(fp)),
			Hint:     "add an explicit column and/or row_field override to disambiguate",
		}}
	}
}

func columnIdentity(c *pb.Column) string {
	if c.OriginalName != "" {
		return c.OriginalName
	}
	return c.Name
}

func typeName(c *pb.Column) string {
	if c.Type == nil {
		return "<unknown>"
	}
	return c.Type.Name
}

func describeMatch(fp config.FieldPolicy) string {
	switch {
	case fp.Column != "" && fp.RowField != "":
		return fmt.Sprintf("column=%q row_field=%q", fp.Column, fp.RowField)
	case fp.Column != "":
		return fmt.Sprintf("column=%q", fp.Column)
	case fp.RowField != "":
		return fmt.Sprintf("row_field=%q", fp.RowField)
	default:
		return "automatic match against field name"
	}
}

// commonInitialisms is Go's own well-known initialism list (a subset of
// golint/revive's table) — a field key part matching one of these,
// case-insensitively, is rendered fully upper-case rather than merely
// title-cased, so generated identifiers read idiomatically (ID, not Id).
var commonInitialisms = map[string]bool{
	"id": true, "ids": true, "uuid": true, "url": true, "uri": true,
	"api": true, "html": true, "json": true, "xml": true, "sql": true,
	"http": true, "https": true, "tcp": true, "udp": true, "ip": true,
	"db": true, "ttl": true, "utf8": true,
}

// PascalCase converts a snake_case identifier (the generator's own
// convention for field keys and generated Go names) to PascalCase, e.g.
// "created_at" -> "CreatedAt", applying Go's common-initialism convention
// so "id" -> "ID" rather than "Id".
func PascalCase(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if commonInitialisms[strings.ToLower(p)] {
			b.WriteString(strings.ToUpper(p))
			continue
		}
		r := []rune(p)
		b.WriteRune(unicode.ToUpper(r[0]))
		if len(r) > 1 {
			b.WriteString(string(r[1:]))
		}
	}
	return b.String()
}
