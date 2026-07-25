package golden

import (
	"testing"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// userBasicOptions is the quickstart.md `sqlc.yaml` options block, verbatim
// in shape (see specs/001-model-generation-config/quickstart.md).
const userBasicOptions = `{
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
            "insert": "CreateUser",
            "update": "UpdateUser",
            "delete": "DeleteUser",
            "refresh": "GetUser"
          },
          "fields": {
            "id": {"readable": true, "generated": "insert"},
            "name": {"readable": true, "fillable": true, "mutable": true},
            "email": {"readable": true, "fillable": true, "mutable": true},
            "active": {"readable": true, "fillable": true, "mutable": true},
            "created_at": {"readable": true, "generated": "insert"},
            "updated_at": {"readable": true, "generated": "save"}
          }
        }
      }
    }
  ]
}`

func col(name, pgType string, notNull bool) *pb.Column {
	return &pb.Column{
		Name:         name,
		OriginalName: name,
		NotNull:      notNull,
		Type:         &pb.Identifier{Name: pgType},
	}
}

func param(number int32, name, pgType string, notNull bool) *pb.Parameter {
	return &pb.Parameter{Number: number, Column: col(name, pgType, notNull)}
}

// userBasicRequest builds the sqlc plugin request corresponding to
// quickstart.md's `users` table and its four queries.
func userBasicRequest(t *testing.T) *pb.GenerateRequest {
	t.Helper()

	rowColumns := []*pb.Column{
		col("id", "uuid", true),
		col("name", "text", true),
		col("email", "text", true),
		col("active", "bool", true),
		col("created_at", "timestamptz", true),
		col("updated_at", "timestamptz", true),
	}

	queries := []*pb.Query{
		{
			Name:    "GetUser",
			Cmd:     ":one",
			Text:    "SELECT id, name, email, active, created_at, updated_at FROM users WHERE id = $1;",
			Columns: rowColumns,
			Params:  []*pb.Parameter{param(1, "id", "uuid", true)},
		},
		{
			Name:    "CreateUser",
			Cmd:     ":one",
			Text:    "INSERT INTO users (name, email, active) VALUES ($1, $2, $3) RETURNING id, name, email, active, created_at, updated_at;",
			Columns: rowColumns,
			Params: []*pb.Parameter{
				param(1, "name", "text", true),
				param(2, "email", "text", true),
				param(3, "active", "bool", true),
			},
		},
		{
			Name:    "UpdateUser",
			Cmd:     ":one",
			Text:    "UPDATE users SET name = $1, email = $2, active = $3, updated_at = NOW() WHERE id = $4 RETURNING id, name, email, active, created_at, updated_at;",
			Columns: rowColumns,
			Params: []*pb.Parameter{
				param(1, "name", "text", true),
				param(2, "email", "text", true),
				param(3, "active", "bool", true),
				param(4, "id", "uuid", true),
			},
		},
		{
			Name:   "DeleteUser",
			Cmd:    ":execrows",
			Text:   "DELETE FROM users WHERE id = $1;",
			Params: []*pb.Parameter{param(1, "id", "uuid", true)},
		},
	}

	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		Queries:       queries,
		PluginOptions: []byte(userBasicOptions),
	}
}

// fieldPolicyMatrixOptions exercises every documented field-policy flag on
// one model: readable-only, fillable-only, mutable-only, both, generated
// insert/save, sensitive, version, immutable_after_insert, and an explicit
// column override (spec 001 User Story 2's independent test).
const fieldPolicyMatrixOptions = `{
  "version": 1,
  "sqlc": {"package": "sqlcdb", "import": "example.com/project/sqlc", "driver": "pgx/v5"},
  "contexts": [
    {
      "name": "inventory",
      "package": "inventory",
      "directory": "inventory",
      "models": {
        "Widget": {
          "row": "Widget",
          "operations": {
            "find": "GetWidget",
            "insert": "CreateWidget",
            "update": "UpdateWidget",
            "delete": "DeleteWidget",
            "refresh": "GetWidget"
          },
          "fields": {
            "id": {"readable": true, "generated": "insert"},
            "version_no": {"version": true},
            "readonly_field": {"readable": true},
            "fillable_field": {"readable": true, "fillable": true},
            "mutable_field": {"readable": true, "mutable": true},
            "fillable_and_mutable": {"readable": true, "fillable": true, "mutable": true},
            "generated_insert_field": {"readable": true, "generated": "insert"},
            "generated_save_field": {"readable": true, "generated": "save"},
            "sensitive_field": {"readable": true, "sensitive": true},
            "explicit_mapped": {"readable": true, "fillable": true, "column": "actual_column_name"},
            "immutable_field": {"readable": true, "fillable": true, "mutable": true, "immutable_after_insert": true}
          }
        }
      }
    }
  ]
}`

// fieldPolicyMatrixColumns are shared by every hydrating Widget query.
func fieldPolicyMatrixColumns() []*pb.Column {
	return []*pb.Column{
		col("id", "uuid", true),
		col("version_no", "int4", true),
		col("readonly_field", "text", true),
		col("fillable_field", "text", true),
		col("mutable_field", "text", true),
		col("fillable_and_mutable", "text", true),
		col("generated_insert_field", "text", true),
		col("generated_save_field", "text", true),
		col("sensitive_field", "text", true),
		col("actual_column_name", "text", true),
		col("immutable_field", "text", true),
	}
}

func fieldPolicyMatrixRequest(t *testing.T) *pb.GenerateRequest {
	t.Helper()

	rowColumns := fieldPolicyMatrixColumns()
	writableParams := []*pb.Parameter{
		param(1, "fillable_field", "text", true),
		param(2, "mutable_field", "text", true),
		param(3, "fillable_and_mutable", "text", true),
		param(4, "actual_column_name", "text", true),
		param(5, "immutable_field", "text", true),
	}
	updateParams := append(append([]*pb.Parameter{}, writableParams...), param(6, "id", "uuid", true))

	queries := []*pb.Query{
		{
			Name:    "GetWidget",
			Cmd:     ":one",
			Text:    "SELECT id, version_no, readonly_field, fillable_field, mutable_field, fillable_and_mutable, generated_insert_field, generated_save_field, sensitive_field, actual_column_name, immutable_field FROM widgets WHERE id = $1;",
			Columns: rowColumns,
			Params:  []*pb.Parameter{param(1, "id", "uuid", true)},
		},
		{
			Name:    "CreateWidget",
			Cmd:     ":one",
			Text:    "INSERT INTO widgets (fillable_field, mutable_field, fillable_and_mutable, actual_column_name, immutable_field) VALUES ($1, $2, $3, $4, $5) RETURNING id, version_no, readonly_field, fillable_field, mutable_field, fillable_and_mutable, generated_insert_field, generated_save_field, sensitive_field, actual_column_name, immutable_field;",
			Columns: rowColumns,
			Params:  writableParams,
		},
		{
			Name:    "UpdateWidget",
			Cmd:     ":one",
			Text:    "UPDATE widgets SET fillable_field = $1, mutable_field = $2, fillable_and_mutable = $3, actual_column_name = $4, immutable_field = $5 WHERE id = $6 RETURNING id, version_no, readonly_field, fillable_field, mutable_field, fillable_and_mutable, generated_insert_field, generated_save_field, sensitive_field, actual_column_name, immutable_field;",
			Columns: rowColumns,
			Params:  updateParams,
		},
		{
			Name:   "DeleteWidget",
			Cmd:    ":execrows",
			Text:   "DELETE FROM widgets WHERE id = $1;",
			Params: []*pb.Parameter{param(1, "id", "uuid", true)},
		},
	}

	return &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		Queries:       queries,
		PluginOptions: []byte(fieldPolicyMatrixOptions),
	}
}
