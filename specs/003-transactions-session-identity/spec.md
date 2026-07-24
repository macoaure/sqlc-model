# Feature Specification: Transactions & Session Identity

**Feature Branch**: `001-transactions-session-identity`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Developers need a safe way to run multiple model operations as a single atomic database transaction. A developer opens a transaction by calling Transaction(ctx, callback) on their root model session; this begins a database transaction, immediately registers rollback-on-panic cleanup, and invokes the callback with a new transaction-scoped session whose model collections issue queries bound to that transaction. The callback's return value determines the outcome: returning nil commits the transaction, returning an error rolls it back and propagates the error to the caller, and a panic triggers the deferred rollback before being rethrown. Every session -- root or transaction-scoped -- carries a private, distinct identity established at construction. Any model instance created or loaded through a session permanently retains that session's identity; it is never silently reattached to a different session. This means models fetched before a transaction started remain bound to the root session and must not be reused inside the transaction callback -- all participating models must be created or loaded through the transaction session itself. When code attempts to associate two model instances that belong to different sessions (e.g., linking a transaction-scoped post to a root-session-loaded user), the operation is rejected with a session-mismatch error rather than silently succeeding. Associating a related model that has no persisted identifier yet is also rejected, since automatic cascade saves are out of scope. This identity guarantee prevents subtle bugs: operations silently executing outside their intended transaction, stale model state leaking across transaction boundaries, and hidden persistence-boundary violations that would otherwise be invisible. It lets developers reason confidently about exactly which models participate in which transaction."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run multiple model operations as one atomic unit of work (Priority: P1)

A developer needs to make several related changes (e.g., creating a record and updating another) that must all succeed together or not happen at all. They wrap the operations in a transaction so the database only reflects the change if every step completes successfully.

**Why this priority**: This is the core value of the feature — without atomic, all-or-nothing execution, there is no transaction capability at all. Every other guarantee in this spec exists to protect this behavior.

**Independent Test**: Can be fully tested by opening a transaction, performing two or more model saves inside the callback, and verifying: (a) when the callback succeeds, all changes are visible afterward; (b) when the callback returns an error partway through, none of the changes are visible afterward.

**Acceptance Scenarios**:

1. **Given** a developer has an active model session, **When** they open a transaction and the callback performs multiple model changes and returns without error, **Then** all of those changes are committed together and are visible once the transaction call returns.
2. **Given** a transaction is in progress and some changes have already been made inside the callback, **When** the callback returns an error, **Then** none of the changes made inside the callback are persisted, and the error is returned to the code that opened the transaction.
3. **Given** a transaction is in progress, **When** the callback panics after making some changes, **Then** none of those changes are persisted, the transaction is cleanly closed, and the panic continues to propagate to the caller.

---

### User Story 2 - Be protected from accidentally mixing models across transaction boundaries (Priority: P2)

A developer loads or creates model instances while working with a session (either the root session or a transaction's session). They need confidence that a model tied to one session cannot be silently used as if it belonged to another, so they cannot accidentally write outside the transaction they intended, or use half-committed data.

**Why this priority**: This guarantee is what makes the atomic transaction from User Story 1 trustworthy in real code. Without it, a developer could accidentally believe an operation is protected by a transaction when it is not, silently defeating the atomicity guarantee.

**Independent Test**: Can be fully tested by loading a model from the root session, attempting to relate it to a model created inside a transaction callback, and confirming the operation is rejected rather than silently succeeding.

**Acceptance Scenarios**:

1. **Given** a model instance was loaded or created through a particular session, **When** that instance is used for further operations, **Then** it continues to operate against that same session for the rest of its lifetime, and is never silently switched to a different session.
2. **Given** two model instances that belong to different sessions (for example, one loaded before a transaction started and one created inside the transaction callback), **When** a developer attempts to associate them with one another, **Then** the operation is rejected and does not modify either model or the database.
3. **Given** a developer is working inside a transaction callback, **When** they load or create all the models they intend to use directly through the transaction's session, **Then** those models can be freely associated and saved together within that transaction.

---

### User Story 3 - Get a clear, distinguishable error when a session or association rule is violated (Priority: P3)

A developer who accidentally violates a session rule (mixing sessions, or associating an unsaved related model) needs to be able to detect and handle that specific condition in code, rather than have it fail in a generic or ambiguous way.

**Why this priority**: This improves the safety and debuggability of the guarantees from User Story 2, but the core protection (rejecting the unsafe operation) already exists without it. Distinguishable errors are a refinement that makes the feature practical to build reliable applications on top of.

**Independent Test**: Can be fully tested by independently triggering a cross-session association and an association with an unsaved related model, and confirming each produces its own identifiable outcome.

**Acceptance Scenarios**:

1. **Given** two models that belong to different sessions, **When** a developer attempts to associate them, **Then** the resulting failure is identifiable as a session-mismatch condition, distinct from other failure types.
2. **Given** a newly created related model that has not yet been saved and has no persisted identifier, **When** a developer attempts to associate it with another model, **Then** the resulting failure is identifiable as an unsaved-related-model condition, distinct from a session-mismatch condition.

---

### Edge Cases

- What happens when a transaction callback makes valid changes but then returns an error anyway? All changes made during the callback are rolled back; none are persisted, regardless of how many succeeded before the error was returned.
- What happens when a panic occurs after some operations inside the callback already succeeded against the database? The already-registered rollback cleanup still runs, discarding all changes, before the panic continues to propagate.
- What happens if a developer tries to reuse a model loaded before a transaction started, inside that transaction's callback? The model remains tied to its original session; it is not treated as part of the transaction, and any attempt to associate it with a model that belongs to the transaction session is rejected.
- What happens if a developer tries to associate a related model that was created but never saved? The association is rejected because the related model has no persisted identifier for the relationship to reference.
- What happens to a transaction-scoped session and its models once the transaction callback has returned (committed or rolled back)? They are considered closed; continued use outside the callback is not a supported pattern.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow a developer to run a group of model operations as a single atomic database transaction by supplying a callback function.
- **FR-002**: The system MUST begin the underlying database transaction before the developer-supplied callback executes.
- **FR-003**: The system MUST register the transaction's rollback cleanup before the callback executes, so that any abnormal termination of the callback cannot leave the transaction open.
- **FR-004**: The system MUST supply the callback with a session dedicated to that transaction, through which model operations are executed within the transaction's scope.
- **FR-005**: The system MUST commit the transaction only when the callback completes and returns no error.
- **FR-006**: The system MUST roll back the transaction and return the callback's error to the caller when the callback returns a non-nil error.
- **FR-007**: The system MUST roll back the transaction when the callback panics, and MUST allow the panic to continue propagating after rollback completes.
- **FR-008**: The system MUST assign every session — whether the root session or a transaction's session — a distinct identity at the moment it is created.
- **FR-009**: The system MUST permanently bind every model instance to the identity of the session that created or loaded it, for the lifetime of that instance.
- **FR-010**: The system MUST NOT silently change which session a model instance is bound to.
- **FR-011**: The system MUST reject an attempt to associate two model instances that belong to different sessions, and MUST NOT allow the association to silently succeed.
- **FR-012**: The system MUST reject an attempt to associate a related model instance that has no persisted identifier, and MUST NOT allow the association to silently succeed.
- **FR-013**: The system MUST make the rejection reasons in FR-011 and FR-012 distinguishable from one another, so a developer can tell which rule was violated.
- **FR-014**: The system MUST allow a developer to load or create model instances directly through a transaction's session, so that all models intended to participate in the transaction can share its identity.

### Key Entities

- **Session**: A bound context for model work that groups related model collections, has a distinct identity, and determines which transaction (if any) operations run against. A root session exists outside any transaction; a transaction session exists only for the duration of one transaction callback.
- **Model instance**: A concrete, individually addressable record-backed object. It is bound to exactly one session's identity from the moment it is created or loaded, for its entire lifetime.
- **Transaction**: A callback-scoped unit of work with its own session identity. It commits when its callback succeeds and rolls back when the callback errors or panics.
- **Association**: An operation that links two model instances together (e.g., relating a child record to a parent record). Its validity depends on both models sharing the same session identity, and on the related model already having a persisted identifier.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: When a multi-step operation fails partway through, zero partial changes are observed afterward, across all tested failure-injection scenarios (error returns and panics).
- **SC-002**: When a callback panics mid-transaction, the transaction is fully closed with no leaked or dangling open transaction, in 100% of tested panic scenarios.
- **SC-003**: Every attempt to associate model instances from two different sessions is rejected before any data is written, with zero instances of silent cross-session data leakage across repeated testing.
- **SC-004**: Developers can determine, from the rejection alone, whether a failed association was caused by a session mismatch or by an unsaved related model, without needing to inspect internal implementation details.
- **SC-005**: A developer unfamiliar with the feature can correctly identify, after reading the documented behavior, which models are safe to use together inside a given transaction, in a first-attempt comprehension check.

## Assumptions

- Nested transactions (opening a new transaction from within an already-running transaction's callback) are out of scope for this feature; each transaction is a standalone unit of work.
- Concurrent use of a single session, or of a single model instance, from multiple threads of execution at the same time is not guaranteed safe and is out of scope for this feature.
- A transaction's session and the model instances bound to it are intended for use only within that transaction's callback; using them after the callback has returned is an unsupported pattern, not a guarantee this feature provides.
- Automatic cascading saves of related models (having the system save an entire graph of related, unsaved models automatically) are out of scope; developers are responsible for saving related models explicitly and in the correct order.
- The underlying persistence layer supports standard atomic transaction semantics (begin, commit, rollback) that this feature builds on rather than reimplements.
