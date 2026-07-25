// Package relation validates the relation graph declared across a
// bounded context's models: target/inverse resolution and kind
// compatibility, per-kind query command/cardinality requirements, and scope
// parameter/type/name-collision compatibility. It is the home for
// spec 001's previously-reserved diagnostics stages 7 ("relation graph") and
// 8 ("scope compatibility") — see specs/002-relations-lazy-eager-loading/
// data-model.md and contracts/relation-diagnostics-contract.md. Decoding
// RelationConfiguration/ScopeConfiguration itself lives in internal/config,
// alongside the rest of the plugin's `options` schema; this package only
// validates the already-decoded shape against other models and real sqlc
// query metadata.
package relation
