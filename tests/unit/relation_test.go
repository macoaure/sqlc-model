package unit

import (
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/generate"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// relParentCols/relChildCols are shared by the relation-graph test fixtures
// below: a minimal Parent/Child pair related by has_many/belongs_to.
func relParentCols() []*pb.Column {
	return []*pb.Column{pcol("id", "uuid", true), pcol("name", "text", true)}
}
func relChildCols() []*pb.Column {
	return []*pb.Column{pcol("id", "uuid", true), pcol("parent_id", "uuid", true)}
}

func relBaseQueries() []*pb.Query {
	return []*pb.Query{
		{Name: "GetParent", Cmd: ":one", Columns: relParentCols(), Params: []*pb.Parameter{pparam(1, "id", "uuid", true)}, Text: "SELECT id, name FROM parents WHERE id = $1;"},
		{Name: "CreateParent", Cmd: ":one", Columns: relParentCols(), Params: []*pb.Parameter{pparam(1, "name", "text", true)}, Text: "INSERT INTO parents (name) VALUES ($1) RETURNING id, name;"},
		{Name: "GetChild", Cmd: ":one", Columns: relChildCols(), Params: []*pb.Parameter{pparam(1, "id", "uuid", true)}, Text: "SELECT id, parent_id FROM children WHERE id = $1;"},
		{Name: "CreateChild", Cmd: ":one", Columns: relChildCols(), Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true)}, Text: "INSERT INTO children (parent_id) VALUES ($1) RETURNING id, parent_id;"},
		{Name: "ListChildrenByParent", Cmd: ":many", Columns: relChildCols(), Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true)}, Text: "SELECT id, parent_id FROM children WHERE parent_id = $1;"},
		{Name: "ListChildrenByParentIDs", Cmd: ":many", Columns: relChildCols(), Params: []*pb.Parameter{pparam(1, "parent_ids", "uuid", true)}, Text: "SELECT id, parent_id FROM children WHERE parent_id = ANY($1);"},
	}
}

// relOptions builds a full options document from a "relations" JSON
// fragment for Parent's Children relation, so each test case only needs to
// supply the part it's exercising.
func relOptions(childRelations string) string {
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
						"Children": ` + childRelations + `
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

func relRequest(t *testing.T, childRelations string, mutateQueries func([]*pb.Query) []*pb.Query) *pb.GenerateRequest {
	t.Helper()
	queries := relBaseQueries()
	if mutateQueries != nil {
		queries = mutateQueries(queries)
	}
	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(relOptions(childRelations)),
		Queries:       queries,
	}
}

func TestRelation_ValidHasManyGenerates(t *testing.T) {
	req := relRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent"
	}`, nil)
	resp, diags := generate.Generate(req)
	if resp == nil {
		t.Fatalf("expected successful generation, got diagnostics: %+v", diags)
	}
}

func TestRelation_UnknownKindFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_none", "model": "Child", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_MissingModelFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_MissingLazyQueryFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_MissingLocalKeyFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "Child", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_MissingForeignKeyFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "Child", "local_key": "id", "lazy_query": "ListChildrenByParent"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_UnresolvedTargetModelFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "NoSuchModel", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_UnresolvedLocalKeyFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "Child", "local_key": "nonexistent", "foreign_key": "parent_id", "lazy_query": "ListChildrenByParent"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_LazyQueryWrongCommandFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "GetChild"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_LazyQueryUnresolvedFails(t *testing.T) {
	req := relRequest(t, `{"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id", "lazy_query": "NoSuchQuery"}`, nil)
	assertRelationError(t, req)
}

func TestRelation_ImplicitParamCountMismatchFails(t *testing.T) {
	req := relRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParentIDs"
	}`, func(qs []*pb.Query) []*pb.Query {
		for _, q := range qs {
			if q.Name == "ListChildrenByParentIDs" {
				q.Params = append(q.Params, pparam(2, "extra", "uuid", true))
			}
		}
		return qs
	})
	assertRelationError(t, req)
}

// --- Inverse validation ---

func TestRelation_ValidInverseGenerates(t *testing.T) {
	req := relRequestWithChildInverse(t, `"inverse": "Parent"`, `
		"kind": "belongs_to", "model": "Parent", "local_key": "parent_id", "foreign_key": "id",
		"nullable": false, "lazy_query": "GetParent"
	`)
	resp, diags := generate.Generate(req)
	if resp == nil {
		t.Fatalf("expected successful generation with a valid inverse pair, got: %+v", diags)
	}
}

func TestRelation_UnknownInverseFails(t *testing.T) {
	req := relRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"inverse": "NoSuchRelation", "lazy_query": "ListChildrenByParent"
	}`, nil)
	assertRelationError(t, req)
}

func TestRelation_IncompatibleInverseKindFails(t *testing.T) {
	req := relRequestWithChildInverse(t, `"inverse": "Parent"`, `
		"kind": "has_many", "model": "Parent", "local_key": "parent_id", "foreign_key": "id",
		"lazy_query": "GetParent"
	`)
	assertRelationError(t, req)
}

// relRequestWithChildInverse declares Parent.Children (has_many, with the
// given inverse fragment) and a matching relation on Child (childRelation
// body), so inverse-pairing tests can vary the Child side independently.
func relRequestWithChildInverse(t *testing.T, inverseFragment, childRelation string) *pb.GenerateRequest {
	t.Helper()
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
						"Children": {
							"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
							"lazy_query": "ListChildrenByParent", ` + inverseFragment + `
						}
					}
				},
				"Child": {
					"row": "Child",
					"operations": {"find": "GetChild", "insert": "CreateChild"},
					"fields": {"id": {"readable": true}, "parent_id": {"readable": true}},
					"relations": {
						"Parent": {` + childRelation + `}
					}
				}
			}
		}]
	}`
	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries:       relBaseQueries(),
	}
}

// --- Eager-query validation ---

func TestRelation_EagerQueryWrongParamCountFails(t *testing.T) {
	req := relRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent", "eager_query": "ListChildrenByParentIDs"
	}`, func(qs []*pb.Query) []*pb.Query {
		for _, q := range qs {
			if q.Name == "ListChildrenByParentIDs" {
				q.Params = append(q.Params, pparam(2, "extra", "uuid", true))
			}
		}
		return qs
	})
	assertRelationError(t, req)
}

func TestRelation_EagerQueryMissingForeignKeyColumnFails(t *testing.T) {
	req := relRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent", "eager_query": "ListChildrenByParentIDs"
	}`, func(qs []*pb.Query) []*pb.Query {
		for _, q := range qs {
			if q.Name == "ListChildrenByParentIDs" {
				q.Columns = []*pb.Column{pcol("id", "uuid", true)} // drop parent_id
			}
		}
		return qs
	})
	assertRelationError(t, req)
}

func TestRelation_EagerQueryValidGenerates(t *testing.T) {
	req := relRequest(t, `{
		"kind": "has_many", "model": "Child", "local_key": "id", "foreign_key": "parent_id",
		"lazy_query": "ListChildrenByParent", "eager_query": "ListChildrenByParentIDs"
	}`, nil)
	resp, diags := generate.Generate(req)
	if resp == nil {
		t.Fatalf("expected successful generation with a valid eager_query, got: %+v", diags)
	}
}

func TestRelation_ManyToManyEagerQueryUnsupportedFails(t *testing.T) {
	req := relRequest(t, `{
		"kind": "many_to_many", "model": "Child", "local_key": "id",
		"lazy_query": "ListChildrenByParent", "eager_query": "ListChildrenByParentIDs"
	}`, nil)
	assertRelationError(t, req)
}

func assertRelationError(t *testing.T, req *pb.GenerateRequest) {
	t.Helper()
	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected generation to fail, got %d files", len(resp.Files))
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic, got %+v", diags)
	}
}
