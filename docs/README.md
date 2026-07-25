# sqlc-gen-richmodel documentation

`sqlc-gen-richmodel` is an Eloquent-inspired rich model layer generated over sqlc. sqlc remains the source of truth for SQL, query signatures, parameter types, result types, and driver integration. The rich-model generator adds mutable model objects, lifecycle state, fluent behavior, relations, validation, dirty tracking, transactions, and model-oriented persistence APIs.

This archive contains the proposed product and implementation documentation organized with the Diátaxis framework.

## Documentation areas

- [Tutorials](docs/tutorials/index.md) teach the system through complete, guided examples.
- [How-to guides](docs/how-to/index.md) solve specific implementation and application tasks.
- [Reference](docs/reference/index.md) defines exact configuration, API, lifecycle, query, and compatibility contracts.
- [Explanation](docs/explanation/index.md) records the architectural reasoning and rejected alternatives.
- [Project implementation](docs/project/index.md) defines the delivery plan, release boundary, and definition of done.

Start with [Documentation index](docs/index.md).

## Project definition

> `sqlc-gen-richmodel` generates fluent Active Record models, typed relationships, lifecycle state, and persistence adapters over statically declared sqlc queries.

The public abstraction consists of sessions, collections, models, relations, typed scopes, and terminal persistence operations. The persistence foundation consists of explicit SQL, sqlc query analysis, sqlc-generated Go types, and transaction-bound sqlc query objects.

## Core boundary

```text
The rich-model layer owns model behavior and lifecycle.

sqlc owns SQL contracts and database access.
```

## Intended API

```go
models := content.New(pool)

user, err := models.Users.Find(ctx, userID)
if err != nil {
    return err
}

err = user.
    Rename("Marcos Aurelio").
    ChangeEmail("marcos@example.com").
    Activate().
    Save(ctx)
```

Relations use explicit terminal operations:

```go
posts, err := user.
    Posts().
    Published().
    Latest().
    Limit(10).
    Get(ctx)
```

Calling `Posts()` or a scope method does not execute SQL. `Get(ctx)` is the operation boundary that may perform lazy loading.
