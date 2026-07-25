package unit

import (
	"encoding/json"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/codegen"
	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// minimalHasManyRelation builds the smallest plan.ResolvedRelation/
// ResolvedModel pair RenderRelation needs, for tests that exercise a
// specific codegen branch directly rather than through the full
// config-decode -> plan-build pipeline (which would reject an
// incompatible scope value before codegen ever runs it — see
// relation.ValidateScopes's own valueCompatible check, already exercised
// by tests/unit/relation_scope_test.go).
func minimalHasManyRelation(scopes []config.ScopeConfiguration) (plan.ResolvedContext, plan.ResolvedModel, plan.ResolvedRelation) {
	target := plan.ResolvedModel{Name: "Child", Row: "Child"}
	local := rfield("ID", "pgtype.UUID")
	rel := plan.ResolvedRelation{
		Name:       "Children",
		Kind:       config.HasMany,
		Target:     &target,
		LocalKey:   local,
		ForeignKey: "parent_id",
		LazyQuery: &pb.Query{
			Name: "ListChildrenByParent", Cmd: ":many",
			Params: []*pb.Parameter{{Number: 1, Column: &pb.Column{Name: "parent_id", NotNull: true, Type: &pb.Identifier{Name: "uuid"}}}},
		},
		LazyParams: []plan.RelationParam{{Number: 1, Name: "parent_id", Source: plan.ParamFromParent, ParentField: local}},
		Scopes:     scopes,
	}
	m := plan.ResolvedModel{Name: "Parent", Row: "Parent", Relations: []plan.ResolvedRelation{rel}}
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	return ctx, m, rel
}

// TestRenderRelation_ScopeValueUnsupportedTypeFails exercises RenderRelation's
// own goLiteral-error diagnostic path directly: a "value" scope whose JSON
// value is an array. In the full pipeline this is already rejected earlier
// by relation.ValidateScopes (tested in relation_scope_test.go), so
// RenderRelation's own defensive check is only reachable by constructing
// the plan directly, as here.
func TestRenderRelation_ScopeValueUnsupportedTypeFails(t *testing.T) {
	scopes := []config.ScopeConfiguration{
		{Name: "Bogus", Parameter: "parent_id", Value: json.RawMessage(`[1,2,3]`), ValueSet: true},
	}
	ctx, m, rel := minimalHasManyRelation(scopes)

	_, diags := codegen.RenderRelation(ctx, m, rel)
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic for an unsupported scope value literal, got %+v", diags)
	}
}

// TestRenderRelation_NoScopesGenerates is a baseline sanity check that
// minimalHasManyRelation's hand-built plan renders successfully with zero
// scopes, so the failing case above is attributable to the scope value
// alone.
func TestRenderRelation_NoScopesGenerates(t *testing.T) {
	ctx, m, rel := minimalHasManyRelation(nil)
	_, diags := codegen.RenderRelation(ctx, m, rel)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
}
