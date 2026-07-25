# Errors reference

## Framework errors

```go
var ErrDetachedModel error
var ErrDeletedModel error
var ErrModelDetached error
var ErrModelDeleted error
var ErrModelNotPersisted error
var ErrNotFound error
var ErrSessionMismatch error
var ErrUnsavedRelated error
var ErrUnsavedRelatedModel error
var ErrLazyLoadingPrevented error
var ErrInvalidModelState error
var ErrUnsupportedQueryContract error
```

## Validation errors

```go
type ValidationError struct {
    Model string
    Field string
    Err   error
}
```

Field validation uses replacement semantics. Multiple current errors may be returned through `errors.Join`.

## Database errors

```go
type DatabaseErrorKind uint8

const (
    DatabaseErrorUnknown DatabaseErrorKind = iota
    DatabaseErrorUniqueViolation
    DatabaseErrorForeignKeyViolation
    DatabaseErrorNotNullViolation
    DatabaseErrorCheckViolation
    DatabaseErrorSerializationFailure
    DatabaseErrorDeadlock
)
```

Translated errors preserve the original driver error through `Unwrap`.

## Relation errors

- `ErrSessionMismatch`: models belong to different sessions.
- `ErrUnsavedRelatedModel`: a relation requires an identifier unavailable on a new model. `ErrUnsavedRelated` is a compatibility alias.
- `ErrLazyLoadingPrevented`: strict mode rejected an uncached relation query.

## Generation errors

Generation diagnostics identify severity, configuration path, context, model, relation, query, message, and remediation hint. Invalid contracts prevent output emission.
