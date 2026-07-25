package unit

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/codegen"
	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/contract"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/mapping"
	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func rfield(goField, goType string) *plan.ResolvedField {
	return &plan.ResolvedField{
		ResolvedField: mapping.ResolvedField{
			Name:       strings.ToLower(goField),
			GoField:    goField,
			ColumnName: strings.ToLower(goField),
			GoType:     mapping.GoType{Expr: goType},
			NotNull:    true,
		},
	}
}

// TestRenderCollection_FindParamNaming exercises lowerFirst's mixed-case
// branch and sortedParams' swap branch, via Find operation params
// deliberately out of declared-number order.
func TestRenderCollection_FindParamNaming(t *testing.T) {
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name: "Widget",
		Row:  "Widget",
		Operations: plan.ResolvedOperations{
			Find: &plan.ResolvedOperation{
				Kind:      contract.Find,
				QueryName: "FindWidget",
				Query:     &pb.Query{Name: "FindWidget", Cmd: ":one"},
				Params: []plan.ParamBinding{
					{Number: 3, Field: rfield("Name", "string")},
					{Number: 1, Field: rfield("ID", "int64")},
					{Number: 2, Field: rfield("Name", "string")},
				},
			},
		},
	}

	out, diags := codegen.RenderCollection(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	if !strings.Contains(src, "name") || !strings.Contains(src, "id") {
		t.Fatalf("expected lowered param names in generated source, got:\n%s", src)
	}
	if !strings.Contains(src, "name2") {
		t.Fatalf("expected a disambiguated second %q param name, got:\n%s", "name2", src)
	}
}

// TestRenderModel_SliceField exercises the NeedsSlices branch for a field
// whose Go type is a slice.
func TestRenderModel_SliceField(t *testing.T) {
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name: "Widget",
		Row:  "Widget",
		Fields: []plan.ResolvedField{
			{
				ResolvedField: mapping.ResolvedField{
					Name: "tags", GoField: "Tags", ColumnName: "tags",
					GoType: mapping.GoType{Expr: "[]string"}, NotNull: true,
				},
				Policy: config.FieldPolicy{Name: "tags", Readable: true},
			},
		},
	}

	out, diags := codegen.RenderModel(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if !strings.Contains(string(out), "[]string") {
		t.Fatalf("expected slice field type in generated source, got:\n%s", out)
	}
}

// TestRender_PropagatesCodegenFailure exercises Render's error-aggregation
// branch: a model with an invalid Go identifier for Row produces generated
// source that fails to format, so Render must return a nil file list.
func TestRender_PropagatesCodegenFailure(t *testing.T) {
	p := &plan.Plan{
		Contexts: []plan.ResolvedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []plan.ResolvedModel{
					{
						Name: "BadName",
						Row:  "Bad-Name",
						Operations: plan.ResolvedOperations{
							Insert: &plan.ResolvedOperation{
								Kind:      contract.Insert,
								QueryName: "CreateBadName",
								Query:     &pb.Query{Name: "CreateBadName", Cmd: ":one"},
							},
						},
					},
				},
			},
		},
	}

	files, diags := codegen.Render(p, t.TempDir())
	if files != nil {
		t.Fatalf("expected nil files for a codegen failure, got %d files", len(files))
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic, got %+v", diags)
	}
}
