package golden

import pb "github.com/sqlc-dev/plugin-sdk-go/plugin"

const valueObjectOptions = `{
  "version": 1,
  "sqlc": {"package": "sqlcdb", "import": "example.com/project/internal/database/sqlc", "driver": "pgx/v5"},
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
            "insert": "CreateUser",
            "update": "UpdateUser",
            "refresh": "GetUser"
          },
          "fields": {
            "id": {"readable": true, "generated": "insert"},
            "email": {
              "readable": true,
              "fillable": true,
              "mutable": true,
              "value_object": {"type": "Email", "constructor": "NewEmail", "accessor": "String"}
            }
          }
        }
      }
    }
  ]
}`

func valueObjectRequest() *pb.GenerateRequest {
	rowColumns := []*pb.Column{
		col("id", "uuid", true),
		col("email", "text", true),
	}
	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(valueObjectOptions),
		Queries: []*pb.Query{
			{
				Name:    "GetUser",
				Cmd:     ":one",
				Text:    "SELECT id, email FROM users WHERE id = $1;",
				Columns: rowColumns,
				Params:  []*pb.Parameter{param(1, "id", "uuid", true)},
			},
			{
				Name:    "CreateUser",
				Cmd:     ":one",
				Text:    "INSERT INTO users (email) VALUES ($1) RETURNING id, email;",
				Columns: rowColumns,
				Params:  []*pb.Parameter{param(1, "email", "text", true)},
			},
			{
				Name:    "UpdateUser",
				Cmd:     ":one",
				Text:    "UPDATE users SET email = $1 WHERE id = $2 RETURNING id, email;",
				Columns: rowColumns,
				Params: []*pb.Parameter{
					param(1, "email", "text", true),
					param(2, "id", "uuid", true),
				},
			},
		},
	}
}
