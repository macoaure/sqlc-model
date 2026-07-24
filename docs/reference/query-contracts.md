# Query contracts reference

The generator validates operation semantics against sqlc query metadata. It does not guess method signatures from query names or parameter counts.

## Find

Required command: `:one`.

- One row hydrates a clean persisted model.
- No row becomes `ErrNotFound`.
- Other errors pass through the database-error translator.

The result must contain every canonical model field.

## Insert

Default required command: `:one` with `RETURNING`.

The result must provide canonical persisted state, including generated identifiers, defaults, timestamps, and trigger-modified fields.

An `:exec` insert is rejected under the default policy.

## Update

Default required command: `:one` with `RETURNING`.

The returned row hydrates the same model instance and synchronizes the original snapshot.

## Delete

Preferred command: `:execrows`, allowing missing-row detection.

Fallback command: `:exec`. With `:exec`, the framework cannot distinguish zero affected rows unless the driver reports an error.

## Refresh

Required command: `:one` returning canonical model fields.

## Lazy has-many

Required command: `:many` with a configured single-parent key parameter.

## Eager has-many

Required command: `:many` accepting a collection of parent identifiers and returning a grouping field.

## Unsupported contracts

Generation fails when a query is missing, has an incompatible annotation, lacks required parameters, returns insufficient fields, has ambiguous mappings, or uses types unsupported by the configured compatibility baseline.

## Parameter signatures

The recommended sqlc setting is:

```yaml
query_parameter_limit: 0
```

This reduces signature variability and simplifies stable adapter generation. The generator must still read the actual metadata rather than relying solely on the recommendation.
