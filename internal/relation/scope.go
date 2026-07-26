package relation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/macoaure/sqlc-model/internal/config"
	"github.com/macoaure/sqlc-model/internal/contract"
	"github.com/macoaure/sqlc-model/internal/diagnostics"
	"github.com/macoaure/sqlc-model/internal/mapping"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// reservedRelationMethodNames are the fixed relation-operation identifiers a
// scope name must not collide with (data-model.md "ScopeConfiguration"
// validation; contracts/relation-api.md).
var reservedRelationMethodNames = map[string]bool{
	"Get": true, "Reload": true, "Cached": true, "Loaded": true, "Forget": true,
	"New": true, "Associate": true, "Dissociate": true,
	"Attach": true, "Detach": true, "Sync": true,
}

// ValidateScopes validates rel's `parameters` bindings and `scopes` against
// g's already-matched queries and req's full query list (data-model.md
// "Relation Graph Validation" stage 8; FR-009 through FR-011). It also
// returns each query-variant scope's resolved query (scope name -> query),
// so internal/plan doesn't need to re-resolve query names it already
// validated here.
func ValidateScopes(rel config.RelationConfiguration, g Graph, queries []*pb.Query, path, contextName, modelName string) ([]diagnostics.Diagnostic, map[string]*pb.Query) {
	var diags []diagnostics.Diagnostic
	variants := make(map[string]*pb.Query)
	errf := func(suffix, format string, args ...any) {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + suffix,
			Context:  contextName,
			Model:    modelName,
			Relation: rel.Name,
			Message:  fmt.Sprintf(format, args...),
		})
	}

	// lazyParamNames is used for the "every parameter has a binding" check:
	// eager_query's single parameter is always implicitly bound to the
	// batch of parent local_key values (research.md), never through the
	// `parameters` map, so it's deliberately excluded here. allParamNames
	// (lazy + eager) is used only to check that a configured `parameters`
	// entry references a real parameter of *some* configured query.
	lazyParamNames := queryParamNames(g.LazyQuery)
	allParamNames := queryParamNames(g.LazyQuery)
	for n := range queryParamNames(g.EagerQuery) {
		allParamNames[n] = true
	}

	if len(rel.Parameters) > 0 {
		for _, pbind := range rel.Parameters {
			if !allParamNames[pbind.Name] {
				errf(".parameters."+pbind.Name, "relation %q: parameter %q does not match any parameter declared by lazy_query/eager_query", rel.Name, pbind.Name)
				continue
			}
			switch {
			case strings.HasPrefix(pbind.Source, "scope."):
				scopeName := strings.TrimPrefix(pbind.Source, "scope.")
				if !hasScope(rel.Scopes, scopeName) {
					errf(".parameters."+pbind.Name, "relation %q: parameter %q source references scope %q, which is not declared on this relation", rel.Name, pbind.Name, scopeName)
				}
			case strings.HasPrefix(pbind.Source, "parent."):
				// key existence against the owning model's fields is
				// validated at plan-resolution time (internal/plan), where
				// the model's resolved field list is available.
			default:
				errf(".parameters."+pbind.Name, "relation %q: parameter %q source %q must be \"parent.<key>\" or \"scope.<ScopeName>\"", rel.Name, pbind.Name, pbind.Source)
			}
		}

		bound := make(map[string]bool, len(rel.Parameters))
		for _, pbind := range rel.Parameters {
			bound[pbind.Name] = true
		}
		for name := range lazyParamNames {
			if !bound[name] {
				errf(".parameters", "relation %q: query parameter %q has no corresponding entry in \"parameters\"", rel.Name, name)
			}
		}
	}
	paramNames := allParamNames

	lazyKind := contract.LazyToOne
	if rel.Kind == config.HasMany || rel.Kind == config.ManyToMany {
		lazyKind = contract.LazyToMany
	}

	for _, sc := range rel.Scopes {
		scPath := path + ".scopes." + sc.Name
		if reservedRelationMethodNames[sc.Name] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     scPath,
				Context:  contextName,
				Model:    modelName,
				Relation: rel.Name,
				Message:  fmt.Sprintf("scope %q collides with a fixed relation operation name", sc.Name),
				Hint:     "rename the scope",
			})
		}

		switch {
		case sc.Query != "":
			variant, qdiags := contract.ValidateOperation(lazyKind, sc.Query, queries, scPath+".query", contextName, modelName)
			diags = append(diags, withRelation(qdiags, rel.Name)...)
			if variant != nil {
				variants[sc.Name] = variant
			}
			if variant != nil && g.LazyQuery != nil && !sameColumnSet(variant.GetColumns(), g.LazyQuery.GetColumns()) {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     scPath + ".query",
					Context:  contextName,
					Model:    modelName,
					Relation: rel.Name,
					Query:    sc.Query,
					Message:  fmt.Sprintf("scope %q: query %q's result shape does not match relation %q's lazy_query result shape", sc.Name, sc.Query, rel.Name),
					Hint:     "a query-variant scope must return the same columns as lazy_query",
				})
			}
		case sc.Parameter != "":
			if !paramNames[sc.Parameter] {
				errf(scPath, "scope %q: parameter %q does not exist for relation %q", sc.Name, sc.Parameter, rel.Name)
				continue
			}
			// paramNames (checked above) is built from exactly these two
			// queries' own parameters, and paramGoType's case-insensitive
			// match is a superset of that exact-match lookup — so whichever
			// query contributed sc.Parameter to paramNames is guaranteed to
			// resolve it here; there's no third "not found" outcome to
			// guard against.
			gt, ok := paramGoType(g.LazyQuery, sc.Parameter)
			if !ok {
				gt, _ = paramGoType(g.EagerQuery, sc.Parameter)
			}
			if sc.ValueSet && !valueCompatible(sc.Value, gt) {
				errf(scPath, "scope %q: value is not compatible with parameter %q's type %s", sc.Name, sc.Parameter, gt.Expr)
			}
			if sc.Argument != "" && sc.Argument != gt.Expr {
				errf(scPath, "scope %q: argument type %q does not match parameter %q's type %s", sc.Name, sc.Argument, sc.Parameter, gt.Expr)
			}
		}
	}

	return diags, variants
}

func hasScope(scopes []config.ScopeConfiguration, name string) bool {
	for _, sc := range scopes {
		if sc.Name == name {
			return true
		}
	}
	return false
}

func queryParamNames(q *pb.Query) map[string]bool {
	out := make(map[string]bool)
	if q == nil {
		return out
	}
	for _, p := range q.GetParams() {
		if col := p.GetColumn(); col != nil && col.Name != "" {
			out[col.Name] = true
		}
	}
	return out
}

func paramGoType(q *pb.Query, name string) (mapping.GoType, bool) {
	if q == nil {
		return mapping.GoType{}, false
	}
	for _, p := range q.GetParams() {
		col := p.GetColumn()
		if col != nil && strings.EqualFold(col.Name, name) {
			return mapping.ResolveGoType(col), true
		}
	}
	return mapping.GoType{}, false
}

func sameColumnSet(a, b []*pb.Column) bool {
	if len(a) != len(b) {
		return false
	}
	names := make(map[string]bool, len(a))
	for _, c := range a {
		names[strings.ToLower(c.Name)] = true
	}
	for _, c := range b {
		if !names[strings.ToLower(c.Name)] {
			return false
		}
	}
	return true
}

// valueCompatible checks a fixed scope value's JSON literal kind against a
// query parameter's resolved Go type — a best-effort structural check
// (FR-011's "incompatible value type"), not full type inference: JSON
// strings are accepted broadly since many pgtype wrapper types (UUID,
// timestamps, numeric) are conventionally supplied as text-like literals in
// configuration.
func valueCompatible(raw json.RawMessage, gt mapping.GoType) bool {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return false
	}
	switch v.(type) {
	case nil:
		return true
	case bool:
		return gt.Expr == "bool" || gt.Expr == "pgtype.Bool"
	case float64:
		switch gt.Expr {
		case "int16", "int32", "int64", "float32", "float64",
			"pgtype.Int2", "pgtype.Int4", "pgtype.Int8", "pgtype.Float4", "pgtype.Float8", "pgtype.Numeric":
			return true
		default:
			return false
		}
	case string:
		switch gt.Expr {
		case "string", "pgtype.Text", "pgtype.UUID", "pgtype.Timestamptz", "pgtype.Timestamp",
			"pgtype.Date", "pgtype.Time", "pgtype.Interval", "pgtype.Inet":
			return true
		default:
			return false
		}
	default:
		return false
	}
}
