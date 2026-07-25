# Compatibility reference

## Initial baseline

The proposed first stable release targets:

```text
Go 1.24 or later
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

## Version coupling

The rich-model generator runs alongside sqlc-gen-go and targets the symbols and type behavior expected from a supported configuration. Compatibility tests execute generation and compilation against every supported sqlc release.
