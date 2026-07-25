package relation

import (
	"encoding/json"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/mapping"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// TestParamGoType_NilQuery covers paramGoType's nil-query guard directly.
// In ValidateScopes' real call pattern, the fallback call to
// paramGoType(g.EagerQuery, ...) is only reached when the parameter is
// already known (via paramNames' union) to exist in one of the two
// queries — making a genuinely nil second query at that point unreachable
// through the public config/generate entry points. Covered directly here
// since paramGoType is unexported.
func TestParamGoType_NilQuery(t *testing.T) {
	if _, ok := paramGoType(nil, "anything"); ok {
		t.Fatal("expected ok=false for a nil query")
	}
}

// TestParamGoType_NotFound covers paramGoType's not-found return when the
// query is non-nil but has no parameter by that name.
func TestParamGoType_NotFound(t *testing.T) {
	q := &pb.Query{Params: []*pb.Parameter{{Number: 1, Column: &pb.Column{Name: "id", Type: &pb.Identifier{Name: "uuid"}}}}}
	if _, ok := paramGoType(q, "nonexistent"); ok {
		t.Fatal("expected ok=false for an unknown parameter name")
	}
}

// TestValueCompatible_InvalidJSON covers valueCompatible's json.Unmarshal
// error branch. Unreachable through the public API since a
// ScopeConfiguration.Value is always itself valid JSON by construction
// (decoded from the same options document decodeScope already parsed).
func TestValueCompatible_InvalidJSON(t *testing.T) {
	if valueCompatible(json.RawMessage(`{not valid`), mapping.GoType{Expr: "bool"}) {
		t.Fatal("expected false for malformed JSON")
	}
}

// TestSameColumnSet_Direct is a direct sanity check of sameColumnSet's
// three branches (length mismatch, content mismatch, exact match) —
// already exercised indirectly via tests/unit/relation_edge_test.go, kept
// here for a fast, isolated regression check on the helper itself.
func TestSameColumnSet_Direct(t *testing.T) {
	a := []*pb.Column{{Name: "id"}, {Name: "name"}}
	b := []*pb.Column{{Name: "id"}, {Name: "name"}}
	if !sameColumnSet(a, b) {
		t.Fatal("expected identical column sets to match")
	}
	if sameColumnSet(a, []*pb.Column{{Name: "id"}}) {
		t.Fatal("expected different-length column sets to differ")
	}
	if sameColumnSet(a, []*pb.Column{{Name: "id"}, {Name: "other"}}) {
		t.Fatal("expected same-length, different-content column sets to differ")
	}
}
