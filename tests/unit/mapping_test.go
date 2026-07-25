package unit

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/mapping"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func mcol(name, originalName, pgType string, notNull bool) *pb.Column {
	return &pb.Column{Name: name, OriginalName: originalName, NotNull: notNull, Type: &pb.Identifier{Name: pgType}}
}

func TestResolve_AutomaticMatch(t *testing.T) {
	cols := []*pb.Column{mcol("name", "name", "text", true)}
	fp := config.FieldPolicy{Name: "name", Readable: true}
	rf, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if rf.GoField != "Name" || rf.ColumnName != "name" || rf.GoType.Expr != "string" {
		t.Fatalf("unexpected resolution: %+v", rf)
	}
}

func TestResolve_AutomaticMatchCaseInsensitive(t *testing.T) {
	cols := []*pb.Column{mcol("Name", "Name", "text", true)}
	fp := config.FieldPolicy{Name: "name", Readable: true}
	_, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("expected case-insensitive automatic match to succeed, got %+v", diags)
	}
}

func TestResolve_NoMatchIsAmbiguityError(t *testing.T) {
	cols := []*pb.Column{mcol("email", "email", "text", true)}
	fp := config.FieldPolicy{Name: "name", Readable: true}
	_, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error when no column matches, got %+v", diags)
	}
}

func TestResolve_AmbiguousMultipleMatches(t *testing.T) {
	cols := []*pb.Column{
		mcol("name", "name", "text", true),
		mcol("name", "other_name", "text", true), // two columns exposed as "name"
	}
	fp := config.FieldPolicy{Name: "name", Readable: true}
	_, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected ambiguous match to be an error, got %+v", diags)
	}
}

func TestResolve_ExplicitColumnOverride(t *testing.T) {
	cols := []*pb.Column{mcol("nm", "name", "text", true)}
	fp := config.FieldPolicy{Name: "full_name", Column: "name"}
	rf, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if rf.ColumnName != "nm" {
		t.Fatalf("expected explicit column override to resolve via OriginalName, got %+v", rf)
	}
}

func TestResolve_ExplicitColumnAndRowFieldMatch(t *testing.T) {
	cols := []*pb.Column{mcol("nm", "name", "text", true)}
	fp := config.FieldPolicy{Name: "full_name", Column: "name", RowField: "nm"}
	rf, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if rf.ColumnName != "nm" {
		t.Fatalf("expected combined column+row_field override to resolve, got %+v", rf)
	}
}

func TestResolve_ExplicitRowFieldOverride(t *testing.T) {
	cols := []*pb.Column{mcol("nm", "name", "text", true)}
	fp := config.FieldPolicy{Name: "full_name", RowField: "nm"}
	rf, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if rf.ColumnName != "nm" {
		t.Fatalf("expected explicit row_field override to resolve, got %+v", rf)
	}
}

func TestResolveGoType_NullableUsesWrapper(t *testing.T) {
	notNull := mapping.ResolveGoType(mcol("email", "email", "text", true))
	if notNull.Expr != "string" {
		t.Fatalf("expected NOT NULL text to map to string, got %s", notNull.Expr)
	}
	nullable := mapping.ResolveGoType(mcol("email", "email", "text", false))
	if nullable.Expr != "pgtype.Text" {
		t.Fatalf("expected nullable text to map to pgtype.Text, got %s", nullable.Expr)
	}
}

func TestResolveGoType_UnmappedFallsBackToString(t *testing.T) {
	gt := mapping.ResolveGoType(mcol("x", "x", "some_exotic_type", true))
	if gt.Expr != "string" || !gt.Unmapped {
		t.Fatalf("expected unmapped fallback to string, got %+v", gt)
	}
}

func TestResolveGoType_EmptyTypeNameFallsBackToString(t *testing.T) {
	col := &pb.Column{Name: "x", OriginalName: "x", NotNull: true, Type: &pb.Identifier{Name: ""}}
	gt := mapping.ResolveGoType(col)
	if gt.Expr != "string" || !gt.Unmapped {
		t.Fatalf("expected empty type name fallback to string, got %+v", gt)
	}
}

func TestResolveGoType_NullableWithoutOverrideKeepsNotNullMapping(t *testing.T) {
	gt := mapping.ResolveGoType(mcol("n", "n", "numeric", false))
	if gt.Expr != "pgtype.Numeric" {
		t.Fatalf("expected nullable numeric with no nullableGo override to keep notNullGo mapping, got %+v", gt)
	}
}

func TestResolveGoType_ArrayColumn(t *testing.T) {
	col := &pb.Column{Name: "tags", OriginalName: "tags", NotNull: true, Type: &pb.Identifier{Name: "text"}, IsArray: true}
	gt := mapping.ResolveGoType(col)
	if gt.Expr != "[]string" {
		t.Fatalf("expected array column to map to []string, got %+v", gt)
	}

	col2 := &pb.Column{Name: "tags2", OriginalName: "tags2", NotNull: true, Type: &pb.Identifier{Name: "text"}, ArrayDims: 1}
	gt2 := mapping.ResolveGoType(col2)
	if gt2.Expr != "[]string" {
		t.Fatalf("expected ArrayDims > 0 column to map to []string, got %+v", gt2)
	}
}

func TestResolve_ColumnIdentityFallsBackToName(t *testing.T) {
	col := &pb.Column{Name: "name", OriginalName: "", NotNull: true, Type: &pb.Identifier{Name: "text"}}
	fp := config.FieldPolicy{Name: "full_name", Column: "name"}
	rf, diags := mapping.Resolve(fp, []*pb.Column{col}, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if rf.ColumnName != "name" {
		t.Fatalf("expected column identity to fall back to Name when OriginalName is empty, got %+v", rf)
	}
}

func TestResolve_UnmappedColumnEmitsErrorWithTypeName(t *testing.T) {
	cols := []*pb.Column{mcol("x", "x", "some_exotic_type", true)}
	fp := config.FieldPolicy{Name: "x", Readable: true}
	_, diags := mapping.Resolve(fp, cols, "path", "ctx", "Model")
	if len(diags) != 1 || diags[0].Severity != diagnostics.SeverityError {
		t.Fatalf("expected a single error diagnostic, got %+v", diags)
	}
	if !strings.Contains(diags[0].Message, "some_exotic_type") {
		t.Fatalf("expected error message to include the type name, got %q", diags[0].Message)
	}
	if !strings.Contains(diags[0].Hint, "supported PostgreSQL type") {
		t.Fatalf("expected remediation hint, got %q", diags[0].Hint)
	}
}

func TestResolve_UnmappedColumnWithNilTypeUsesUnknown(t *testing.T) {
	col := &pb.Column{Name: "x", OriginalName: "x", NotNull: true, Type: nil}
	fp := config.FieldPolicy{Name: "x", Readable: true}
	_, diags := mapping.Resolve(fp, []*pb.Column{col}, "path", "ctx", "Model")
	if len(diags) != 1 || diags[0].Severity != diagnostics.SeverityError {
		t.Fatalf("expected a single error diagnostic, got %+v", diags)
	}
	if !strings.Contains(diags[0].Message, "<unknown>") {
		t.Fatalf("expected error message to include <unknown> for nil type, got %q", diags[0].Message)
	}
}

func TestResolve_DescribeMatchBranches(t *testing.T) {
	cols := []*pb.Column{mcol("email", "email", "text", true)}

	_, diags := mapping.Resolve(config.FieldPolicy{Name: "x", Column: "missing"}, cols, "p", "c", "M")
	if !diagnostics.HasError(diags) || !strings.Contains(diags[0].Message, `column="missing"`) {
		t.Fatalf("expected column-only describeMatch, got %+v", diags)
	}

	_, diags = mapping.Resolve(config.FieldPolicy{Name: "x", RowField: "missing"}, cols, "p", "c", "M")
	if !diagnostics.HasError(diags) || !strings.Contains(diags[0].Message, `row_field="missing"`) {
		t.Fatalf("expected row_field-only describeMatch, got %+v", diags)
	}

	_, diags = mapping.Resolve(config.FieldPolicy{Name: "x", Column: "missing", RowField: "missing"}, cols, "p", "c", "M")
	if !diagnostics.HasError(diags) || !strings.Contains(diags[0].Message, `column="missing" row_field="missing"`) {
		t.Fatalf("expected column+row_field describeMatch, got %+v", diags)
	}
}

func TestPascalCase_Initialisms(t *testing.T) {
	cases := map[string]string{
		"id":         "ID",
		"created_at": "CreatedAt",
		"user_uuid":  "UserUUID",
		"api_key":    "APIKey",
		"_leading":   "Leading",
		"double__us": "DoubleUs",
	}
	for in, want := range cases {
		if got := mapping.PascalCase(in); got != want {
			t.Errorf("PascalCase(%q) = %q, want %q", in, got, want)
		}
	}
}
