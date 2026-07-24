# Risk register

## sqlc-gen-go compatibility drift

**Risk:** The rich-model generator targets symbols and type behavior produced by sqlc-gen-go, while both generators run in parallel.

**Mitigation:** Pin supported sqlc versions, require a narrow configuration baseline, preserve complete metadata in the IR, and compile fixtures against every supported release.

## Type-mapping divergence

**Risk:** Manual mappings differ from sqlc overrides, nullable policies, enums, arrays, JSON types, or driver representations.

**Mitigation:** Avoid speculative conversion, require explicit value-object conversion when necessary, and maintain a broad compile matrix.

## Relation graph cycles

**Risk:** Package-per-model or package-per-relation layouts create Go import cycles.

**Mitigation:** Keep concrete related models and public relations in one bounded-context package; use nested directories only for internal loaders.

## Hidden N+1 queries

**Risk:** Lazy relation access in loops produces excessive database calls.

**Mitigation:** Provide batch eager loading, strict lazy-loading modes, logging hooks, and documented execution semantics.

## Cache incoherence

**Risk:** Association changes or duplicate model instances leave stale relation caches.

**Mitigation:** Cache only canonical relations initially, invalidate local and known inverse caches, document independent snapshots, and defer identity-map guarantees.

## Transaction leakage

**Risk:** Models from a root session execute writes inside code that appears transactional.

**Mitigation:** Assign session identity, reject cross-session association, avoid implicit reattachment, and require transaction-session model creation/loading.

## Fluent error persistence

**Risk:** One invalid setter permanently poisons the model after a corrected value.

**Mitigation:** Store errors by field and clear them on valid correction.

## Generator surface explosion

**Risk:** Excessive generated files, methods, scopes, or traversal types make code unreadable and slow.

**Mitigation:** Generate only configured direct relations and scopes, use filenames for public organization, and avoid traversal-path packages.

## ORM scope creep

**Risk:** Dynamic SQL, cascade persistence, events, migrations, and identity maps expand the project into a full ORM.

**Mitigation:** Require explicit architectural proposals and query/lifecycle contracts before broadening the initial boundary.
