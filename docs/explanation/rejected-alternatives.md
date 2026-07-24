# Explanation: rejected alternatives

## Context-based repository lookup

Rejected because context becomes a service locator, dependencies become invisible, missing wiring fails at runtime, and transaction ownership becomes ambiguous.

Replacement: models retain a private session attachment.

## Public repository interfaces

Rejected because they undermine the intended `model.Save(ctx)` Active Record API and duplicate sqlc abstractions.

Replacement: public collections retrieve models; private stores adapt sqlc.

## Public restore or mark-persisted methods

Rejected because callers could forge lifecycle state and bypass invariants.

Replacement: hydration and snapshot synchronization remain private to the bounded-context package.

## Package per relation traversal

Rejected because of import cycles, unbounded traversal paths, duplicated concepts, ambiguous semantics, and inability to hydrate private target state.

Replacement: public relations share the model package; internal loaders may use nested directories.

## Automatic cascading persistence

Rejected for the first release because it requires hidden query ordering, implicit transactions, cycle detection, cascade policies, and difficult error recovery.

Replacement: reject unsaved related models without keys and require explicit saves.

## Dynamic SQL generation

Rejected because it would create a second query engine and weaken sqlc's static-analysis boundary.

Replacement: typed scopes map to predefined queries and parameters.

## Mandatory identity map

Deferred because it introduces memory retention, session lifetime, stale-object policy, transaction isolation, and concurrency concerns.

Replacement: document independent mutable snapshots and hydrate known inverse instances where possible.
