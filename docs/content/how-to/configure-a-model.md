# How to configure a model

## Objective

Map a sqlc row and a set of named sqlc queries to one lifecycle-aware model.

## Define the model entry

```yaml
contexts:
  - name: content
    package: content
    directory: content
    models:
      User:
        row: User
        operations:
          find: GetUser
          insert: CreateUser
          update: UpdateUser
          delete: DeleteUser
          refresh: GetUser
```

`row` identifies the canonical sqlc result shape used to hydrate the model. An operation can use a different result type only when the generator can prove that its fields map completely to the canonical model.

## Configure field policies

```yaml
fields:
  id:
    readable: true
    mutable: false
    generated: insert

  email:
    readable: true
    fillable: true
    mutable: true

  created_at:
    readable: true
    mutable: false
    generated: insert

  updated_at:
    readable: true
    mutable: false
    generated: save
```

Use these policies deliberately:

- `readable` generates a getter.
- `fillable` includes the field in new-model assignment APIs.
- `mutable` generates a public chainable setter.
- `generated: insert` means insert hydration must populate the field.
- `generated: save` means insert and update hydration may change the field.
- `immutable_after_insert` allows assignment while new but prevents mutation after persistence.
- `sensitive` excludes a value from diagnostic formatting.
- `version` marks an optimistic-concurrency field.

## Use explicit mappings when names differ

```yaml
fields:
  display_name:
    column: name
    row_field: Name
```

Do not depend on fuzzy name matching for production schemas. The generator should report an ambiguity rather than guessing.

## Validate operation contracts

The default operation policies are:

| Operation | Required query command | Required result |
|---|---|---|
| `find` | `:one` | Canonical model fields |
| `insert` | `:one` | Canonical persisted row |
| `update` | `:one` | Canonical persisted row |
| `delete` | `:execrows` preferred | Affected row count |
| `refresh` | `:one` | Canonical model fields |

Use `RETURNING` for insert and update. An `:exec` insert cannot populate generated identifiers or timestamps and is rejected by the default lifecycle policy.

## Regenerate and compile

```bash
sqlc generate
go test ./...
```

A successful generator run is not enough. The generated rich-model output must compile against the actual sqlc-gen-go output in a fixture or application build.
