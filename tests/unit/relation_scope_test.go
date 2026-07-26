package unit

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-model/internal/generate"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// scopeChildCols/scopeQueries add a "published" column/parameter for
// scope-compatibility tests, distinct from relation_test.go's plain
// Parent/Child fixture.
func scopeChildCols() []*pb.Column {
	return []*pb.Column{pcol("id", "uuid", true), pcol("parent_id", "uuid", true)}
}

func scopeQueries() []*pb.Query {
	qs := relBaseQueries()
	for _, q := range qs {
		if q.Name == "ListChildrenByParent" {
			q.Params = append(q.Params, pparam(2, "published", "bool", false))
		}
	}
	qs = append(qs, &pb.Query{
		Name: "ListArchivedChildrenByParent", Cmd: ":many", Columns: scopeChildCols(),
		Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true)},
		Text:   "SELECT id, parent_id FROM children WHERE parent_id = $1 AND archived;",
	})
	return qs
}

func scopeRequest(t *testing.T, childRelations string) *pb.GenerateRequest {
	t.Helper()
	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(relOptions(childRelations)),
		Queries:       scopeQueries(),
	}
}

func assertRelationSuccess(t *testing.T, req *pb.GenerateRequest) {
	t.Helper()
	resp, diags := generate.Generate(req)
	if resp == nil {
		t.Fatalf("expected successful generation, got diagnostics: %+v", diags)
	}
}

func TestRelation_ValueScopeGenerates(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published"}},
		"scopes": {"Published": {"parameter": "published", "value": true}}
	}`)
	assertRelationSuccess(t, req)
}

func TestRelation_ArgumentScopeGenerates(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Recent", "default": false}},
		"scopes": {"Recent": {"parameter": "published", "argument": "pgtype.Bool"}}
	}`)
	assertRelationSuccess(t, req)
}

func TestRelation_QueryVariantScopeGenerates(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {
			"Published": {"parameter": "published", "value": true},
			"Archived": {"query": "ListArchivedChildrenByParent"}
		}
	}`)
	assertRelationSuccess(t, req)
}

func TestRelation_ScopeUnknownParameterFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": {"parameter": "nonexistent", "value": true}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ScopeDiagnosticIncludesRelation(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": {"parameter": "nonexistent", "value": true}}
	}`)
	_, diags := generate.Generate(req)
	for _, d := range diags {
		if d.Relation == "Children" && strings.Contains(d.Path, ".scopes.Published") {
			return
		}
	}
	t.Fatalf("expected scope diagnostic to include relation, got %+v", diags)
}

func TestRelation_ScopeNameCollisionFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Get", "default": true}},
		"scopes": {"Get": {"parameter": "published", "value": true}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ScopeArgumentTypeMismatchFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Recent", "default": true}},
		"scopes": {"Recent": {"parameter": "published", "argument": "int32"}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ScopeValueTypeMismatchFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published"}},
		"scopes": {"Published": {"parameter": "published", "value": [1, 2, 3]}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ScopeReferencesUndeclaredScopeFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.NoSuchScope"}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ParameterUnboundQueryParamFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ParameterInvalidSourceFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "bogus.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": {"parameter": "published", "value": true}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ParameterUnresolvedParentKeyFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.nonexistent"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": {"parameter": "published", "value": true}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ParameterUnknownNameFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}, "bogus": {"source": "parent.id"}},
		"scopes": {"Published": {"parameter": "published", "value": true}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_QueryVariantResultShapeMismatchFails(t *testing.T) {
	queries := scopeQueries()
	queries = append(queries, &pb.Query{
		Name: "BadShapeQuery", Cmd: ":many",
		Columns: []*pb.Column{pcol("id", "uuid", true)}, // missing parent_id
		Params:  []*pb.Parameter{pparam(1, "parent_id", "uuid", true)},
		Text:    "SELECT id FROM children WHERE parent_id = $1;",
	})
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(relOptions(`{
			"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
			"lazy_query": "ListChildrenByParent",
			"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
			"scopes": {
				"Published": {"parameter": "published", "value": true},
				"Archived": {"query": "BadShapeQuery"}
			}
		}`)),
		Queries: queries,
	}
	assertRelationError(t, req)
}

func TestRelation_ScopeExactlyOneOfViolationFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": {"parameter": "published", "value": true, "argument": "bool"}}
	}`)
	assertRelationError(t, req)
}

func TestRelation_ScopeMissingParameterAndQueryFails(t *testing.T) {
	req := scopeRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent",
		"parameters": {"parent_id": {"source": "parent.id"}, "published": {"source": "scope.Published", "default": true}},
		"scopes": {"Published": {"value": true}}
	}`)
	assertRelationError(t, req)
}
