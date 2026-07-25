package config

import (
	"encoding/json"
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
	Relations  []RelationConfiguration
	Queries    []QueryConfiguration
	Lookups    []LookupOperation
}

type QueryTerminal string

const (
	QueryTerminalGet     QueryTerminal = "get"
	QueryTerminalFirst   QueryTerminal = "first"
	QueryTerminalFind    QueryTerminal = "find"
	QueryTerminalDelete  QueryTerminal = "delete"
	QueryTerminalRefresh QueryTerminal = "refresh"
)

type QueryConfiguration struct {
	Name      string
	Operation string
	Terminal  QueryTerminal
	Scopes    []QueryScope
}

type QueryScope struct {
	Name      string
	Parameter string
	Value     json.RawMessage
	ValueSet  bool
	Argument  string
	Query     string
	Relation  string
}

type LookupOperation struct {
	Name  string
	Query string
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
	Relations  OrderedObject  `json:"relations"`
	Queries    OrderedObject  `json:"queries"`
	Lookups    OrderedObject  `json:"lookups"`
}

type queryConfigurationJSON struct {
	Operation string        `json:"operation"`
	Terminal  string        `json:"terminal"`
	Scopes    OrderedObject `json:"scopes"`
}

type queryScopeJSON struct {
	Parameter string          `json:"parameter"`
	Value     json.RawMessage `json:"value"`
	Argument  string          `json:"argument"`
	Query     string          `json:"query"`
	Relation  string          `json:"relation"`
}

type lookupOperationJSON struct {
	Query string `json:"query"`
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

	seenRelations := make(map[string]bool, len(mj.Relations))
	for _, entry := range mj.Relations {
		if seenRelations[entry.Key] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fmt.Sprintf("%s.relations.%s", path, entry.Key),
				Context:  contextName,
				Model:    name,
				Relation: entry.Key,
				Message:  fmt.Sprintf("model %q: duplicate relation %q", name, entry.Key),
			})
			continue
		}
		seenRelations[entry.Key] = true

		relPath := fmt.Sprintf("%s.relations.%s", path, entry.Key)
		rc, rdiags := decodeRelation(entry.Key, entry.Value, relPath, contextName, name)
		diags = append(diags, rdiags...)
		mc.Relations = append(mc.Relations, rc)
	}

	seenQueries := make(map[string]bool, len(mj.Queries))
	for _, entry := range mj.Queries {
		if seenQueries[entry.Key] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fmt.Sprintf("%s.queries.%s", path, entry.Key),
				Context:  contextName,
				Model:    name,
				Message:  fmt.Sprintf("model %q: duplicate query %q", name, entry.Key),
			})
			continue
		}
		seenQueries[entry.Key] = true

		queryPath := fmt.Sprintf("%s.queries.%s", path, entry.Key)
		qc, qdiags := decodeQuery(entry.Key, entry.Value, queryPath, contextName, name)
		diags = append(diags, qdiags...)
		mc.Queries = append(mc.Queries, qc)
	}

	seenLookups := make(map[string]bool, len(mj.Lookups))
	for _, entry := range mj.Lookups {
		if seenLookups[entry.Key] {
			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Path:     fmt.Sprintf("%s.lookups.%s", path, entry.Key),
				Context:  contextName,
				Model:    name,
				Message:  fmt.Sprintf("model %q: duplicate lookup %q", name, entry.Key),
			})
			continue
		}
		seenLookups[entry.Key] = true

		lookupPath := fmt.Sprintf("%s.lookups.%s", path, entry.Key)
		lo, ldiags := decodeLookup(entry.Key, entry.Value, lookupPath, contextName, name)
		diags = append(diags, ldiags...)
		mc.Lookups = append(mc.Lookups, lo)
	}

	return mc, diags
}

func decodeQuery(name string, raw []byte, path, context, model string) (QueryConfiguration, []diagnostics.Diagnostic) {
	var qj queryConfigurationJSON
	if err := DecodeStrict(raw, &qj); err != nil {
		return QueryConfiguration{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("query %q: invalid query configuration: %v", name, err),
		}}
	}

	qc := QueryConfiguration{Name: name, Operation: qj.Operation, Terminal: QueryTerminal(qj.Terminal)}
	var diags []diagnostics.Diagnostic
	errf := func(suffix, format string, args ...any) {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + suffix,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf(format, args...),
		})
	}
	if qj.Operation == "" {
		errf(".operation", "query %q: \"operation\" is required", name)
	}
	switch qc.Terminal {
	case QueryTerminalGet, QueryTerminalFirst, QueryTerminalFind, QueryTerminalDelete, QueryTerminalRefresh:
	case "":
		errf(".terminal", "query %q: \"terminal\" is required", name)
	default:
		errf(".terminal", "query %q: unsupported terminal %q", name, qj.Terminal)
	}

	seenScopes := make(map[string]bool, len(qj.Scopes))
	for _, entry := range qj.Scopes {
		if seenScopes[entry.Key] {
			errf(".scopes."+entry.Key, "query %q: duplicate scope %q", name, entry.Key)
			continue
		}
		seenScopes[entry.Key] = true

		scopePath := path + ".scopes." + entry.Key
		sc, sdiags := decodeQueryScope(entry.Key, entry.Value, scopePath, context, model, name)
		diags = append(diags, sdiags...)
		qc.Scopes = append(qc.Scopes, sc)
	}

	return qc, diags
}

func decodeQueryScope(name string, raw []byte, path, context, model, queryName string) (QueryScope, []diagnostics.Diagnostic) {
	var sj queryScopeJSON
	if err := DecodeStrict(raw, &sj); err != nil {
		return QueryScope{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("query %q scope %q: invalid scope configuration: %v", queryName, name, err),
		}}
	}

	sc := QueryScope{
		Name:      name,
		Parameter: sj.Parameter,
		Value:     sj.Value,
		ValueSet:  sj.Value != nil,
		Argument:  sj.Argument,
		Query:     sj.Query,
		Relation:  sj.Relation,
	}

	kinds := 0
	if sc.ValueSet {
		kinds++
	}
	if sc.Argument != "" {
		kinds++
	}
	if sc.Query != "" {
		kinds++
	}
	if sc.Relation != "" {
		kinds++
	}

	var diags []diagnostics.Diagnostic
	if kinds != 1 {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("query %q scope %q: exactly one of value, argument, query, or relation is required", queryName, name),
		})
	}
	if (sc.ValueSet || sc.Argument != "") && sc.Parameter == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     path + ".parameter",
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("query %q scope %q: \"parameter\" is required for parameter scopes", queryName, name),
		})
	}
	return sc, diags
}

func decodeLookup(name string, raw []byte, path, context, model string) (LookupOperation, []diagnostics.Diagnostic) {
	var lj lookupOperationJSON
	if err := DecodeStrict(raw, &lj); err != nil {
		return LookupOperation{}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("lookup %q: invalid lookup configuration: %v", name, err),
		}}
	}
	if lj.Query == "" {
		return LookupOperation{Name: name}, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path + ".query",
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("lookup %q: \"query\" is required", name),
		}}
	}
	return LookupOperation{Name: name, Query: lj.Query}, nil
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
