# sqlc-model

<!-- Logo intentionally omitted: this repository does not include a project logo yet. -->

[![CI](https://github.com/macoaure/sqlc-model/actions/workflows/ci.yml/badge.svg)](https://github.com/macoaure/sqlc-model/actions/workflows/ci.yml)
[![MIT License](https://img.shields.io/github/license/macoaure/sqlc-model)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/macoaure/sqlc-model)](https://github.com/macoaure/sqlc-model/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/macoaure/sqlc-model.svg)](https://pkg.go.dev/github.com/macoaure/sqlc-model)

`sqlc-model` is an Eloquent-inspired rich model layer generated over [sqlc](https://sqlc.dev) for Go applications using PostgreSQL and pgx/v5.

sqlc remains the source of truth for SQL, query signatures, parameter types, result types, and driver integration. `sqlc-model` adds generated sessions, collections, mutable model objects, lifecycle state, dirty tracking, validation, relations, eager loading, transactions, and model-oriented persistence APIs.

The generator is not a dynamic ORM. Every terminal database operation resolves to a statically declared sqlc query.

## Features

- Generated Active Record-style models over sqlc queries.
- Session and collection APIs for model construction, lookup, persistence, and transactions.
- Dirty tracking, original snapshots, lifecycle state, and validation errors.
- Typed relations with lazy loading, eager loading, scopes, cache inspection, and explicit terminal operations.
- Value-object field conversion hooks.
- Deterministic generated files and golden snapshot coverage.
- PostgreSQL integration tests, compile fixtures, race-detector checks, and documentation example validation.

## Installation

Install the sqlc process plugin:

```bash
go install github.com/macoaure/sqlc-model/cmd/sqlc-model@latest
```

Add the plugin to `sqlc.yaml`:

```yaml
plugins:
  - name: richmodel
    process:
      cmd: sqlc-model

sql:
  - schema: schema.sql
    queries: query.sql
    engine: postgresql
    codegen:
      - plugin: richmodel
        out: internal/models
        options:
          version: 1
          sqlc:
            package: sqlcdb
            import: example.com/project/internal/database/sqlc
            driver: pgx/v5
          contexts:
            - name: content
              package: content
              directory: content
              models: {}
```

See the [configuration reference](docs/content/reference/configuration.md) for the complete options contract.

## Usage/Examples

Create a session and save a model:

```go
models := content.New(pool)

user, err := models.Users.Find(ctx, userID)
if err != nil {
    return err
}

user.SetName("Ada Lovelace")
if err := user.Save(ctx); err != nil {
    return err
}
```

Run related changes atomically:

```go
err := models.Transaction(ctx, func(tx *content.Session) error {
    user := tx.Users.New().SetName("Ada Lovelace")
    if err := user.Save(ctx); err != nil {
        return err
    }

    post := tx.Posts.New().SetTitle("Notes")
    if err := post.Author().Associate(user); err != nil {
        return err
    }
    return post.Save(ctx)
})
```

Load a scoped relation:

```go
posts, err := user.
    Posts().
    Published().
    Latest().
    Limit(10).
    Get(ctx)
```

Calling `Posts()` or a scope method does not execute SQL. `Get(ctx)` is the operation boundary that may perform lazy loading.

## API Reference

#### Session

```go
func New(pool *pgxpool.Pool, options ...SessionOption) *Session
```

The generated session owns collections, transaction capability, runtime policies, and model identity.

```go
func (s *Session) Transaction(ctx context.Context, fn func(*Session) error) error
```

Returning nil commits. Returning an error rolls back. Panics roll back before continuing.

#### Collection

```go
func (c *UserCollection) New() *User
func (c *UserCollection) Find(ctx context.Context, id UserID) (*User, error)
```

Collections create attached models and expose configured lookup/query operations.

#### Model

```go
func (u *User) SetName(value string) *User
func (u *User) IsDirty(fields ...UserField) bool
func (u *User) Save(ctx context.Context) error
func (u *User) Delete(ctx context.Context) error
func (u *User) Refresh(ctx context.Context) error
```

Models contain current values, original persisted snapshots, lifecycle flags, validation errors, session attachment, and relation caches.

#### Relation

```go
func (r UserPostsRelation) Get(ctx context.Context) ([]*Post, error)
func (r UserPostsRelation) Reload(ctx context.Context) ([]*Post, error)
func (r UserPostsRelation) Cached() ([]*Post, bool)
func (r UserPostsRelation) Forget() *User
```

Only the canonical unconstrained relation is cached by default. Scoped results do not overwrite canonical caches.

## Running Tests

Run the normal Go test suite:

```bash
go test ./...
```

Run the generator-focused checks:

```bash
go test -race ./...
go test ./tests/compile/...
go test ./tests/golden
```

Run PostgreSQL-backed integration checks with a disposable database:

```bash
SQLC_RICHMODEL_TEST_DATABASE_URL='postgres://user:pass@localhost:5432/postgres?sslmode=disable' \
    go test ./tests/integration/...
```

See [How to test generated models](docs/content/how-to/test-generated-models.md) for the full test-level guide.

## FAQ

#### Is sqlc-model an ORM?

No. It generates model-oriented Go code over named sqlc queries. It does not build dynamic SQL at runtime.

#### Does sqlc still own database access?

Yes. sqlc owns SQL parsing, static analysis, parameter types, result shapes, driver integration, and transaction-bound query objects.

#### When does generated code execute SQL?

Only terminal operations execute SQL. Examples include `Find(ctx)`, `Save(ctx)`, `Delete(ctx)`, `Refresh(ctx)`, relation `Get(ctx)`, and relation `Reload(ctx)`.

#### Where should I start?

Start with [Generate the first rich model](docs/content/tutorials/first-model.md), then read the [configuration reference](docs/content/reference/configuration.md).

## Feedback

Open a [GitHub issue](https://github.com/macoaure/sqlc-model/issues) for bugs, feature requests, and documentation gaps.

## Acknowledgements

- [sqlc](https://sqlc.dev/)
- [pgx](https://github.com/jackc/pgx)
- [Diataxis](https://diataxis.fr/)
- [Laravel Eloquent](https://laravel.com/docs/eloquent)

## License

[MIT](LICENSE)
