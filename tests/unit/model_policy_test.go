package unit

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-model/internal/generate"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// pcol/pparam build minimal sqlc column/parameter metadata for these tests.
func pcol(name, pgType string, notNull bool) *pb.Column {
	return &pb.Column{Name: name, OriginalName: name, NotNull: notNull, Type: &pb.Identifier{Name: pgType}}
}
func pparam(number int32, name, pgType string, notNull bool) *pb.Parameter {
	return &pb.Parameter{Number: number, Column: pcol(name, pgType, notNull)}
}

// generateSecretModel builds a single-field-of-interest model ("secret",
// sensitive + immutable_after_insert + mutable) alongside a plain "name"
// field, and returns the rendered model file's source.
func generateSecretModel(t *testing.T) string {
	t.Helper()

	cols := []*pb.Column{
		pcol("id", "uuid", true),
		pcol("secret", "text", true),
	}
	queries := []*pb.Query{
		{Name: "GetThing", Cmd: ":one", Text: "SELECT id, secret FROM things WHERE id = $1;", Columns: cols, Params: []*pb.Parameter{pparam(1, "id", "uuid", true)}},
		{Name: "CreateThing", Cmd: ":one", Text: "INSERT INTO things (secret) VALUES ($1) RETURNING id, secret;", Columns: cols, Params: []*pb.Parameter{pparam(1, "secret", "text", true)}},
	}

	opts := `{
		"version": 1,
		"sqlc": {"package": "sqlcdb", "import": "example.com/x", "driver": "pgx/v5"},
		"contexts": [{
			"name": "c", "package": "c", "directory": "c",
			"models": {"Thing": {
				"row": "Thing",
				"operations": {"insert": "CreateThing", "find": "GetThing"},
				"fields": {
					"id": {"readable": true, "generated": "insert"},
					"secret": {"readable": true, "mutable": true, "sensitive": true, "immutable_after_insert": true}
				}
			}}
		}]
	}`

	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Codegen: &pb.Codegen{Out: ""}},
		Queries:       queries,
		PluginOptions: []byte(opts),
	}

	resp, diags := generate.Generate(req)
	if resp == nil {
		t.Fatalf("generation failed: %+v", diags)
	}
	for _, f := range resp.Files {
		if f.Name == "c/thing_gen.go" {
			return string(f.Contents)
		}
	}
	t.Fatalf("expected c/thing_gen.go in response, got %v", fileNames(resp.Files))
	return ""
}

func fileNames(files []*pb.File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Name)
	}
	return out
}

func TestGeneratedModel_SensitiveFieldRedactedInString(t *testing.T) {
	src := generateSecretModel(t)
	if !strings.Contains(src, `"Secret=<redacted>"`) {
		t.Fatalf("expected sensitive field to be redacted in String(), got:\n%s", src)
	}
	if strings.Contains(src, `fmt.Sprintf("Secret=%v"`) {
		t.Fatalf("sensitive field's real value must not appear in String() formatting, got:\n%s", src)
	}
}

func TestGeneratedModel_SensitiveFieldStillHasRealGetter(t *testing.T) {
	src := generateSecretModel(t)
	if !strings.Contains(src, "func (u *Thing) Secret() string { return u.current.Secret }") {
		t.Fatalf("expected sensitive field to still have a normal getter, got:\n%s", src)
	}
}

func TestGeneratedModel_ImmutableAfterInsertChecksIsNew(t *testing.T) {
	src := generateSecretModel(t)
	if !strings.Contains(src, "if !u.IsNew() {") || !strings.Contains(src, "ErrImmutableField") {
		t.Fatalf("expected SetSecret to guard against mutation once persisted, got:\n%s", src)
	}
}
