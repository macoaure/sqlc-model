# Initial release boundary

## Included

- PostgreSQL and pgx/v5.
- sqlc configuration version 2.
- One or more independent bounded-context packages.
- Single-column primary keys.
- UUID, integer, and application-assigned identifiers within tested mappings.
- Mutable models with private fields.
- Generated chainable setters.
- Handwritten domain behavior.
- Field validation state.
- Original snapshots and dirty tracking.
- Find, insert, update, delete, and refresh.
- Belongs-to, has-one, and has-many.
- Lazy loading and canonical relation caching.
- Batch eager loading and inverse hydration.
- Configured relation scopes.
- Transaction sessions and strict session identity.
- Explicit database-error translation.

## Deferred

- Identity map.
- Automatic cascade persistence.
- Dynamic SQL builder.
- Database migrations.
- Polymorphic relations.
- Cross-context concrete relation graphs.
- Composite primary keys.
- Transparent soft deletes.
- Model events and observers.
- Automatic timestamp conventions.
- General many-to-many sync without configured queries.
- Dynamic partial SQL updates.
- `database/sql`, MySQL, and SQLite.

## Scope rule

Deferral is not rejection. A feature moves into scope only after its query contract, lifecycle semantics, package implications, transaction behavior, and test matrix are defined.
