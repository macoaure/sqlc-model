# Feature Specification: Static Query Composition & Contracts

**Feature Branch**: `001-query-composition`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Developers need a way to build queries against generated models using chained, statically-typed method calls instead of a dynamic runtime query builder. Starting from a model's collection entry point (e.g. `models.Users.Query()`), a developer chains typed scopes such as filters, ordering, pagination limits, and eager-load directives (e.g. `Active().WithPosts().Limit(50)`), then executes the chain with a context-aware terminal call (e.g. `Get(ctx)`). Each chainable method must correspond to a statically known, pre-declared sqlc query capability — a known parameter with a constant or typed argument, a configured query variant, or an eager-load plan. There is no free-form predicate, operator, or join composition at runtime; every valid chain is fully determined and validated at generation time, and any chain that cannot resolve to a configured query contract is rejected before code is generated (a build-time failure, not a runtime one), so developers always know exactly which declared SQL will execute. Each generated query and collection operation carries an explicit behavioral contract depending on its kind: fetching a single record returns a fully hydrated model or a not-found error; inserting or updating returns canonical persisted state (including generated identifiers, defaults, and trigger-modified fields) and synchronizes the model's tracked snapshot; deleting reports whether a row was actually affected when possible; refreshing re-hydrates a model from canonical fields. Collections also expose direct lookup methods (find by identifier, find by other configured keys) and a way to construct new, session-attached, not-yet-persisted models. The result of executing a composed query is a typed, iterable collection of hydrated models that developers can consume directly, with guarantees about what fields are populated and what errors mean."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Compose a statically validated query chain (Priority: P1)

A developer working against a generated model wants to retrieve a filtered, ordered, paginated set of records, optionally with related data loaded, without hand-writing new SQL or using a runtime query builder. They start from the model's collection entry point and chain a sequence of typed, named scope calls, then execute the chain to get results.

**Why this priority**: This is the core capability the feature exists to deliver — composing everyday read queries safely and ergonomically is the primary reason a developer reaches for this layer instead of writing raw SQL or a dynamic builder.

**Independent Test**: Can be fully tested by taking a model with at least one declared filter, one declared ordering, a pagination capability, and one declared relation, chaining all four together, executing the chain, and confirming the returned collection matches the expected filtered/ordered/paginated/related-data result — with no new SQL written by the developer.

**Acceptance Scenarios**:

1. **Given** a model whose collection exposes pre-declared filter, ordering, pagination, and eager-load scopes, **When** a developer chains those scopes together and executes the chain, **Then** the system returns a typed, iterable collection of fully hydrated model instances reflecting the composed filter, order, limit, and related data.
2. **Given** a developer attempts to chain a scope name that has no corresponding pre-declared query capability for that model, **When** the project is built/generated, **Then** generation fails with an error identifying the unresolved chain, and no application code is produced that would execute an undefined query.
3. **Given** a valid composed chain, **When** a developer inspects the chain's scope names, **Then** they can determine which declared query operation will execute without needing to read generated internals.

---

### User Story 2 - Rely on explicit, predictable operation contracts (Priority: P2)

A developer performing a single-record fetch, insert, update, delete, or refresh needs to know, without guessing, what the operation guarantees: whether the result is fully populated, how "not found" is signaled, and whether the in-memory model reflects true persisted state afterward.

**Why this priority**: Predictable contracts are what let developers write correct error-handling and state-management code around generated operations; without this, the query composition capability from Story 1 would be unsafe to build on.

**Independent Test**: Can be fully tested by performing one of each operation kind (find, insert, update, delete, refresh) against a test model and verifying the documented outcome (hydration completeness, error type, snapshot synchronization, affected-row reporting) matches the contract for that operation kind.

**Acceptance Scenarios**:

1. **Given** a single-record fetch by identifier, **When** no matching row exists, **Then** the system returns a distinct not-found error rather than a partially populated model or a generic error.
2. **Given** a single-record fetch by identifier, **When** a matching row exists, **Then** the returned model instance has every canonical field populated.
3. **Given** an insert or update operation, **When** it completes successfully, **Then** the affected model instance's tracked state is synchronized with the canonical persisted values, including any generated identifiers, defaults, or trigger-modified fields.
4. **Given** a delete operation, **When** it completes, **Then** the system reports whether a row was actually affected whenever the underlying capability supports that detection.
5. **Given** a refresh operation on an existing model instance, **When** it completes, **Then** the instance's fields reflect the current canonical persisted state.

---

### User Story 3 - Use collection-level shortcuts for common single-record needs (Priority: P3)

A developer who just needs one record — by primary identifier, by another pre-configured key, or a brand-new unsaved instance — wants a direct method rather than composing a full query chain for a one-record lookup or creation.

**Why this priority**: This is a convenience layer on top of Stories 1 and 2; useful for ergonomics and common cases, but the feature is still coherent without it if query composition and contracts exist.

**Independent Test**: Can be fully tested by calling a direct find-by-identifier method and a configured find-by-other-key method against seeded data, and by constructing a new unsaved instance, verifying each behaves per its contract independent of any query chain.

**Acceptance Scenarios**:

1. **Given** a collection for a model, **When** a developer calls its direct find-by-identifier method with a matching identifier, **Then** the matching model instance is returned; with a non-matching identifier, a not-found error is returned.
2. **Given** a model with an additional pre-configured lookup key, **When** a developer calls the corresponding named retrieval method, **Then** the system behaves per the same single-record fetch contract as identifier-based lookup.
3. **Given** a collection for a model, **When** a developer requests a new instance, **Then** the system returns a model instance that is associated with the current session but does not yet exist in the database.

---

### Edge Cases

- What happens when a developer tries to chain a scope that has no corresponding pre-declared query capability? The chain MUST fail to generate (a build-time failure), not fail silently or throw only at runtime.
- What happens when an insert or update's underlying declared query does not return enough fields to satisfy the operation's contract (e.g., missing generated/default fields)? Generation MUST fail rather than produce a model with an incomplete synchronized snapshot.
- What happens when a delete operation's underlying declared query cannot report affected-row count? The outcome MUST be treated as ambiguous per the documented fallback, not silently reported as "one row affected."
- What happens when a query chain combines pagination with eager-loading of a to-many relation? The composed result MUST still return a typed, iterable collection where each returned model's eagerly loaded relation data is complete for that page of parent records.
- What happens when a single-record fetch chain (e.g., by identifier) resolves to zero matching rows versus more than one matching row unexpectedly? Zero rows MUST yield the not-found error; contracts for fetch operations assume at most one row is possible by construction (e.g., unique key), so multi-row ambiguity is out of scope for this feature.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide, for each model, a single well-defined entry point from which developers begin composing a query.
- **FR-002**: The system MUST allow developers to chain multiple typed scope calls — covering filters, ordering, pagination limits, and eager-load directives — onto that entry point before executing the query.
- **FR-003**: The system MUST resolve every chainable scope to one of: a known query parameter with a constant value, a known query parameter with a typed argument, a configured query variant, or an eager-load plan.
- **FR-004**: The system MUST reject, at generation/build time rather than at runtime, any query chain that cannot be resolved to a statically configured query capability, so that invalid chains never reach compiled application code.
- **FR-005**: The system MUST NOT offer free-form/dynamic predicate, operator, or join composition (e.g., arbitrary "field, operator, value" construction) as part of query composition.
- **FR-006**: The system MUST execute a composed query chain only via an explicit, context-aware terminal call, and MUST return results as a typed, iterable collection of hydrated model instances.
- **FR-007**: The system MUST, for a single-record fetch, return either a fully hydrated model instance with every canonical field populated, or a distinct not-found error — never a partially populated model, and never conflate not-found with other error types.
- **FR-008**: The system MUST, for insert and update operations, return canonical persisted state — including generated identifiers, default values, and trigger-modified fields — and synchronize the affected model instance's tracked snapshot with that returned state.
- **FR-009**: The system MUST, for delete operations, report whether a row was actually affected whenever the underlying declared query capability allows it, and MUST treat the outcome as ambiguous (not falsely report success/failure) when that detection is not possible.
- **FR-010**: The system MUST, for refresh operations, re-hydrate a model instance's fields from canonical persisted state.
- **FR-011**: The system MUST allow developers to construct, from a model's collection, a new model instance that is associated with the current session but not yet persisted to the database.
- **FR-012**: The system MUST allow developers to perform direct single-record lookups (by identifier and by other pre-configured keys) from a model's collection as an alternative to composing a full query chain.
- **FR-013**: The system MUST validate every generated operation's contract — required result shape, required fields, and required command type — at generation time, and MUST fail generation for any operation that is missing, ambiguously mapped, or insufficient to satisfy its contract.

### Key Entities

- **Query Chain**: An ordered sequence of typed scope calls made against a model's collection entry point, terminated by an execution call; each valid chain resolves, at generation time, to exactly one declared query operation.
- **Scope**: A single typed, chainable method representing one pre-declared query capability (a filter, an ordering, a pagination limit, or an eager-load directive) available for a given model.
- **Collection**: The per-model entry point that exposes query composition, direct single-record lookups, and construction of new unpersisted model instances.
- **Model Instance**: A hydrated (or newly constructed, unpersisted) record produced or affected by a query or operation, carrying a tracked snapshot of its persisted state.
- **Operation Contract**: The documented guarantee — required populated fields, error semantics, hydration behavior, and snapshot-synchronization behavior — associated with each kind of generated operation (single-record fetch, insert, update, delete, refresh, related-data load).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can express common retrieval needs — filtering, ordering, pagination, and related-data loading — entirely through chained method calls, without writing new query logic, for any capability that has been pre-declared for a model.
- **SC-002**: 100% of query chains referencing an undeclared filter, ordering, or relation fail before the application can run, rather than failing or misbehaving at runtime.
- **SC-003**: Given only the names of the scopes used in a composed chain, a developer can correctly state which declared query operation will execute, without inspecting generated source code.
- **SC-004**: For single-record lookups, developers can distinguish "no matching record" from all other failure conditions 100% of the time using only the returned error, without additional checks.
- **SC-005**: After any successful insert or update, 100% of the fields on the returned/synchronized model instance match the actual persisted database state, including generated, default, and trigger-modified fields.
- **SC-006**: Developers new to the generator can correctly compose a filtered, sorted, paginated, related-data-loading query on their first attempt after reading the reference documentation, at a target first-attempt success rate of 90% or higher.

## Assumptions

- The set of scopes (filters, orderings, eager-load plans) available for a given model is determined by that project's own underlying query definitions and generator configuration; this feature covers how those pre-declared capabilities are composed and what guarantees they carry, not how they are authored or declared in the first place.
- "Statically-typed method calls" and "typed scopes" refer to generated, model-specific methods rather than a shared generic query DSL; each model's collection and query type exposes only the scopes meaningful for that model.
- Pagination behavior (e.g., limit-based paging) is bounded by whatever forms are supported by the underlying pre-declared query capabilities; this feature does not introduce new pagination mechanisms beyond composing existing ones.
- When a delete operation's underlying capability cannot report the number of affected rows, surfacing that as a documented limitation (ambiguous outcome) is acceptable behavior rather than a defect.
- Single-record fetch operations are assumed to be backed by queries that can return at most one row by construction (e.g., a unique key); handling of unexpected multi-row results for such operations is out of scope for this feature.
