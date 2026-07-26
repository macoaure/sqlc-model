package relation

import (
	"fmt"
	"strings"

	"github.com/macoaure/sqlc-model/internal/config"
	"github.com/macoaure/sqlc-model/internal/contract"
	"github.com/macoaure/sqlc-model/internal/diagnostics"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Graph is stage 7's resolved output for one relation: the matched sqlc
// queries and the target/inverse relation's position within the context,
// ready for internal/plan to wire into ResolvedRelation pointers
// (data-model.md "Relation Graph Validation").
type Graph struct {
	TargetModel *config.ModelConfiguration    // pointer into the owning context's Models
	Inverse     *config.RelationConfiguration // pointer into TargetModel.Relations; nil if unset
	LazyQuery   *pb.Query
	EagerQuery  *pb.Query // nil unless configured
	AttachQuery *pb.Query // nil unless configured (standalone attach_query)
	DetachQuery *pb.Query // nil unless configured (standalone detach_query)
	SyncList    *pb.Query // nil unless sync_queries configured
	SyncAttach  *pb.Query // nil unless sync_queries configured
	SyncDetach  *pb.Query // nil unless sync_queries configured
}

// ValidateGraph validates rel's target/inverse resolution and its
// configured queries' command/cardinality against ctx's other models and
// req's query metadata (data-model.md "Relation Graph Validation" stage 7).
// Target/inverse resolution is same-context only (research.md Assumptions):
// an unresolved target is also how a cross-context relation attempt is
// rejected, since there is no cross-context lookup to fall back to.
func ValidateGraph(ctx config.BoundedContext, rel config.RelationConfiguration, queries []*pb.Query, path, contextName, modelName string) (Graph, []diagnostics.Diagnostic) {
	var diags []diagnostics.Diagnostic
	var g Graph

	var target *config.ModelConfiguration
	for i := range ctx.Models {
		if ctx.Models[i].Name == rel.Model {
			target = &ctx.Models[i]
			break
		}
	}
	if target == nil {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".model",
			Context:  contextName,
			Model:    modelName,
			Relation: rel.Name,
			Message:  fmt.Sprintf("relation %q: target model %q was not found in context %q", rel.Name, rel.Model, contextName),
			Hint:     "check the model name, and that it's declared in the same bounded context — cross-context relations are not supported",
		})
	} else {
		g.TargetModel = target
	}

	if rel.Inverse != "" && target != nil {
		var inv *config.RelationConfiguration
		for i := range target.Relations {
			if target.Relations[i].Name == rel.Inverse {
				inv = &target.Relations[i]
				break
			}
		}
		switch {
		case inv == nil:
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path + ".inverse",
				Context:  contextName,
				Model:    modelName,
				Relation: rel.Name,
				Message:  fmt.Sprintf("relation %q declares inverse %q, but %s has no relation named %q", rel.Name, rel.Inverse, target.Name, rel.Inverse),
				Hint:     fmt.Sprintf("declare a %q relation on %s, or remove the inverse reference from %s.%s", rel.Inverse, target.Name, modelName, rel.Name),
			})
		case !kindsCompatible(rel.Kind, inv.Kind):
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path + ".inverse",
				Context:  contextName,
				Model:    modelName,
				Relation: rel.Name,
				Message:  fmt.Sprintf("relation %q (kind %s) declares inverse %q (kind %s) on %s, but these kinds are not compatible", rel.Name, rel.Kind, rel.Inverse, inv.Kind, target.Name),
				Hint:     "belongs_to's inverse must be has_many or has_one (and vice versa); many_to_many's inverse must be many_to_many",
			})
		default:
			g.Inverse = inv
		}
	}

	lazyKind := contract.LazyToOne
	if rel.Kind == config.HasMany || rel.Kind == config.ManyToMany {
		lazyKind = contract.LazyToMany
	}
	if lazyQuery, qdiags := contract.ValidateOperation(lazyKind, rel.LazyQuery, queries, path+".lazy_query", contextName, modelName); len(qdiags) > 0 {
		diags = append(diags, withRelation(qdiags, rel.Name)...)
	} else {
		g.LazyQuery = lazyQuery
	}

	if rel.EagerQuery != "" {
		eagerQuery, qdiags := contract.ValidateOperation(contract.Eager, rel.EagerQuery, queries, path+".eager_query", contextName, modelName)
		diags = append(diags, withRelation(qdiags, rel.Name)...)
		g.EagerQuery = eagerQuery
		if eagerQuery != nil {
			if len(eagerQuery.GetParams()) != 1 {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     path + ".eager_query",
					Context:  contextName,
					Model:    modelName,
					Relation: rel.Name,
					Query:    rel.EagerQuery,
					Message:  fmt.Sprintf("eager_query %q must declare exactly one parameter (the batch of parent keys), found %d", rel.EagerQuery, len(eagerQuery.GetParams())),
				})
			}
			if rel.ForeignKey != "" && !hasColumn(eagerQuery.GetColumns(), rel.ForeignKey) {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     path + ".eager_query",
					Context:  contextName,
					Model:    modelName,
					Relation: rel.Name,
					Query:    rel.EagerQuery,
					Message:  fmt.Sprintf("eager_query %q must return a column matching foreign_key %q, so results can be grouped by parent", rel.EagerQuery, rel.ForeignKey),
				})
			}
		}
	}

	if rel.AttachQuery != "" {
		attachQuery, qdiags := contract.ValidateOperation(contract.Attach, rel.AttachQuery, queries, path+".attach_query", contextName, modelName)
		diags = append(diags, withRelation(qdiags, rel.Name)...)
		g.AttachQuery = attachQuery
		diags = append(diags, requireParamCount(attachQuery, rel.AttachQuery, 2, path+".attach_query", contextName, modelName, rel.Name)...)
	}
	if rel.DetachQuery != "" {
		detachQuery, qdiags := contract.ValidateOperation(contract.Detach, rel.DetachQuery, queries, path+".detach_query", contextName, modelName)
		diags = append(diags, withRelation(qdiags, rel.Name)...)
		g.DetachQuery = detachQuery
		diags = append(diags, requireParamCount(detachQuery, rel.DetachQuery, 2, path+".detach_query", contextName, modelName, rel.Name)...)
	}
	if rel.SyncQueries != nil {
		listQuery, qdiags := contract.ValidateOperation(contract.SyncList, rel.SyncQueries.List, queries, path+".sync_queries.list", contextName, modelName)
		diags = append(diags, withRelation(qdiags, rel.Name)...)
		g.SyncList = listQuery
		diags = append(diags, requireParamCount(listQuery, rel.SyncQueries.List, 1, path+".sync_queries.list", contextName, modelName, rel.Name)...)
		// sync_queries.list's single result column is scanned positionally
		// into a target_key-typed variable (codegen's Sync()), not matched
		// by column name — so unlike ForeignKey/eager_query, its name need
		// not equal target_key.
		if listQuery != nil && len(listQuery.GetColumns()) != 1 {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     path + ".sync_queries.list",
				Context:  contextName,
				Model:    modelName,
				Relation: rel.Name,
				Query:    rel.SyncQueries.List,
				Message:  fmt.Sprintf("sync_queries.list %q must return exactly one column (the related identifier), found %d", rel.SyncQueries.List, len(listQuery.GetColumns())),
			})
		}

		syncAttach, adiags := contract.ValidateOperation(contract.Attach, rel.SyncQueries.Attach, queries, path+".sync_queries.attach", contextName, modelName)
		diags = append(diags, withRelation(adiags, rel.Name)...)
		g.SyncAttach = syncAttach
		diags = append(diags, requireParamCount(syncAttach, rel.SyncQueries.Attach, 2, path+".sync_queries.attach", contextName, modelName, rel.Name)...)

		syncDetach, ddiags := contract.ValidateOperation(contract.Detach, rel.SyncQueries.Detach, queries, path+".sync_queries.detach", contextName, modelName)
		diags = append(diags, withRelation(ddiags, rel.Name)...)
		g.SyncDetach = syncDetach
		diags = append(diags, requireParamCount(syncDetach, rel.SyncQueries.Detach, 2, path+".sync_queries.detach", contextName, modelName, rel.Name)...)
	}

	// When no explicit `parameters` map is configured, lazy_query must
	// accept exactly the implicit single parameter (the parent's local_key
	// value) — the only binding this generator infers automatically.
	// Relations with scopes/multiple parameters must configure `parameters`
	// explicitly (stage 8, scope.go).
	if len(rel.Parameters) == 0 && g.LazyQuery != nil && len(g.LazyQuery.GetParams()) != 1 {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".lazy_query",
			Context:  contextName,
			Model:    modelName,
			Relation: rel.Name,
			Query:    rel.LazyQuery,
			Message:  fmt.Sprintf("lazy_query %q declares %d parameters, but relation %q has no \"parameters\" map to bind them — with no explicit parameters, exactly one implicit parameter (bound to local_key) is expected", rel.LazyQuery, len(g.LazyQuery.GetParams()), rel.Name),
			Hint:     "add a \"parameters\" map naming a source for each query parameter",
		})
	}

	return g, diags
}

// kindsCompatible reports whether a relation of kind a may legitimately
// declare an inverse of kind b (FR-004): belongs_to pairs with has_many/
// has_one and vice versa; many_to_many pairs only with many_to_many.
func kindsCompatible(a, b config.RelationKind) bool {
	switch a {
	case config.BelongsTo:
		return b == config.HasMany || b == config.HasOne
	case config.HasMany, config.HasOne:
		return b == config.BelongsTo
	case config.ManyToMany:
		return b == config.ManyToMany
	default:
		return false
	}
}

// requireParamCount checks that q (already matched by name/command) declares
// exactly n parameters — nil q means the query itself already failed
// validation, in which case there is nothing further to check here.
func requireParamCount(q *pb.Query, queryName string, n int, path, contextName, modelName, relationName string) []diagnostics.Diagnostic {
	if q == nil || len(q.GetParams()) == n {
		return nil
	}
	return []diagnostics.Diagnostic{{
		Severity: diagnostics.SeverityError,
		Path:     path,
		Context:  contextName,
		Model:    modelName,
		Relation: relationName,
		Query:    queryName,
		Message:  fmt.Sprintf("query %q must declare exactly %d parameter(s), found %d", queryName, n, len(q.GetParams())),
	}}
}

func hasColumn(cols []*pb.Column, name string) bool {
	for _, c := range cols {
		if strings.EqualFold(c.Name, name) || strings.EqualFold(c.OriginalName, name) {
			return true
		}
	}
	return false
}

// withRelation copies diags with Relation set to name — contract.
// ValidateOperation doesn't know about the relation it's being validated
// for, so callers here fill in the field internal/diagnostics reserved for
// this feature.
func withRelation(diags []diagnostics.Diagnostic, name string) []diagnostics.Diagnostic {
	out := make([]diagnostics.Diagnostic, len(diags))
	for i, d := range diags {
		d.Relation = name
		out[i] = d
	}
	return out
}
