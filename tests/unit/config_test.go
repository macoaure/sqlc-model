package unit

import (
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
)

func hasErrorPath(diags []diagnostics.Diagnostic, path string) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError && d.Path == path {
			return true
		}
	}
	return false
}

func TestDecode_MissingVersion(t *testing.T) {
	root, diags := config.Decode([]byte(`{"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[]}`))
	if root != nil {
		t.Fatalf("expected nil root for missing version, got %+v", root)
	}
	if !hasErrorPath(diags, "options.version") {
		t.Fatalf("expected error at options.version, got %+v", diags)
	}
}

func TestDecode_UnsupportedVersion(t *testing.T) {
	root, diags := config.Decode([]byte(`{"version":2,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[]}`))
	if root != nil {
		t.Fatalf("expected nil root for unsupported version, got %+v", root)
	}
	if !hasErrorPath(diags, "options.version") {
		t.Fatalf("expected error at options.version, got %+v", diags)
	}
}

func TestDecode_MissingSqlcPackage(t *testing.T) {
	_, diags := config.Decode([]byte(`{"version":1,"sqlc":{"import":"i","driver":"pgx/v5"},"contexts":[]}`))
	if !hasErrorPath(diags, "options.sqlc.package") {
		t.Fatalf("expected error at options.sqlc.package, got %+v", diags)
	}
}

func TestDecode_MissingSqlcImport(t *testing.T) {
	_, diags := config.Decode([]byte(`{"version":1,"sqlc":{"package":"p","driver":"pgx/v5"},"contexts":[]}`))
	if !hasErrorPath(diags, "options.sqlc.import") {
		t.Fatalf("expected error at options.sqlc.import, got %+v", diags)
	}
}

func TestDecode_UnsupportedDriver(t *testing.T) {
	_, diags := config.Decode([]byte(`{"version":1,"sqlc":{"package":"p","import":"i","driver":"mysql"},"contexts":[]}`))
	if !hasErrorPath(diags, "options.sqlc.driver") {
		t.Fatalf("expected error at options.sqlc.driver, got %+v", diags)
	}
}

func TestDecode_ValidMinimal(t *testing.T) {
	root, diags := config.Decode([]byte(`{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[]}`))
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if root == nil || root.Version != 1 || len(root.Contexts) != 0 {
		t.Fatalf("unexpected root: %+v", root)
	}
}

func TestDecode_DuplicateContextName(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{}},
		{"name":"a","package":"b","directory":"b","models":{}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected duplicate context name to be an error, got %+v", diags)
	}
}

func TestDecode_InvalidContextJSON(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{},"bogus":true}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic for an unknown context field, got %+v", diags)
	}
}

func TestDecode_ContextMissingName(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"package":"a","directory":"a","models":{}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts[0].name") {
		t.Fatalf("expected error at contexts[0].name, got %+v", diags)
	}
}

func TestDecode_ContextMissingPackage(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","directory":"a","models":{}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts.a.package") {
		t.Fatalf("expected error at contexts.a.package, got %+v", diags)
	}
}

func TestDecode_ContextMissingDirectory(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","models":{}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts.a.directory") {
		t.Fatalf("expected error at contexts.a.directory, got %+v", diags)
	}
}

func TestDecode_DuplicateModelKey(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{}},
			"User":{"row":"User2","operations":{"insert":"CreateUser2"},"fields":{}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts.a.models.User") {
		t.Fatalf("expected duplicate-model error at contexts.a.models.User, got %+v", diags)
	}
}

func TestDecode_InvalidModelJSON(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{},"bogus":true}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic for an unknown model field, got %+v", diags)
	}
}

func TestDecode_ModelMissingRow(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"operations":{"insert":"CreateUser"},"fields":{}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts.a.models.User.row") {
		t.Fatalf("expected error at contexts.a.models.User.row, got %+v", diags)
	}
}

func TestDecode_DuplicateFieldKey(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"a":{"readable":true},
				"a":{"readable":true}
			}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts.a.models.User.fields.a") {
		t.Fatalf("expected duplicate-field error at contexts.a.models.User.fields.a, got %+v", diags)
	}
}

func TestDecode_InvalidFieldJSON(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"a":{"readable":true,"bogus":true}
			}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic for an unknown field key, got %+v", diags)
	}
}

func TestDecode_FieldOrderPreserved(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"zeta":{"readable":true},
				"alpha":{"readable":true},
				"middle":{"readable":true}
			}}
		}}
	]}`
	root, diags := config.Decode([]byte(src))
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	fields := root.Contexts[0].Models[0].Fields
	want := []string{"zeta", "alpha", "middle"}
	if len(fields) != len(want) {
		t.Fatalf("expected %d fields, got %d", len(want), len(fields))
	}
	for i, name := range want {
		if fields[i].Name != name {
			t.Fatalf("field order not preserved: index %d = %q, want %q", i, fields[i].Name, name)
		}
	}
}

func TestDecode_InsertRequired(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{},"fields":{}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !hasErrorPath(diags, "contexts.a.models.User.operations.insert") {
		t.Fatalf("expected error requiring operations.insert, got %+v", diags)
	}
}

func TestDecode_GeneratedAndFillableConflict(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"id":{"generated":"insert","fillable":true}
			}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected generated+fillable conflict to be an error, got %+v", diags)
	}
}

func TestDecode_MultipleVersionFields(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"a":{"version":true},
				"b":{"version":true}
			}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected at-most-one-version-field to be an error, got %+v", diags)
	}
}

func TestDecode_ImmutableWithoutFillableOrMutableIsWarning(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"a":{"immutable_after_insert":true}
			}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if diagnostics.HasError(diags) {
		t.Fatalf("expected no error-severity diagnostics, got %+v", diags)
	}
	found := false
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a warning diagnostic, got %+v", diags)
	}
}

func TestDecode_ValueObjectField(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"email":{"readable":true,"fillable":true,"value_object":{"type":"Email","constructor":"NewEmail","accessor":"String"}}
			}}
		}}
	]}`
	root, diags := config.Decode([]byte(src))
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	vo := root.Contexts[0].Models[0].Fields[0].ValueObject
	if vo == nil || vo.Type != "Email" || vo.Constructor != "NewEmail" || vo.Accessor != "String" {
		t.Fatalf("unexpected value object mapping: %+v", vo)
	}
}

func TestDecode_ValueObjectRequiresTypeConstructorAndAccessor(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{
				"email":{"value_object":{}}
			}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	for _, path := range []string{
		"contexts.a.models.User.fields.email.value_object.type",
		"contexts.a.models.User.fields.email.value_object.constructor",
		"contexts.a.models.User.fields.email.value_object.accessor",
	} {
		if !hasErrorPath(diags, path) {
			t.Fatalf("expected error at %s, got %+v", path, diags)
		}
	}
}

func TestDecode_UnknownFieldRejected(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[],"bogus":true}`
	root, diags := config.Decode([]byte(src))
	if root != nil {
		t.Fatalf("expected decode to fail on unknown top-level field, got %+v", root)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic, got %+v", diags)
	}
}
