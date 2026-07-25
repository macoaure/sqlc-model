# Explanation: transactions and session identity

A model's `Save(ctx)` needs access to persistence infrastructure. The selected design attaches a private session instead of retrieving repositories from `context.Context`.

## Why session identity matters

A transaction session uses sqlc queries rebound through `WithTx`. A model loaded before the transaction still points to the root session.

Silently adopting the transaction would hide a major persistence boundary. Silently keeping the root session would execute outside the expected transaction. Both are unsafe.

The framework therefore assigns each session an identity and rejects cross-session associations. Models participating in a transaction must be loaded or created from the callback session.

## Transaction lifecycle

The transaction implementation begins a driver transaction, immediately defers rollback cleanup, creates a new session over `Queries.WithTx`, runs the callback, and commits only on `nil`.

Panic cleanup is non-negotiable. Rollback registration occurs before user code runs. The panic may be rethrown after cleanup is secured.

## No implicit reattachment

An explicit future API could clone or reload a model into another session, but the first release does not mutate a model's session behind the caller's back.
