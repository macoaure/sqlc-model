package plan

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/macoaure/sqlc-model/internal/config"
	"github.com/macoaure/sqlc-model/internal/contract"
	"github.com/macoaure/sqlc-model/internal/diagnostics"
	"github.com/macoaure/sqlc-model/internal/mapping"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Plan is the fully-resolved, order-preserved, ready-to-render
// representation of one successful RootConfiguration (data-model.md
// "GenerationPlan").
type Plan struct {
	Contexts []ResolvedContext
}

// ResolvedContext mirrors config.BoundedContext with every model resolved.
type ResolvedContext struct {
	Name      string
	Package   string
	Directory string
	Models    []ResolvedModel
}

// ResolvedModel mirrors config.ModelConfiguration with every operation and
// field resolved against real sqlc query metadata.
type ResolvedModel struct {
	Name       string
	Row        string
	Operations ResolvedOperations
	Fields     []ResolvedField
	Relations  []ResolvedRelation
	Queries    []ResolvedQuery
	Lookups    []ResolvedLookup
}

// ResolvedOperations holds the matched sqlc query for each configured
// lifecycle operation; a nil entry means that operation was not configured
// (no corresponding generated method), except Insert which config-stage
// validation guarantees is always present in a valid configuration.
type ResolvedOperations struct {
	Find    *ResolvedOperation
	Insert  *ResolvedOperation
	Update  *ResolvedOperation
	Delete  *ResolvedOperation
	Refresh *ResolvedOperation
}

// ResolvedOperation is one lifecycle operation resolved to its concrete
// sqlc query, plus each of that query's parameters bound to the resolved
// field that supplies its value (stage 5, "parameter mapping validation").
type ResolvedOperation struct {
	Kind      contract.OperationKind
	QueryName string
	Query     *pb.Query
	Params    []ParamBinding
	// Scan is the field to hydrate for each returned column, in the
	// query's own column order — this is the exact positional order a
	// generated Scan(...) call must use, which need not match the model's
	// declared field order. Empty for operations that don't return rows
	// (Delete via :execrows/:exec).
	Scan []*ResolvedField
}

// ParamBinding is one query parameter (by its 1-based positional Number)
// bound to the field that supplies its value at execution time.
type ParamBinding struct {
	Number int32
	Field  *ResolvedField
}

type ResolvedQuery struct {
	Name      string
	Terminal  config.QueryTerminal
	Operation *ResolvedOperation
	Scopes    []ResolvedQueryScope
	Scan      []*ResolvedField
}

type ResolvedQueryScope struct {
	Name      string
	Parameter string
	Value     json.RawMessage
	ValueSet  bool
	Argument  string
	QueryName string
	Relation  string
	GoType    mapping.GoType
	Number    int32
	Variant   *pb.Query
}

type ResolvedLookup struct {
	Name      string
	QueryName string
	Operation *ResolvedOperation
}

// ResolvedField is one field fully resolved: its mapping.ResolvedField
// (column/Go-type identity) plus the field policy governing its generated
// API surface.
type ResolvedField struct {
	mapping.ResolvedField
	Policy config.FieldPolicy
}

// operationOrder is the fixed evaluation order for a model's lifecycle
// operations — deliberately not a map, so diagnostic collection order
// (before the final sort) and, more importantly, reasoning about the
// pipeline stays deterministic and easy to follow.
var operationOrder = []contract.OperationKind{
	contract.Find, contract.Insert, contract.Update, contract.Delete, contract.Refresh,
}

// Build constructs a GenerationPlan from validated configuration and the
// sqlc request's query metadata. Any error-severity diagnostic anywhere —
// from this stage or carried in from an earlier one — means Build returns a
// nil Plan: this is the mechanism behind FR-017's all-or-nothing guarantee.
func Build(root *config.RootConfiguration, req *pb.GenerateRequest, priorDiags []diagnostics.Diagnostic) (*Plan, []diagnostics.Diagnostic) {
	diags := append([]diagnostics.Diagnostic{}, priorDiags...)

	p := &Plan{}
	for _, ctx := range root.Contexts {
		rc := ResolvedContext{Name: ctx.Name, Package: ctx.Package, Directory: ctx.Directory}
		for _, m := range ctx.Models {
			rm, mdiags := buildModel(ctx, m, req.GetQueries())
			diags = append(diags, mdiags...)
			rc.Models = append(rc.Models, rm)
		}
		// Relations are resolved in a second pass, after every model in
		// this context has its fields resolved (research.md "Two-pass
		// resolution") — a relation's target model may be declared later
		// than the relation referencing it.
		for i := range ctx.Models {
			diags = append(diags, buildRelationsPass2(ctx, ctx.Models[i], &rc.Models[i], &rc, req.GetQueries())...)
		}
		buildRelationsPass3(ctx, &rc)
		diags = append(diags, checkContextCollisions(ctx.Name, rc)...)
		p.Contexts = append(p.Contexts, rc)
	}

	if diagnostics.HasError(diags) {
		return nil, diags
	}
	return p, diags
}

func buildModel(ctx config.BoundedContext, m config.ModelConfiguration, queries []*pb.Query) (ResolvedModel, []diagnostics.Diagnostic) {
	var diags []diagnostics.Diagnostic
	path := fmt.Sprintf("contexts.%s.models.%s", ctx.Name, m.Name)

	if !isExportedIdent(m.Row) {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".row",
			Context:  ctx.Name,
			Model:    m.Name,
			Message:  fmt.Sprintf("model %q: \"row\" %q must be a valid exported Go identifier (it becomes the generated struct name)", m.Name, m.Row),
		})
	}

	rm := ResolvedModel{Name: m.Name, Row: m.Row}

	queryNameFor := map[contract.OperationKind]string{
		contract.Find:    m.Operations.Find,
		contract.Insert:  m.Operations.Insert,
		contract.Update:  m.Operations.Update,
		contract.Delete:  m.Operations.Delete,
		contract.Refresh: m.Operations.Refresh,
	}

	var insertQuery *pb.Query
	for _, kind := range operationOrder {
		queryName := queryNameFor[kind]
		if queryName == "" {
			continue
		}
		opPath := fmt.Sprintf("%s.operations.%s", path, kind)
		q, qdiags := contract.ValidateOperation(kind, queryName, queries, opPath, ctx.Name, m.Name)
		diags = append(diags, qdiags...)
		if q == nil {
			continue
		}
		ro := &ResolvedOperation{Kind: kind, QueryName: queryName, Query: q}
		switch kind {
		case contract.Find:
			rm.Operations.Find = ro
		case contract.Insert:
			rm.Operations.Insert = ro
			insertQuery = q
		case contract.Update:
			rm.Operations.Update = ro
		case contract.Delete:
			rm.Operations.Delete = ro
		case contract.Refresh:
			rm.Operations.Refresh = ro
		}
	}

	if insertQuery == nil {
		// Insert failed contract validation (already diagnosed above); the
		// model has no canonical column set to resolve fields against.
		return rm, diags
	}

	for _, fp := range m.Fields {
		fieldPath := fmt.Sprintf("%s.fields.%s", path, fp.Name)
		resolved, fdiags := mapping.Resolve(fp, insertQuery.GetColumns(), fieldPath, ctx.Name, m.Name)
		diags = append(diags, fdiags...)
		rf := ResolvedField{ResolvedField: resolved, Policy: fp}
		if rf.ValueObject != nil && !rf.NotNull {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fieldPath + ".value_object",
				Context:  ctx.Name,
				Model:    m.Name,
				Message:  fmt.Sprintf("field %q: value_object does not support nullable columns", fp.Name),
				Hint:     "remove value_object or make the underlying column non-nullable",
			})
		}
		diags = append(diags, validateGeneratedHydration(rf, rm.Operations, fieldPath, ctx.Name, m.Name)...)
		rm.Fields = append(rm.Fields, rf)
	}

	for _, ro := range []*ResolvedOperation{rm.Operations.Find, rm.Operations.Insert, rm.Operations.Update, rm.Operations.Delete, rm.Operations.Refresh} {
		if ro == nil {
			continue
		}
		opPath := fmt.Sprintf("%s.operations.%s", path, ro.Kind)
		diags = append(diags, resolveParams(&rm, ro, opPath, ctx.Name, m.Name)...)
		if ro.Kind != contract.Delete {
			diags = append(diags, resolveScanOrder(&rm, ro, opPath, ctx.Name, m.Name)...)
		}
	}

	diags = append(diags, resolveQueries(ctx, m, &rm, queries, path)...)
	diags = append(diags, resolveLookups(ctx, m, &rm, queries, path)...)

	return rm, diags
}

func resolveQueries(ctx config.BoundedContext, m config.ModelConfiguration, rm *ResolvedModel, queries []*pb.Query, path string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	for _, qc := range m.Queries {
		qpath := fmt.Sprintf("%s.queries.%s", path, qc.Name)
		kind := operationKindForTerminal(qc.Terminal)
		opQuery, qdiags := contract.ValidateOperation(kind, qc.Operation, queries, qpath+".operation", ctx.Name, m.Name)
		diags = append(diags, qdiags...)
		if opQuery == nil {
			continue
		}

		ro := &ResolvedOperation{Kind: kind, QueryName: qc.Operation, Query: opQuery}
		rq := ResolvedQuery{Name: qc.Name, Terminal: qc.Terminal, Operation: ro}
		if kind != contract.Delete {
			diags = append(diags, resolveScanOrder(rm, ro, qpath, ctx.Name, m.Name)...)
			rq.Scan = ro.Scan
		}
		diags = append(diags, resolveQueryScopes(ctx, m, rm, &rq, qc, queries, qpath)...)
		diags = append(diags, validateQueryParamsCovered(&rq, qpath, ctx.Name, m.Name)...)
		rm.Queries = append(rm.Queries, rq)
	}
	return diags
}

func resolveLookups(ctx config.BoundedContext, m config.ModelConfiguration, rm *ResolvedModel, queries []*pb.Query, path string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	for _, lo := range m.Lookups {
		lpath := fmt.Sprintf("%s.lookups.%s", path, lo.Name)
		q, qdiags := contract.ValidateOperation(contract.Lookup, lo.Query, queries, lpath+".query", ctx.Name, m.Name)
		diags = append(diags, qdiags...)
		if q == nil {
			continue
		}
		ro := &ResolvedOperation{Kind: contract.Lookup, QueryName: lo.Query, Query: q}
		diags = append(diags, resolveParams(rm, ro, lpath, ctx.Name, m.Name)...)
		diags = append(diags, resolveScanOrder(rm, ro, lpath, ctx.Name, m.Name)...)
		rm.Lookups = append(rm.Lookups, ResolvedLookup{Name: lo.Name, QueryName: lo.Query, Operation: ro})
	}
	return diags
}

func operationKindForTerminal(t config.QueryTerminal) contract.OperationKind {
	switch t {
	case config.QueryTerminalGet:
		return contract.List
	case config.QueryTerminalFirst, config.QueryTerminalFind:
		return contract.Single
	case config.QueryTerminalDelete:
		return contract.Delete
	case config.QueryTerminalRefresh:
		return contract.Refresh
	default:
		return contract.List
	}
}

func resolveQueryScopes(ctx config.BoundedContext, m config.ModelConfiguration, rm *ResolvedModel, rq *ResolvedQuery, qc config.QueryConfiguration, queries []*pb.Query, path string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	for _, sc := range qc.Scopes {
		rs := ResolvedQueryScope{
			Name:      sc.Name,
			Parameter: sc.Parameter,
			Value:     sc.Value,
			ValueSet:  sc.ValueSet,
			Argument:  sc.Argument,
			QueryName: sc.Query,
			Relation:  sc.Relation,
		}
		spath := path + ".scopes." + sc.Name

		if sc.Parameter != "" {
			param := findQueryParam(rq.Operation.Query, sc.Parameter)
			if param == nil {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     spath + ".parameter",
					Context:  ctx.Name,
					Model:    m.Name,
					Query:    rq.Operation.QueryName,
					Message:  fmt.Sprintf("query scope %q references parameter %q, but query %q has no such parameter", sc.Name, sc.Parameter, rq.Operation.QueryName),
				})
			} else {
				rs.Number = param.GetNumber()
				rs.GoType = mapping.ResolveGoType(param.GetColumn())
				if sc.Argument != "" && rs.GoType.Expr != sc.Argument {
					diags = append(diags, diagnostics.Diagnostic{
						Severity: diagnostics.SeverityError,
						Path:     spath + ".argument",
						Context:  ctx.Name,
						Model:    m.Name,
						Query:    rq.Operation.QueryName,
						Message:  fmt.Sprintf("query scope %q argument type %q does not match parameter %q type %q", sc.Name, sc.Argument, sc.Parameter, rs.GoType.Expr),
					})
				}
			}
		}

		if sc.Query != "" {
			q, qdiags := contract.ValidateOperation(rq.Operation.Kind, sc.Query, queries, spath+".query", ctx.Name, m.Name)
			diags = append(diags, qdiags...)
			rs.Variant = q
		}

		if sc.Relation != "" && !hasEagerRelationConfig(m, sc.Relation) {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     spath + ".relation",
				Context:  ctx.Name,
				Model:    m.Name,
				Message:  fmt.Sprintf("eager-load scope %q references relation %q, but that relation has no eager_query", sc.Name, sc.Relation),
			})
		}

		rq.Scopes = append(rq.Scopes, rs)
	}
	return diags
}

func validateQueryParamsCovered(rq *ResolvedQuery, path, context, model string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	covered := make(map[int32]bool)
	for _, sc := range rq.Scopes {
		if sc.Parameter != "" && sc.Number != 0 {
			covered[sc.Number] = true
		}
	}
	for _, p := range rq.Operation.Query.GetParams() {
		if !covered[p.GetNumber()] {
			name := ""
			if col := p.GetColumn(); col != nil {
				name = col.Name
			}
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path,
				Context:  context,
				Model:    model,
				Query:    rq.Operation.QueryName,
				Message:  fmt.Sprintf("query %q parameter %d (%q) is not supplied by any configured scope", rq.Operation.QueryName, p.GetNumber(), name),
			})
		}
	}
	return diags
}

func findQueryParam(q *pb.Query, name string) *pb.Parameter {
	for _, p := range q.GetParams() {
		if col := p.GetColumn(); col != nil && strings.EqualFold(col.Name, name) {
			return p
		}
	}
	return nil
}

func hasEagerRelationConfig(m config.ModelConfiguration, name string) bool {
	for _, rel := range m.Relations {
		if rel.Name == name && rel.EagerQuery != "" {
			return true
		}
	}
	return false
}

// resolveScanOrder binds each column an operation's query returns, in that
// query's own column order, to the field it hydrates (stage 6,
// "result-hydration validation"). Every one of the model's fields must
// appear exactly once among the query's returned columns — a hydrating
// operation whose query is missing (or adds) a column relative to the
// model's declared fields fails generation (FR-004 for insert/update; the
// same full-row requirement extends to find/refresh so every operation
// that hydrates a model actually can).
func resolveScanOrder(rm *ResolvedModel, ro *ResolvedOperation, path, context, model string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	matched := make(map[string]bool, len(rm.Fields))

	for _, col := range ro.Query.GetColumns() {
		var field *ResolvedField
		for i := range rm.Fields {
			if strings.EqualFold(rm.Fields[i].ColumnName, col.Name) {
				field = &rm.Fields[i]
				break
			}
		}
		if field == nil {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path,
				Context:  context,
				Model:    model,
				Query:    ro.QueryName,
				Message:  fmt.Sprintf("%s query %q returns column %q, which does not match any configured field", ro.Kind, ro.QueryName, col.Name),
				Hint:     "add a field for this column, or remove it from the query's result list",
			})
			continue
		}
		matched[field.Name] = true
		ro.Scan = append(ro.Scan, field)
	}

	for _, f := range rm.Fields {
		if !matched[f.Name] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path,
				Context:  context,
				Model:    model,
				Query:    ro.QueryName,
				Message:  fmt.Sprintf("%s query %q does not return field %q's column %q — every hydrating operation must return the full row", ro.Kind, ro.QueryName, f.Name, f.ColumnName),
			})
		}
	}

	return diags
}

// resolveParams binds each of a query's parameters (in declared position)
// to the resolved field whose column identity matches, case-insensitively
// (stage 5, "parameter mapping validation"). A parameter with no matching
// field is an error — the store adapter has no value to supply for it.
func resolveParams(rm *ResolvedModel, ro *ResolvedOperation, path, context, model string) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	for _, param := range ro.Query.GetParams() {
		col := param.GetColumn()
		name := ""
		if col != nil {
			name = col.Name
		}
		var match *ResolvedField
		for i := range rm.Fields {
			if strings.EqualFold(rm.Fields[i].ColumnName, name) {
				match = &rm.Fields[i]
				break
			}
		}
		if match == nil {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path,
				Context:  context,
				Model:    model,
				Query:    ro.QueryName,
				Message:  fmt.Sprintf("%s query %q parameter %d (%q) does not match any configured field", ro.Kind, ro.QueryName, param.GetNumber(), name),
				Hint:     "add a field whose column/row_field resolves to this parameter, or check the query's WHERE/SET clause",
			})
			continue
		}
		ro.Params = append(ro.Params, ParamBinding{Number: param.GetNumber(), Field: match})
	}
	return diags
}

// validateGeneratedHydration checks that a field marked generated: save has
// an update operation to be repopulated from — otherwise "save" is
// indistinguishable from "insert" and likely a configuration mistake.
func validateGeneratedHydration(rf ResolvedField, ops ResolvedOperations, path, context, model string) []diagnostics.Diagnostic {
	if rf.Policy.Generated == config.GeneratedSave && ops.Update == nil {
		return []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityWarning,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("field %q: generated: save has no effect without a configured update operation", rf.Name),
			Hint:     "use generated: insert instead, or configure operations.update",
		}}
	}
	return nil
}

func isExportedIdent(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)
	if !unicode.IsUpper(r[0]) {
		return false
	}
	for _, c := range r {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
	}
	return true
}
