# Compatibility reference

## Initial baseline

The proposed first stable release targets:

```text
Go 1.25.0 module baseline
sqlc plugin protocol compatible with the selected sqlc 1.31 baseline
PostgreSQL
pgx/v5
sqlc configuration version 2
```

The exact supported sqlc version range must be pinned and tested before release.

## Recommended sqlc-gen-go configuration

```yaml
emit_interface: true
query_parameter_limit: 0
sql_package: pgx/v5
```

## Deferred platforms

`database/sql`, MySQL, and SQLite remain deferred until the internal query/type abstraction no longer assumes PostgreSQL and pgx representations.

## Type compatibility

Compile fixtures must cover UUIDs, integer identifiers, timestamps, booleans, text, numeric values, JSON, JSONB, byte arrays, arrays, enums, nullable values, custom overrides, and renamed fields.

Unsupported mappings produce explicit diagnostics. The generator must never emit speculative wrapper access such as `.Valid` unless the actual type contract supports it.

## Fixture matrix coverage

Generated-model test fixtures use explicit case names so failures identify the supported surface that broke:

| Dimension | Required cases |
|---|---|
| Identifier style | UUID, serial integer, bigint, application-generated |
| Nullability | pointers, pgtype wrappers, nullable custom types |
| Data type | text, boolean, numeric, JSON, JSONB, byte arrays, enums, arrays, timestamps |
| Parameter count | zero, one, multiple |
| Result shape | table rows, custom rows, joined rows |
| Query command | `:one`, `:many`, `:exec`, `:execrows` |
| Configuration | renames, overrides, aliases |
| Relation kind | belongs-to, has-one, has-many, many-to-many |
| Session state | root, transaction, mismatched |

Unsupported combinations must be listed in `tests/compile/matrix.md` with the diagnostic or reason that excludes them.

## Version coupling

The rich-model generator runs alongside sqlc-gen-go and targets the symbols and type behavior expected from a supported configuration. Compatibility tests execute generation and compilation against every supported sqlc release.
