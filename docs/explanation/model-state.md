# Explanation: model state and dirty tracking

A model is not a sqlc row. It contains current values, an original persisted snapshot, lifecycle flags, validation errors, session attachment, and relation caches.

## Why a boolean change map is insufficient

A setter may change a value and then restore the original:

```go
original := user.OriginalName()
user.SetName("Temporary")
user.SetName(original)
```

The final model is clean. A permanent `changed[name] = true` flag would be wrong.

## Snapshot model

Each generated model receives a comparable or field-wise snapshot type. Setters compare current and original values. The dirty set is updated as a derived optimization.

```text
current != original → dirty
current == original → clean
```

After successful restore, insert, update, or refresh, the current state becomes the new original state.

## Validation state

Validation errors use per-field replacement semantics. A valid correction removes the previous error. This supports fluent chains without making a model permanently invalid after one failed attempt.

## Independent snapshots

The first release has no identity map. Loading the same row twice may produce distinct objects with independent dirty state and caches. This must be documented rather than hidden behind implied object identity.

## Concurrency

Models are mutable and not safe for concurrent use. Adding internal locks would complicate hydration, relation loading, and cloning while encouraging unsafe object sharing. External synchronization is required when a model crosses goroutines.
