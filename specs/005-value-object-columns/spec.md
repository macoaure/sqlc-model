# Feature Specification: Value Object Field Mapping

**Feature Branch**: `001-value-object-columns`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Allow a developer to map a database column to a richer, domain-specific value object type on a generated model instead of exposing the raw primitive type that sqlc produces. The developer hand-writes the value object type in their own Go source (e.g. a struct with a private field, a parsing/validating constructor function that returns the value and an error, and an accessor method that returns the underlying primitive). In the model's field configuration, the developer declares a value_object mapping for the column, naming the value object type, the constructor function used to build it from the sqlc primitive, and the method used to convert it back to the sqlc primitive for persistence. Using this configuration, the generator emits hydration code that calls the constructor when loading a row, propagating any validation error with context identifying the model and field, and emits persistence code that calls the conversion method when writing query parameters. The generator must not silently assume every type mismatch between the model field and the sqlc column is a nullable wrapper needing .Valid-style handling — conversion is only produced when the types are directly assignable, match a known supported mapping, or are explicitly configured as a value object. Validation rules and other domain behavior stay entirely in developer-owned code; the generator only ever emits the conversion plumbing. This preserves type safety and expressive domain modeling on generated models while avoiding primitive obsession, without weakening or duplicating the underlying sqlc persistence layer."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Expose a domain type on a generated model field (Priority: P1)

A developer has a database column that holds a primitive value (e.g. a string) but represents a domain concept with its own rules (e.g. an email address). They want the generated model to expose that column as their own value object type instead of the raw primitive, so application code works with meaningful, self-validating types rather than loose strings or numbers.

**Why this priority**: This is the core capability of the feature. Without it, nothing else in this spec has a purpose — every other behavior (error propagation, safe-conversion guarding) exists to support this outcome.

**Independent Test**: Can be fully tested by defining a value object type with a constructor and accessor, declaring a value-object mapping for one model field, regenerating the model, and confirming the generated model field's type is the value object rather than the primitive.

**Acceptance Scenarios**:

1. **Given** a developer has written a value object type with a validating constructor and a primitive-returning accessor, **When** they declare a value-object mapping for a model field naming that type, constructor, and accessor, **Then** the generated model exposes the field using the value object type.
2. **Given** a generated model with a value-object-mapped field, **When** a row is loaded from the database, **Then** the generated code constructs the value object from the underlying primitive using the configured constructor before the model is usable.
3. **Given** a generated model with a value-object-mapped field holding a valid value object, **When** the model is persisted, **Then** the generated code converts the value object back to the underlying primitive using the configured accessor before it is written to the query parameters.

---

### User Story 2 - Get actionable errors when stored data fails domain validation (Priority: P2)

A developer needs to know, immediately and unambiguously, when a row in the database contains a value that fails the domain rules encoded in a value object's constructor (e.g. a malformed email address that was inserted outside the application). They want the failure to identify exactly which model and field caused it, not a generic or silent failure.

**Why this priority**: Without reliable error propagation, invalid data becomes a silent correctness bug or an unhelpful crash, undermining the trust developers place in the generated layer. This is essential to the feature being safe to use, but depends on User Story 1 existing first.

**Independent Test**: Can be fully tested by seeding a row with a value that fails a value object's constructor, loading it through the generated model, and confirming the resulting error message names the specific model and field.

**Acceptance Scenarios**:

1. **Given** a stored column value that fails the configured constructor's validation, **When** the generated model attempts to hydrate that row, **Then** hydration fails with an error that identifies both the model and the field involved.
2. **Given** a hydration failure caused by a value object constructor, **When** the underlying constructor error is inspected, **Then** the original validation error is preserved and reachable from the surfaced error (not replaced or discarded).

---

### User Story 3 - Avoid incorrect automatic conversions for unconfigured or mismatched fields (Priority: P3)

A developer wants confidence that the generator only performs a field conversion when it is safe to do so — either because the types already match, because the generator recognizes the conversion as a supported pattern, or because the developer explicitly configured a value object. They do not want the generator guessing that any mismatched type is a nullable wrapper and silently generating incorrect plumbing.

**Why this priority**: This is a safety/correctness guardrail rather than new capability — it prevents the feature from misfiring on fields the developer never intended to customize. It matters less on day one than being able to configure a value object at all, but it protects every model in the project once value objects are in use.

**Independent Test**: Can be fully tested by introducing a model field whose type does not match its column's underlying type, is not a recognized supported mapping, and has no value-object configuration, then confirming the generator does not produce a nullable-wrapper-style conversion for it (and instead reports the mismatch rather than guessing).

**Acceptance Scenarios**:

1. **Given** a model field type that differs from its underlying column type but is directly assignable, **When** the model is generated, **Then** the generator produces the assignment without requiring value-object configuration.
2. **Given** a model field type that differs from its underlying column type and matches a generator-recognized supported mapping, **When** the model is generated, **Then** the generator produces that recognized conversion.
3. **Given** a model field type that differs from its underlying column type, is not directly assignable, and is not a recognized supported mapping, **When** no value-object configuration is provided for that field, **Then** the generator does not fabricate a nullable-wrapper conversion and instead treats the field as unresolved.

---

### Edge Cases

- What happens when a value-object mapping names a constructor or accessor that does not exist or does not match the expected shape (input/output types)? The generator should report this as a configuration problem before producing generated output, rather than emitting code that silently misbehaves or fails to build.
- What happens when a stored value is missing entirely (e.g. the underlying column is nullable) but the value object has no concept of absence? This spec assumes non-nullable columns are the baseline case for value-object mapping; nullable-column handling in combination with value objects is out of scope here.
- How does the system handle two different fields on the same model each mapped to different value object types? Each mapping is independent and applies only to its own field.
- How does the system handle a developer changing a value object's constructor logic (e.g. tightening validation) after data already exists in the database? Existing rows that no longer satisfy the constructor will fail hydration with the model/field-identifying error described in User Story 2; this is expected behavior, not a defect.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow a developer to declare, per model field, a value-object mapping that names a value object type, a constructor used to build that type from the field's underlying persisted value, and an accessor used to convert the value object back to the underlying persisted value.
- **FR-002**: The system MUST allow the value object type, its constructor, and its accessor to be entirely developer-authored; the system does not require or generate the value object's own source code.
- **FR-003**: The system MUST cause a generated model field with a value-object mapping to expose the value object type to application code, not the underlying primitive type.
- **FR-004**: The system MUST invoke the configured constructor when a row is loaded, so that the generated model field is populated with the value object (not the raw primitive).
- **FR-005**: When the configured constructor reports a failure, the system MUST propagate that failure as part of hydration and MUST include context identifying which model and which field produced it.
- **FR-006**: The system MUST invoke the configured accessor to obtain the underlying persisted value from the value object when writing the field's value for persistence.
- **FR-007**: The system MUST NOT generate a field conversion unless the field's type is directly assignable to/from the underlying persisted type, matches a generator-recognized supported mapping, or has an explicit value-object mapping.
- **FR-008**: The system MUST NOT assume a nullable-wrapper conversion for a field solely because its type differs from the underlying persisted type; that treatment is reserved for types the generator specifically recognizes as such.
- **FR-009**: The system MUST keep all validation rules and other domain behavior for a value object in developer-owned source; generated code MUST contain only the calls needed to convert between the value object and the underlying persisted value.
- **FR-010**: The system MUST detect and report an invalid or unusable value-object mapping (e.g. a referenced constructor or accessor that cannot be used as configured) before producing generated output for that field.
- **FR-011**: The system MUST support independent value-object mappings for multiple fields on the same model without the mappings interfering with one another.

### Key Entities

- **Value-Object Field Mapping**: Per-field configuration that associates a model field with a value object type, the constructor used to hydrate it, and the accessor used to persist it. Owned by the developer, read by the generator.
- **Value Object Type**: A developer-authored type representing a domain concept (e.g. an email address) that wraps an underlying primitive value, enforces its own validation/normalization through its constructor, and exposes the primitive back out through an accessor.
- **Generated Model Field**: The attribute exposed on a generated model to application code; when a value-object mapping exists for it, its type is the value object rather than the underlying persisted primitive.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can introduce a new value-object-typed model field by writing only their own value object type and a field-level mapping, without modifying generator internals or the underlying persistence layer.
- **SC-002**: 100% of hydration attempts against stored values that fail a value object's constructor produce an error that identifies the specific model and field, with zero silent data corruption or unexplained panics.
- **SC-003**: Round-tripping a record through hydration and persistence preserves the original data for every value-object-mapped field, with no observed data loss or mutation introduced by the conversion plumbing.
- **SC-004**: Zero cases where the generator applies a nullable-wrapper-style conversion to a field type it does not recognize and that has no explicit value-object mapping.
- **SC-005**: A developer can locate and modify the validation or domain behavior of any value object entirely within their own source files, without ever needing to edit generated code.

## Assumptions

- The underlying sqlc-generated persistence layer and its column/parameter types are treated as fixed; value-object mapping is an additive layer on top of it and does not require changes to sqlc-generated code.
- A value-object mapping applies to exactly one model field mapped from exactly one underlying persisted value; mapping a single value object across multiple columns (composite value objects) is out of scope for this feature.
- The developer is fully responsible for authoring the value object type, its constructor, and its accessor; the generator's role is limited to invoking them correctly and wiring the resulting conversions into generated code.
- "Known supported mappings" (e.g. standard nullable-wrapper conversions) are a separate, existing capability of the generator's configuration surface; this feature does not redefine which conversions qualify as recognized, only how explicit value-object mappings coexist with them.
- Nullable underlying columns in combination with value-object mapping are treated as a distinct concern; this feature assumes the baseline case of a non-nullable, directly-hydratable column.
