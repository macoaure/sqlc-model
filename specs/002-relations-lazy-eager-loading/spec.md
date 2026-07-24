# Feature Specification: Relations — Lazy and Eager Loading

**Feature Branch**: `001-relations-lazy-eager-loading`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Developers building on generated models need to express relationships between database tables -- belongs-to, has-many, has-one, and many-to-many -- and to read related data without hand-writing SQL for every access pattern. A relation is declared per model by naming its kind, target model, the local/foreign key pair (or pivot queries for many-to-many), an optional inverse relation for back-references, and the underlying queries that fetch one parent's related rows ("lazy") and many parents' related rows at once ("eager"). Belongs-to and has-one relations also support associating or dissociating a related row when the foreign key is nullable; many-to-many relations support attach/detach/sync operations backed by explicit pivot queries. A relation can be scoped: named, chainable constraints -- backed by declared query parameters, not dynamically generated SQL -- that narrow or order the related rows (e.g., only published posts, most recent first, limited to N). Scopes compose by value, leaving the parent model and its canonical (unconstrained) cache untouched. Loading is lazy by default: calling a relation performs no I/O until a terminal method like Get is called, giving local, chainable ergonomics. Because lazy access inside a loop over parent records causes one query per parent (N+1), eager loading lets a developer request related data for a whole result set in a single batch query, including nested relations, populating each parent's cache up front. A strict mode can flag or block uncached lazy access during development to catch accidental N+1 patterns without removing lazy loading as an option."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Declare a relation and read related records (Priority: P1)

A developer working on a generated model wants to navigate from one model to a related model (e.g., from a user to that user's posts, or from a post to its author) without hand-writing joins or lookup queries. They declare the relationship once — its kind, the target model, how parent and related records are matched, and which data-access operations retrieve one parent's related records versus many parents' at once — and then use a generated, chainable accessor to read the related data.

**Why this priority**: This is the foundational capability. Without the ability to declare and read a relation at all, scoping and eager loading have nothing to operate on. It delivers immediate value on its own: developers stop writing repetitive join/lookup code by hand.

**Independent Test**: Can be fully tested by declaring a single relation between two models, generating the related code, and calling the generated accessor to retrieve the related record(s) for a known parent — delivering correct related data with zero hand-written SQL.

**Acceptance Scenarios**:

1. **Given** a model with a declared one-to-many relation to another model, **When** a developer calls the generated relation accessor on a parent instance and requests the result, **Then** the related records for that parent are returned and no data access has occurred before the request.
2. **Given** a model with a declared many-to-one relation whose link is optional, **When** a developer associates a related record with a parent, **Then** the parent reflects the new association, and a disassociation operation is available because the link may be absent.
3. **Given** a model with a declared many-to-one relation whose link is required, **When** a developer inspects the generated relation, **Then** no disassociation operation is available, because the link can never be legitimately absent.
4. **Given** a model with a declared many-to-many relation backed by attach/detach/synchronize data-access operations, **When** a developer attaches, detaches, or synchronizes a set of related records, **Then** the association is updated accordingly and no association operation is invented beyond what the developer explicitly configured.
5. **Given** a related record that belongs to a different working context than the parent, **When** a developer attempts to associate, attach, or synchronize it, **Then** the operation is rejected rather than silently mixing state from two contexts.

---

### User Story 2 - Scope a relation to a filtered or ordered subset (Priority: P2)

A developer wants to read only part of a related set — for example, only published posts, the most recent ones, or a limited number of them — without writing a new hand-rolled query for every variation. They define named, reusable constraints on a relation once, and any developer can chain them onto the relation accessor to get a narrowed or reordered result, while the default (unconstrained) relation continues to behave and cache exactly as before.

**Why this priority**: Scoping multiplies the value of User Story 1 by making the same relation reusable across many read patterns, but it depends on relations already being declared and readable — hence P2.

**Independent Test**: Can be fully tested by defining one named scope on an existing relation, invoking it, and confirming the returned records satisfy the scope's constraint while an unscoped call on the same relation still returns the full default set.

**Acceptance Scenarios**:

1. **Given** a relation with a named scope that narrows results by a fixed condition, **When** a developer chains that scope onto the relation and reads the result, **Then** only records satisfying the condition are returned.
2. **Given** a relation with a named scope that accepts a developer-supplied value, **When** a developer chains the scope with different values on two separate reads, **Then** each read reflects its own value without affecting the other.
3. **Given** a scoped relation and the same relation's default (unscoped) form, **When** both are read from the same parent, **Then** applying the scope does not alter the parent model or the relation's default cached result.
4. **Given** a scope declaration that references an undeclared parameter, an incompatible value type, a name that collides with an existing relation operation, or a data source whose result shape does not match the relation, **When** the developer generates code for that configuration, **Then** generation fails with a diagnostic identifying the problem, rather than producing code that fails at run time.

---

### User Story 3 - Choose lazy or eager loading, and detect accidental repeated queries (Priority: P3)

A developer reading a relation for a single record is happy to let it load on demand. But when reading the same relation across a whole collection of parent records — for example, every user's posts on a listing page — they want to fetch all the related data in one batch instead of once per parent, and they want related data for nested relations to also be batchable in the same request. During development, they also want a way to be warned or stopped when a relation is read lazily without being pre-populated, so accidental repeated-query patterns are caught before they reach production.

**Why this priority**: This builds on both prior stories (a relation must exist and may be scoped) and addresses a performance/correctness concern that matters most once relations are used inside loops or listing views — a natural next step after the basics work.

**Independent Test**: Can be fully tested by reading a relation across a set of parent records once lazily (observing one data-access operation per parent) and once eagerly (observing exactly one batch data-access operation regardless of parent count), and by enabling the development safeguard and confirming an unpopulated lazy read is flagged or blocked while a pre-populated read is not.

**Acceptance Scenarios**:

1. **Given** a relation accessor that has not yet been read, **When** a developer merely constructs the accessor (without calling a terminal read operation), **Then** no data access has occurred.
2. **Given** a collection of parent records and a relation configured for batch retrieval, **When** a developer requests eager loading of that relation for the whole collection, **Then** the related data for all parents is retrieved in a single batch operation and each parent's default relation result is pre-populated, including parents with zero related records.
3. **Given** an eagerly loaded relation whose related records point back to the original parent, **When** a developer reads the inverse relation from a related record, **Then** it resolves to the already-loaded parent instance without an additional data access.
4. **Given** a relation configured with a nested relation, **When** a developer requests eager loading of the outer relation together with its nested relation, **Then** both levels are populated as part of the same eager-load request.
5. **Given** the development safeguard against uncached lazy access is enabled, **When** a developer reads a relation that has not been loaded or cached, **Then** the read is flagged or prevented; **When** the same relation has already been loaded (via prior read, cache, or eager loading), **Then** the read succeeds normally.
6. **Given** a relation whose default result has already been loaded, **When** a developer checks whether it is loaded without performing a fresh read, **Then** the check reports the loaded state and its cached value (if any) without any data access, and it is possible to explicitly force a fresh reload or explicitly clear the cached value.

---

### Edge Cases

- What happens when a developer tries to eager-load a relation for which no batch data-access source was configured? Generation must fail rather than silently falling back to per-parent access.
- What happens when a has-one relation's target record does not exist for a given parent? The relation must have a well-defined "no related record" outcome distinct from an error.
- What happens when a developer reads the same unscoped relation twice in a row? The second read must return the cached result without a repeated data access.
- What happens when a developer reads a scoped variant of a relation that has already been loaded in its default form? The scoped read must still perform its own data access rather than reusing the default cache, and must not overwrite the default cache.
- What happens when eager loading returns zero related records for a given parent? That parent's relation must be marked as loaded with an empty result, distinguishable from "not yet loaded."
- What happens when a developer tries to associate, attach, or synchronize a related record that has no usable identifier (e.g., not yet persisted)? The operation must be rejected in the initial release.
- What happens when a developer forgets/clears a relation's cache and then reads it again? The next read must perform a fresh data access, since the cache no longer holds a value.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Developers MUST be able to declare a relation from one model to another, specifying its kind (belongs-to-one, has-many, has-one, or many-to-many).
- **FR-002**: For belongs-to-one, has-many, and has-one relations, developers MUST specify how a parent record and its related record(s) are matched (the key each side uses to link records).
- **FR-003**: For many-to-many relations, developers MUST specify the data-access operations needed to read associated records and, when mutation is desired, to attach, detach, and fully synchronize the set of associated records.
- **FR-004**: Developers MUST be able to name an inverse relation so that traversing from a parent to a related record and back resolves to the same in-memory instance where already known, avoiding redundant reads.
- **FR-005**: Every relation MUST have a configured data-access operation that retrieves one parent's related record(s) ("lazy" access); relations that will be eager-loaded MUST additionally have a configured data-access operation that retrieves related records for many parents in a single batch ("eager" access).
- **FR-006**: For belongs-to-one and has-one relations, developers MUST be able to generate an operation that associates a related record with a parent.
- **FR-007**: A disassociation operation MUST be generated only for relations whose link between parent and related record is optional (may legitimately be absent); it MUST NOT be generated when the link is required.
- **FR-008**: For many-to-many relations, attach, detach, and synchronize operations MUST be generated only when the developer has explicitly configured the corresponding data-access operations; none of these operations may be invented without an explicit configuration.
- **FR-009**: Developers MUST be able to define one or more named scopes on a relation, each mapping to a specific query parameter paired with either a fixed value, a developer-supplied argument, or an alternate configured data-access operation.
- **FR-010**: Applying a scope MUST NOT mutate the parent model, the relation's default (unscoped) state, or any other scope invocation; scopes MUST be composable by value.
- **FR-011**: Generation MUST fail with a diagnostic (not a run-time failure) when a scope configuration references an undeclared parameter, supplies an incompatible value type, uses a name that conflicts with an existing relation operation, or selects a data source whose result is incompatible with the relation.
- **FR-012**: Relation access MUST be lazy by default: constructing or chaining a relation accessor MUST perform no data access until an explicit terminal read is requested.
- **FR-013**: Developers MUST be able to request eager loading of a relation for an entire collection of parent records, retrieving the related data in a single batch operation rather than one operation per parent.
- **FR-014**: Eager loading MUST support nested relations, so that a relation's own related data can be populated as part of the same eager-load request.
- **FR-015**: Eager loading MUST populate each parent's default relation cache, including an explicit empty result where no related records exist, and MUST populate the inverse relation's cache on related records with the already-known parent instance where an inverse is configured.
- **FR-016**: Developers MUST be able to check whether a relation's default result is already loaded, and retrieve its cached value if so, without triggering any data access.
- **FR-017**: Developers MUST be able to force a relation to discard its cached default result and perform a fresh data access.
- **FR-018**: Developers MUST be able to clear a relation's cached default result without performing any data access.
- **FR-019**: Only the default, unconstrained form of a relation is cached; reading a scoped/constrained variant MUST NOT overwrite the default cache.
- **FR-020**: Developers MUST be able to enable a development-time safeguard that flags or blocks a lazy read of a relation that has not yet been loaded or cached, without disabling lazy loading as a capability; reads of relations that are already loaded or cached MUST remain unaffected by this safeguard.
- **FR-021**: Association, attach, and synchronize operations MUST be rejected when the related record belongs to a different working context than the parent, or, in the initial release, when the related record has no usable identifier (e.g., is not yet persisted).

### Key Entities

- **Relation**: A named, typed link declared on a model, pointing to a target model. Carries its kind (belongs-to-one / has-many / has-one / many-to-many), how parent and related records are matched, an optional inverse relation name, and the data-access operations used for lazy reads, eager reads, and (where applicable) association/attach/detach/synchronize mutations.
- **Scope**: A named, reusable constraint attached to a relation, mapping to a query parameter with a fixed value, a developer-supplied argument, or an alternate data source. Scopes compose by value and never alter the relation's default cached state.
- **Relation cache (per parent instance)**: The default (unscoped) loaded state of a relation on a specific parent record — whether it has been loaded, and if so, its current value (which may be an empty result).
- **Eager-load request**: A batch request, issued against a collection of parent records, to populate one or more relations (optionally including their own nested relations) in a single data-access operation per relation level.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can declare a new relation and successfully read related records through the generated accessor using a single relation declaration, with zero hand-written join or lookup queries.
- **SC-002**: Reading the same relation across a collection of parent records via eager loading requires exactly one batch data-access operation per relation level, regardless of how many parent records are in the collection — compared to one operation per parent for lazy access.
- **SC-003**: 100% of invalid scope configurations (undeclared parameter, incompatible type, conflicting name, incompatible data source) are caught at generation time, before any generated code is executed.
- **SC-004**: A developer can determine whether a relation's default result is already loaded, and read its cached value, with zero data-access operations, every time.
- **SC-005**: With the development-time safeguard enabled, 100% of reads of unloaded, uncached relations are flagged or blocked, while reads of already-loaded or eager-loaded relations complete normally and are never blocked.
- **SC-006**: Applying a scope to a relation never changes the outcome of a subsequent unscoped read of the same relation on the same parent.

## Assumptions

- The underlying data-access operations referenced by a relation (lazy read, eager read, attach, detach, synchronize) are supplied by the developer as pre-existing, named queries; this feature does not generate ad hoc join or filter SQL on the developer's behalf.
- The initial release caches only the default, unconstrained result of a relation per parent instance; caching of arbitrary scoped/constrained variants is out of scope and may be considered in a future iteration.
- The initial release requires related records used in association, attach, or synchronize operations to already have a usable identifier; supporting not-yet-persisted related records in these operations is out of scope for this iteration.
- A has-one relation shares the same single-record load/cache/inspect behavior as a belongs-to-one relation, differing only in how the related record is located relative to the parent.
- "Working context" refers to the logical session or unit of work under which model instances are loaded and tracked; association-type operations across two different contexts are disallowed to prevent inconsistent or unintended state changes.
- Many-to-many relations always require an explicit join/pivot representation configured by the developer; the feature does not infer or synthesize pivot semantics automatically.
