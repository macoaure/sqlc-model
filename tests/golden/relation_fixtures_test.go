package golden

import (
	"testing"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// userRelationsOptions declares User has_many Posts / Post belongs_to
// Author, per specs/002-relations-lazy-eager-loading/quickstart.md steps
// 1-2, with a Published scope, an eager_query + nested Tags eager
// relation on Posts, and a Tags many-to-many relation on Post for
// attach/detach/sync coverage.
const userRelationsOptions = `{
  "version": 1,
  "sqlc": {
    "package": "sqlcdb",
    "import": "example.com/project/internal/database/sqlc",
    "driver": "pgx/v5"
  },
  "contexts": [
    {
      "name": "content",
      "package": "content",
      "directory": "content",
      "models": {
        "User": {
          "row": "User",
          "operations": {
            "find": "GetUser",
            "insert": "CreateUser"
          },
          "fields": {
            "id": {"readable": true, "generated": "insert"},
            "name": {"readable": true, "fillable": true, "mutable": true}
          },
          "relations": {
            "Posts": {
              "kind": "has_many",
              "model": "Post",
              "local_key": "id",
              "foreign_key": "user_id",
              "inverse": "Author",
              "lazy_query": "ListPostsByUser",
              "eager_query": "ListPostsByUserIDs",
              "parameters": {
                "user_id": {"source": "parent.id"},
                "published": {"source": "scope.Published"}
              },
              "scopes": {
                "Published": {"parameter": "published", "value": true}
              }
            }
          }
        },
        "Post": {
          "row": "Post",
          "operations": {
            "find": "GetPost",
            "insert": "CreatePost"
          },
          "fields": {
            "id": {"readable": true, "generated": "insert"},
            "user_id": {"readable": true, "fillable": true},
            "title": {"readable": true, "fillable": true, "mutable": true}
          },
          "relations": {
            "Author": {
              "kind": "belongs_to",
              "model": "User",
              "local_key": "user_id",
              "foreign_key": "id",
              "inverse": "Posts",
              "nullable": false,
              "lazy_query": "GetUser"
            },
            "Tags": {
              "kind": "many_to_many",
              "model": "Tag",
              "local_key": "id",
              "target_key": "id",
              "lazy_query": "ListTagsByPost",
              "attach_query": "AttachTagToPost",
              "detach_query": "DetachTagFromPost",
              "sync_queries": {
                "list": "ListTagIDsByPost",
                "attach": "AttachTagToPost",
                "detach": "DetachTagFromPost"
              }
            }
          }
        },
        "Tag": {
          "row": "Tag",
          "operations": {
            "find": "GetTag",
            "insert": "CreateTag"
          },
          "fields": {
            "id": {"readable": true, "generated": "insert"},
            "name": {"readable": true, "fillable": true}
          }
        }
      }
    }
  ]
}`

func userRelationsColumns() []*pb.Column {
	return []*pb.Column{col("id", "uuid", true), col("name", "text", true)}
}

func postColumns() []*pb.Column {
	return []*pb.Column{
		col("id", "uuid", true),
		col("user_id", "uuid", true),
		col("title", "text", true),
	}
}

func tagColumns() []*pb.Column {
	return []*pb.Column{col("id", "uuid", true), col("name", "text", true)}
}

// userRelationsRequest builds the sqlc plugin request for
// userRelationsOptions: User/Post/Tag models with has_many/belongs_to/
// many_to_many relations, a scope, and eager loading.
func userRelationsRequest(t *testing.T) *pb.GenerateRequest {
	t.Helper()

	userCols := userRelationsColumns()
	postCols := postColumns()
	tagCols := tagColumns()

	queries := []*pb.Query{
		{
			Name:    "GetUser",
			Cmd:     ":one",
			Text:    "SELECT id, name FROM users WHERE id = $1;",
			Columns: userCols,
			Params:  []*pb.Parameter{param(1, "id", "uuid", true)},
		},
		{
			Name:    "CreateUser",
			Cmd:     ":one",
			Text:    "INSERT INTO users (name) VALUES ($1) RETURNING id, name;",
			Columns: userCols,
			Params:  []*pb.Parameter{param(1, "name", "text", true)},
		},
		{
			Name:    "GetPost",
			Cmd:     ":one",
			Text:    "SELECT id, user_id, title FROM posts WHERE id = $1;",
			Columns: postCols,
			Params:  []*pb.Parameter{param(1, "id", "uuid", true)},
		},
		{
			Name:    "CreatePost",
			Cmd:     ":one",
			Text:    "INSERT INTO posts (user_id, title) VALUES ($1, $2) RETURNING id, user_id, title;",
			Columns: postCols,
			Params:  []*pb.Parameter{param(1, "user_id", "uuid", true), param(2, "title", "text", true)},
		},
		{
			Name:    "GetTag",
			Cmd:     ":one",
			Text:    "SELECT id, name FROM tags WHERE id = $1;",
			Columns: tagCols,
			Params:  []*pb.Parameter{param(1, "id", "uuid", true)},
		},
		{
			Name:    "CreateTag",
			Cmd:     ":one",
			Text:    "INSERT INTO tags (name) VALUES ($1) RETURNING id, name;",
			Columns: tagCols,
			Params:  []*pb.Parameter{param(1, "name", "text", true)},
		},
		{
			Name:    "ListPostsByUser",
			Cmd:     ":many",
			Text:    "SELECT id, user_id, title FROM posts WHERE user_id = $1 AND (published = $2 OR $2 IS NULL);",
			Columns: postCols,
			Params: []*pb.Parameter{
				param(1, "user_id", "uuid", true),
				param(2, "published", "bool", false),
			},
		},
		{
			Name:    "ListPostsByUserIDs",
			Cmd:     ":many",
			Text:    "SELECT id, user_id, title FROM posts WHERE user_id = ANY($1);",
			Columns: postCols,
			Params:  []*pb.Parameter{param(1, "user_ids", "uuid", true)},
		},
		{
			Name:    "ListTagsByPost",
			Cmd:     ":many",
			Text:    "SELECT tags.id, tags.name FROM tags INNER JOIN post_tags ON post_tags.tag_id = tags.id WHERE post_tags.post_id = $1;",
			Columns: tagCols,
			Params:  []*pb.Parameter{param(1, "post_id", "uuid", true)},
		},
		{
			Name:   "AttachTagToPost",
			Cmd:    ":exec",
			Text:   "INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2);",
			Params: []*pb.Parameter{param(1, "post_id", "uuid", true), param(2, "tag_id", "uuid", true)},
		},
		{
			Name:   "DetachTagFromPost",
			Cmd:    ":exec",
			Text:   "DELETE FROM post_tags WHERE post_id = $1 AND tag_id = $2;",
			Params: []*pb.Parameter{param(1, "post_id", "uuid", true), param(2, "tag_id", "uuid", true)},
		},
		{
			Name:    "ListTagIDsByPost",
			Cmd:     ":many",
			Text:    "SELECT tag_id FROM post_tags WHERE post_id = $1;",
			Columns: []*pb.Column{col("tag_id", "uuid", true)},
			Params:  []*pb.Parameter{param(1, "post_id", "uuid", true)},
		},
	}

	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		Queries:       queries,
		PluginOptions: []byte(userRelationsOptions),
	}
}
