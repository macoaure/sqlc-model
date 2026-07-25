# Documentation index

This documentation follows Diátaxis. Each section answers a different kind of question and should be maintained according to that purpose.

## Tutorials: learning through completion

Tutorials are guided learning experiences. They provide a controlled path from an empty project to a working result. A tutorial should minimize optional branches and should not become an exhaustive API specification.

Start with:

1. [Generate the first rich model](tutorials/first-model.md)
2. [Add the first lazy relation](tutorials/first-relation.md)
3. [Run model operations in a transaction](tutorials/first-transaction.md)

## How-to guides: solving a concrete problem

How-to guides assume the reader understands the basic concepts. Each guide answers a specific operational question, such as configuring a relation scope or testing generated output.

See [How-to guides](how-to/index.md).

## Reference: exact contracts

Reference material is descriptive and normative. It defines configuration keys, generated APIs, lifecycle states, cache behavior, errors, supported query contracts, and compatibility boundaries.

See [Reference](reference/index.md).

## Explanation: understanding the design

Explanation material discusses why the architecture is shaped this way. It covers Active Record semantics, sqlc integration, package-cycle constraints, static query composition, session identity, and deliberately rejected alternatives.

See [Explanation](explanation/index.md).

## Project implementation

The implementation section is project governance rather than a Diátaxis quadrant. It defines phases, acceptance criteria, scope boundaries, risks, and release readiness.

See [Project implementation](project/index.md).

## Architectural statement

The project implements a statically generated, Eloquent-inspired Active Record layer for Go. It uses sqlc-generated queries as the exclusive persistence engine.

```text
Application code
        ↓
Generated rich models and relations
        ↓
Generated model adapters
        ↓
sqlc-generated Querier
        ↓
pgx
        ↓
PostgreSQL
```

The generator is not a dynamic ORM. Every terminal database operation resolves to a statically declared sqlc query.
