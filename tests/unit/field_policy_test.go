package unit

import (
	"testing"

	"github.com/macoaure/sqlc-model/internal/config"
	"github.com/macoaure/sqlc-model/internal/diagnostics"
)

func decodeOneField(t *testing.T, fieldJSON string) *config.RootConfiguration {
	t.Helper()
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{"x":` + fieldJSON + `}}
		}}
	]}`
	root, diags := config.Decode([]byte(src))
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors for %s: %+v", fieldJSON, diags)
	}
	return root
}

func TestFieldPolicy_VersionIsImplicitlyReadable(t *testing.T) {
	root := decodeOneField(t, `{"version":true}`)
	f := root.Contexts[0].Models[0].Fields[0]
	if !f.Readable {
		t.Fatalf("expected version field to be implicitly readable, got %+v", f)
	}
}

func TestFieldPolicy_VersionAndFillableConflict(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{"x":{"version":true,"fillable":true}}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected version+fillable conflict to be an error, got %+v", diags)
	}
}

func TestFieldPolicy_VersionAndMutableConflict(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{"x":{"version":true,"mutable":true}}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected version+mutable conflict to be an error, got %+v", diags)
	}
}

func TestFieldPolicy_GeneratedInsertVsSave(t *testing.T) {
	insertField := decodeOneField(t, `{"generated":"insert"}`).Contexts[0].Models[0].Fields[0]
	if insertField.Generated != config.GeneratedInsert {
		t.Fatalf("expected GeneratedInsert, got %v", insertField.Generated)
	}
	saveField := decodeOneField(t, `{"generated":"save"}`).Contexts[0].Models[0].Fields[0]
	if saveField.Generated != config.GeneratedSave {
		t.Fatalf("expected GeneratedSave, got %v", saveField.Generated)
	}
}

func TestFieldPolicy_UnsupportedGeneratedValue(t *testing.T) {
	src := `{"version":1,"sqlc":{"package":"p","import":"i","driver":"pgx/v5"},"contexts":[
		{"name":"a","package":"a","directory":"a","models":{
			"User":{"row":"User","operations":{"insert":"CreateUser"},"fields":{"x":{"generated":"update"}}}
		}}
	]}`
	_, diags := config.Decode([]byte(src))
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected unsupported generated value to be an error, got %+v", diags)
	}
}

func TestFieldPolicy_NeitherFillableNorMutableGeneratesSuccessfully(t *testing.T) {
	root := decodeOneField(t, `{"readable":true}`)
	f := root.Contexts[0].Models[0].Fields[0]
	if f.Fillable || f.Mutable {
		t.Fatalf("expected no setter capability, got %+v", f)
	}
}

func TestFieldPolicy_FillableAndMutableBothAllowed(t *testing.T) {
	root := decodeOneField(t, `{"fillable":true,"mutable":true}`)
	f := root.Contexts[0].Models[0].Fields[0]
	if !f.Fillable || !f.Mutable {
		t.Fatalf("expected both fillable and mutable, got %+v", f)
	}
}

func TestFieldPolicy_SensitiveIndependentOfOtherFlags(t *testing.T) {
	root := decodeOneField(t, `{"readable":true,"sensitive":true}`)
	f := root.Contexts[0].Models[0].Fields[0]
	if !f.Sensitive || !f.Readable {
		t.Fatalf("expected sensitive+readable to coexist, got %+v", f)
	}
}
