# Model lifecycle reference

## New and attached

Created through `models.Users.New()`.

```text
attached = true
exists = false
deleted = false
original = zero or construction baseline
dirty = assigned fields that differ from original
```

`Save` executes the configured insert query.

## Persisted and clean

Created through retrieval or successful persistence.

```text
attached = true
exists = true
deleted = false
original = current snapshot
dirty = empty
```

`Save` performs no query.

## Persisted and dirty

At least one current field differs from the original snapshot.

```text
attached = true
exists = true
deleted = false
dirty = one or more fields
```

`Save` executes the configured update query. The initial implementation sends the complete configured update parameter structure rather than dynamically generating partial SQL.

## Deleted

After a successful delete:

```text
attached = true
exists = false
deleted = true
dirty = empty
```

A second delete is idempotent. `Save` and `Refresh` return `ErrDeletedModel`.

## Detached

A model without a persistence session is detached. Persistence and lazy loading return `ErrDetachedModel`.

## Snapshot synchronization

After successful restore, insert, update, or refresh:

```text
original = current
dirty = empty
```

## Reverting changes

```go
original := user.OriginalName()
user.SetName("Temporary")
user.SetName(original)
```

The user becomes clean again because dirty state is derived from current-versus-original values, not from a permanent mutation flag.
