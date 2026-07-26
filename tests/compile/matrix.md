# Compile Fixture Matrix

## Covered

- `identifier-styles`: UUID, serial integer, bigint, application-generated identifiers.
- `type-matrix`: pointer nullability, pgtype wrapper nullability, nullable custom type, text, boolean, numeric, JSON, JSONB, byte arrays, enums, arrays, timestamps.
- `query-matrix`: zero, one, and multiple parameters; table, custom, and joined result rows; `:one`, `:many`, `:exec`, and `:execrows`.
- `config-matrix`: renames, overrides, aliases.
- `relation-session-matrix`: belongs-to, has-one, has-many, many-to-many, root session, transaction session, mismatched session.

## Explicit Exclusions

No unsupported matrix combinations are excluded yet. Add exclusions here with the expected diagnostic before removing a case from a fixture.
