# Explanation: system architecture

## Product shape

The system is an Eloquent-inspired model layer generated over sqlc. It is intentionally neither a pure DDD entity generator nor a general-purpose ORM.

```text
Application
    ↓
Bounded-context model package
    ↓
Generated model stores and relation loaders
    ↓
sqlc-generated Querier
    ↓
pgx/v5
    ↓
PostgreSQL
```

## Parallel generation

During `sqlc generate`, sqlc-gen-go and sqlc-gen-richmodel consume representations derived from the same schema and query analysis.

```text
                           ┌── sqlc-gen-go ─────────→ sqlcdb
Schema and named queries ──┤
                           └── sqlc-gen-richmodel ──→ models
```

The rich-model plugin must not assume sqlc-generated `.go` files already exist during the same invocation. It produces code targeting the expected public contract of the configured sqlc-gen-go output.

## Responsibilities

sqlc owns SQL parsing, static analysis, parameter types, result shapes, driver integration, and transaction-bound query objects.

The rich-model layer owns lifecycle-aware model state, fluent mutation, validation, dirty tracking, hydration, collections, typed relations, relation caches, eager-loading plans, and persistence ergonomics.

## Public concepts

The public API is organized around:

```text
Session
Collection
Model
Relation
Typed scope
Terminal operation
```

Generated stores and loaders are internal implementation details.

## Explicit I/O boundary

Fluent setters and scope methods do not execute SQL. Methods that may perform I/O accept `context.Context` and return an error. This makes persistence visible even while retaining a class-like model experience.
