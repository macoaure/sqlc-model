package config

import (
	"encoding/json"
	"fmt"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
)

// RelationKind identifies one of a relation's four supported shapes
// (data-model.md "RelationConfiguration", FR-001).
type RelationKind string

const (
	BelongsTo  RelationKind = "belongs_to"
	HasMany    RelationKind = "has_many"
	HasOne     RelationKind = "has_one"
	ManyToMany RelationKind = "many_to_many"
)

// SyncQueries names the three queries backing a many-to-many relation's
// Sync operation: list currently-attached related IDs, then attach/detach
// the diff (data-model.md "RelationConfiguration.sync_queries").
type SyncQueries struct {
	List   string
	Attach string
	Detach string
}

// ParameterBinding is one entry in a relation's `parameters` map: how one
// sqlc query parameter gets its value at call time (data-model.md
// "ParameterBinding").
type ParameterBinding struct {
	// Name is the query parameter name this binding supplies (map key).
	Name string
	// Source is either "parent.<key>" or "scope.<ScopeName>", unparsed.
	Source string
	// Default is the raw JSON default value; DefaultSet is false when no
	// default was configured.
	Default    json.RawMessage
	DefaultSet bool
}

// ScopeConfiguration is a named, reusable constraint on a relation
// (data-model.md "ScopeConfiguration").
type ScopeConfiguration struct {
	Name      string
	Parameter string
	// Value is the raw JSON fixed value for a fixed-value scope; ValueSet
	// distinguishes "unset" from a legitimate JSON `null`/`false`/`0`.
	Value    json.RawMessage
	ValueSet bool
	// Argument is the declared Go type name for a developer-supplied-value
	// scope; empty means unset.
	Argument string
	// Query names an alternate configured query for a query-variant scope;
	// empty means unset.
	Query string
}

// RelationConfiguration is one named relation declared on a model
// (data-model.md "RelationConfiguration", Key Entity in spec).
type RelationConfiguration struct {
	Name       string
	Kind       RelationKind
	Model      string
	LocalKey   string
	ForeignKey string
	// TargetKey identifies the target model's own field used to bind
	// many-to-many pivot queries (attach_query/detach_query's second
	// parameter, sync_queries.list's diff key) — required only when a
	// many_to_many relation configures Attach/Detach/Sync. Not part of the
	// original relations.schema.json sketch; added during implementation
	// once many-to-many pivot parameter binding needed an explicit target-
	// side field, the same way local_key already identifies the parent
	// side.
	TargetKey   string
	Nullable    bool
	Inverse     string
	LazyQuery   string
	EagerQuery  string
	AttachQuery string
	DetachQuery string
	SyncQueries *SyncQueries
	Parameters  []ParameterBinding
	Scopes      []ScopeConfiguration
}

type syncQueriesJSON struct {
	List   string `json:"list"`
	Attach string `json:"attach"`
	Detach string `json:"detach"`
}

type parameterBindingJSON struct {
	Source  string          `json:"source"`
	Default json.RawMessage `json:"default"`
}

type scopeConfigurationJSON struct {
	Parameter string          `json:"parameter"`
	Value     json.RawMessage `json:"value"`
	Argument  string          `json:"argument"`
	Query     string          `json:"query"`
}

type relationConfigurationJSON struct {
	Kind        string           `json:"kind"`
	Model       string           `json:"model"`
	LocalKey    string           `json:"local_key"`
	ForeignKey  string           `json:"foreign_key"`
	TargetKey   string           `json:"target_key"`
	Nullable    bool             `json:"nullable"`
	Inverse     string           `json:"inverse"`
	LazyQuery   string           `json:"lazy_query"`
	EagerQuery  string           `json:"eager_query"`
	AttachQuery string           `json:"attach_query"`
	DetachQuery string           `json:"detach_query"`
	SyncQueries *syncQueriesJSON `json:"sync_queries"`
	Parameters  OrderedObject    `json:"parameters"`
	Scopes      OrderedObject    `json:"scopes"`
}

var validRelationKinds = map[string]RelationKind{
	string(BelongsTo):  BelongsTo,
	string(HasMany):    HasMany,
	string(HasOne):     HasOne,
	string(ManyToMany): ManyToMany,
}

// decodeRelation decodes and structurally validates one relation entry.
// Cross-model concerns (target/inverse resolution, kind compatibility) and
// query-metadata concerns (command/cardinality, parameter existence, type
// compatibility) are internal/relation's responsibility (data-model.md
// "Relation Graph Validation" stages 7/8) — this function only validates
// what's knowable from the configuration text alone.
func decodeRelation(name string, raw []byte, path, context, model string) (RelationConfiguration, []diagnostics.Diagnostic) {
	var rj relationConfigurationJSON
	if err := DecodeStrict(raw, &rj); err != nil {
		return RelationConfiguration{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Relation: name,
			Message:  fmt.Sprintf("relation %q: invalid relation configuration: %v", name, err),
		}}
	}

	var diags []diagnostics.Diagnostic
	errf := func(suffix, format string, args ...any) {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + suffix,
			Context:  context,
			Model:    model,
			Relation: name,
			Message:  fmt.Sprintf(format, args...),
		})
	}

	kind, ok := validRelationKinds[rj.Kind]
	if !ok {
		errf(".kind", "relation %q: unsupported kind %q, expected one of belongs_to, has_many, has_one, many_to_many", name, rj.Kind)
	}
	if rj.Model == "" {
		errf(".model", "relation %q: \"model\" is required", name)
	}
	if rj.LazyQuery == "" {
		errf(".lazy_query", "relation %q: \"lazy_query\" is required", name)
	}
	// local_key identifies the parent-side field whose current value
	// parameterizes lazy_query/eager_query — needed for every kind,
	// including many_to_many, even though many_to_many's *matching* is
	// entirely delegated to the configured pivot queries (research.md).
	if kind != "" && rj.LocalKey == "" {
		errf(".local_key", "relation %q: \"local_key\" is required", name)
	}
	if kind == BelongsTo || kind == HasMany || kind == HasOne {
		if rj.ForeignKey == "" {
			errf(".foreign_key", "relation %q: \"foreign_key\" is required for kind %q", name, kind)
		}
	}
	if kind == ManyToMany && rj.EagerQuery != "" {
		errf(".eager_query", "relation %q: eager loading is not supported for many_to_many relations in this release", name)
	}
	if kind == ManyToMany && rj.TargetKey == "" && (rj.AttachQuery != "" || rj.DetachQuery != "" || rj.SyncQueries != nil) {
		errf(".target_key", "relation %q: \"target_key\" is required when attach_query/detach_query/sync_queries is configured", name)
	}

	rc := RelationConfiguration{
		Name:        name,
		Kind:        kind,
		Model:       rj.Model,
		LocalKey:    rj.LocalKey,
		ForeignKey:  rj.ForeignKey,
		TargetKey:   rj.TargetKey,
		Nullable:    rj.Nullable,
		Inverse:     rj.Inverse,
		LazyQuery:   rj.LazyQuery,
		EagerQuery:  rj.EagerQuery,
		AttachQuery: rj.AttachQuery,
		DetachQuery: rj.DetachQuery,
	}

	if rj.SyncQueries != nil {
		if rj.SyncQueries.List == "" || rj.SyncQueries.Attach == "" || rj.SyncQueries.Detach == "" {
			errf(".sync_queries", "relation %q: \"sync_queries\" requires \"list\", \"attach\", and \"detach\" all together", name)
		} else {
			rc.SyncQueries = &SyncQueries{List: rj.SyncQueries.List, Attach: rj.SyncQueries.Attach, Detach: rj.SyncQueries.Detach}
		}
	}

	seenParams := make(map[string]bool, len(rj.Parameters))
	for _, entry := range rj.Parameters {
		if seenParams[entry.Key] {
			errf(".parameters."+entry.Key, "relation %q: duplicate parameter %q", name, entry.Key)
			continue
		}
		seenParams[entry.Key] = true

		var pj parameterBindingJSON
		if err := DecodeStrict(entry.Value, &pj); err != nil {
			errf(".parameters."+entry.Key, "relation %q: parameter %q: invalid binding: %v", name, entry.Key, err)
			continue
		}
		if pj.Source == "" {
			errf(".parameters."+entry.Key, "relation %q: parameter %q: \"source\" is required", name, entry.Key)
			continue
		}
		rc.Parameters = append(rc.Parameters, ParameterBinding{
			Name:       entry.Key,
			Source:     pj.Source,
			Default:    pj.Default,
			DefaultSet: pj.Default != nil,
		})
	}

	seenScopes := make(map[string]bool, len(rj.Scopes))
	for _, entry := range rj.Scopes {
		if seenScopes[entry.Key] {
			errf(".scopes."+entry.Key, "relation %q: duplicate scope %q", name, entry.Key)
			continue
		}
		seenScopes[entry.Key] = true

		sc, sdiags := decodeScope(entry.Key, entry.Value, path+".scopes."+entry.Key, context, model, name)
		diags = append(diags, sdiags...)
		rc.Scopes = append(rc.Scopes, sc)
	}

	return rc, diags
}

// decodeScope decodes and structurally validates one scope entry: exactly
// one of value/argument/query must be set, and parameter is required unless
// query is set (data-model.md "ScopeConfiguration" validation; FR-011).
func decodeScope(name string, raw []byte, path, context, model, relationName string) (ScopeConfiguration, []diagnostics.Diagnostic) {
	var sj scopeConfigurationJSON
	if err := DecodeStrict(raw, &sj); err != nil {
		return ScopeConfiguration{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Relation: relationName,
			Message:  fmt.Sprintf("scope %q: invalid scope configuration: %v", name, err),
		}}
	}

	sc := ScopeConfiguration{
		Name:      name,
		Parameter: sj.Parameter,
		Value:     sj.Value,
		ValueSet:  sj.Value != nil,
		Argument:  sj.Argument,
		Query:     sj.Query,
	}

	var diags []diagnostics.Diagnostic
	set := 0
	if sc.ValueSet {
		set++
	}
	if sc.Argument != "" {
		set++
	}
	if sc.Query != "" {
		set++
	}
	if set != 1 {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Relation: relationName,
			Message:  fmt.Sprintf("scope %q: exactly one of \"value\", \"argument\", or \"query\" must be set (got %d)", name, set),
		})
	}
	if sc.Query == "" && sc.Parameter == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Relation: relationName,
			Message:  fmt.Sprintf("scope %q: \"parameter\" is required unless \"query\" is set", name),
		})
	}

	return sc, diags
}
