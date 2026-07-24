# Feature Specification: Model Lifecycle & State Tracking

**Feature Branch**: `001-model-lifecycle-state`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Generated model instances move through a defined lifecycle so developers always know what a Save or Refresh call will actually do. A newly constructed model (via the generator's New() constructor) starts attached to a session but not yet persisted: it has no backing row, and any assigned fields are compared against a construction baseline to determine what is dirty. Calling Save on a new model runs the configured insert query and, on success, promotes it to persisted-and-clean, capturing current field values as the new \"original\" snapshot. Loading a model from the database also produces a persisted-and-clean instance. As field setters are called, the model becomes persisted-and-dirty whenever any current value differs from its original snapshot; Save then issues the configured update query, and on success the model becomes clean again. Deleting a persisted model transitions it to deleted, after which further Save or Refresh calls fail predictably rather than silently no-op or resurrect the row; a second delete is idempotent. A model with no persistence session attached is detached, and persistence or lazy-loading operations on it fail predictably. Crucially, dirtiness is derived by comparing current values to the original snapshot rather than tracked as a one-way \"was this field ever touched\" flag, so setting a field and then restoring its original value leaves the model clean again. This keeps saves minimal, avoids overwriting untouched fields, helps detect conflicts, and makes save behavior predictable for developers reasoning about model state."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Persist a newly created record (Priority: P1)

A developer constructs a new model instance, assigns field values, and saves it. The system must recognize the model has no backing row yet, perform a single insert, and afterward treat the model as an accurate reflection of the persisted row.

**Why this priority**: This is the entry point of the entire lifecycle — without a reliable new-to-persisted transition, no other state behavior matters. It is the first thing any developer using a generated model will do.

**Independent Test**: Can be fully tested by constructing a new model, setting fields, calling Save, and verifying exactly one insert occurs and the model subsequently reports no pending changes.

**Acceptance Scenarios**:

1. **Given** a newly constructed model with no backing row, **When** the developer sets several fields and calls Save, **Then** the system performs an insert and the model afterward reflects the saved values with no pending changes.
2. **Given** a newly constructed model that has just been saved, **When** the developer calls Save again without changing any fields, **Then** no additional write occurs.
3. **Given** a model retrieved by a query instead of constructed, **When** the developer inspects it, **Then** it already reflects the retrieved values with no pending changes, without requiring a Save call.

---

### User Story 2 - Save only when something actually changed (Priority: P1)

A developer loads an existing record, changes one or more fields, and saves it. The system must detect that the record has unsaved changes and update it; if nothing was changed, saving must do nothing.

**Why this priority**: Avoiding redundant writes and knowing in advance whether a save will touch the database is the core value proposition of this capability — it's what separates predictable save behavior from blind "always write" persistence.

**Independent Test**: Can be fully tested by loading a record, changing a field, saving, and confirming an update occurred; then saving again unchanged and confirming no update occurred.

**Acceptance Scenarios**:

1. **Given** a loaded record with no changes, **When** the developer calls Save, **Then** no write occurs.
2. **Given** a loaded record where the developer changes one field, **When** the developer calls Save, **Then** the system performs a single update and the model afterward reflects the new values with no pending changes.
3. **Given** a loaded record where the developer changes a field and then sets it back to its original value, **When** the developer calls Save, **Then** no write occurs, because the record is no longer considered changed.
4. **Given** a record with unsaved changes, **When** the developer asks what the original value of a changed field was, **Then** the system returns the value from before the change (not the current value).

---

### User Story 3 - Delete a record and get predictable behavior afterward (Priority: P2)

A developer deletes a record they previously loaded or saved. Afterward, any attempt to save, refresh, or delete it again must behave predictably rather than silently doing nothing unexpected or reviving data.

**Why this priority**: Deletion is a common but destructive operation; developers need certainty that a deleted record cannot be accidentally resurrected or silently re-saved, which matters for data integrity even though it's used less often than create/update.

**Independent Test**: Can be fully tested by deleting a loaded record, then attempting Save, Refresh, and a second Delete, and verifying each behaves as specified.

**Acceptance Scenarios**:

1. **Given** an existing persisted record, **When** the developer deletes it successfully, **Then** the model reports that it no longer exists and has no pending changes.
2. **Given** a record that was just deleted, **When** the developer calls Delete on it again, **Then** the second call completes without error.
3. **Given** a record that was just deleted, **When** the developer calls Save or Refresh on it, **Then** the system returns a clear error indicating the record is deleted, rather than performing a write or silently doing nothing.

---

### User Story 4 - Get a clear error instead of a confusing failure on a disconnected model (Priority: P3)

A developer holds a model instance that has no active persistence session attached (for example, one that was never connected or whose session ended). Attempting to save, refresh, or lazy-load related data on it must fail with a clear, identifiable error.

**Why this priority**: This is a guardrail against a class of programmer error; it protects correctness but is encountered far less often than the core save/update flows, so it is lower priority than the primary lifecycle paths.

**Independent Test**: Can be fully tested by creating a model instance with no session attached and verifying that persistence and lazy-load operations each return a distinct, identifiable error.

**Acceptance Scenarios**:

1. **Given** a model instance with no persistence session attached, **When** the developer calls Save, **Then** the system returns a clear error rather than attempting a write.
2. **Given** a model instance with no persistence session attached, **When** the developer attempts to lazily load related data, **Then** the system returns a clear error rather than attempting a query.

---

### Edge Cases

- What happens when a developer calls Save on a model that has just been constructed and never had any fields changed from their construction baseline? (No changed fields relative to the construction baseline still requires an insert, since the record does not yet exist — this differs from the "no write" case, which applies only to already-persisted records with no changes.)
- What happens when a developer changes a field, saves, changes it back to a different value, and saves again? Each save must only occur when there is an actual difference from the most recent original snapshot, so the second save both is triggered and updates the snapshot again.
- How does the system handle a developer reading the "original" value of a field that was never changed? It must return the current value, since original and current are identical when clean.
- What happens if the same underlying record is loaded twice into two separate model instances and only one is changed and saved? The two instances track state independently; the unchanged instance is unaffected by the other's save.
- How does the system handle an insert or update that fails (e.g., a database-level rejection)? The model must remain in its pre-save state (still new/dirty as applicable) so the developer can inspect, correct, and retry without having silently lost the unsaved changes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST represent a newly constructed model as not yet backed by a persisted row, distinct from a model that has been loaded from or successfully saved to storage.
- **FR-002**: System MUST compare each field's current value against a baseline value (the value at construction, or the value at last successful load/save) to determine whether that field has unsaved changes.
- **FR-003**: When a developer saves a model that has no backing row yet, the system MUST create the record and, on success, mark the model as reflecting the persisted state with no unsaved changes, updating the baseline to the current values.
- **FR-004**: When a model is produced by loading an existing record, the system MUST initialize it as reflecting the persisted state with no unsaved changes.
- **FR-005**: System MUST automatically consider an already-persisted model to have unsaved changes whenever any field's current value differs from its baseline, without requiring the developer to explicitly flag the change.
- **FR-006**: When a developer saves an already-persisted model that has unsaved changes, the system MUST update the record and, on success, mark the model as having no unsaved changes, updating the baseline to the current values.
- **FR-007**: When a developer saves a model that has no unsaved changes, the system MUST NOT perform any write to storage.
- **FR-008**: When a developer successfully deletes a persisted model, the system MUST mark it as no longer existing, with no unsaved changes.
- **FR-009**: A repeated delete call on an already-deleted model MUST complete without returning an error.
- **FR-010**: Save and Refresh calls on a deleted model MUST return a distinct, identifiable error rather than performing a write, silently doing nothing, or restoring the deleted record.
- **FR-011**: A model with no persistence session attached MUST return a distinct, identifiable error when the developer attempts a save, delete, refresh, or lazy-load operation on it, rather than attempting the operation.
- **FR-012**: System MUST determine unsaved-changes status by comparing current values to the baseline at the time of the check, not by recording a one-way "this field was modified" flag — so that changing a field and then restoring it to its baseline value results in no unsaved changes.
- **FR-013**: After any successful create, update, or refresh operation, the system MUST update the baseline to match the model's current field values.
- **FR-014**: Developers MUST be able to retrieve the pre-change (baseline) value of an individual field while the model has unsaved changes.
- **FR-015**: System MUST allow independent model instances loaded from the same underlying record to track their own unsaved-changes state independently of one another.

### Key Entities

- **Model Instance**: A generated, in-memory representation of one record. Holds current field values, a baseline snapshot of previously-persisted values, and lifecycle flags indicating whether it is attached to a session, whether it exists as a persisted row, and whether it has been deleted.
- **Baseline Snapshot ("original" values)**: The last known persisted (or construction-time) values for the model's fields, used as the comparison point for determining unsaved changes. Updated on every successful create, update, or refresh.
- **Unsaved-Changes State**: A derived condition — not a stored flag — indicating whether any field's current value differs from the baseline snapshot at the moment it is checked.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can predict, before calling Save, whether it will perform a write, with the model's reported change state matching actual save behavior in 100% of cases.
- **SC-002**: Saving a model that has no unsaved changes never results in a write to storage, eliminating redundant persistence operations in normal use.
- **SC-003**: Reverting a field to its baseline value before saving results in no write for that save call, in 100% of tested cases.
- **SC-004**: Attempting to save, refresh, or delete a record through a deleted or session-less model produces a clear, identifiable error in 100% of cases, with zero instances of silent no-ops or unintended data changes.
- **SC-005**: Repeated delete calls against the same already-deleted record complete successfully without error in 100% of cases.
- **SC-006**: After any successful create, update, or refresh, a developer immediately querying the model's change state finds it has no unsaved changes, in 100% of cases.

## Assumptions

- Each model instance is used by a single caller at a time; concurrent use of the same instance across multiple threads of execution is out of scope for this feature and is the caller's responsibility.
- There is no shared cache of records by identity in this feature's scope: loading the same underlying record more than once produces independent model instances, each with its own baseline snapshot and unsaved-changes state.
- When a save operation fails (e.g., the underlying write is rejected), the model retains its pre-save state — including its unsaved changes — so the developer can inspect and retry rather than losing the attempted changes.
- Calling Refresh on a model that has no backing row yet (never successfully saved) is not a supported operation, since there is no persisted record to refresh from; this is treated the same as any other operation attempted on a model without a valid persisted counterpart.
- The initial implementation updates a full set of configured fields on every write rather than generating a query that touches only the individual changed columns; this is a performance/implementation characteristic and does not change the developer-facing change-tracking behavior described above.
- Field-level validation error state is a related but distinct concern from lifecycle/dirty-state tracking and is out of scope for this feature's requirements.
