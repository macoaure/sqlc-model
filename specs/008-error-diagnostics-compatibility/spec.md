# Feature Specification: Errors, Generation Diagnostics & Compatibility

**Feature Branch**: `001-error-diagnostics-compatibility`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "As a developer using the sqlc-model generator, I need a well-defined, predictable system of runtime errors, generation-time diagnostics, and compatibility guarantees so I can write reliable code against generated models and trust the generator to fail loudly and clearly rather than silently producing broken output. At runtime, generated code must raise a consistent taxonomy of typed errors I can detect and handle: framework-level state errors (detached model, deleted model, not-found, session mismatch, unsaved related model, lazy-loading prevented, invalid model state, unsupported query contract), structured field-level validation errors (identifying the model and field, using replacement semantics per field, with multiple concurrent errors combinable via Go's errors.Join), and database errors classified by kind (unique violation, foreign key violation, not-null violation, check violation, serialization failure, deadlock, or unknown) while preserving the original driver error through Unwrap so I can still inspect the underlying cause. When code generation runs, misconfiguration or invalid input must produce structured diagnostics identifying severity, the offending configuration path, context, model, relation, query, a human-readable message, and a remediation hint. Diagnostics must be produced deterministically (stable sort order) across a fixed sequence of validation stages (schema, context/package, model/field mapping, query/annotation, parameter mapping, result hydration, relation graph, scope compatibility, file/declaration collisions, Go formatting). Any error-severity diagnostic must block all output; warnings may still allow generation to succeed. I also need explicit, tested compatibility boundaries: which Go, sqlc, sqlc config, and PostgreSQL/pgx versions are supported, which platforms are explicitly out of scope for now, and which data types are guaranteed to compile and behave correctly."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Handling runtime failures predictably (Priority: P1)

A developer writing application code against generated models needs to detect and respond to specific failure conditions (a stale model, a not-found row, a constraint violation) using a consistent, documented set of error types rather than parsing driver-specific strings or guessing at behavior.

**Why this priority**: Runtime error handling is the most frequent touchpoint every developer using generated code will hit, on every request path. Without a predictable taxonomy, applications either crash ungracefully or silently mishandle failures, which undermines the entire premise of a "reliable" generated layer.

**Independent Test**: Can be fully tested by exercising each documented failure condition (e.g., saving a deleted model, violating a unique constraint, querying a related model across sessions) against generated code and confirming the returned error matches the documented category and can be distinguished programmatically from other categories.

**Acceptance Scenarios**:

1. **Given** a model instance that has been deleted, **When** the developer attempts to save it again, **Then** the operation returns an error identifiable as the deleted-model condition, distinct from all other error conditions.
2. **Given** a query that violates a database constraint (e.g., a duplicate unique key), **When** the query executes, **Then** the returned error is classified by constraint kind and still exposes the original underlying database error for inspection.
3. **Given** a model with multiple invalid fields, **When** validation runs, **Then** the developer receives a single combined error that identifies every invalid field and its associated model, without losing information about any individual field.
4. **Given** two model instances loaded from different sessions, **When** the developer attempts to use one as a related model of the other, **Then** the operation returns an error identifying the session mismatch rather than succeeding with inconsistent state or failing with an unrelated error.

---

### User Story 2 - Diagnosing generation failures (Priority: P2)

A developer configuring the generator (mapping queries to models, wiring relations, adjusting scopes) needs clear, actionable feedback when their configuration is invalid or incompatible, so they can fix the problem without reverse-engineering the generator's internals.

**Why this priority**: Generation-time feedback is what makes the tool usable during setup and ongoing schema evolution. Poor diagnostics turn every configuration change into a trial-and-error loop, which is the single biggest adoption barrier for a code-generation tool with many moving configuration parts.

**Independent Test**: Can be fully tested by intentionally introducing a known category of misconfiguration (e.g., an insert query missing the required `RETURNING` clause) and confirming the generator emits a diagnostic that names the exact configuration location, describes the problem, and suggests a concrete fix, while producing no generated output.

**Acceptance Scenarios**:

1. **Given** a query configured with an annotation that is incompatible with the model's lifecycle requirements, **When** generation runs, **Then** the generator reports a diagnostic naming the specific configuration path, model, and query involved, along with a remediation hint, and emits no output files.
2. **Given** a configuration containing both a blocking problem and a discouraged-but-valid pattern, **When** generation runs, **Then** the blocking problem is reported as an error that prevents all output, while the discouraged pattern is reported as a warning that does not prevent generation from succeeding.
3. **Given** an unchanged, invalid configuration, **When** generation is run multiple times, **Then** the reported diagnostics are identical in content and order every time.
4. **Given** a configuration with multiple independent problems spanning different validation concerns (e.g., a field mapping issue and a relation graph issue), **When** generation runs, **Then** diagnostics for all applicable concerns are reported together rather than stopping at the first one found.

---

### User Story 3 - Confirming compatibility before adopting the generator (Priority: P3)

A developer evaluating or upgrading the generator needs to know, without trial and error, which language runtime, query-compiler, database, and driver versions are supported, which platforms are intentionally out of scope, and which data types are guaranteed to work correctly.

**Why this priority**: Compatibility clarity prevents wasted evaluation effort and avoids subtle runtime breakage from unsupported combinations. It is lower frequency than daily error handling or generation diagnostics, but it gates whether a team can safely adopt or upgrade at all.

**Independent Test**: Can be fully tested by comparing a project's toolchain (language runtime version, query-compiler version, database engine, driver, and the data types used in its schema) against the published compatibility boundaries and confirming the result (supported or not) matches actual generation and compilation behavior.

**Acceptance Scenarios**:

1. **Given** a project using a supported combination of language runtime, query-compiler, database, and driver versions, **When** the developer checks the published compatibility information, **Then** the combination is clearly listed as supported.
2. **Given** a project targeting a platform that is explicitly deferred (e.g., a different database engine or driver), **When** the developer checks compatibility, **Then** the platform is clearly listed as unsupported rather than silently omitted.
3. **Given** a database schema using a data type outside the guaranteed-compatible set, **When** generation runs, **Then** the developer receives an explicit diagnostic about the unsupported type mapping instead of generated code that compiles but misbehaves at runtime.
4. **Given** a database schema using only guaranteed-compatible data types, **When** generation and compilation run, **Then** the generated code compiles and the resulting model code accesses the fields in a way that matches the actual type's contract (no speculative access to features the type does not provide).

---

### Edge Cases

- What happens when a developer combines multiple current validation errors on the same model into one Go error via joining, and later needs to test for one specific field's error among them?
- How does the system behave when a database error's kind cannot be determined from the driver response (falls into an "unknown" classification) — is the original error still fully inspectable?
- How does generation behave when a configuration produces only warning-level diagnostics and no errors — does output still get produced, and are the warnings still visible to the developer?
- What happens when a developer runs generation against a schema type that is nominally supported but has been renamed or given a custom override — is compatibility still honored?
- How does the system communicate when the installed sqlc version falls outside the tested/pinned range, versus merely being untested?
- What happens when two diagnostics apply to the exact same configuration path (e.g., both a field-mapping problem and a downstream hydration problem caused by it) — are both reported, or only the root cause?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a fixed set of distinguishable framework-level error conditions covering at minimum: a detached model, a deleted model, a not-found result, a session mismatch between related models, an unsaved related model lacking an identifier, lazy-loading prevented by strict mode, an invalid model state, and an unsupported query contract.
- **FR-002**: Each framework-level error condition MUST be distinguishable from every other condition through a documented, stable identity that application code can check against, independent of the error's message text.
- **FR-003**: The system MUST report field-level validation failures with enough structure to identify both the model and the specific field that failed, plus the underlying validation problem.
- **FR-004**: When a field fails validation more than once, only the most recent validation error for that field MUST be retained (replacement semantics) rather than accumulating duplicates for the same field.
- **FR-005**: When multiple different fields on the same model fail validation, the system MUST be able to combine their errors into a single returned error without losing the ability to inspect each individual field failure.
- **FR-006**: The system MUST classify database errors returned from persistence operations into a fixed set of kinds, at minimum: unique violation, foreign key violation, not-null violation, check violation, serialization failure, deadlock, and an explicit unknown/unclassified fallback.
- **FR-007**: A classified database error MUST retain access to the original underlying driver error so a developer can inspect implementation-specific detail beyond the classification.
- **FR-008**: The system MUST report a distinguishable error when a relation operation involves models belonging to different sessions.
- **FR-009**: The system MUST report a distinguishable error when a relation requires an identifier that is not yet available on a newly created, unsaved related model.
- **FR-010**: The system MUST report a distinguishable error when a relation would require an uncached, on-demand load while the caller has disabled that behavior (strict/no-lazy-loading mode).
- **FR-011**: When code generation encounters a misconfiguration or invalid input, the system MUST produce a diagnostic identifying at minimum: severity, the offending configuration location, the surrounding context, the affected model (when applicable), the affected relation (when applicable), the affected query (when applicable), a human-readable explanation, and a suggested remediation.
- **FR-012**: The system MUST validate generator input through a fixed, ordered sequence of concerns: configuration schema, context/package structure, model and field mapping, query name and annotation, parameter mapping, result hydration, relation graph, scope compatibility, file/declaration collisions, and generated code formatting.
- **FR-013**: Any diagnostic marked as an error MUST prevent the generator from emitting any output for that run.
- **FR-014**: Diagnostics marked as warnings MUST NOT prevent generation from completing and producing output, but MUST still be surfaced to the developer.
- **FR-015**: Given unchanged input, repeated generation runs MUST produce diagnostics in the same content and the same order every time.
- **FR-016**: The system MUST publish, and keep current, the supported combination of language runtime version, query-compiler version and protocol, query-compiler configuration version, and target database/driver for the current stable release line.
- **FR-017**: The system MUST publish a pinned and tested range of supported query-compiler (sqlc) versions prior to each stable release, rather than leaving the range open-ended.
- **FR-018**: The system MUST explicitly document platforms and database engines that are not yet supported, rather than leaving unsupported combinations undocumented.
- **FR-019**: The system MUST publish the set of data types guaranteed to generate correctly compiling, correctly behaving model code, covering at minimum: unique identifiers, integer identifiers, timestamps, booleans, text, numeric values, JSON, JSONB, byte arrays, arrays, enumerations, nullable values, custom type overrides, and renamed fields.
- **FR-020**: When a schema uses a data type mapping outside the guaranteed-compatible set, the system MUST produce an explicit diagnostic rather than generating code that compiles but behaves incorrectly at runtime.
- **FR-021**: Generated code MUST NOT access type-specific wrapper behavior (such as a nullability indicator) unless the actual resolved type contract supports it.
- **FR-022**: The system MUST verify, through repeatable compatibility testing, that generation and compilation succeed against every version within its currently supported query-compiler range.

### Key Entities

- **Framework Error Condition**: A distinguishable runtime failure category representing an invalid operation against a model's or relation's current state (e.g., detached, deleted, not-found, session-mismatch, unsaved-related, lazy-loading-prevented, invalid-state, unsupported-query-contract).
- **Validation Error**: A structured failure tied to one model and one field, describing why the field's value was rejected; multiple validation errors for different fields on the same model can be combined into a single reportable error.
- **Database Error**: A classified failure returned from a persistence operation, carrying a fixed "kind" (e.g., unique violation, foreign key violation) plus a reference back to the original underlying error.
- **Generation Diagnostic**: A structured piece of feedback from the generator describing a single problem or discouraged pattern found while processing configuration, including its severity, location, affected model/relation/query, explanation, and suggested fix.
- **Compatibility Boundary**: The published statement of which language runtime, query-compiler, query-compiler configuration, database engine, and driver versions are supported, and which platforms/data types are explicitly out of scope or guaranteed-compatible.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can determine the category of any runtime error produced by generated code (state error, validation error, or database error, and its specific kind within that category) using only the error's documented type, without needing to inspect internal implementation details, in 100% of documented failure conditions.
- **SC-002**: When a developer misconfigures the generator, they can locate and fix the problem after a single generation run, using only the reported diagnostic's location and remediation hint, without needing to consult external support.
- **SC-003**: Running generation twice against the same unchanged, invalid configuration produces byte-for-byte identical diagnostic output both times.
- **SC-004**: 100% of generation runs that produce at least one error-severity diagnostic result in zero output files being written or updated.
- **SC-005**: A developer can determine, before writing any code, whether their intended language runtime, query-compiler, database, driver, and schema data types are supported, purely from published compatibility documentation, without needing to run the generator to find out.
- **SC-006**: Across every version within the currently supported query-compiler range, generation and subsequent compilation of the generated code succeed without manual intervention.

## Assumptions

- "Developer" refers to an engineer integrating the generator into a Go project and writing application code against its generated output; this spec does not address end-users of the applications built with it.
- Error conditions are exposed as distinguishable, checkable values or types in the host language, consistent with idiomatic error-handling patterns for that language, rather than as opaque strings.
- The generator's compatibility scope for this initial stable release line is a single target database engine and driver combination; support for additional database engines or drivers is intentionally deferred and tracked separately.
- The exact pinned range of supported query-compiler versions is finalized and tested as part of release preparation rather than fixed at specification time; this spec requires that such a range be published and tested, not what the specific version numbers are.
- Diagnostics and compatibility documentation are considered part of the generator's contract with developers and are expected to be kept in sync with each release.
