# sqlc-gen-richmodel

[![CI](https://github.com/macoaure/sqlc-model/actions/workflows/ci.yml/badge.svg)](https://github.com/macoaure/sqlc-model/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/github/license/macoaure/sqlc-model)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/macoaure/sqlc-model)](https://github.com/macoaure/sqlc-model/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/macoaure/sqlc-gen-richmodel.svg)](https://pkg.go.dev/github.com/macoaure/sqlc-gen-richmodel)

`sqlc-gen-richmodel` is an Eloquent-inspired rich model layer generated over [sqlc](https://sqlc.dev). sqlc remains the source of truth for SQL, query signatures, parameter types, and result types; this plugin adds mutable model objects, lifecycle state, dirty tracking, relations, validation, and model-oriented persistence APIs on top of it — no dynamic queries, no ORM magic, just generated code.

## Metadata

- Repository: <https://github.com/macoaure/sqlc-model>
- Default branch: `main`
- Visibility: public
- Go module: `github.com/macoaure/sqlc-gen-richmodel`
- Go baseline: `1.25.0`
- sqlc plugin SDK: `github.com/sqlc-dev/plugin-sdk-go v1.23.0`

## Install

Add the plugin to your `sqlc.yaml`:

```yaml
plugins:
  - name: richmodel
    process:
      cmd: sqlc-gen-richmodel

sql:
  - schema: schema.sql
    queries: query.sql
    engine: postgresql
    codegen:
      - plugin: richmodel
        out: internal/models
```

Then install the plugin binary:

```sh
go install github.com/macoaure/sqlc-gen-richmodel/cmd/sqlc-gen-richmodel@latest
```

## Usage

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

Run related model operations atomically with `Transaction`:

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

## Status

This project is under active development. The current test strategy covers golden snapshots, compile fixtures, PostgreSQL integration tests, race-detector checks, and documentation fixture validation.

## Development Checks

```sh
go test ./...
go test -race ./...
go test ./tests/compile
go test ./tests/golden
go test ./tests/integration
```

Run PostgreSQL-backed integration checks with a disposable database:

```sh
SQLC_RICHMODEL_TEST_DATABASE_URL='postgres://user:pass@localhost:5432/postgres?sslmode=disable' \
	go test ./tests/integration/...
```

See [docs/content/how-to/test-generated-models.md](docs/content/how-to/test-generated-models.md) for the full test-level guide.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to set up a development environment, run tests, and submit changes.

## License

MIT — see [LICENSE](LICENSE).
