// Package contract validates each configured lifecycle operation against
// sqlc query metadata: the required command (:one, :execrows, ...), whether
// insert/update return the full persisted row via RETURNING, and whether a
// query's parameters and result columns can hydrate the fields it needs to.
package contract
