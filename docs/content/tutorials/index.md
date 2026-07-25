# Tutorials

Tutorials teach the project by guiding the reader through a complete result. Follow them in order when evaluating the proposed API.

## Learning path

1. [Generate the first rich model](first-model.md): define schema and queries, configure the plugin, generate a model, create a session, and persist a user.
2. [Add the first lazy relation](first-relation.md): configure `User has many Posts`, define lazy and eager queries, and inspect relation caching.
3. [Run model operations in a transaction](first-transaction.md): create a transaction-bound session and enforce session identity.

These tutorials intentionally use PostgreSQL, pgx/v5, single-column UUID identifiers, and a single bounded-context package. Broader compatibility belongs in reference and implementation documentation.
