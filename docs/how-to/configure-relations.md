# How to configure direct relations

## Belongs to

```yaml
models:
  Post:
    relations:
      Author:
        kind: belongs_to
        model: User
        local_key: user_id
        foreign_key: id
        inverse: Posts
        lazy_query: GetUser
        eager_query: ListUsersByIDs
        nullable: false
        cache: canonical
```

Generated operations:

```go
author, err := post.Author().Get(ctx)
post.Author().Associate(user)
```

`Dissociate` is generated only when the foreign key is nullable.

## Has many

```yaml
models:
  User:
    relations:
      Posts:
        kind: has_many
        model: Post
        local_key: id
        foreign_key: user_id
        inverse: Author
        lazy_query: ListPostsByUser
        eager_query: ListPostsByUserIDs
        cache: canonical
```

Generated operations:

```go
posts, err := user.Posts().Get(ctx)
post := user.Posts().New()
```

`New()` attaches the new child to the same session and assigns the parent key when available. It does not persist the child.

## Has one

Configure similarly to `has_many`, but require a `:one` lazy query and define the expected behavior when no row exists.

## Many to many

```yaml
relations:
  Tags:
    kind: many_to_many
    model: Tag
    lazy_query: ListTagsByPost
    eager_query: ListTagsByPostIDs
    attach_query: AttachTagToPost
    detach_query: DetachTagFromPost
    sync_queries:
      list: ListTagIDsByPost
      attach: AttachTagToPost
      detach: DetachTagFromPost
```

Terminal mutation operations require explicit sqlc queries:

```go
err := post.Tags().Attach(ctx, tagA, tagB)
err := post.Tags().Detach(ctx, tagA)
err := post.Tags().Sync(ctx, tagA, tagC)
```

The generator does not invent pivot-table SQL.

## Enforce session identity

`Associate`, `Attach`, and `Sync` reject models attached to a different session. The initial release also rejects unsaved related models without usable identifiers.
