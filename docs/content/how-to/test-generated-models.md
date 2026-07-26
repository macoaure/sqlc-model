# How to test generated models

Use three test levels. None is sufficient alone.

## Commands

```sh
go test ./tests/golden
go test ./tests/compile/...
go test -race ./...
SQLC_RICHMODEL_TEST_DATABASE_URL='postgres://user:pass@localhost:5432/postgres?sslmode=disable' \
	go test ./tests/integration/...
```

The PostgreSQL command is optional locally unless you are changing generated persistence behavior. It must use a disposable database or schema, never production data.

## Generator golden tests

Feed a fixture `CodeGenRequest` and options to the generator. Compare deterministic output for model structure, collections, adapters, relations, eager loaders, transactions, and diagnostics.

Golden tests prove rendering stability, not compilability.

To accept an intentional output change:

```sh
go test ./tests/golden -update
```

Review the snapshot diff before committing it.

## Compile fixtures

Run sqlc-gen-go and sqlc-model against the same schema and query set, then compile the result.

The fixture matrix must include:

| Dimension | Cases |
|---|---|
| Identifiers | UUID, serial integer, bigint, application-generated |
| Nullability | pointers, pgtype wrappers, nullable custom types |
| Data | text, boolean, numeric, JSON, JSONB, byte arrays, enums, arrays, timestamps |
| Parameters | zero, one, multiple |
| Results | table rows, custom rows, joined rows |
| Commands | `:one`, `:many`, `:exec`, `:execrows` |
| Configuration | renames, overrides, aliases |
| Relations | belongs-to, has-one, has-many, many-to-many |
| Sessions | root, transaction, mismatched |

## PostgreSQL integration tests

Validate database-generated identifiers and timestamps, insert/update hydration, affected-row semantics, constraint translation, relation caching, eager loading, inverse hydration, commit, rollback, panic cleanup, and session mismatch.

## Race tests

Model instances are not concurrency-safe, but session and collection behavior should still be exercised under `go test -race` to detect accidental shared mutable generator/runtime state.

## Documentation tests

Compile every public Go example and validate SQL/configuration fixtures in CI. Documentation must not describe APIs that generated code does not provide.
