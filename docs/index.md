---
template: landing.html
title: sqlc-model
hero_title: Fluent Active Record models, generated over sqlc.
hero_tagline: >-
  sqlc stays the source of truth for SQL, query signatures, parameter types, and driver
  integration. <code>sqlc-model</code> adds behavior, lifecycle, relations, validation, and
  persistence on top — statically generated, nothing dynamic, nothing improvised.
---

<div class="section section--boundary" markdown>

```text
The rich-model layer owns model behavior and lifecycle.

sqlc owns SQL contracts and database access.
```

</div>

<div class="section section--code" markdown>

## What it looks like

<div class="code-showcase" markdown>
<div class="code-showcase__item" markdown>

**Mutate a model**

```go
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

</div>
<div class="code-showcase__item" markdown>

**Traverse a relation**

```go
posts, err := user.
    Posts().
    Published().
    Latest().
    Limit(10).
    Get(ctx)
```

Calling `Posts()` or a scope method does not execute SQL. `Get(ctx)` is the operation boundary that may perform lazy loading.

</div>
</div>

</div>

<div class="section section--features" markdown>

## What it adds

<div class="feature-list" markdown>

- **Active Record–style models.** Generated over sqlc queries — not a dynamic ORM, every terminal operation resolves to a statically declared query.
- **Sessions and collections.** APIs for model construction, lookup, persistence, and transactions.
- **Dirty tracking and validation.** Original snapshots, lifecycle state, and validation errors on every model.
- **Typed relations.** Lazy loading, eager loading, scopes, cache inspection, and explicit terminal operations.
- **Value-object conversion hooks.** Map database columns to real domain types at the field level.
- **Deterministic generation.** Golden snapshot coverage keeps generated output stable across runs.

</div>

</div>

<div class="section section--install" markdown>

## Install

Install the sqlc process plugin:

```bash
go install github.com/macoaure/sqlc-model/cmd/sqlc-model@latest
```

Add it to `sqlc.yaml`:

```yaml
plugins:
  - name: richmodel
    process:
      cmd: sqlc-model

sql:
  - schema: schema.sql
    queries: query.sql
    engine: postgresql
    codegen:
      - plugin: richmodel
        out: internal/models
        options:
          version: 1
          sqlc:
            package: sqlcdb
            import: example.com/project/internal/database/sqlc
            driver: pgx/v5
          contexts:
            - name: content
              package: content
              directory: content
              models: {}
```

See the [configuration reference](content/reference/configuration.md) for the complete options contract.

</div>

<div class="section section--closing" markdown>

## Documentation

- [Tutorials](content/tutorials/index.md) teach the system through complete, guided examples.
- [How-to guides](content/how-to/index.md) solve specific implementation and application tasks.
- [Reference](content/reference/index.md) defines exact configuration, API, lifecycle, query, and compatibility contracts.
- [Explanation](content/explanation/index.md) records the architectural reasoning and rejected alternatives.
- [Project implementation](content/project/index.md) defines the delivery plan, release boundary, and definition of done.

Start with the [documentation index](content/index.md).

</div>
