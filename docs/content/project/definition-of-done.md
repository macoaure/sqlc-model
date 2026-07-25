# Definition of done

The first stable release is ready only when all conditions below are satisfied.

## Identity and installation

- One canonical repository and Go module path exist.
- A clean project can install and invoke the plugin.
- Plugin and runtime compatibility is versioned.

## Generation architecture

- The plugin consumes sqlc's CodeGenRequest without reading generated Go files during the same pass.
- Generated models compile against actual sqlc-gen-go output.
- Output is deterministic and gofmt-compliant.
- Generation is atomic.

## Model lifecycle

- Insert and update hydrate database-returned values.
- Dirty state reflects current-versus-original values.
- Correcting a field removes its previous validation error.
- Handwritten model files survive repeated generation.
- Deleted and detached states are enforced.

## Relations

- Bidirectional relations compile without import cycles.
- Lazy loading has explicit cache semantics.
- Scoped results cannot corrupt canonical caches.
- Eager loading uses configured batch queries.
- Empty eager-loaded relations are marked loaded.
- Inverse caches are hydrated when the parent instance is known.

## Transactions

- Transaction sessions use sqlc `WithTx`.
- Session mismatches are rejected.
- Error callbacks roll back.
- Panic paths execute rollback cleanup.
- Models are never silently reattached.

## Quality

- The fixture matrix covers ordinary and adversarial schemas.
- PostgreSQL integration tests cover generated identifiers, timestamps, constraints, relations, and transactions.
- Unsupported contracts produce actionable diagnostics.
- Every public documentation example is compiled in CI.
