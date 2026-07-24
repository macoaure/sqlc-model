# Feature Specification: Fluent Behavior & Validation

**Feature Branch**: `001-fluent-behavior-validation`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Developers using the generator need to extend a generated model with custom, domain-oriented behavior without ever touching generated files, so their additions survive regeneration. They do this by adding chainable methods to a separate handwritten file in the same package as the model. Each method must use a pointer receiver and return the model pointer, so custom methods compose with generated setters and with each other in a single fluent chain (e.g. user.Rename(...).Activate().Save(ctx)). Custom methods may normalize input, validate values, update fields (preferably by composing generated setters, to keep dirty-state tracking centralized), mark dirty state, clear or set field-level errors, and invalidate local relation caches — but they must never perform I/O: no Save, no transaction, no direct query execution, no lazy relation loading, no session switching. I/O stays confined to terminal methods. Validation is expressed per field: the model stores validation errors in a per-field map so a new valid value replaces (not appends to) any prior error for that field, letting a developer correct a bad value later in the same chain and end up with a clean error state. Cross-field rules are expressed by overriding/extending a Validate() method that combines existing field errors with additional cross-field checks. Save(ctx) always runs validation first and aborts before issuing any database query if validation fails. Generated models expose inspection APIs — Err(), HasErrors(), FieldError(field), and an explicit ClearErrors() escape hatch that does not alter field values or dirty state — so calling code can surface failures without inspecting internals."

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
-->

### User Story 1 - Add custom domain behavior that survives regeneration (Priority: P1)

A developer wants to add a domain-oriented, chainable operation (e.g. "suspend this user") to a generated model. They place the new method in a handwritten file alongside — but separate from — the generated model file, so that re-running the generator never touches, overwrites, or deletes their addition.

**Why this priority**: This is the foundational value proposition of the capability: without a safe, regeneration-proof extension point, developers cannot layer custom behavior onto generated models at all, and every regeneration becomes a manual merge exercise.

**Independent Test**: Add a custom method to a model, regenerate the model layer, and confirm the custom method is still present, unmodified, and callable — with no manual reapplication step.

**Acceptance Scenarios**:

1. **Given** a generated model with no custom behavior, **When** a developer adds a chainable custom method in a separate handwritten file in the same package, **Then** the method is callable on the model and returns the model itself so it can be chained with other methods.
2. **Given** a model with existing custom methods, **When** the model layer is regenerated, **Then** all custom methods remain present, unchanged, and functional.
3. **Given** a custom method that composes generated field setters, **When** the custom method runs, **Then** the model's change/dirty tracking reflects the underlying field changes exactly as if the generated setters had been called directly.
4. **Given** a custom method, **When** it executes, **Then** it does not perform any database I/O (no persistence, no transaction, no query execution, no on-demand loading of related data, no session change) — I/O only happens through dedicated terminal operations such as save.

---

### User Story 2 - Correct an invalid field value within the same chain (Priority: P1)

A developer calls a chainable method with an invalid value, sees a field-level error, then calls a correcting method later in the same chain with a valid value. The correction clears the earlier error for that field so the model ends up in a clean, savable state without restarting the chain or manually clearing errors.

**Why this priority**: Field-level validation with correction semantics is what makes fluent chains usable for real input handling — without it, one bad value would permanently and silently invalidate the model, or errors would accumulate and misrepresent the model's true state.

**Independent Test**: Call a validating method with an invalid value, confirm a field error is recorded, then call the same (or another) validating method for that field with a valid value, and confirm the field's error is gone while unrelated errors are untouched.

**Acceptance Scenarios**:

1. **Given** a model with no errors, **When** a chainable method is called with an invalid value for a field, **Then** the model records an error associated with that specific field.
2. **Given** a model with a recorded error for a field, **When** a chainable method is called again with a valid value for that same field, **Then** the previously recorded error for that field is replaced with no error (not appended to).
3. **Given** a model with errors on two different fields, **When** the invalid value on one field is corrected, **Then** the error on the other, unrelated field remains recorded.
4. **Given** a chain that sets an invalid value and then a valid value for the same field in sequence, **When** the chain completes, **Then** the model's overall error state for that field is clean.

---

### User Story 3 - Validation blocks persistence and failures are inspectable (Priority: P2)

A developer calls the model's save operation on a model that currently has one or more field errors, or that violates a cross-field business rule. The save operation runs validation first, refuses to issue any database query, and returns an error the developer can inspect — overall or per field — using the model's public inspection methods.

**Why this priority**: Preventing invalid data from reaching the database, and giving developers a clear way to explain *why* a save was rejected, is what makes the validation mechanism trustworthy and actionable; without it, validation would be advisory rather than enforced.

**Independent Test**: Construct a model with a known field error (or a cross-field rule violation), call save, and confirm no database write occurs and the returned/inspectable error identifies the failure. Then clear the error via the explicit reset operation and confirm field values and change-tracking state are unaffected.

**Acceptance Scenarios**:

1. **Given** a model with at least one field error, **When** save is called, **Then** no database query is issued and an error is returned before any insert/update decision is made.
2. **Given** a model whose individual fields are each independently valid but which violates a cross-field rule (a rule spanning more than one field), **When** save is called, **Then** the save is rejected and the cross-field violation is reflected in the model's overall error.
3. **Given** a model with recorded errors, **When** a developer calls the inspection methods, **Then** they can determine whether any error exists, retrieve the overall error, and retrieve the error for one specific field, without needing to know how errors are stored internally.
4. **Given** a model with recorded errors, **When** the developer explicitly clears all errors, **Then** the model reports no errors afterward, while its field values and pending-change state remain exactly as they were before clearing.
5. **Given** a model with no field errors and no cross-field violations, **When** save is called, **Then** validation passes and the save proceeds to its normal insert/update behavior.

---

### Edge Cases

- What happens when a cross-field validation rule fails but none of the individual fields involved have their own field-level error? The overall validation result must still reflect the failure even though no single field is flagged.
- What happens when a developer clears all errors and then immediately calls save without correcting the underlying invalid data? The model has no recorded errors, so it is a valid design outcome that save may re-validate and may re-surface the same error(s) if a validating method is not re-invoked to update the field's error state; the explicit clear operation is documented as not re-validating on its own.
- What happens when two different custom methods set errors on the same field in sequence, each with a different message? The later call's outcome (error or clean) must be what the field reflects — no error list grows unbounded for a single field.
- What happens when a custom method needs data from a related record that has not yet been loaded? Because custom methods must not trigger on-demand loading of related data, this scenario is out of scope for a single fluent chain and must be handled by the developer loading the related data beforehand.
- How does the system behave if a developer accidentally places custom behavior inside a generated file instead of the separate handwritten file? The next regeneration is expected to overwrite that file, and any custom behavior placed there is lost — the safety guarantee applies only to code placed in the designated non-generated location.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The generated model layer MUST provide a designated, non-generated extension point (a file or location separate from generated output, sharing the model's namespace) where developers add custom behavior.
- **FR-002**: Regenerating the model layer MUST NOT modify, remove, or require reapplication of any custom behavior placed in the designated extension point.
- **FR-003**: Custom methods added at the extension point MUST be able to return the model instance itself, so a developer can chain multiple custom and/or generated methods in a single expression.
- **FR-004**: Custom methods MUST be able to compose generated field-setting operations and other custom methods within the same chain, and doing so MUST update the model's change/dirty-tracking state consistently with calling those generated operations directly.
- **FR-005**: Custom (non-terminal) methods MUST NOT perform database I/O of any kind: no persistence/save, no transaction management, no direct query execution, no on-demand loading of related data, and no session/connection switching. Such operations are reserved for designated terminal methods.
- **FR-006**: The model MUST record validation failures per field, such that setting a new value for a field replaces any previously recorded error for that same field rather than accumulating multiple errors for one field.
- **FR-007**: When a field's value is successfully validated after a prior failure, the model MUST clear the previously recorded error for that field.
- **FR-008**: The model MUST support cross-field validation rules — checks that depend on more than one field — evaluated in addition to, and combined with, per-field validation results.
- **FR-009**: The save/persist operation MUST evaluate all validation (per-field and cross-field) before deciding whether to insert or update, and MUST NOT issue any database query if validation fails.
- **FR-010**: The generated model MUST expose public inspection operations that let calling code determine: whether any error exists, the combined overall error, and the error (if any) for one specific field — without requiring knowledge of internal error storage.
- **FR-011**: The generated model MUST provide an explicit operation to clear all recorded validation errors, and this operation MUST NOT alter current field values or the model's change/dirty-tracking state.
- **FR-012**: Cross-field validation failures MUST be reflected in the model's overall/combined error even when no single field carries its own field-level error for that failure.

### Key Entities

- **Model**: The generated, record-representing object that holds current field values, tracks which fields have pending changes, and accumulates validation state; the target that both generated and custom behavior operate on.
- **Custom Behavior**: Developer-authored, chainable operations layered onto a model at a designated non-generated extension point; may read/normalize/validate input and mutate model state, but must not perform I/O.
- **Field Error**: A validation failure associated with exactly one field of a model, held with replacement (not accumulation) semantics — a new evaluation of that field supersedes its previous error.
- **Cross-Field Validation Rule**: A validation check that depends on more than one field's value and contributes to the model's overall/combined validation outcome independently of individual field errors.
- **Validation Outcome**: The combined result of all field errors and cross-field rules for a model at a point in time; consulted by the save/persist operation before any database query is issued.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of custom behavior added at the designated extension point remains present and functional after the model layer is regenerated, with zero manual reapplication steps required.
- **SC-002**: A developer can correct an invalid field value later in the same call chain and reach a state with zero recorded errors for that field, without restarting the chain or calling a separate error-clearing operation.
- **SC-003**: 100% of save attempts on a model with one or more outstanding validation errors (field-level or cross-field) result in zero database writes.
- **SC-004**: A developer can identify which field caused a save failure, and read a human-readable description of the failure, using only the model's public inspection operations — with no need to inspect internal error storage.
- **SC-005**: After clearing all errors via the explicit clear operation, 100% of a model's field values and change-tracking state remain identical to their pre-clear values, verified by comparison before and after.

## Assumptions

- Custom behavior is added in files that live alongside the generated model but are excluded from the generator's write/overwrite scope by file location or naming convention; the exact convention is a generator implementation detail outside this spec's scope.
- "Terminal" operations (the only place I/O is permitted) include, at minimum, the save/persist operation; other terminal operations (e.g. explicit reload, explicit relation loading) are assumed to exist but are not detailed by this spec.
- Cross-field validation is expressed as an extension/override of a single overall validation step that a developer can add to, rather than a plugin list of independent cross-field rule objects; this matches the simplest mechanism described in the source material and keeps ordering and combination behavior deterministic.
- Validation errors and field errors are scoped to a single model instance in memory and are not persisted or shared across requests/sessions.
- This spec covers the developer-facing contract for adding custom behavior and declaring/enforcing validation; it does not cover the internal code-generation mechanics (e.g. how the generator decides which files to (re)write).
