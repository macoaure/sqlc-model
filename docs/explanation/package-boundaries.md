# Explanation: package boundaries

## Why package-per-relation fails

A visually scoped hierarchy such as:

```text
models/user
models/user/posts
models/post
models/post/author
```

creates separate Go packages. Bidirectional relations create a cycle:

```text
user → user/posts → post → post/author → user
```

Import cycles do not compile. Passing only IDs avoids one direct dependency but does not solve the target model return type or inverse relation.

## Why traversal paths are not ownership

`models/user/posts/tags` describes one route through a relation graph, not a stable domain object. Cyclic graphs have no natural traversal depth, and generated directories would proliferate for repeated paths.

## Chosen structure

Concrete related model types and public relation builders share a bounded-context package:

```text
internal/models/content/
├── user_gen.go
├── user_posts_relation_gen.go
├── post_gen.go
├── post_author_relation_gen.go
├── post_tags_relation_gen.go
└── tag_gen.go
```

All use `package content`.

Persistence internals may remain hierarchically scoped:

```text
internal/models/content/internal/
├── user/posts/
├── post/author/
└── post/tags/
```

These internal packages return neutral generated records or sqlc results and do not import the public model package. The public package owns private hydration and inverse-cache coordination.

## Bounded-context boundary

Cross-context concrete relations are deferred. If two models require bidirectional concrete relations, they belong to the same generated context unless the architecture introduces a deliberate interface or identifier boundary.
