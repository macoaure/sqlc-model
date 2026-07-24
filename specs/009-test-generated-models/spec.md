# Feature Specification: Testing Generated Models

**Feature Branch**: `001-test-generated-models`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Developers who generate Active Record-style Go models from sqlc need confidence that the generated persistence code behaves correctly and that changes to the generator don't silently break existing model behavior. The project must let a developer test generated models at three complementary levels, since no single level is sufficient alone. First, generator golden tests let a developer feed a fixture code-generation request and options into the generator and compare the deterministic rendered output (model structs, collections, adapters, relations, eager loaders, transactions, and diagnostics) against known-good snapshots, proving the generator's output is stable even though this does not prove the output compiles. Second, compile fixtures let a developer run the sqlc and rich-model generators together against a shared schema and query set, then compile the result, exercising a fixture matrix spanning identifier styles, nullability, data types, parameter counts, result shapes, query command types, configuration overrides, relation kinds, and session states. Third, PostgreSQL integration tests let a developer validate real database behavior against generated models: database-generated identifiers and timestamps, insert/update hydration, affected-row semantics, constraint error translation, relation caching, eager loading, inverse hydration, transaction commit/rollback/panic cleanup, and session mismatch handling. Additionally, a developer must be able to run model, session, and collection code under Go's race detector to catch accidental shared mutable state, and CI must compile every public documentation example and validate its SQL/configuration fixtures so documentation never describes APIs the generated code doesn't actually provide."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Validate real database behavior of generated models (Priority: P1)

A developer who has changed the generator, the schema, or the query set needs to know that generated models still behave correctly against a real PostgreSQL database: identifiers and timestamps populate as expected, records hydrate correctly after insert/update, affected-row counts are accurate, database constraint violations translate into predictable errors, relations cache and eager-load correctly (including the inverse direction), and transactions commit, roll back, and clean up after a panic exactly as documented. This is the strongest, most direct evidence that the generated persistence layer is trustworthy.

**Why this priority**: This is the ultimate proof of correctness — everything else (golden tests, compile fixtures) is a cheaper proxy for this. Without it, a developer has no real assurance that generated code correctly talks to PostgreSQL.

**Independent Test**: Can be fully tested by pointing the test suite at a disposable, test-scoped PostgreSQL database, running the integration test suite against generated models built from a representative schema, and confirming every scenario (identifier/timestamp population, hydration, affected rows, constraint errors, relation caching/eager loading/inverse hydration, commit/rollback/panic cleanup, session mismatch) passes without needing the other test levels to run first.

**Acceptance Scenarios**:

1. **Given** a disposable PostgreSQL database and generated models built from a representative schema, **When** a developer inserts a new record through a generated model, **Then** the database-generated identifier and timestamp are hydrated back onto the model instance.
2. **Given** an open transaction, **When** the developer's code panics before committing, **Then** the transaction is rolled back and underlying resources are cleaned up automatically.
3. **Given** a model instance created under one session, **When** it is used with a mismatched session, **Then** the system reports a clear session-mismatch error instead of silently operating on the wrong connection/transaction.
4. **Given** a parent record with an eager-loaded has-many relation, **When** the developer loads the parent, **Then** the related records are populated and a subsequent inverse lookup from a child correctly resolves back to the cached parent.

---

### User Story 2 - Confirm generated code compiles across supported type and relation combinations (Priority: P2)

A developer needs assurance that generated code compiles correctly for every supported identifier style, nullability representation, data type, parameter count, result shape, query command type, configuration override, relation kind, and session state — not just the handful of cases exercised by a single sample schema.

**Why this priority**: Compilation failures are cheaper to catch than runtime failures but still require running both generators together against real schemas; this is a prerequisite for the integration tests in User Story 1 to even be meaningful, and it catches an entire class of breakage (type-mapping and code-generation bugs) that database tests alone would not isolate as clearly.

**Independent Test**: Can be fully tested by running the query generator and the rich-model generator together against a shared schema/query fixture set covering the full support matrix, then compiling the combined output — independent of whether any database is available.

**Acceptance Scenarios**:

1. **Given** a schema and query set covering every supported identifier style (UUID, serial integer, bigint, application-generated), **When** both generators run against it, **Then** the combined output compiles without error for each identifier style.
2. **Given** query fixtures covering `:one`, `:many`, `:exec`, and `:execrows` commands with zero, one, and multiple parameters, **When** the generators run, **Then** the generated code compiles for every command/parameter combination.
3. **Given** configuration fixtures using renames, overrides, and aliases, **When** the generators run with that configuration, **Then** the generated code reflects the configuration and still compiles.
4. **Given** fixtures covering belongs-to, has-one, has-many, and many-to-many relations, **When** the generators run, **Then** the generated relation code compiles for each relation kind.

---

### User Story 3 - Detect unintended changes to generated output (Priority: P3)

A developer changing the generator needs fast feedback on whether their change altered the rendered output (model structs, collections, adapters, relations, eager loaders, transactions, diagnostics) compared to the last known-good version, without waiting for a full compile-and-database cycle.

**Why this priority**: This is the fastest and cheapest signal, valuable for rapid iteration, but it only proves output stability, not correctness — a stable snapshot can still fail to compile or behave incorrectly against a real database, which is why it ranks below the other two levels.

**Independent Test**: Can be fully tested by feeding a fixture code-generation request into the generator and diffing the rendered output against a stored known-good snapshot, without compiling anything or touching a database.

**Acceptance Scenarios**:

1. **Given** a fixture code-generation request and a stored known-good snapshot, **When** the generator runs unchanged, **Then** the newly rendered output matches the snapshot exactly.
2. **Given** a generator change that alters rendered model, relation, or transaction code, **When** the golden test runs, **Then** the mismatch is reported clearly enough for the developer to identify what changed and decide whether to accept it as an intentional update or treat it as a regression.

---

### User Story 4 - Catch concurrency defects and documentation drift (Priority: P4)

A developer needs to know that generated session and collection code has no accidental shared mutable state that would break under concurrent use, and that every public documentation example still compiles and every documented SQL/configuration fixture remains valid, so documentation never claims behavior the generated code doesn't actually provide.

**Why this priority**: These are safety nets rather than primary correctness checks — concurrency bugs and documentation drift are real risks but narrower in scope than the behaviors covered by the other three levels.

**Independent Test**: Can be fully tested by running model, session, and collection tests with race detection enabled, and separately by compiling every public documentation code example and validating its referenced SQL/configuration fixtures in an automated CI step.

**Acceptance Scenarios**:

1. **Given** session and collection code exercised concurrently by tests, **When** the tests run with race detection enabled, **Then** any accidental shared mutable state is reported as a failure.
2. **Given** a documentation page containing a public code example, **When** CI runs, **Then** the example compiles successfully as part of the build.
3. **Given** a documentation page referencing a SQL or configuration fixture, **When** CI runs, **Then** the fixture is validated and any drift from the actual generated API causes CI to fail.

---

### Edge Cases

- What happens when a generator change is an intentional output improvement rather than a regression? The developer must be able to review the golden-test diff and explicitly accept a new snapshot as the new known-good baseline.
- How does the system handle a fixture-matrix combination that is not yet supported by the generator (e.g., a relation kind paired with an identifier style that has no defined behavior)? The matrix must either include an explicit expected outcome for it or explicitly exclude it with a documented reason, so untested combinations are never silently assumed to work.
- How does the system handle an integration test database that already has leftover data from a previous run? Each integration test run must use a clean, isolated, disposable database state so results aren't affected by prior runs.
- What happens when a documentation example references an API that was removed or renamed during a refactor? CI must fail the build rather than let the documentation continue describing a nonexistent API.
- What happens when a panic occurs partway through a multi-statement transaction? The rollback and cleanup behavior must be verified to leave no open transaction or leaked connection/resource behind.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The project MUST provide golden tests that feed a fixture code-generation request and generator options into the generator and capture the deterministic rendered output for models, collections, adapters, relations, eager loaders, transactions, and diagnostics.
- **FR-002**: The project MUST let a developer compare freshly rendered generator output against a stored known-good snapshot and clearly report any difference.
- **FR-003**: The project MUST provide a compile-fixture harness that runs the query generator and the rich-model generator together against a shared schema and query set, and then compiles the combined output.
- **FR-004**: The compile-fixture harness MUST exercise, at minimum, fixtures covering: identifier styles (UUID, serial integer, bigint, application-generated); nullability representations (pointers, wrapper types, nullable custom types); data types (text, boolean, numeric, JSON, JSONB, byte arrays, enums, arrays, timestamps); parameter counts (zero, one, multiple); result shapes (table rows, custom rows, joined rows); query command types (`:one`, `:many`, `:exec`, `:execrows`); configuration variations (renames, overrides, aliases); relation kinds (belongs-to, has-one, has-many, many-to-many); and session states (root, transaction, mismatched).
- **FR-005**: The project MUST provide integration tests that run against a real or test-scoped PostgreSQL database and validate: database-generated identifiers and timestamps, insert/update hydration, affected-row semantics, constraint-violation error translation, relation caching, eager loading, inverse relation hydration, transaction commit, transaction rollback, panic-triggered cleanup, and session mismatch handling.
- **FR-006**: Each PostgreSQL integration test run MUST operate against a clean, isolated, disposable database state so results are not affected by data left over from prior runs.
- **FR-007**: The project MUST allow a developer to run model, session, and collection tests with race detection enabled to surface accidental shared mutable state in generated or runtime code.
- **FR-008**: The project's CI MUST compile every public documentation code example as part of the build.
- **FR-009**: The project's CI MUST validate every SQL and configuration fixture referenced in documentation.
- **FR-010**: The project's CI MUST fail when documentation describes an API or behavior that the generated code does not actually provide.
- **FR-011**: A developer MUST be able to run each test level (golden, compile-fixture, integration, race, documentation) independently, without requiring the other levels to have run first, except that the integration tests depend on a successful compile of the generated fixtures.
- **FR-012**: When a test at any level fails, the failure report MUST identify which specific case failed (e.g., which fixture-matrix dimension, which golden-snapshot diff, which database assertion) so the developer can diagnose it without re-running the entire suite.
- **FR-013**: When a golden-test snapshot mismatch reflects an intentional output change, the project MUST allow a developer to explicitly update the stored snapshot to the new known-good output.

### Key Entities

- **Fixture Code-Generation Request**: The input (schema, query set, and generator options) fed into the generator to produce deterministic output for golden and compile-fixture testing.
- **Golden Snapshot**: The stored, known-good rendered output for a given fixture request, used as the comparison baseline for detecting unintended output changes.
- **Fixture Matrix Case**: One combination of identifier style, nullability, data type, parameter count, result shape, command type, configuration variation, relation kind, and session state used to exercise the compile-fixture harness.
- **Test-Scoped Database**: A disposable, isolated PostgreSQL instance/state used to run integration tests without affecting or being affected by other data.
- **Documentation Example**: A public code snippet, SQL fixture, or configuration fixture embedded in project documentation that must stay compilable/valid and consistent with the generated API.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Any change to the generator that alters rendered output for models, collections, adapters, relations, eager loaders, transactions, or diagnostics is caught by a golden-test diff before that change is merged.
- **SC-002**: Generated code compiles successfully for 100% of the defined fixture-matrix cases (identifier styles, nullability, data types, parameter counts, result shapes, command types, configuration variations, relation kinds, session states) before a generator or schema change is considered mergeable.
- **SC-003**: Every documented real-database behavior (identifier/timestamp population, hydration, affected-row semantics, constraint error translation, relation caching, eager loading, inverse hydration, transaction commit/rollback/panic cleanup, session mismatch) is verified automatically against a real PostgreSQL database with zero manual verification steps required per release.
- **SC-004**: Concurrency defects in session and collection code (accidental shared mutable state) are detected automatically by an existing, repeatable test step before a change reaches production, rather than discovered later through intermittent production failures.
- **SC-005**: 100% of public documentation code examples and referenced SQL/configuration fixtures pass automated CI validation; any documentation describing an API the generated code does not provide is caught before merge, not after a developer relying on the docs hits it.
- **SC-006**: A developer introducing a generator or schema change can determine pass/fail confidence across all test levels (output stability, compilability, real-database behavior, concurrency safety, documentation accuracy) from a single test run, without hand-writing ad hoc verification scripts.

## Assumptions

- A disposable or test-scoped PostgreSQL database instance is available to the developer (locally and in CI) for integration testing; production or shared databases are never used for this purpose.
- Golden snapshots are stored as versioned artifacts alongside the code and reviewed like any other code change, so intentional output changes go through normal review rather than being silently accepted.
- "Public documentation example" refers to any code, SQL, or configuration snippet in project documentation that describes how to use generated models, sessions, collections, or relations.
- The compile-fixture matrix dimensions listed in this spec represent the minimum required coverage; the project may extend the matrix over time as new generator capabilities are added.
- Race-detection testing targets the hand-written and generated session/collection/model runtime code, not the underlying database driver or PostgreSQL server itself.
