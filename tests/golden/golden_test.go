package golden

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/generate"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

var update = flag.Bool("update", false, "regenerate testdata/golden fixtures instead of comparing against them")

// TestUserBasic validates User Story 1's independent test end to end: from
// quickstart.md's `users` table and queries, generation produces the exact
// file set the quickstart documents, byte-for-byte stable across repeated
// runs against unchanged configuration (FR-018, SC-002).
func TestUserBasic(t *testing.T) {
	assertGolden(t, "user-basic", userBasicRequest(t))
}

// TestFieldPolicyMatrix validates User Story 2's independent test: a model
// configured with a mix of field policies exposes exactly the getters/
// setters each policy implies.
func TestFieldPolicyMatrix(t *testing.T) {
	assertGolden(t, "field-policy-matrix", fieldPolicyMatrixRequest(t))
}

// TestFieldPolicyMatrix_AmbiguousFieldFails validates the other half of User
// Story 2's independent test: a field with no explicit override and no
// matching column fails generation with a diagnostic identifying the
// ambiguity (FR-007), rather than guessing.
func TestFieldPolicyMatrix_AmbiguousFieldFails(t *testing.T) {
	req := fieldPolicyMatrixRequest(t)
	for _, q := range req.Queries {
		if q.Name == "CreateWidget" || q.Name == "GetWidget" || q.Name == "UpdateWidget" {
			q.Columns = append(q.Columns, col("mystery_field", "text", true))
		}
	}

	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected generation to fail for an unresolvable column, got %d files", len(resp.Files))
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error diagnostic identifying the ambiguity, got %+v", diags)
	}
}

// TestTransactionSessionAPI validates specs/003-transactions-session-identity's
// generated public surface: sessions expose Transaction, generated queries
// route through one executor abstraction, and the exported association errors
// remain stable.
func TestTransactionSessionAPI(t *testing.T) {
	resp, diags := generate.Generate(userRelationsRequest(t))
	if resp == nil {
		t.Fatalf("generation failed: %v", diags)
	}

	session := generatedFile(t, resp.Files, "content/session_gen.go")
	for _, want := range []string{
		"type richmodelExecutor interface",
		"type sessionIdentity struct{}",
		"identity    *sessionIdentity",
		"func (s *Session) Transaction(ctx context.Context, fn func(*Session) error) error",
		"tx, err := s.pool.Begin(ctx)",
		"defer func() {",
		"_ = tx.Rollback(ctx)",
		"tx.Commit(ctx)",
		"ErrSessionMismatch = errors.New(\"richmodel: related model belongs to a different session\")",
		"ErrUnsavedRelated = errors.New(\"richmodel: related model has no persisted identifier\")",
	} {
		if !strings.Contains(session, want) {
			t.Fatalf("expected generated session to contain %q:\n%s", want, session)
		}
	}

	store := generatedFile(t, resp.Files, "content/user_store_gen.go")
	if !strings.Contains(store, "executor richmodelExecutor") || strings.Contains(store, "pool *pgxpool.Pool") {
		t.Fatalf("expected generated store to use richmodelExecutor instead of *pgxpool.Pool:\n%s", store)
	}
}

func TestValueObjectFieldMapping(t *testing.T) {
	resp, diags := generate.Generate(valueObjectRequest())
	if resp == nil {
		t.Fatalf("generation failed: %v", diags)
	}
	if *update {
		writeFiles(t, filepath.Join("..", "compile", "user-value-object"), resp.Files)
		t.Log("updated value-object compile fixture")
	}

	model := generatedFile(t, resp.Files, "content/user_gen.go")
	for _, want := range []string{
		"func (u *User) Email() Email",
		"func (u *User) SetEmail(value Email) *User",
		"func (u *User) OriginalEmail() Email",
	} {
		if !strings.Contains(model, want) {
			t.Fatalf("expected generated model to contain %q:\n%s", want, model)
		}
	}

	store := generatedFile(t, resp.Files, "content/user_store_gen.go")
	for _, want := range []string{
		"rec.Email.String()",
		"var scanEmail string",
		"email, err := NewEmail(scanEmail)",
		`fmt.Errorf("User.Email: %w", err)`,
	} {
		if !strings.Contains(store, want) {
			t.Fatalf("expected generated store to contain %q:\n%s", want, store)
		}
	}
}

func TestQueryCompositionAPI(t *testing.T) {
	resp, diags := generate.Generate(queryCompositionRequest(t))
	if resp == nil {
		t.Fatalf("generation failed: %v", diags)
	}
	if *update {
		writeFiles(t, filepath.Join("..", "compile", "user-query"), resp.Files)
		t.Log("updated query-composition compile fixture")
	}

	collection := generatedFile(t, resp.Files, "content/user_collection_gen.go")
	for _, want := range []string{
		"func (c *UserCollection) Query() UserQuery",
		"func (q UserQuery) Active() UserQuery",
		"func (q UserQuery) Limit(limit int32) UserQuery",
		"func (q UserQuery) OrderByName() UserQuery",
		"func (q UserQuery) WithPosts() UserQuery",
		"func (q UserQuery) Get(ctx context.Context) ([]*User, error)",
		"q.coll.session.executor.Query(ctx, sql, q.active, q.limit)",
		"q.coll.EagerLoad(ctx, out, WithPosts())",
		"func (c *UserCollection) FindByName(ctx context.Context, name string) (*User, error)",
	} {
		if !strings.Contains(collection, want) {
			t.Fatalf("expected generated collection to contain %q:\n%s", want, collection)
		}
	}
}

// assertGolden is TestUserBasic's comparison logic, generalized so other
// fixtures (e.g. the field-policy matrix) can reuse it against their own
// testdata/golden/<name>/ directory.
func assertGolden(t *testing.T, name string, req *pb.GenerateRequest) {
	t.Helper()

	resp1, diags1 := generate.Generate(req)
	if resp1 == nil {
		t.Fatalf("generation failed: %v", diags1)
	}
	resp2, diags2 := generate.Generate(req)
	if resp2 == nil {
		t.Fatalf("second generation run failed: %v", diags2)
	}
	if len(resp1.Files) != len(resp2.Files) {
		t.Fatalf("non-deterministic file count: run1=%d run2=%d", len(resp1.Files), len(resp2.Files))
	}
	for i := range resp1.Files {
		if resp1.Files[i].Name != resp2.Files[i].Name || string(resp1.Files[i].Contents) != string(resp2.Files[i].Contents) {
			t.Fatalf("non-deterministic output across repeated runs (SC-002 violation) for %s", resp1.Files[i].Name)
		}
	}

	goldenDir := filepath.Join("..", "..", "testdata", "golden", name)

	if *update {
		for _, f := range resp1.Files {
			path := filepath.Join(goldenDir, f.Name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, f.Contents, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("updated golden fixtures under %s", goldenDir)
		return
	}

	wantPaths := map[string]bool{}
	err := filepath.WalkDir(goldenDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(goldenDir, path)
		if err != nil {
			return err
		}
		wantPaths[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walking golden dir: %v", err)
	}

	gotPaths := map[string]bool{}
	for _, f := range resp1.Files {
		gotPaths[f.Name] = true
		want, err := os.ReadFile(filepath.Join(goldenDir, f.Name))
		if err != nil {
			t.Errorf("unexpected generated file %s not present in golden fixture (run with -update to add it): %v", f.Name, err)
			continue
		}
		if string(f.Contents) != string(want) {
			t.Errorf("generated file %s does not match golden fixture (run with -update to refresh)\n--- got ---\n%s", f.Name, f.Contents)
		}
	}
	for path := range wantPaths {
		if !gotPaths[path] {
			t.Errorf("golden fixture file %s was not produced by generation (run with -update to refresh)", path)
		}
	}
}

func generatedFile(t *testing.T, files []*pb.File, name string) string {
	t.Helper()
	for _, f := range files {
		if f.Name == name {
			return string(f.Contents)
		}
	}
	t.Fatalf("generated file %s not found; got %v", name, fileNames(files))
	return ""
}

// TestUserBasicReturningRequired validates the quickstart's negative case:
// an insert/update query without a full-row RETURNING fails generation with
// zero output (FR-004, FR-017).
func TestUserBasicReturningRequired(t *testing.T) {
	req := userBasicRequest(t)
	for _, q := range req.Queries {
		if q.Name == "CreateUser" {
			q.Cmd = ":exec"
		}
	}

	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected generation to fail when CreateUser lacks RETURNING, got %d files", len(resp.Files))
	}
	if len(diags) == 0 {
		t.Fatal("expected at least one diagnostic explaining the failure")
	}
}
