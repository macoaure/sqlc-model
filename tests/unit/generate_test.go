package unit

import (
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/generate"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestGenerate_InvalidConfigReturnsNil(t *testing.T) {
	req := &pb.GenerateRequest{PluginOptions: []byte(`not json`)}
	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected nil response for invalid configuration, got %+v", resp)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic, got %+v", diags)
	}
}

// TestGenerate_PlanBuildFailureReturnsNil covers Generate's "plan build
// failed" branch: a row that isn't a valid Go identifier passes config
// decode but fails plan.Build's identifier check.
func TestGenerate_PlanBuildFailureReturnsNil(t *testing.T) {
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"User": {"row": "not-exported", "operations": {"insert": "CreateUser"}, "fields": {}}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
	}
	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected nil response for a plan-build failure, got %+v", resp)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic, got %+v", diags)
	}
}

// TestGenerate_CodegenFailureReturnsNil covers Generate's "codegen render
// failed" branch. Neither config decode nor plan.Build validates that a
// field's declared key produces a valid Go identifier once PascalCased —
// only the model's "row" gets that check — so a field key containing a
// character PascalCase doesn't strip (anything other than "_") sails
// through both stages and only breaks once codegen tries to format the
// resulting (invalid) Go source.
func TestGenerate_CodegenFailureReturnsNil(t *testing.T) {
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Widget": {
					"row": "Widget",
					"operations": {"insert": "CreateWidget"},
					"fields": {"bad-field": {"readable": true}}
				}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		Settings:      &pb.Settings{Engine: "postgresql", Codegen: &pb.Codegen{Out: ""}},
		PluginOptions: []byte(options),
		Queries: []*pb.Query{
			{
				Name: "CreateWidget",
				Cmd:  ":one",
				Columns: []*pb.Column{
					{Name: "bad-field", OriginalName: "bad-field", NotNull: true, Type: &pb.Identifier{Name: "text"}},
				},
			},
		},
	}
	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected nil response for a codegen failure, got %+v", resp)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic, got %+v", diags)
	}
}
