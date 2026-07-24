# How to define relation scopes

## Expose parameters in SQL

```sql
-- name: ListTagsByPost :many
SELECT tags.id, tags.name, tags.usage_count
FROM tags
INNER JOIN post_tags ON post_tags.tag_id = tags.id
WHERE post_tags.post_id = sqlc.arg(post_id)
  AND (
      sqlc.narg(minimum_usage_count)::BIGINT IS NULL
      OR tags.usage_count >= sqlc.narg(minimum_usage_count)
  )
ORDER BY tags.usage_count DESC;
```

## Map a constant scope

```yaml
relations:
  Tags:
    scopes:
      Popular:
        parameter: minimum_usage_count
        value: 100
```

Generated API:

```go
tags, err := post.Tags().Popular().Get(ctx)
```

## Map an argument scope

```yaml
scopes:
  MinimumUsage:
    parameter: minimum_usage_count
    argument: int64
```

Generated API:

```go
tags, err := post.Tags().MinimumUsage(500).Get(ctx)
```

## Map query variants when one query cannot express a scope

```yaml
scopes:
  Archived:
    query: ListArchivedTagsByPost
```

A query-variant scope selects a different configured sqlc method rather than generating SQL dynamically.

## Validate combinations

Generation must fail when a scope parameter is missing, the type is incompatible, the method name conflicts, the selected query returns an incompatible result, or the scope implies arbitrary SQL that no configured query can execute.

## Keep builders immutable by value

Scope methods use value receivers:

```go
published := user.Posts().Published()
all := user.Posts()
```

Configuring `published` does not mutate `all` or the model's canonical relation state.
