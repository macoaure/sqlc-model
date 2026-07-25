package unit

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func findDiag(diags []diagnostics.Diagnostic, substr string) *diagnostics.Diagnostic {
	for i := range diags {
		if strings.Contains(diags[i].Message, substr) {
			return &diags[i]
		}
	}
	return nil
}

// TestBuild_InvalidRowIdentifiers covers isExportedIdent's three failure
// branches (empty, lowercase-first, invalid character) and the
// checkContextCollisions "skip models with no row" branch, all via models
// with no configured operations so buildModel returns immediately after the
// row check.
func TestBuild_InvalidRowIdentifiers(t *testing.T) {
	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{Name: "EmptyRow", Row: ""},
					{Name: "LowerRow", Row: "lowercase"},
					{Name: "BadCharRow", Row: "Bad-Name"},
				},
			},
		},
	}

	_, diags := plan.Build(root, &pb.GenerateRequest{}, nil)
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected row-identifier errors, got %+v", diags)
	}
	if findDiag(diags, `"row" "" must be a valid exported Go identifier`) == nil {
		t.Errorf("expected diagnostic for empty row, got %+v", diags)
	}
	if findDiag(diags, `"row" "lowercase" must be a valid exported Go identifier`) == nil {
		t.Errorf("expected diagnostic for lowercase row, got %+v", diags)
	}
	if findDiag(diags, `"row" "Bad-Name" must be a valid exported Go identifier`) == nil {
		t.Errorf("expected diagnostic for row with invalid character, got %+v", diags)
	}
}

// TestBuild_RowCollision covers checkContextCollisions' "two models share a
// row" branch.
func TestBuild_RowCollision(t *testing.T) {
	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{Name: "Dup1", Row: "DupRow"},
					{Name: "Dup2", Row: "DupRow"},
				},
			},
		},
	}

	_, diags := plan.Build(root, &pb.GenerateRequest{}, nil)
	d := findDiag(diags, `both declare row "DupRow"`)
	if d == nil {
		t.Fatalf("expected a row-collision diagnostic, got %+v", diags)
	}
	if d.Severity != diagnostics.SeverityError {
		t.Fatalf("expected row collision to be an error, got %+v", d)
	}
}

// TestBuild_ScanOrderMissingField covers resolveScanOrder's "declared field
// not present among the query's returned columns" branch: Find returns
// fewer columns than Insert (which establishes the model's fields).
func TestBuild_ScanOrderMissingField(t *testing.T) {
	idCol := mcol("id", "id", "int8", true)
	nameCol := mcol("name", "name", "text", true)

	req := &pb.GenerateRequest{
		Queries: []*pb.Query{
			{Name: "CreateScanGap", Cmd: ":one", Columns: []*pb.Column{idCol, nameCol}},
			{Name: "FindScanGap", Cmd: ":one", Columns: []*pb.Column{idCol}},
		},
	}

	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{
						Name: "ScanGap",
						Row:  "ScanGapRow",
						Operations: config.Operations{
							Insert: "CreateScanGap",
							Find:   "FindScanGap",
						},
						Fields: []config.FieldPolicy{
							{Name: "id", Readable: true},
							{Name: "name", Readable: true},
						},
					},
				},
			},
		},
	}

	_, diags := plan.Build(root, req, nil)
	d := findDiag(diags, `does not return field "name"`)
	if d == nil {
		t.Fatalf("expected a missing-scan-field diagnostic, got %+v", diags)
	}
	if d.Severity != diagnostics.SeverityError {
		t.Fatalf("expected missing scan field to be an error, got %+v", d)
	}
}

// TestBuild_ParamNoMatchingField covers resolveParams' "parameter doesn't
// match any configured field" branch.
func TestBuild_ParamNoMatchingField(t *testing.T) {
	idCol := mcol("id", "id", "int8", true)
	nameCol := mcol("name", "name", "text", true)
	ghostCol := mcol("ghost", "ghost", "text", true)

	req := &pb.GenerateRequest{
		Queries: []*pb.Query{
			{Name: "CreateParamGap", Cmd: ":one", Columns: []*pb.Column{idCol, nameCol}},
			{
				Name:    "UpdateParamGap",
				Cmd:     ":one",
				Columns: []*pb.Column{idCol, nameCol},
				Params:  []*pb.Parameter{{Number: 1, Column: ghostCol}},
			},
		},
	}

	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{
						Name: "ParamGap",
						Row:  "ParamGapRow",
						Operations: config.Operations{
							Insert: "CreateParamGap",
							Update: "UpdateParamGap",
						},
						Fields: []config.FieldPolicy{
							{Name: "id", Readable: true},
							{Name: "name", Readable: true, Fillable: true, Mutable: true},
						},
					},
				},
			},
		},
	}

	_, diags := plan.Build(root, req, nil)
	d := findDiag(diags, `does not match any configured field`)
	if d == nil {
		t.Fatalf("expected a param-no-match diagnostic, got %+v", diags)
	}
	if d.Severity != diagnostics.SeverityError {
		t.Fatalf("expected unmatched param to be an error, got %+v", d)
	}
}

// TestBuild_GeneratedSaveWithoutUpdateWarns covers
// validateGeneratedHydration's warning branch: a `generated: save` field on
// a model with no configured update operation.
func TestBuild_GeneratedSaveWithoutUpdateWarns(t *testing.T) {
	idCol := mcol("id", "id", "int8", true)
	secretCol := mcol("secret", "secret", "text", true)

	req := &pb.GenerateRequest{
		Queries: []*pb.Query{
			{Name: "CreateGeneratedGap", Cmd: ":one", Columns: []*pb.Column{idCol, secretCol}},
		},
	}

	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{
						Name: "GeneratedGap",
						Row:  "GeneratedGapRow",
						Operations: config.Operations{
							Insert: "CreateGeneratedGap",
						},
						Fields: []config.FieldPolicy{
							{Name: "id", Readable: true},
							{Name: "secret", Readable: true, Generated: config.GeneratedSave},
						},
					},
				},
			},
		},
	}

	p, diags := plan.Build(root, req, nil)
	if p == nil {
		t.Fatalf("expected generation to succeed (warning only), got %+v", diags)
	}
	d := findDiag(diags, `generated: save has no effect without a configured update operation`)
	if d == nil {
		t.Fatalf("expected a generated-save-without-update diagnostic, got %+v", diags)
	}
	if d.Severity != diagnostics.SeverityWarning {
		t.Fatalf("expected generated-save-without-update to be a warning, got %+v", d)
	}
}

func TestBuild_ValueObjectFieldKeepsPersistedType(t *testing.T) {
	req := &pb.GenerateRequest{
		Queries: []*pb.Query{
			{
				Name:    "CreateUser",
				Cmd:     ":one",
				Columns: []*pb.Column{mcol("email", "email", "text", true)},
				Params:  []*pb.Parameter{{Number: 1, Column: mcol("email", "email", "text", true)}},
			},
		},
	}
	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{
						Name:       "User",
						Row:        "User",
						Operations: config.Operations{Insert: "CreateUser"},
						Fields: []config.FieldPolicy{
							{
								Name: "email", Readable: true, Fillable: true,
								ValueObject: &config.ValueObjectMapping{Type: "Email", Constructor: "NewEmail", Accessor: "String"},
							},
						},
					},
				},
			},
		},
	}

	p, diags := plan.Build(root, req, nil)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	field := p.Contexts[0].Models[0].Fields[0]
	if field.GoType.Expr != "Email" {
		t.Fatalf("model field type = %s, want Email", field.GoType.Expr)
	}
	if field.PersistedGoType.Expr != "string" {
		t.Fatalf("persisted field type = %s, want string", field.PersistedGoType.Expr)
	}
	if field.ValueObject == nil || field.ValueObject.Constructor != "NewEmail" || field.ValueObject.Accessor != "String" {
		t.Fatalf("unexpected value object metadata: %+v", field.ValueObject)
	}
}

func TestBuild_ValueObjectFieldRejectsNullableColumn(t *testing.T) {
	req := &pb.GenerateRequest{
		Queries: []*pb.Query{
			{Name: "CreateUser", Cmd: ":one", Columns: []*pb.Column{mcol("email", "email", "text", false)}},
		},
	}
	root := &config.RootConfiguration{
		Version: config.SupportedVersion,
		Contexts: []config.BoundedContext{
			{
				Name:      "content",
				Package:   "content",
				Directory: "content",
				Models: []config.ModelConfiguration{
					{
						Name:       "User",
						Row:        "User",
						Operations: config.Operations{Insert: "CreateUser"},
						Fields: []config.FieldPolicy{
							{
								Name: "email",
								ValueObject: &config.ValueObjectMapping{
									Type: "Email", Constructor: "NewEmail", Accessor: "String",
								},
							},
						},
					},
				},
			},
		},
	}

	_, diags := plan.Build(root, req, nil)
	if findDiag(diags, "value_object does not support nullable columns") == nil {
		t.Fatalf("expected nullable value_object diagnostic, got %+v", diags)
	}
}
