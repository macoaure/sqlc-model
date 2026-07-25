package golden

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/generate"
)

// TestUserRelations validates specs/002-relations-lazy-eager-loading's User
// Stories 1-3 together: has_many/belongs_to/many_to_many relations, a
// scope, and eager loading with a nested many-to-many relation, generated
// deterministically (assertGolden runs generation twice and diffs).
func TestUserRelations(t *testing.T) {
	assertGolden(t, "user-relations", userRelationsRequest(t))
}

func TestUserRelations_AssociationUsesSessionIdentity(t *testing.T) {
	resp, diags := generate.Generate(userRelationsRequest(t))
	if resp == nil {
		t.Fatalf("generation failed: %v", diags)
	}

	for _, path := range []string{
		"content/post_author_relation_gen.go",
		"content/post_tags_relation_gen.go",
	} {
		src := generatedFile(t, resp.Files, path)
		if !strings.Contains(src, "!sameSessionIdentity(") {
			t.Fatalf("expected %s to compare private session identity:\n%s", path, src)
		}
		if strings.Contains(src, "session != ") {
			t.Fatalf("expected %s not to compare Session pointer equality:\n%s", path, src)
		}
	}
}

// TestUserRelations_InvalidInverseFails validates FR-004/FR-011: an
// inverse relation name that doesn't exist on the target model fails
// generation with a diagnostic, producing zero files (FR-017).
func TestUserRelations_InvalidInverseFails(t *testing.T) {
	req := userRelationsRequest(t)
	req.PluginOptions = []byte(mustReplace(t, userRelationsOptions, `"inverse": "Author"`, `"inverse": "NoSuchRelation"`))

	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected generation to fail for an invalid inverse reference, got %d files", len(resp.Files))
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic identifying the bad inverse reference, got %+v", diags)
	}
}

// TestUserRelations_MissingTargetModelFails validates that a relation
// naming a target model absent from the context fails generation
// (research.md — this is also how a cross-context relation attempt is
// rejected).
func TestUserRelations_MissingTargetModelFails(t *testing.T) {
	req := userRelationsRequest(t)
	req.PluginOptions = []byte(mustReplace(t, userRelationsOptions, `"model": "Post"`, `"model": "NoSuchModel"`))

	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected generation to fail for an unresolvable target model, got %d files", len(resp.Files))
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic identifying the unresolved target model, got %+v", diags)
	}
}

func mustReplace(t *testing.T, s, old, new string) string {
	t.Helper()
	if !strings.Contains(s, old) {
		t.Fatalf("replacement %q not found in options fixture", old)
	}
	return strings.Replace(s, old, new, 1)
}
