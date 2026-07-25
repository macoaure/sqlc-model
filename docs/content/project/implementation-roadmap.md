# Implementation roadmap

## Phase 0: identity and baseline

Deliver a canonical repository name, Go module path, binary name, runtime package, versioning policy, and supported Go/sqlc/PostgreSQL/pgx matrix.

Acceptance criteria:

- `go install` resolves the canonical module.
- README commands use one path.
- plugin and runtime versions are compatible.
- CI installs from a clean module.

## Phase 1: protocol and intermediate representation

Implement CodeGenRequest decoding, options parsing, schema validation, query catalog, model and relation IR, diagnostics, and deterministic ordering.

The IR must preserve query name, command, parameters, Go types, result cardinality, result fields, table metadata, nullability, enum data, renames, and available overrides.

Acceptance criteria:

- templates never consume raw protocol structures;
- invalid contracts fail before rendering;
- diagnostics identify a configuration path;
- IR output is deterministic.

## Phase 2: lifecycle runtime

Implement session identity, model attachment, lifecycle flags, original snapshots, dirty tracking, field errors, canonical relation cache, framework errors, and database-error translation interfaces.

Acceptance criteria:

- new models are attached and nonexisting;
- hydrated models are existing and clean;
- reverting a field clears dirty state;
- correcting input clears its field error;
- deleted models cannot save or refresh.

## Phase 3: model and collection generation

Generate fields, constants, getters, fillable setters, original getters, dirty methods, validation methods, lifecycle methods, collections, `New`, `Find`, and configured named retrieval methods.

Acceptance criteria:

- handwritten files survive regeneration;
- generated code is formatted;
- generated packages compile with actual sqlc output;
- no public restore or mark-persisted methods exist.

## Phase 4: persistence adapters

Implement operation-specific adapters for find `:one`, insert `:one`, update `:one`, delete `:execrows`, delete `:exec` fallback, and refresh `:one`.

Acceptance criteria:

- database-generated values hydrate the same model;
- method signatures are derived from actual metadata;
- unsupported commands fail generation;
- affected-row semantics are enforced where available.

## Phase 5: direct relations

Implement belongs-to, has-one, has-many, canonical lazy loading, cache inspection/reset, inverse hydration, associate, dissociate, and child construction.

Acceptance criteria:

- relation construction performs no SQL;
- `Get(ctx)` performs lazy loading;
- `Cached()` never performs SQL;
- constrained results do not replace canonical cache;
- session mismatch and unsaved-parent policies are enforced.

## Phase 6: eager loading

Implement batch parent-key collection, configured eager queries, grouping, canonical cache population, inverse cache population, nested eager-load plans, and strict lazy-loading mode.

Acceptance criteria:

- N parents plus one relation use two queries;
- empty parents receive loaded empty caches;
- nested eager loading does not create traversal packages.

## Phase 7: typed scopes

Implement constant parameter scopes, argument scopes, query variants, compatibility validation, and deterministic relation state.

Acceptance criteria:

- every generated scope maps to a known sqlc capability;
- invalid parameters or result shapes fail generation;
- builder calls are value-based and isolated.

## Phase 8: transactions

Implement transaction sessions through sqlc `WithTx`, unique session identity, deferred rollback, panic-safe cleanup, commit behavior, and mismatch enforcement.

Acceptance criteria:

- nil commits;
- errors roll back;
- panic paths clean up;
- outer-session models cannot persist as transaction models.

## Phase 9: hardening

Add compile fixture matrices, PostgreSQL integration tests, race tests, configuration fuzzing, compatibility tests, benchmarks, generation performance budgets, and documentation compilation.

Acceptance criteria:

- every public example compiles;
- every supported query contract has a compile fixture;
- unsupported types produce actionable diagnostics;
- repeated generation is byte-for-byte deterministic after formatting.
