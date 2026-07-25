# Tasks: Value Object Field Mapping

**Input**: Design documents from `/specs/005-value-object-columns/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/value-object-field-api.md, quickstart.md

**Tests**: Included because the spec defines independent tests and the plan requires unit/golden/compile coverage.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare fixtures and shared test helpers without changing behavior.

- [X] T001 Add value-object fixture request/options helpers in tests/golden/value_object_fixtures_test.go
- [X] T002 [P] Add compile fixture module skeleton in tests/compile/user-value-object/go.mod
- [X] T003 [P] Add developer-owned Email value object fixture in tests/compile/user-value-object/content/email.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the shared data model needed by every story before rendering conversions.

- [X] T004 Add ValueObjectMapping config structs and strict JSON decoding in internal/config/field.go
- [X] T005 Add value_object schema fields to specs/001-model-generation-config/contracts/config.schema.json
- [X] T006 Add exposed/persisted Go type fields and value-object metadata to internal/mapping/resolve.go
- [X] T007 Thread exposed/persisted field metadata through plan.ResolvedField in internal/plan/build.go
- [X] T008 Update existing codegen helpers to use ExposedGoType where model APIs need exposed types in internal/codegen/model.go
- [X] T009 Update existing codegen helpers to use PersistedGoType where query scan/argument code needs sqlc types in internal/codegen/store.go

**Checkpoint**: Field metadata can represent value-object mappings, but no user story behavior is complete until story tasks pass.

---

## Phase 3: User Story 1 - Expose a domain type on a generated model field (Priority: P1) MVP

**Goal**: A configured field exposes the value object type, hydrates through the constructor, and persists through the accessor.

**Independent Test**: Define Email/NewEmail/String, configure email.value_object, regenerate, and confirm generated model APIs expose Email while persistence uses string.

### Tests for User Story 1

- [X] T010 [P] [US1] Add config decode tests for valid value_object and required keys in tests/unit/config_test.go
- [X] T011 [P] [US1] Add mapping/plan tests for exposed Email type plus persisted string type in tests/unit/plan_test.go
- [X] T012 [P] [US1] Add codegen assertions for Email getter/setter/original signatures in tests/unit/codegen_test.go
- [X] T013 [P] [US1] Add golden fixture expectations for constructor and accessor calls in tests/golden/golden_test.go
- [X] T014 [P] [US1] Add compile fixture usage proving User.Email returns Email in tests/compile/user-value-object/content/value_object_test.go

### Implementation for User Story 1

- [X] T015 [US1] Parse and validate value_object.type, value_object.constructor, and value_object.accessor in internal/config/field.go
- [X] T016 [US1] Resolve value-object fields as exposed type plus persisted type in internal/mapping/resolve.go
- [X] T017 [US1] Carry value-object constructor/accessor metadata into plan fields in internal/plan/build.go
- [X] T018 [US1] Render model record fields with exposed value-object types in internal/codegen/record.go
- [X] T019 [US1] Render collection Find hydration through constructors before returning models in internal/codegen/collection.go
- [X] T020 [US1] Render store query arguments through value-object accessors in internal/codegen/store.go
- [X] T021 [US1] Update generated model getters/setters/original methods to use exposed value-object types in internal/codegen/model.go
- [X] T022 [US1] Generate user-value-object fixture files from the golden fixture into tests/compile/user-value-object/content/

**Checkpoint**: User Story 1 is functional and independently testable with `go test ./tests/unit ./tests/golden` plus `go test ./...` inside tests/compile/user-value-object.

---

## Phase 4: User Story 2 - Get actionable errors when stored data fails domain validation (Priority: P2)

**Goal**: Constructor failures fail hydration with model/field context while preserving the original error.

**Independent Test**: Load a row whose email constructor rejects the stored string and assert the returned error includes User.Email and wraps the constructor error.

### Tests for User Story 2

- [X] T023 [P] [US2] Add codegen assertion for fmt.Errorf model/field wrapping around constructor failures in tests/unit/codegen_test.go
- [X] T024 [P] [US2] Add compile fixture test for hydration error context and errors.Is/errors.As preservation in tests/compile/user-value-object/content/hydration_error_test.go

### Implementation for User Story 2

- [X] T025 [US2] Add generated hydration helper that wraps constructor errors with model and field names in internal/codegen/collection.go
- [X] T026 [US2] Reuse the same hydration helper for insert/update/find/refresh returned rows in internal/codegen/store.go
- [X] T027 [US2] Add any required fmt/errors imports for generated hydration error wrapping in internal/codegen/collection.go

**Checkpoint**: User Story 2 is functional when invalid stored data fails before a model is returned and the original constructor error remains wrapped.

---

## Phase 5: User Story 3 - Avoid incorrect automatic conversions for unconfigured or mismatched fields (Priority: P3)

**Goal**: The generator only emits conversions for direct/known mappings or explicit value_object mappings.

**Independent Test**: Configure an unsupported custom type without value_object and confirm generation fails instead of emitting nullable-wrapper-style conversion.

### Tests for User Story 3

- [X] T028 [P] [US3] Add unit test for nullable-column value_object rejection in tests/unit/config_test.go
- [X] T029 [P] [US3] Add plan/mapping test that unconfigured custom exposed type mismatch returns an error diagnostic in tests/unit/plan_test.go
- [X] T030 [P] [US3] Add golden negative fixture for no fabricated nullable wrapper conversion in tests/golden/golden_test.go

### Implementation for User Story 3

- [X] T031 [US3] Reject value_object mappings on nullable columns during plan resolution in internal/plan/build.go
- [X] T032 [US3] Add unresolved type mismatch diagnostics for fields without direct, known, or value_object conversion in internal/plan/build.go
- [X] T033 [US3] Ensure store rendering never emits .Valid or nullable-wrapper conversion for unconfigured mismatches in internal/codegen/store.go

**Checkpoint**: User Story 3 blocks unsafe generated conversions without breaking direct and recognized existing mappings.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finish documentation, fixture regeneration, and full verification.

- [X] T034 [P] Update value-object quickstart expectations if generated API wording changed in specs/005-value-object-columns/quickstart.md
- [X] T035 [P] Update config documentation for value_object in docs/content/how-to/use-value-objects.md
- [X] T036 Run gofmt/go test verification for generator packages with go test ./...
- [X] T037 Run compile fixture verification inside tests/compile/user-value-object with go test ./...

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1 and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Phase 2 and is the MVP.
- **User Story 2 (Phase 4)**: Depends on User Story 1 hydration plumbing.
- **User Story 3 (Phase 5)**: Depends on Phase 2 and can run after or alongside User Story 2 once User Story 1 metadata exists.
- **Polish (Phase 6)**: Depends on selected user stories being complete.

### User Story Dependencies

- **US1 (P1)**: Required first because it creates the explicit conversion path.
- **US2 (P2)**: Requires US1 constructor hydration.
- **US3 (P3)**: Requires shared exposed/persisted metadata from Phase 2 and should be validated before release.

### Parallel Opportunities

- T002 and T003 can run in parallel.
- T010 through T014 can run in parallel after Phase 2.
- T023 and T024 can run in parallel after US1.
- T028 through T030 can run in parallel after Phase 2.
- T034 and T035 can run in parallel after implementation settles.

---

## Parallel Example: User Story 1

```bash
Task: "Add config decode tests for valid value_object and required keys in tests/unit/config_test.go"
Task: "Add mapping/plan tests for exposed Email type plus persisted string type in tests/unit/plan_test.go"
Task: "Add codegen assertions for Email getter/setter/original signatures in tests/unit/codegen_test.go"
Task: "Add golden fixture expectations for constructor and accessor calls in tests/golden/golden_test.go"
Task: "Add compile fixture usage proving User.Email returns Email in tests/compile/user-value-object/content/value_object_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 for User Story 1.
3. Validate with unit/golden tests and the user-value-object compile fixture.
4. Stop before error-polishing and guardrail hardening if only the MVP is needed.

### Incremental Delivery

1. Ship US1 to expose value objects and round-trip valid data.
2. Add US2 to make invalid stored data failures actionable.
3. Add US3 to close unsafe conversion gaps before release.
4. Run Phase 6 verification.

### Notes

- Keep conversion code explicit and local to existing config, plan, and codegen files.
- Do not add a registry, reflection, or source type-checking pass.
- Every task above uses the strict `- [X] T### [P?] [US?] ... path` checklist format.
