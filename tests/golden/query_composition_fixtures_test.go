package golden

import (
	"strings"
	"testing"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func queryCompositionRequest(t *testing.T) *pb.GenerateRequest {
	t.Helper()
	req := userRelationsRequest(t)
	req.PluginOptions = []byte(strings.Replace(userRelationsOptions, `"relations": {
            "Posts": {`, `"queries": {
            "default": {
              "operation": "ListActiveUsers",
              "terminal": "get",
              "scopes": {
                "Active": {"parameter": "active", "value": true},
                "Limit": {"parameter": "limit", "argument": "int32"},
                "OrderByName": {"query": "ListUsersByName"},
                "WithPosts": {"relation": "Posts"}
              }
            }
          },
          "lookups": {
            "FindByName": {"query": "FindUserByName"}
          },
          "relations": {
            "Posts": {`, 1))
	userCols := userRelationsColumns()
	req.Queries = append(req.Queries,
		&pb.Query{
			Name:    "ListActiveUsers",
			Cmd:     ":many",
			Text:    "SELECT id, name FROM users WHERE active = $1 ORDER BY id LIMIT $2;",
			Columns: userCols,
			Params: []*pb.Parameter{
				param(1, "active", "bool", true),
				param(2, "limit", "int4", true),
			},
		},
		&pb.Query{
			Name:    "ListUsersByName",
			Cmd:     ":many",
			Text:    "SELECT id, name FROM users WHERE active = $1 ORDER BY name LIMIT $2;",
			Columns: userCols,
			Params: []*pb.Parameter{
				param(1, "active", "bool", true),
				param(2, "limit", "int4", true),
			},
		},
		&pb.Query{
			Name:    "FindUserByName",
			Cmd:     ":one",
			Text:    "SELECT id, name FROM users WHERE name = $1;",
			Columns: userCols,
			Params:  []*pb.Parameter{param(1, "name", "text", true)},
		},
	)
	return req
}
