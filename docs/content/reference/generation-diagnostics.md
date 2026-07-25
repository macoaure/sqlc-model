# Generation diagnostics reference

## Diagnostic structure

```go
type Diagnostic struct {
    Severity Severity
    Path     string
    Context  string
    Model    string
    Relation string
    Query    string
    Message  string
    Hint     string
}
```

## Example

```text
error: contexts.content.models.User.operations.insert

query CreateUser uses :exec, but the default insert lifecycle requires
:one so the persisted row can hydrate generated fields.

Hint: add RETURNING ... and change the annotation to :one, or select an
explicit non-hydrating insert policy when that policy becomes supported.
```

## Required validation stages

1. Configuration schema validation.
2. Context and package validation.
3. Model and field mapping validation.
4. Query name and annotation validation.
5. Parameter mapping validation.
6. Result-hydration validation.
7. Relation graph validation.
8. Scope compatibility validation.
9. File/declaration collision validation.
10. Go formatting validation.

## Failure behavior

Any error-severity diagnostic prevents all output. Warnings may be emitted for discouraged but valid behavior, such as a delete query using `:exec` without affected-row detection.

## Determinism

Diagnostics are sorted by configuration path and stable identifiers: path, context, model, relation, then query. Repeated runs over unchanged input produce identical output.
