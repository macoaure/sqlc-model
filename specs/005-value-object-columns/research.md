# Phase 0 Research: Value Object Field Mapping

## Decision: Add `value_object` under each field policy

**Rationale**: `FieldPolicy` already owns per-field API and mapping knobs (`column`, `row_field`, mutability, generated state). A nested value-object block keeps the feature scoped to one field and avoids a model-level registry.

**Alternatives considered**: A top-level value-object registry; rejected because mappings are not reused by the generator and would need extra name resolution. Generated value-object source; rejected because domain behavior must stay developer-owned.

## Decision: Split exposed type from persisted type in the resolved plan

**Rationale**: Today `mapping.ResolveGoType` produces one `GoType` used by records, models, and store arguments. Value objects need two types: the persisted sqlc-compatible type for query scanning/parameters and the exposed developer type for model fields.

**Alternatives considered**: Change `mapping.ResolveGoType` to return the value-object type directly; rejected because store parameter rendering would lose the sqlc primitive/wrapper type. Add a new conversion package; rejected because this is just per-field metadata.

## Decision: Hydration calls constructor immediately after scanning persisted values

**Rationale**: The constructor is the developer-owned validation boundary. Calling it before the generated model is returned satisfies the "invalid stored data fails hydration" contract and keeps invalid value objects out of model fields.

**Alternatives considered**: Store the primitive and lazily construct on getter; rejected because it delays errors and makes the generated field type misleading.

## Decision: Persistence calls accessor only at query argument construction

**Rationale**: The model already holds the value object. The only place the persisted value is needed is store execution, so conversion belongs in the generated adapter path that builds query parameters.

**Alternatives considered**: Keep a parallel primitive snapshot beside each value object; rejected because it adds sync bugs and duplicates data the accessor can provide.

## Decision: Validate config shape before generation; enforce handwritten symbol signatures at compile time

**Rationale**: The sqlc plugin request gives reliable query metadata and options, not the full developer package source. The generator can reject missing `type`, `constructor`, or `accessor` names and invalid Go identifiers/selectors before output. The Go compiler is the reliable authority for whether `NewEmail(string) (Email, error)` and `Email.String() string` actually exist.

**Alternatives considered**: Type-check arbitrary project source during plugin execution; rejected as heavy, environment-sensitive, and outside the current generator pipeline. Reflection; rejected because generation happens before runtime and the types may not be compiled yet.

## Decision: Unconfigured mismatches are errors, not nullable guesses

**Rationale**: FR-007 and FR-008 exist to prevent accidental `.Valid`-style plumbing for unrelated custom types. The generator should only convert when types are directly assignable, recognized by the current mapping table, or explicitly configured as a value object.

**Alternatives considered**: Best-effort conversion based on field names or method names; rejected because it recreates the unsafe guessing this feature is meant to stop.
