# Glossary

## Bounded-context package

A Go package containing concrete model types that participate in the same bidirectional relation graph. Keeping these types in one package avoids import cycles and allows private hydration and inverse-cache coordination.

## Canonical relation

The default, unconstrained form of a relation. The initial implementation caches only this form. A scoped relation such as `Posts().Published().Limit(10)` is not canonical.

## Collection

The model-oriented entry point used to create and retrieve model instances, such as `models.Users.New()` or `models.Users.Find(ctx, id)`.

## Dirty field

A mutable field whose current value differs from the value in the model's original persisted snapshot.

## Eager loading

Loading related models for a collection of parent models through a configured batch query, avoiding one relation query per parent.

## Hydration

Applying a sqlc result to a model's private current state and then synchronizing its original snapshot and lifecycle flags.

## Identity map

A session-level cache that guarantees one object instance per database identity. The initial release does not provide an identity map; model instances are independent mutable snapshots.

## Lazy loading

Executing a relation query only when a terminal relation method, normally `Get(ctx)`, is called and no applicable cache is available.

## Model session

The runtime object that owns sqlc query access, model collections, session identity, transaction creation, lazy-loading policy, and error translation.

## Original snapshot

The generated value object containing the model values last known to be persisted. Dirty tracking compares current state against this snapshot.

## Relation builder

A typed, immutable-by-value configuration object returned by methods such as `user.Posts()`. Scope methods return modified copies. Terminal methods execute the configured sqlc query.

## Session identity

An internal token associated with a model session. A transaction session receives a distinct identity. Cross-session associations are rejected.

## Terminal operation

An operation that may perform database I/O and therefore accepts `context.Context` and returns an error, such as `Save`, `Delete`, `Refresh`, `Get`, `Reload`, `Attach`, or `Sync`.
