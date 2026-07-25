# Phase 1 Data Model: Value Object Field Mapping

## ValueObjectMapping

Per-field configuration nested under `FieldPolicy`.

| Field | Type | Notes |
|-------|------|-------|
| `type` | Go type expression | Developer-owned value object exposed by the generated model field. |
| `constructor` | Go function expression | Called with the persisted field value during hydration. Expected shape: `func(PersistedType) (ValueObjectType, error)`. |
| `accessor` | Go method name | Called on the value object during persistence. Expected shape: `func() PersistedType`. |

Validation rules:
- all three fields are required when `value_object` is present;
- `type` and `constructor` may be package-qualified selector expressions;
- `accessor` is a method identifier only;
- nullable columns are out of scope for value-object mapping in this feature.

## ResolvedField

Existing resolved field metadata extended with conversion metadata.

| Field | Type | Notes |
|-------|------|-------|
| `Name` | string | Existing declared field key. |
| `GoField` | string | Existing generated field identifier. |
| `ColumnName` | string | Existing resolved sqlc result/column identity. |
| `PersistedGoType` | `mapping.GoType` | Existing sqlc-compatible type used for scan values and query parameters. |
| `ExposedGoType` | `mapping.GoType` or string equivalent | Type used by generated model getter/setter/original APIs. Equals `PersistedGoType` when no value-object mapping exists. |
| `ValueObject` | optional `ValueObjectMapping` | Present only for explicitly configured value-object fields. |

Validation rules:
- without `ValueObject`, existing assignability/known mapping behavior remains unchanged;
- with `ValueObject`, generated hydration must call the constructor before assigning the exposed field;
- with `ValueObject`, generated persistence must call the accessor before passing query arguments;
- one field's mapping never affects another field.

## Generated Record Shapes

Value-object fields need a persisted scan shape and an exposed model shape.

| Shape | Fields |
|-------|--------|
| persisted scan record | sqlc-compatible persisted types, used inside store scan/parameter code |
| model record | exposed types, used by `User.current`, `User.original`, getters, setters, and dirty checks |

Validation rules:
- constructor errors fail hydration and wrap model/field context;
- accessor conversion happens only when preparing persistence query arguments;
- direct fields keep the current one-shape behavior where exposed and persisted types are identical.

## Hydration Error

Error returned when a constructor rejects a stored value.

| Field | Type | Notes |
|-------|------|-------|
| model | string | Generated model name. |
| field | string | Generated field/config key. |
| cause | error | Original constructor error, preserved for `errors.Is`/`errors.As` wrapping. |

Relationships:

```text
FieldPolicy 0..1 -> ValueObjectMapping
ResolvedField 1 -> PersistedGoType
ResolvedField 1 -> ExposedGoType
ValueObjectMapping 1 -> constructor
ValueObjectMapping 1 -> accessor
```
