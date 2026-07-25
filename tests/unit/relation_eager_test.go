package unit

import (
	"testing"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// eagerOptions declares Parent has_many Child (with eager_query and
// inverse), and Child belongs_to Parent, for nested-eager and inverse
// hydration coverage.
func eagerOptions(nested string) string {
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
							"inverse": "Owner", "lazy_query": "ListChildrenByParent", "eager_query": "ListChildrenByParentIDs"
						}
					}
				},
				"Child": {
					"row": "Child",
					"operations": {"find": "GetChild", "insert": "CreateChild"},
					"fields": {"id": {"readable": true}, "parent_id": {"readable": true}},
					"relations": {
						"Owner": {
							"kind": "belongs_to", "model": "Parent", "local_key": "parent_id", "foreign_key": "id",
							"nullable": false, "lazy_query": "GetParent"
						}` + nested + `
					}
				}` + nestedGrandchildModel(nested) + `
			}
		}]
	}`
}

// nestedGrandchildModel adds a Grandchild model when nested requests a
// second-level eager relation, so TestRelation_NestedEagerGenerates can
// exercise two levels of nesting.
func nestedGrandchildModel(nested string) string {
	if nested == "" {
		return ""
	}
	return `,
	"Grandchild": {
		"row": "Grandchild",
		"operations": {"find": "GetGrandchild", "insert": "CreateGrandchild"},
		"fields": {"id": {"readable": true}, "child_id": {"readable": true}}
	}`
}

func eagerQueries() []*pb.Query {
	qs := relBaseQueries()
	qs = append(qs,
		&pb.Query{Name: "GetGrandchild", Cmd: ":one", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("child_id", "uuid", true)}, Params: []*pb.Parameter{pparam(1, "id", "uuid", true)}, Text: "SELECT id, child_id FROM grandchildren WHERE id = $1;"},
		&pb.Query{Name: "CreateGrandchild", Cmd: ":one", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("child_id", "uuid", true)}, Params: []*pb.Parameter{pparam(1, "child_id", "uuid", true)}, Text: "INSERT INTO grandchildren (child_id) VALUES ($1) RETURNING id, child_id;"},
		&pb.Query{Name: "ListGrandchildrenByChild", Cmd: ":many", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("child_id", "uuid", true)}, Params: []*pb.Parameter{pparam(1, "child_id", "uuid", true)}, Text: "SELECT id, child_id FROM grandchildren WHERE child_id = $1;"},
		&pb.Query{Name: "ListGrandchildrenByChildIDs", Cmd: ":many", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("child_id", "uuid", true)}, Params: []*pb.Parameter{pparam(1, "child_ids", "uuid", true)}, Text: "SELECT id, child_id FROM grandchildren WHERE child_id = ANY($1);"},
	)
	return qs
}

func TestRelation_EagerLoadWithInverseGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(eagerOptions("")),
		Queries:       eagerQueries(),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_NestedEagerGenerates(t *testing.T) {
	nested := `,
		"Grandchildren": {
			"kind": "has_many", "model": "Grandchild", "local_key": "id", "foreign_key": "child_id",
			"lazy_query": "ListGrandchildrenByChild", "eager_query": "ListGrandchildrenByChildIDs"
		}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(eagerOptions(nested)),
		Queries:       eagerQueries(),
	}
	assertRelationSuccess(t, req)
}

// --- Many-to-many pivot coverage ---

func pivotQueries() []*pb.Query {
	qs := relBaseQueries()
	qs = append(qs,
		&pb.Query{Name: "GetTag", Cmd: ":one", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("name", "text", true)}, Params: []*pb.Parameter{pparam(1, "id", "uuid", true)}, Text: "SELECT id, name FROM tags WHERE id = $1;"},
		&pb.Query{Name: "CreateTag", Cmd: ":one", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("name", "text", true)}, Params: []*pb.Parameter{pparam(1, "name", "text", true)}, Text: "INSERT INTO tags (name) VALUES ($1) RETURNING id, name;"},
		&pb.Query{Name: "ListTagsByParent", Cmd: ":many", Columns: []*pb.Column{pcol("id", "uuid", true), pcol("name", "text", true)}, Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true)}, Text: "SELECT id, name FROM tags JOIN parent_tags ON parent_tags.tag_id = tags.id WHERE parent_tags.parent_id = $1;"},
		&pb.Query{Name: "AttachTagToParent", Cmd: ":exec", Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true), pparam(2, "tag_id", "uuid", true)}, Text: "INSERT INTO parent_tags (parent_id, tag_id) VALUES ($1, $2);"},
		&pb.Query{Name: "DetachTagFromParent", Cmd: ":exec", Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true), pparam(2, "tag_id", "uuid", true)}, Text: "DELETE FROM parent_tags WHERE parent_id = $1 AND tag_id = $2;"},
		&pb.Query{Name: "ListTagIDsByParent", Cmd: ":many", Columns: []*pb.Column{pcol("tag_id", "uuid", true)}, Params: []*pb.Parameter{pparam(1, "parent_id", "uuid", true)}, Text: "SELECT tag_id FROM parent_tags WHERE parent_id = $1;"},
	)
	return qs
}

func pivotOptions(tagsRelation string) string {
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
					"relations": {"Tags": ` + tagsRelation + `}
				},
				"Tag": {
					"row": "Tag",
					"operations": {"find": "GetTag", "insert": "CreateTag"},
					"fields": {"id": {"readable": true}, "name": {"readable": true}}
				}
			}
		}]
	}`
}

func TestRelation_ManyToManyAttachDetachSyncGenerates(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
			"lazy_query": "ListTagsByParent",
			"attach_query": "AttachTagToParent", "detach_query": "DetachTagFromParent",
			"sync_queries": {"list": "ListTagIDsByParent", "attach": "AttachTagToParent", "detach": "DetachTagFromParent"}
		}`)),
		Queries: pivotQueries(),
	}
	assertRelationSuccess(t, req)
}

func TestRelation_ManyToManyMissingTargetKeyFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id",
			"lazy_query": "ListTagsByParent",
			"attach_query": "AttachTagToParent", "detach_query": "DetachTagFromParent"
		}`)),
		Queries: pivotQueries(),
	}
	assertRelationError(t, req)
}

func TestRelation_ManyToManyUnresolvedTargetKeyFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "nonexistent",
			"lazy_query": "ListTagsByParent", "attach_query": "AttachTagToParent"
		}`)),
		Queries: pivotQueries(),
	}
	assertRelationError(t, req)
}

func TestRelation_AttachQueryWrongParamCountFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
			"lazy_query": "ListTagsByParent", "attach_query": "ListTagsByParent"
		}`)),
		Queries: pivotQueries(),
	}
	assertRelationError(t, req)
}

func TestRelation_SyncQueriesPartialFails(t *testing.T) {
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
			"lazy_query": "ListTagsByParent",
			"sync_queries": {"list": "ListTagIDsByParent", "attach": "", "detach": "DetachTagFromParent"}
		}`)),
		Queries: pivotQueries(),
	}
	assertRelationError(t, req)
}

func TestRelation_SyncListWrongColumnCountFails(t *testing.T) {
	queries := pivotQueries()
	for _, q := range queries {
		if q.Name == "ListTagIDsByParent" {
			q.Columns = append(q.Columns, pcol("extra", "text", true))
		}
	}
	req := &pb.GenerateRequest{
		Settings: &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(pivotOptions(`{
			"kind": "many_to_many", "model": "Tag", "local_key": "id", "target_key": "id",
			"lazy_query": "ListTagsByParent",
			"sync_queries": {"list": "ListTagIDsByParent", "attach": "AttachTagToParent", "detach": "DetachTagFromParent"}
		}`)),
		Queries: queries,
	}
	assertRelationError(t, req)
}
