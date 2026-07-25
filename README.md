# sqlc-gen-richmodel

[![CI](https://github.com/macoaure/sqlc-gen-richmodel/actions/workflows/ci.yml/badge.svg)](https://github.com/macoaure/sqlc-gen-richmodel/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/github/license/macoaure/sqlc-gen-richmodel)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/macoaure/sqlc-gen-richmodel)](https://github.com/macoaure/sqlc-gen-richmodel/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/macoaure/sqlc-gen-richmodel.svg)](https://pkg.go.dev/github.com/macoaure/sqlc-gen-richmodel)

`sqlc-gen-richmodel` is an Eloquent-inspired rich model layer generated over [sqlc](https://sqlc.dev). sqlc remains the source of truth for SQL, query signatures, parameter types, and result types; this plugin adds mutable model objects, lifecycle state, dirty tracking, relations, validation, and model-oriented persistence APIs on top of it — no dynamic queries, no ORM magic, just generated code.

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

## Status

This project is under active development. See [specs/](specs/) for the specifications driving implementation.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to set up a development environment, run tests, and submit changes.

## License

MIT — see [LICENSE](LICENSE).
