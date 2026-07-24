# Feature Specification: Model Generation & Configuration

**Feature Branch**: `001-model-generation-config`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Developers using this Go code generator need to configure which database tables become fully-featured Active-Record-style models, without hand-writing CRUD boilerplate or dirty-tracking logic. Configuration is declared per bounded context (a name, Go package, and output directory) inside the sqlc codegen plugin options. Within a context, each model maps a canonical sqlc row type to a set of lifecycle operations (find, insert, update, delete, refresh), each pointing to a named sqlc query; insert and update operations must return the full persisted row via RETURNING so generated/database-computed values (IDs, timestamps) flow back into the model. For each column, the developer declares a field policy: whether it is readable (getter), fillable (settable on a new model), mutable (chainable setter after creation), generated at insert or on every save, immutable after insert, sensitive (hidden from diagnostics), a version field for optimistic concurrency, or explicitly mapped to a differently-named column/row field when names diverge. From this configuration, the generator atomically produces, per context, a session type, field-identifier constants, per-model generated files (model struct, collection, internal store/record adapters) plus a single editable file per model that is created once and never overwritten, so developers can add handwritten domain methods alongside generated ones. The generated model exposes getters, chainable setters with automatic dirty tracking, original-value accessors, dirty/clean queries, field-level validation errors, lifecycle state (new/existing/deleted/attached), and terminal Save/Delete/Refresh methods that resolve to the configured sqlc queries. This gives developers Eloquent-like ergonomics while sqlc remains the sole, statically verified persistence engine — no dynamic queries, and misconfiguration (ambiguous mappings, missing RETURNING, unknown schema version) fails generation rather than emitting partial or incorrect output."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Turn a database table into a working Active Record model (Priority: P1)

A developer who already has a database table and a set of sqlc queries for it wants a working, ergonomic Go model — with getters, setters, dirty tracking, and save/delete/refresh behavior — without writing any of that logic by hand. They declare a bounded context and a model entry that names the canonical sqlc row and maps the find/insert/update/delete/refresh lifecycle operations to their named sqlc queries, then run generation.

**Why this priority**: This is the core value proposition of the generator. Without it, there is no product — every other configuration capability (field policies, handwritten extensions) only matters once a model can be generated and used at all.

**Independent Test**: Can be fully tested by configuring a single model against an existing table and its sqlc queries, running generation, and confirming the application can construct, save, retrieve, and delete an instance of that model using only generated code — with zero hand-written persistence logic.

**Acceptance Scenarios**:

1. **Given** a table with a corresponding sqlc row type and named `:one` queries for find/insert/update/refresh and an `:execrows` or `:one` query for delete, **When** the developer maps these in a model entry under a bounded context and runs generation, **Then** the generator produces a model whose constructor, save, delete, and refresh operations work end-to-end against the database without additional hand-written code.
2. **Given** an insert or update query that uses `RETURNING` to return the full persisted row, **When** a new model is saved, **Then** database-computed values (generated identifiers, default values, timestamps) are populated back onto the same model instance and its clean/original state is synchronized.
3. **Given** a model configuration that is regenerated after the sqlc queries or schema change in a compatible way, **Then** the generated model output is replaced atomically — the developer never observes a partially-updated or broken generated model layer.

---

### User Story 2 - Control each field's exposed API surface via policy (Priority: P2)

A developer wants fine-grained control over what a generated model exposes per column: which fields are readable, which can be set on a new instance, which can be changed after creation, which are populated only by the database, and which should never appear in diagnostic output. They express this as a declarative policy per field rather than writing per-field accessor code.

**Why this priority**: Field policy is what makes the generated model safe and idiomatic (e.g., preventing an identifier from being manually overwritten, or a password hash from leaking into logs) — it depends on model generation (Story 1) already working, so it is the natural next increment.

**Independent Test**: Can be fully tested by configuring a model with a mix of field policies (readable-only, fillable, mutable, generated-on-insert, sensitive) and confirming the generated model exposes exactly the getters/setters implied by each policy, and that generation fails clearly when a field mapping is ambiguous.

**Acceptance Scenarios**:

1. **Given** a field marked `readable`, **When** the model is generated, **Then** a getter for that field is produced; fields not marked `readable` have no getter.
2. **Given** a field marked `mutable` or `fillable`, **When** the model is generated, **Then** a chainable setter is produced that updates dirty state when the value differs from the original snapshot; fields marked neither receive no setter.
3. **Given** a field marked `generated: insert`, **When** a new model is saved for the first time, **Then** the field is populated from the database response rather than requiring the developer to supply it.
4. **Given** a database column name that does not match the sqlc result field name and no explicit mapping is provided, **When** generation runs, **Then** generation fails with an ambiguity error rather than guessing a mapping.
5. **Given** a field marked `sensitive`, **When** the model is formatted for diagnostics or logging, **Then** its value is excluded from that output.

---

### User Story 3 - Extend a generated model with handwritten behavior (Priority: P3)

A developer wants to add domain-specific methods (e.g., `Activate()`, `Rename()`) on top of a generated model without those additions being lost the next time the model is regenerated, and without editing generated files directly.

**Why this priority**: This is what makes the generated layer usable as a long-lived foundation rather than a one-shot scaffold; it matters once a model already exists and needs to evolve, so it naturally follows Stories 1 and 2.

**Independent Test**: Can be fully tested by generating a model, adding a handwritten method to the designated editable file, regenerating, and confirming the handwritten method still compiles and works while generated files reflect any configuration changes.

**Acceptance Scenarios**:

1. **Given** a model generated for the first time, **When** generation completes, **Then** exactly one editable file for that model is created if it does not already exist, and the developer can add methods to it.
2. **Given** an editable file that already exists with developer-added methods, **When** the model is regenerated after a configuration change, **Then** the editable file is left untouched while generated files are refreshed.
3. **Given** a developer needs to change field state from handwritten code, **When** they use the protected field-error helpers exposed for this purpose, **Then** they can set or clear field-level validation errors without bypassing the model's public setter/dirty-tracking behavior.

---

### Edge Cases

- What happens when an insert or update operation is mapped to a query that does not return the full persisted row (e.g., an `:exec` insert instead of `:one` with `RETURNING`)? Generation MUST fail rather than produce a model that silently omits database-generated values.
- What happens when the configuration declares an unsupported or unspecified schema `version`? Generation MUST fail rather than proceed with an ambiguous or best-guess interpretation.
- What happens when two fields, or a field and a column, cannot be unambiguously matched by name? Generation MUST fail with a diagnostic identifying the ambiguity rather than silently choosing a mapping.
- What happens when a partial configuration or mapping error occurs partway through generation? No generated files for that run MUST be emitted or left in a partially updated state — generation is all-or-nothing.
- What happens when a developer manually edits a generator-owned file? Those edits MUST be discarded the next time generation runs, since generator-owned files are always replaced atomically.
- What happens when a model or field is removed from configuration and generation is re-run? The generated output for that model/field MUST no longer be produced, without affecting the developer-owned editable file or other unrelated models.
- What happens when a model is configured without any `mutable` or `fillable` fields? The model MUST still generate successfully, simply exposing no setters.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST let a developer declare one or more named bounded contexts, each with a Go package name and an output directory, as the organizing unit for related generated models.
- **FR-002**: The system MUST let a developer declare, within a context, one model per canonical sqlc row/result type that the model will be hydrated from.
- **FR-003**: The system MUST let a developer map each of the model's lifecycle operations (find, insert, update, delete, refresh) to a distinct named sqlc query.
- **FR-004**: The system MUST require that insert and update operations map to queries that return the full persisted row, and MUST fail generation when this requirement is not met.
- **FR-005**: The system MUST let a developer declare, per field, whether it is readable, fillable, mutable, generated at insert, generated on every save, immutable after insert, sensitive, or an optimistic-concurrency version field.
- **FR-006**: The system MUST let a developer explicitly map a field to a differently-named database column and/or sqlc result field when automatic name matching would be ambiguous.
- **FR-007**: The system MUST fail generation, with a diagnostic identifying the ambiguity, when a field-to-column or field-to-result-field mapping cannot be determined unambiguously and no explicit mapping is provided.
- **FR-008**: The system MUST fail generation when the declared configuration schema version is missing or not supported, rather than proceeding with a guessed interpretation.
- **FR-009**: The system MUST generate, for each configured model, a getter for every field marked readable and no getter for fields not marked readable.
- **FR-010**: The system MUST generate, for each configured model, a chainable setter for every field marked mutable or fillable, and no setter otherwise; each setter MUST update the model's dirty state by comparing the new value against the field's original snapshot.
- **FR-011**: The system MUST generate, for each configured model, accessors for each field's original (last-persisted) value, plus operations to query whether specific fields or the whole model are dirty or clean, in the model's declared field order.
- **FR-012**: The system MUST generate, for each configured model, field-level and model-level validation error accessors, and lifecycle-state queries indicating whether an instance is new, persisted, deleted, or attached to a parent collection.
- **FR-013**: The system MUST generate, for each configured model, terminal Save, Delete, and Refresh operations that resolve to the configured sqlc queries, where Save validates first, inserts new instances, skips already-clean persisted instances, updates dirty persisted instances, and rejects deleted or detached instances.
- **FR-014**: The system MUST exclude any field marked sensitive from generated diagnostic/formatted output of the model.
- **FR-015**: The system MUST produce, for each configured model, exactly one developer-owned file that is created only if it does not already exist and is never overwritten by subsequent generation runs, so handwritten behavior persists across regeneration.
- **FR-016**: The system MUST treat all other generated output as fully owned by the generator, replacing it in full on every successful generation run.
- **FR-017**: The system MUST perform generation as an atomic, all-or-nothing operation: any configuration or validation failure MUST prevent any generated output for that run from being emitted or left partially updated.
- **FR-018**: The system MUST produce generated output (contexts, models, fields, operations, and files) in a deterministic order, so re-running generation against unchanged configuration produces identical output.
- **FR-019**: The system MUST expose protected, package-internal helpers for setting and clearing field-level validation errors so handwritten model extensions can integrate with the model's error/dirty-tracking state without bypassing its public setter behavior.

### Key Entities

- **Bounded Context**: A named grouping of related models that share a Go package and output directory; models connected to each other must live in the same context.
- **Model Configuration**: The mapping for one database entity — its canonical sqlc row/result type, its lifecycle-operation-to-query mapping, and its field policies — that the generator turns into a rich model.
- **Field Policy**: Per-column configuration describing what API surface a field gets (readable/fillable/mutable), how it is populated (generated at insert/on save, immutable after insert), how it is treated in diagnostics (sensitive), whether it participates in optimistic concurrency (version), and how it maps to the underlying column/result field when names diverge.
- **Lifecycle Operation**: One of find, insert, update, delete, or refresh — a named point in a model's life mapped to a specific sqlc query, with a required query result shape (e.g., insert/update must return the full row).
- **Generated Model**: The Go type produced from a model configuration, exposing getters, setters, dirty/original-value tracking, validation state, lifecycle state, and terminal persistence methods.
- **Developer-Owned Extension File**: The single per-model file the generator creates once and never modifies again, used to add handwritten domain behavior alongside generated methods.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can go from an existing table and its sqlc queries to a fully working, saveable/retrievable/deletable model by writing only declarative configuration — zero hand-written CRUD or dirty-tracking code.
- **SC-002**: Re-running generation against unchanged configuration produces byte-for-byte identical generated output, with no ordering-related differences.
- **SC-003**: 100% of handwritten additions to a model's developer-owned file survive any number of subsequent regenerations unchanged.
- **SC-004**: 100% of the identified misconfiguration cases (missing full-row return on insert/update, ambiguous field mapping, unsupported schema version) are caught and reported at generation time, with no generated output produced for that run.
- **SC-005**: A developer can determine a field's entire exposed API surface (getter presence, setter presence, diagnostic visibility) by reading its configuration alone, without inspecting generated code.
- **SC-006**: Changing a field's policy or an operation's target query and regenerating updates the model's exposed behavior without requiring any change to the developer-owned extension file.

## Assumptions

- Configuration targets a single relational database engine and driver combination per project in this initial scope (PostgreSQL accessed through the currently supported driver), consistent with the documented default.
- Only one configuration schema version is valid at a time; the generator is expected to be strict rather than best-effort about recognizing it.
- Automatic name matching between database columns and sqlc result fields is a convenience for the common case; production configurations are expected to use explicit mappings whenever names diverge, and the generator does not attempt fuzzy or partial matching.
- The default required query shape per lifecycle operation (e.g., `:one` for find/insert/update/refresh, row-count-returning for delete) is a fixed generator policy for this capability, not something each model configuration redefines.
- Defining relations between models (e.g., belongs-to/has-many associations, lazy/eager loaders) is a related but separate capability and is out of scope for this specification, which covers single-model configuration and generation only.
- The developer-owned extension file is scoped one-to-one with a model; it is created empty (or with a minimal package declaration) on first generation and is otherwise untouched by the generator.
- "Diagnostic output" for the `sensitive` policy refers to any generated human-readable formatting/string representation of the model (e.g., debug printing), not to values a developer explicitly reads via a getter and logs themselves.
