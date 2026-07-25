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

func TestRenderCollection_QueryChainGet(t *testing.T) {
	id := rfield("ID", "int64")
	active := rfield("Active", "bool")
	limitParam := &plan.ResolvedQueryScope{Name: "Limit", Parameter: "limit", Argument: "int32", GoType: mapping.GoType{Expr: "int32"}, Number: 2}
	activeParam := &plan.ResolvedQueryScope{Name: "Active", Parameter: "active", Value: []byte("true"), ValueSet: true, GoType: mapping.GoType{Expr: "bool"}, Number: 1}
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name:   "User",
		Row:    "User",
		Fields: []plan.ResolvedField{*id, *active},
		Queries: []plan.ResolvedQuery{
			{
				Name:     "default",
				Terminal: config.QueryTerminalGet,
				Operation: &plan.ResolvedOperation{
					Kind:      contract.List,
					QueryName: "ListActiveUsers",
					Query:     &pb.Query{Name: "ListActiveUsers", Cmd: ":many", Text: "SELECT id, active FROM users WHERE active = $1 LIMIT $2"},
				},
				Scopes: []plan.ResolvedQueryScope{*activeParam, *limitParam},
				Scan:   []*plan.ResolvedField{id, active},
			},
		},
	}

	out, diags := codegen.RenderCollection(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"func (c *UserCollection) Query() UserQuery",
		"type UserQuery struct",
		"func (q UserQuery) Active() UserQuery",
		"func (q UserQuery) Limit(limit int32) UserQuery",
		"func (q UserQuery) Get(ctx context.Context) ([]*User, error)",
		"q.coll.session.executor.Query(ctx",
		"rows.Scan(&rec.ID, &rec.Active)",
		"out = append(out, u)",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated collection to contain %q:\n%s", want, src)
		}
	}
}

func TestRenderCollection_QueryVariantScope(t *testing.T) {
	id := rfield("ID", "int64")
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name:   "User",
		Row:    "User",
		Fields: []plan.ResolvedField{*id},
		Queries: []plan.ResolvedQuery{
			{
				Name:     "default",
				Terminal: config.QueryTerminalGet,
				Operation: &plan.ResolvedOperation{
					Kind:      contract.List,
					QueryName: "ListUsers",
					Query:     &pb.Query{Name: "ListUsers", Cmd: ":many", Text: "SELECT id FROM users"},
				},
				Scopes: []plan.ResolvedQueryScope{
					{Name: "OrderByName", QueryName: "ListUsersByName", Variant: &pb.Query{Name: "ListUsersByName", Cmd: ":many", Text: "SELECT id FROM users ORDER BY name"}},
				},
				Scan: []*plan.ResolvedField{id},
			},
		},
	}

	out, diags := codegen.RenderCollection(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"querySQL string",
		"func (q UserQuery) OrderByName() UserQuery",
		`q.querySQL = "SELECT id FROM users ORDER BY name"`,
		`sql := "SELECT id FROM users"`,
		"if q.querySQL != \"\"",
		"rows, err := q.coll.session.executor.Query(ctx, sql)",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated collection to contain %q:\n%s", want, src)
		}
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

func TestRenderModel_ValueObjectUsesExposedType(t *testing.T) {
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name: "User",
		Row:  "User",
		Fields: []plan.ResolvedField{
			{
				ResolvedField: mapping.ResolvedField{
					Name: "email", GoField: "Email", ColumnName: "email",
					GoType: mapping.GoType{Expr: "Email"}, PersistedGoType: mapping.GoType{Expr: "string"},
					ValueObject: &config.ValueObjectMapping{Type: "Email", Constructor: "NewEmail", Accessor: "String"},
					NotNull:     true,
				},
				Policy: config.FieldPolicy{Name: "email", Readable: true, Fillable: true, Mutable: true},
			},
		},
	}

	out, diags := codegen.RenderModel(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"func (u *User) Email() Email",
		"func (u *User) SetEmail(value Email) *User",
		"func (u *User) OriginalEmail() Email",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated model to contain %q:\n%s", want, src)
		}
	}
}

func TestRenderModel_LifecycleAPISurface(t *testing.T) {
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name: "User",
		Row:  "User",
		Fields: []plan.ResolvedField{
			{
				ResolvedField: mapping.ResolvedField{
					Name: "id", GoField: "ID", ColumnName: "id",
					GoType: mapping.GoType{Expr: "int64"}, NotNull: true,
				},
				Policy: config.FieldPolicy{Name: "id", Readable: true},
			},
			{
				ResolvedField: mapping.ResolvedField{
					Name: "name", GoField: "Name", ColumnName: "name",
					GoType: mapping.GoType{Expr: "string"}, NotNull: true,
				},
				Policy: config.FieldPolicy{Name: "name", Readable: true, Fillable: true, Mutable: true},
			},
		},
	}

	model, diags := codegen.RenderModel(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	session, diags := codegen.RenderSession(plan.ResolvedContext{
		Name:      "content",
		Package:   "content",
		Directory: "content",
		Models:    []plan.ResolvedModel{m},
	})
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}

	src := string(model) + "\n" + string(session)
	for _, want := range []string{
		"func (u *User) IsPersisted() bool",
		"func (u *User) IsDetached() bool",
		"func (u *User) HasChanges() bool",
		"ErrModelDetached = errors.New(\"richmodel: model is not attached to a session\")",
		"ErrModelDeleted = errors.New(\"richmodel: model is deleted\")",
		"ErrModelNotPersisted = errors.New(\"richmodel: model is not persisted\")",
		"return ErrModelDeleted",
		"return ErrModelDetached",
		"return ErrModelNotPersisted",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated lifecycle API to contain %q:\n%s", want, src)
		}
	}
}

func TestRenderStore_ValueObjectConvertsAtBoundaries(t *testing.T) {
	field := &plan.ResolvedField{
		ResolvedField: mapping.ResolvedField{
			Name: "email", GoField: "Email", ColumnName: "email",
			GoType: mapping.GoType{Expr: "Email"}, PersistedGoType: mapping.GoType{Expr: "string"},
			ValueObject: &config.ValueObjectMapping{Type: "Email", Constructor: "NewEmail", Accessor: "String"},
			NotNull:     true,
		},
		Policy: config.FieldPolicy{Name: "email", Readable: true, Fillable: true, Mutable: true},
	}
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name:   "User",
		Row:    "User",
		Fields: []plan.ResolvedField{*field},
		Operations: plan.ResolvedOperations{
			Insert: &plan.ResolvedOperation{
				Kind:  contract.Insert,
				Query: &pb.Query{Name: "CreateUser", Cmd: ":one", Text: "INSERT ..."},
				Params: []plan.ParamBinding{
					{Number: 1, Field: field},
				},
				Scan: []*plan.ResolvedField{field},
			},
		},
	}

	out, diags := codegen.RenderStore(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"rec.Email.String()",
		"var scanEmail string",
		"&scanEmail",
		"email, err := NewEmail(scanEmail)",
		`fmt.Errorf("User.Email: %w", err)`,
		"out.Email = email",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated store to contain %q:\n%s", want, src)
		}
	}
}

func TestRenderStore_MapsNoRowsToErrNotFound(t *testing.T) {
	field := rfield("ID", "int64")
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name:   "User",
		Row:    "User",
		Fields: []plan.ResolvedField{*field},
		Operations: plan.ResolvedOperations{
			Find: &plan.ResolvedOperation{
				Kind:  contract.Find,
				Query: &pb.Query{Name: "FindUser", Cmd: ":one", Text: "SELECT id FROM users WHERE id = $1"},
				Scan:  []*plan.ResolvedField{field},
			},
		},
	}

	out, diags := codegen.RenderStore(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		`errors.Is(err, pgx.ErrNoRows)`,
		"return userRecord{}, ErrNotFound",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated store to contain %q:\n%s", want, src)
		}
	}
}

func TestRenderCollection_LookupMapsNoRowsToErrNotFound(t *testing.T) {
	id := rfield("ID", "int64")
	name := rfield("Name", "string")
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name:   "User",
		Row:    "User",
		Fields: []plan.ResolvedField{*id, *name},
		Lookups: []plan.ResolvedLookup{
			{
				Name:      "FindByName",
				QueryName: "FindUserByName",
				Operation: &plan.ResolvedOperation{
					Kind:  contract.Lookup,
					Query: &pb.Query{Name: "FindUserByName", Cmd: ":one", Text: "SELECT id, name FROM users WHERE name = $1"},
					Params: []plan.ParamBinding{
						{Number: 1, Field: name},
					},
					Scan: []*plan.ResolvedField{id, name},
				},
			},
		},
	}

	out, diags := codegen.RenderCollection(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"func (c *UserCollection) FindByName(ctx context.Context, name string) (*User, error)",
		`errors.Is(err, pgx.ErrNoRows)`,
		"return nil, ErrNotFound",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated collection to contain %q:\n%s", want, src)
		}
	}
}

func TestRenderModel_ValidateBeforeSave(t *testing.T) {
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name: "User",
		Row:  "User",
		Fields: []plan.ResolvedField{
			{
				ResolvedField: mapping.ResolvedField{
					Name: "name", GoField: "Name", ColumnName: "name",
					GoType: mapping.GoType{Expr: "string"}, NotNull: true,
				},
				Policy: config.FieldPolicy{Name: "name", Readable: true, Mutable: true},
			},
		},
		Operations: plan.ResolvedOperations{
			Insert: &plan.ResolvedOperation{Kind: contract.Insert, Query: &pb.Query{Name: "CreateUser", Cmd: ":one"}},
			Update: &plan.ResolvedOperation{Kind: contract.Update, Query: &pb.Query{Name: "UpdateUser", Cmd: ":one"}},
		},
	}

	out, diags := codegen.RenderModel(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"func (u *User) Validate() error",
		"interface{ validateUser() error }",
		"if err := u.Validate(); err != nil {\n\t\treturn err\n\t}",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated model to contain %q:\n%s", want, src)
		}
	}

	validatePos := strings.Index(src, "if err := u.Validate(); err != nil")
	insertPos := strings.Index(src, "u.coll.store.insert")
	updatePos := strings.Index(src, "u.coll.store.update")
	if validatePos < 0 || insertPos < 0 || updatePos < 0 || validatePos > insertPos || validatePos > updatePos {
		t.Fatalf("expected Save to validate before insert/update:\n%s", src)
	}
}

func TestRenderModel_FieldErrorHelpersReplaceAndClear(t *testing.T) {
	ctx := plan.ResolvedContext{Name: "content", Package: "content", Directory: "content"}
	m := plan.ResolvedModel{
		Name: "User",
		Row:  "User",
		Fields: []plan.ResolvedField{
			{
				ResolvedField: mapping.ResolvedField{
					Name: "name", GoField: "Name", ColumnName: "name",
					GoType: mapping.GoType{Expr: "string"}, NotNull: true,
				},
				Policy: config.FieldPolicy{Name: "name", Readable: true, Mutable: true},
			},
		},
	}

	out, diags := codegen.RenderModel(ctx, m)
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	src := string(out)
	for _, want := range []string{
		"u.errs[f] = err",
		"delete(u.errs, f)",
		"u.clearFieldError(UserFieldName)",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected generated model to contain %q:\n%s", want, src)
		}
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
