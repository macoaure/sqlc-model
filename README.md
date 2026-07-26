# sqlc-model

[![CI](https://github.com/macoaure/sqlc-model/actions/workflows/ci.yml/badge.svg)](https://github.com/macoaure/sqlc-model/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/github/license/macoaure/sqlc-model)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/macoaure/sqlc-model)](https://github.com/macoaure/sqlc-model/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/macoaure/sqlc-model.svg)](https://pkg.go.dev/github.com/macoaure/sqlc-model)

`sqlc-model` is an Eloquent-inspired rich model layer generated over [sqlc](https://sqlc.dev). sqlc remains the source of truth for SQL, query signatures, parameter types, and result types; the `sqlc-model` plugin adds mutable model objects, lifecycle state, dirty tracking, relations, validation, and model-oriented persistence APIs on top of it.

The generator does not build dynamic queries at runtime. Every terminal database operation is backed by a named sqlc query.

## Metadata

- Repository: <https://github.com/macoaure/sqlc-model>
- Default branch: `main`
- Visibility: public
- Go module: `github.com/macoaure/sqlc-model`
- Go baseline: `1.25.0`
- sqlc plugin SDK: `github.com/sqlc-dev/plugin-sdk-go v1.23.0`

## Install

Install the plugin binary:

```sh
go install github.com/macoaure/sqlc-model/cmd/sqlc-model@latest
```

Add the plugin to `sqlc.yaml` and pass the rich-model options:

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

See the [configuration reference](docs/content/reference/configuration.md) for model, field, relation, query, and lookup options.

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

## Documentation

- [Tutorials](docs/content/tutorials/index.md): learn the generated model workflow through one path.
- [How-to guides](docs/content/how-to/index.md): solve focused tasks such as configuring relations or testing generated output.
- [Reference](docs/content/reference/index.md): look up configuration keys, generated APIs, errors, and compatibility contracts.
- [Explanation](docs/content/explanation/index.md): understand the design choices and boundaries.

## Status

This project is under active development. The current test strategy covers golden snapshots, compile fixtures, PostgreSQL integration tests, race-detector checks, and documentation fixture validation.

## Development Checks

```sh
go test ./...
go test -race ./...
go test ./tests/compile/...
go test ./tests/golden
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
