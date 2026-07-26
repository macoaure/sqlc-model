package plan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/macoaure/sqlc-model/internal/config"
	"github.com/macoaure/sqlc-model/internal/diagnostics"
	"github.com/macoaure/sqlc-model/internal/mapping"
	"github.com/macoaure/sqlc-model/internal/relation"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// RelationParamSource identifies where a lazy_query parameter's value
// comes from at call time (data-model.md "ParameterBinding").
type RelationParamSource int

const (
	// ParamFromParent binds the parameter to a field on the parent model.
	ParamFromParent RelationParamSource = iota
	// ParamFromScope binds the parameter to whichever scope call (if any)
	// sets it in the current call chain, falling back to Default.
	ParamFromScope
)

// RelationParam is one of LazyQuery's declared parameters, resolved to its
// binding source, in the query's own positional order.
type RelationParam struct {
	Number int32
	// Name is the query parameter's own declared name (e.g. "published") —
	// also the key a scope method writes into the builder's scopeValues
	// map, so the two stay in sync regardless of the scope's own name.
	Name        string
	Source      RelationParamSource
	ParentField *ResolvedField  // set when Source == ParamFromParent
	ScopeName   string          // set when Source == ParamFromScope: the scope name whose Default/type this parameter uses
	Default     json.RawMessage // set when Source == ParamFromScope and a default was configured
	GoType      mapping.GoType  // the parameter's resolved Go type
}

// ResolvedRelation mirrors config.RelationConfiguration with every query
// and cross-model reference resolved against real sqlc metadata
// (data-model.md "ResolvedRelation").
type ResolvedRelation struct {
	Name        string
	Kind        config.RelationKind
	Target      *ResolvedModel
	Inverse     *ResolvedRelation
	LocalKey    *ResolvedField
	TargetKey   *ResolvedField
	ForeignKey  string
	Nullable    bool
	LazyQuery   *pb.Query
	EagerQuery  *pb.Query
	AttachQuery *pb.Query
	DetachQuery *pb.Query
	SyncList    *pb.Query
	SyncAttach  *pb.Query
	SyncDetach  *pb.Query
	Parameters  []config.ParameterBinding
	Scopes      []config.ScopeConfiguration
	// LazyParams is LazyQuery's declared parameters, resolved to their
	// binding source, in the query's own positional order.
	LazyParams []RelationParam
	// LazyScan/EagerScan are the target model's fields, in LazyQuery's/
	// EagerQuery's own returned-column order — the exact positional Scan(...)
	// order for hydrating a target instance from that query's result.
	LazyScan  []*ResolvedField
	EagerScan []*ResolvedField
	// ForeignKeyField is the target model's field matching ForeignKey,
	// resolved when EagerQuery is configured (needed to read each returned
	// row's grouping value for batch hydration).
	ForeignKeyField *ResolvedField
	// ScopeVariantQueries maps a query-variant scope's name to its already
	// command/cardinality-validated query (internal/relation's stage 8).
	ScopeVariantQueries map[string]*pb.Query

	// resolvedInverseName is scratch state between pass 2 and pass 3: the
	// validated inverse relation name (empty if unset or invalid), used to
	// decide whether pass 3 should wire Inverse at all.
	resolvedInverseName string
}

// buildRelationsPass2 resolves every model's relations against
// already-field-resolved models in the same context (research.md "Two-pass
// resolution"): queries, local_key, and scopes/parameters are resolved
// here. Target/Inverse cross-relation pointers are wired in
// buildRelationsPass3, once every model's Relations slice in rc is
// complete and stable — taking pointers into a slice still being appended
// to risks pointing into a backing array a later append reallocates away
// from.
func buildRelationsPass2(ctx config.BoundedContext, m config.ModelConfiguration, rm *ResolvedModel, rc *ResolvedContext, queries []*pb.Query) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	path := fmt.Sprintf("contexts.%s.models.%s", ctx.Name, m.Name)

	for _, rel := range m.Relations {
		relPath := fmt.Sprintf("%s.relations.%s", path, rel.Name)
		g, gdiags := relation.ValidateGraph(ctx, rel, queries, relPath, ctx.Name, m.Name)
		diags = append(diags, gdiags...)
		sdiags, variants := relation.ValidateScopes(rel, g, queries, relPath, ctx.Name, m.Name)
		diags = append(diags, sdiags...)

		rr := ResolvedRelation{
			Name:                rel.Name,
			Kind:                rel.Kind,
			ForeignKey:          rel.ForeignKey,
			Nullable:            rel.Nullable,
			LazyQuery:           g.LazyQuery,
			EagerQuery:          g.EagerQuery,
			AttachQuery:         g.AttachQuery,
			DetachQuery:         g.DetachQuery,
			SyncList:            g.SyncList,
			SyncAttach:          g.SyncAttach,
			SyncDetach:          g.SyncDetach,
			Parameters:          rel.Parameters,
			Scopes:              rel.Scopes,
			ScopeVariantQueries: variants,
		}

		if g.Inverse != nil {
			rr.resolvedInverseName = rel.Inverse
		}

		if rel.LocalKey != "" {
			if lf := findResolvedField(rm.Fields, rel.LocalKey); lf == nil {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     relPath + ".local_key",
					Context:  ctx.Name,
					Model:    m.Name,
					Relation: rel.Name,
					Message:  fmt.Sprintf("relation %q: local_key %q does not match any configured field on %s", rel.Name, rel.LocalKey, m.Name),
				})
			} else {
				rr.LocalKey = lf
			}
		}

		var targetFields []ResolvedField
		for i := range ctx.Models {
			if ctx.Models[i].Name == rel.Model {
				targetFields = rc.Models[i].Fields
				break
			}
		}

		if rel.TargetKey != "" {
			if tf := findResolvedField(targetFields, rel.TargetKey); tf == nil {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     relPath + ".target_key",
					Context:  ctx.Name,
					Model:    m.Name,
					Relation: rel.Name,
					Message:  fmt.Sprintf("relation %q: target_key %q does not match any configured field on %s", rel.Name, rel.TargetKey, rel.Model),
				})
			} else {
				rr.TargetKey = tf
			}
		}

		if rr.LazyQuery != nil {
			scan, sdiags := resolveTargetScan(rr.LazyQuery, targetFields, rel.LazyQuery, relPath+".lazy_query", ctx.Name, m.Name, rel.Name)
			diags = append(diags, sdiags...)
			rr.LazyScan = scan
		}
		if rr.EagerQuery != nil {
			scan, sdiags := resolveTargetScan(rr.EagerQuery, targetFields, rel.EagerQuery, relPath+".eager_query", ctx.Name, m.Name, rel.Name)
			diags = append(diags, sdiags...)
			rr.EagerScan = scan
		}
		// ForeignKeyField (the target field matching foreign_key) is needed
		// both for eager grouping and for New()'s auto-assignment of a
		// newly constructed child's own foreign-key field to the parent's
		// local_key value — resolved whenever foreign_key is configured
		// (belongs_to/has_many/has_one), not only when eager_query is set.
		if rel.ForeignKey != "" {
			if fkf := findResolvedField(targetFields, rel.ForeignKey); fkf == nil {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     relPath + ".foreign_key",
					Context:  ctx.Name,
					Model:    m.Name,
					Relation: rel.Name,
					Message:  fmt.Sprintf("relation %q: foreign_key %q does not match any configured field on %s", rel.Name, rel.ForeignKey, rel.Model),
				})
			} else {
				rr.ForeignKeyField = fkf
			}
		}

		lpdiags := resolveLazyParams(rel, rm, &rr, relPath, ctx.Name, m.Name)
		diags = append(diags, lpdiags...)

		rm.Relations = append(rm.Relations, rr)
	}

	return diags
}

// resolveLazyParams binds every parameter LazyQuery declares to its source
// (data-model.md "ParameterBinding"): with no explicit `parameters`
// configured, the single implicit parameter binds to LocalKey (already
// validated to be exactly one by internal/relation's stage 7); with
// `parameters` configured, each declared parameter is matched by name to
// its configured binding — "parent.<key>" resolves against rm's own
// already-resolved fields, "scope.<ScopeName>" carries its configured
// default through for codegen to apply when the scope isn't chained.
func resolveLazyParams(rel config.RelationConfiguration, rm *ResolvedModel, rr *ResolvedRelation, relPath, ctxName, modelName string) []diagnostics.Diagnostic {
	if rr.LazyQuery == nil {
		return nil
	}

	if len(rel.Parameters) == 0 {
		if len(rr.LazyQuery.GetParams()) == 1 && rr.LocalKey != nil {
			p := rr.LazyQuery.GetParams()[0]
			name := ""
			if col := p.GetColumn(); col != nil {
				name = col.Name
			}
			rr.LazyParams = []RelationParam{{Number: p.GetNumber(), Name: name, Source: ParamFromParent, ParentField: rr.LocalKey}}
		}
		return nil
	}

	var diags []diagnostics.Diagnostic
	for _, p := range rr.LazyQuery.GetParams() {
		col := p.GetColumn()
		name := ""
		if col != nil {
			name = col.Name
		}
		var binding *config.ParameterBinding
		for i := range rel.Parameters {
			if rel.Parameters[i].Name == name {
				binding = &rel.Parameters[i]
				break
			}
		}
		if binding == nil {
			continue // missing-binding diagnostic already raised by internal/relation's scope.go
		}

		switch {
		case strings.HasPrefix(binding.Source, "parent."):
			key := strings.TrimPrefix(binding.Source, "parent.")
			pf := findResolvedField(rm.Fields, key)
			if pf == nil {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     relPath + ".parameters." + binding.Name,
					Context:  ctxName,
					Model:    modelName,
					Relation: rel.Name,
					Message:  fmt.Sprintf("relation %q: parameter %q source parent.%s does not match any configured field on %s", rel.Name, binding.Name, key, modelName),
				})
				continue
			}
			rr.LazyParams = append(rr.LazyParams, RelationParam{Number: p.GetNumber(), Name: name, Source: ParamFromParent, ParentField: pf})
		case strings.HasPrefix(binding.Source, "scope."):
			gt := mapping.GoType{}
			if col != nil {
				gt = mapping.ResolveGoType(col)
			}
			rr.LazyParams = append(rr.LazyParams, RelationParam{
				Number:    p.GetNumber(),
				Name:      name,
				Source:    ParamFromScope,
				ScopeName: strings.TrimPrefix(binding.Source, "scope."),
				Default:   binding.Default,
				GoType:    gt,
			})
		}
	}
	return diags
}

// buildRelationsPass3 wires Target/Inverse pointers for every relation in
// every model of ctx/rc, now that no further appends to any model's
// Relations (or rc.Models itself) will occur.
func buildRelationsPass3(ctx config.BoundedContext, rc *ResolvedContext) {
	modelIndex := make(map[string]int, len(ctx.Models))
	for i, m := range ctx.Models {
		modelIndex[m.Name] = i
	}

	for i, m := range ctx.Models {
		for j, rel := range m.Relations {
			targetIdx, ok := modelIndex[rel.Model]
			if !ok {
				continue // unresolved target already diagnosed in pass 2
			}
			rc.Models[i].Relations[j].Target = &rc.Models[targetIdx]

			inverseName := rc.Models[i].Relations[j].resolvedInverseName
			if inverseName == "" {
				continue
			}
			for k, inv := range ctx.Models[targetIdx].Relations {
				if inv.Name == inverseName {
					rc.Models[i].Relations[j].Inverse = &rc.Models[targetIdx].Relations[k]
					break
				}
			}
		}
	}
}

// resolveTargetScan binds each column query returns, in that query's own
// column order, to the target field it hydrates — the relation-query
// analogue of build.go's resolveScanOrder, matched against the *target*
// model's fields rather than the query's own owning model. Unlike
// lifecycle operations, a relation query need not return every target
// field (a relation MAY hydrate a partial view of the target) — but every
// column it does return must match some target field, catching typos
// rather than silently discarding a column.
func resolveTargetScan(query *pb.Query, targetFields []ResolvedField, queryName, path, ctxName, modelName, relName string) ([]*ResolvedField, []diagnostics.Diagnostic) {
	if query == nil {
		return nil, nil
	}
	var out []*ResolvedField
	var diags []diagnostics.Diagnostic
	for _, col := range query.GetColumns() {
		var match *ResolvedField
		for i := range targetFields {
			if strings.EqualFold(targetFields[i].ColumnName, col.Name) {
				match = &targetFields[i]
				break
			}
		}
		if match == nil {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path,
				Context:  ctxName,
				Model:    modelName,
				Relation: relName,
				Query:    queryName,
				Message:  fmt.Sprintf("query %q returns column %q, which does not match any configured field on the target model", queryName, col.Name),
			})
			continue
		}
		out = append(out, match)
	}
	return out, diags
}

func findResolvedField(fields []ResolvedField, name string) *ResolvedField {
	for i := range fields {
		if strings.EqualFold(fields[i].Name, name) {
			return &fields[i]
		}
	}
	return nil
}
