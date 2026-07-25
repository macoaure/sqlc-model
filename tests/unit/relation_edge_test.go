package unit

import (
	"strings"
	"testing"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// --- kindsCompatible / requireParamCount additional branches ---

func TestRelation_HasOneInverseCompatible(t *testing.T) {
	// has_one's inverse (like has_many's) must be belongs_to — the child
	// side always "belongs to" the one-record-owning parent regardless of
	// whether the parent side is itself has_one or has_many.
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Parent": {
					"row": "Parent",
					"operations": {"find": "GetParent", "insert": "CreateParent"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}},
					"relations": {
						"Child": {
							"kind": "has_one", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
							"inverse": "Parent", "lazy_query": "GetChild"
						}
					}
				},
				"Child": {
					"row": "Child",
					"operations": {"find": "GetChild", "insert": "CreateChild"},
					"fields": {"id": {"readable": true}, "parent_id": {"readable": true}},
					"relations": {
						"Parent": {
							"kind": "belongs_to", "model": "Parent", "local_key": "parent_id", "foreign_key": "id",
							"nullable": false, "lazy_query": "GetParent"
						}
					}
				}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       relBaseQueries(),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_ManyToManyInverseCompatible(t *testing.T) {
	queries := pivotQueries()
	queries = append(queries,
		&pb.Query{Name: "ListParentsByTag", Cmd: ":many", Columns: relParentCols(), Params: []*pb.Parameter{pparam(1, "tag_id", "uuid", true)}, Text: "SELECT id, name FROM parents JOIN parent_tags ON parent_tags.parent_id = parents.id WHERE parent_tags.tag_id = $1;"},
	)
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Parent": {
					"row": "Parent",
					"operations": {"find": "GetParent", "insert": "CreateParent"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}},
					"relations": {
						"Tags": {
							"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
							"inverse": "Parents", "lazy_query": "ListTagsByParent"
						}
					}
				},
				"Tag": {
					"row": "Tag",
					"operations": {"find": "GetTag", "insert": "CreateTag"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}},
					"relations": {
						"Parents": {
							"kind": "many_to_many", "model": "Parent", "local_key": "id", "target_key": "id",
							"lazy_query": "ListParentsByTag"
						}
					}
				}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       queries,
	}
	assertRelationSuccess(t, req)
}

func TestRelation_UnknownKindWithInverseFails(t *testing.T) {
	req := relRequestWithChildInverse(t, `"inverse": "Parent"`, `
		"kind": "belongs_to", "model": "Parent", "local_key": "parent_id", "foreign_key": "id",
		"nullable": false, "lazy_query": "GetParent"
	`)
	// relRequestWithChildInverse always declares Parent.Children as
	// has_many; swap its kind to an invalid value to exercise
	// kindsCompatible's unreachable-in-well-formed-config default branch
	// (an unknown kind that still declares an inverse).
	options := string(req.PluginOptions)
	options = strings.Replace(options, `"kind": "has_many", "model": "Child"`, `"kind": "has_none", "model": "Child"`, 1)
	req.PluginOptions = []byte(options)
	assertRelationError(t, req)
}

func TestRelation_ManyToManyInverseIncompatibleFails(t *testing.T) {
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Parent": {
					"row": "Parent",
					"operations": {"find": "GetParent", "insert": "CreateParent"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}},
					"relations": {
						"Tags": {
							"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
							"inverse": "Parent", "lazy_query": "ListTagsByParent"
						}
					}
				},
				"Tag": {
					"row": "Tag",
					"operations": {"find": "GetTag", "insert": "CreateTag"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}},
					"relations": {
						"Parent": {
							"kind": "belongs_to", "model": "Parent", "local_key": "id", "foreign_key": "id",
							"nullable": false, "lazy_query": "GetParent"
						}
					}
				}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       pivotQueries(),
	}
	assertRelationError(t, req)
}

// TestRelation_ScopeQueryVariantSameColumnsDifferentTypesGenerates exercises
// sameColumnSet's true branch with a variant query sharing the exact same
// column set as lazy_query (already covered), alongside a three-parameter
// relation to exercise sortedRelationParams' insertion-sort swap path with
// more than one out-of-order pair.
func TestRelation_ThreeParameterScopeGenerates(t *testing.T) {
	qs := relBaseQueries()
	qs = append(qs, &pb.Query{
		Name: "ListChildrenFiltered", Cmd: ":many", Columns: relChildCols(),
		Params: []*pb.Parameter{
			pparam(1, "limit_count", "int4", true),
			pparam(2, "parent_id", "uuid", true),
			pparam(3, "min_rank", "int4", true),
		},
		Text: "SELECT id, parent_id FROM children WHERE parent_id = $2 AND rank >= $3 LIMIT $1;",
	})
	options := relOptions(`{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenFiltered",
		"parameters": {
			"parent_id": {"source": "parent.id"},
			"min_rank": {"source": "scope.MinRank", "default": 0},
			"limit_count": {"source": "scope.Limit", "default": 100}
		},
		"scopes": {
			"MinRank": {"parameter": "min_rank", "argument": "int32"},
			"Limit": {"parameter": "limit_count", "argument": "int32"}
		}
	}`)
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       qs,
	}
	assertRelationSuccess(t, req)
}

// TestRelation_ScopeOnParentBoundParamFallsBackToStringType exercises
// paramGoTypeForScope's fallback branch: a scope's `parameter` may
// structurally reference a query parameter that's bound via `parent.<key>`
// rather than `scope.<Name>` (unusual, but not itself invalid — nothing
// requires a scope's target parameter to be scope-sourced). Since no
// LazyParam entry then has a matching ScopeName, the Go-type lookup falls
// back to "string" rather than panicking or erroring.
func TestRelation_ScopeOnParentBoundParamFallsBackToStringType(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "parent.id"}},
		"scopes": {"Weird": {"parameter": "published", "value": true}}
	}`)
	assertRelationSuccess(t, req)
}

// TestRelation_ScopeNumericValueOnUnmappedGoTypeGenerates exercises
// goLiteral's numeric-literal fallback (a JSON number bound to a Go type
// that's neither an int nor a float family member, e.g. pgtype.Numeric) —
// still renders as an integer literal rather than failing.
func TestRelation_ScopeNumericValueOnUnmappedGoTypeGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("numeric", "5")),
		Queries:       numericScopeQueries("numeric"),
	}
	assertRelationSuccess(t, req)
}

// TestRelation_ThreeParameterScopeOutOfOrderGenerates declares a query
// whose *positional* Number order differs from its declaration order in
// the Params slice, forcing sortedRelationParams' insertion sort to
// actually swap elements (relation_test.go's ThreeParameterScopeGenerates
// already had Numbers 1,2,3 in that same order, so no swap occurred).
func TestRelation_ThreeParameterScopeOutOfOrderGenerates(t *testing.T) {
	qs := relBaseQueries()
	qs = append(qs, &pb.Query{
		Name: "ListChildrenFilteredOutOfOrder", Cmd: ":many", Columns: relChildCols(),
		Params: []*pb.Parameter{
			pparam(3, "min_rank", "int4", true),
			pparam(1, "limit_count", "int4", true),
			pparam(2, "parent_id", "uuid", true),
		},
		Text: "SELECT id, parent_id FROM children WHERE parent_id = $2 AND rank >= $3 LIMIT $1;",
	})
	options := relOptions(`{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenFilteredOutOfOrder",
		"parameters": {
			"parent_id": {"source": "parent.id"},
			"min_rank": {"source": "scope.MinRank", "default": 0},
			"limit_count": {"source": "scope.Limit", "default": 100}
		},
		"scopes": {
			"MinRank": {"parameter": "min_rank", "argument": "int32"},
			"Limit": {"parameter": "limit_count", "argument": "int32"}
		}
	}`)
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       qs,
	}
	assertRelationSuccess(t, req)
}

// TestRelation_ScopeNumericValueIncompatibleWithBoolFails exercises
// valueCompatible's float64-value-against-non-numeric-type branch: a JSON
// number bound to a bool-typed parameter is incompatible.
func TestRelation_ScopeNumericValueIncompatibleWithBoolFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published"}},
		"scopes": {"Published": {"parameter": "published", "value": 5}}
	}`)
	assertRelationError(t, req)
}

// TestRelation_QueryVariantSameLengthDifferentColumnsFails exercises
// sameColumnSet's same-length-but-different-content branch (the earlier
// QueryVariantResultShapeMismatchFails only covers a length mismatch).
func TestRelation_QueryVariantSameLengthDifferentColumnsFails(t *testing.T) {
	queries := scopeQueries()
	queries = append(queries, &pb.Query{
		Name: "SameLengthDifferentColumns", Cmd: ":many",
		Columns: []*pb.Column{pcol("id", "uuid", true), pcol("other_id", "uuid", true)},
		Params:  []*pb.Parameter{pparam(1, "parent_id", "uuid", true)},
		Text:    "SELECT id, other_id FROM children WHERE parent_id = $1;",
	})
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(relOptions(`{
			"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
			"lazy_query": "ListChildrenByParent",
			"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
			"scopes": {
				"Published": {"parameter": "published", "value": true},
				"Weird": {"query": "SameLengthDifferentColumns"}
			}
		}`)),
		Queries: queries,
	}
	assertRelationError(t, req)
}

// TestRelation_AttachQueryExecRowsAltCommandGenerates exercises
// ValidateOperation's altCmd branch: attach_query/detach_query accept
// :execrows equally alongside :exec, with no warning.
func TestRelation_AttachQueryExecRowsAltCommandGenerates(t *testing.T) {
	queries := pivotQueries()
	for _, q := range queries {
		if q.Name == "AttachTagToParent" || q.Name == "DetachTagFromParent" {
			q.Cmd = ":execrows"
		}
	}
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
			"lazy_query": "ListTagsByParent",
			"attach_query": "AttachTagToParent", "detach_query": "DetachTagFromParent"
		}`)),
		Queries: queries,
	}
	assertRelationSuccess(t, req)
}

// TestRelation_ScopeOnEagerOnlyParameterGenerates exercises ValidateScopes'
// fallback-to-eager-query branch: a scope whose `parameter` names
// eager_query's sole parameter (not one of lazy_query's) is still resolved
// via a second, eager-query lookup rather than failing outright.
func TestRelation_ScopeOnEagerOnlyParameterGenerates(t *testing.T) {
	options := relOptions(`{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent", "eager_query": "ListChildrenByParentIDs",
		"scopes": {"ByIDs": {"parameter": "parent_ids", "argument": "pgtype.UUID"}}
	}`)
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       relBaseQueries(),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_DetachQueryWrongParamCountFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
			"lazy_query": "ListTagsByParent", "attach_query": "AttachTagToParent", "detach_query": "AttachTagToParent"
		}`)),
		Queries: withExtraParam(pivotQueries(), "AttachTagToParent", "extra"),
	}
	assertRelationError(t, req)
}

func withExtraParam(qs []*pb.Query, name, paramName string) []*pb.Query {
	for _, q := range qs {
		if q.Name == name {
			q.Params = append(q.Params, pparam(int32(len(q.Params)+1), paramName, "uuid", true))
		}
	}
	return qs
}

// --- Scope value-type coverage (valueCompatible / goLiteral) ---

func numericScopeOptions(colType, jsonValue string) string {
	return `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Parent": {
					"row": "Parent",
					"operations": {"find": "GetParent", "insert": "CreateParent"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}},
					"relations": {
						"Children": {
							"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
							"lazy_query": "ListChildrenByRank",
							"parameters": {"parent_id": {"source": "parent.id"}, "min_rank": {"source": "scope.MinRank"}},
							"scopes": {"MinRank": {"parameter": "min_rank", "value": ` + jsonValue + `}}
						}
					}
				},
				"Child": {
					"row": "Child",
					"operations": {"find": "GetChild", "insert": "CreateChild"},
					"fields": {"id": {"readable": true}, "parent_id": {"readable": true}}
				}
			}
		}]
	}`
}

func numericScopeQueries(colType string) []*pb.Query {
	qs := relBaseQueries()
	qs = append(qs, &pb.Query{
		Name: "ListChildrenByRank", Cmd: ":many", Columns: relChildCols(),
		Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true), pparam(2, "min_rank", colType, true)},
		Text:   "SELECT id, parent_id FROM children WHERE parent_id = $1 AND rank >= $2;",
	})
	return qs
}

func TestRelation_ScopeIntegerValueGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("int4", "5")),
		Queries:       numericScopeQueries("int4"),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_ScopeFloatValueGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("float8", "1.5")),
		Queries:       numericScopeQueries("float8"),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_ScopeStringValueGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("text", `"gold"`)),
		Queries:       numericScopeQueries("text"),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_ScopeNullValueGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("int4", "null")),
		Queries:       numericScopeQueries("int4"),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_ScopeStringValueIncompatibleWithIntFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("int4", `"not-a-number"`)),
		Queries:       numericScopeQueries("int4"),
	}
	assertRelationError(t, req)
}

func TestRelation_ScopeBoolValueIncompatibleWithIntFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(numericScopeOptions("int4", "true")),
		Queries:       numericScopeQueries("int4"),
	}
	assertRelationError(t, req)
}

// --- decode-level structural edge cases ---

// TestRelation_DuplicateRelationNameFails relies on OrderedObject's custom
// UnmarshalJSON, which tokenizes the raw JSON stream itself and does not
// deduplicate keys the way encoding/json's map decoding would — so a
// literal repeated object key in the raw options text *does* reach
// decodeModel's duplicate-relation guard as two distinct entries.
func TestRelation_DuplicateRelationNameFails(t *testing.T) {
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Parent": {
					"row": "Parent",
					"operations": {"find": "GetParent", "insert": "CreateParent"},
					"fields": {"id": {"readable": true}},
					"relations": {
						"Children": {"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"},
						"Children": {"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"}
					}
				},
				"Child": {"row": "Child", "operations": {"find": "GetChild", "insert": "CreateChild"}, "fields": {"id": {"readable": true}, "parent_id": {"readable": true}}}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       relBaseQueries(),
	}
	assertRelationError(t, req)
}

// TestRelation_DuplicateParameterNameFails / TestRelation_DuplicateScopeNameFails
// exercise decodeRelation's analogous duplicate-key guards for `parameters`
// and `scopes`, using the same literal-duplicate-JSON-key technique.
func TestRelation_DuplicateParameterNameFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {
			"parent_id": {"source": "parent.id"},
			"parent_id": {"source": "parent.id"},
			"published": {"source": "scope.Published", "default": true}
		},
		"scopes": {"Published": {"parameter": "published", "value": true}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_DuplicateScopeNameFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {
			"Published": {"parameter": "published", "value": true},
			"Published": {"parameter": "published", "value": false}
		}
	}`)
	assertRelationError(t, req)
}

func TestRelation_InvalidRelationJSONFails(t *testing.T) {
	req := relRequest(t, `"not an object"`, nil)
	assertRelationError(t, req)
}

func TestRelation_InvalidScopeJSONFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": "not an object"}
	}`)
	assertRelationError(t, req)
}

func TestRelation_InvalidParameterJSONFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": "not an object"}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ParameterMissingSourceFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {}}
	}`)
	assertRelationError(t, req)
}
