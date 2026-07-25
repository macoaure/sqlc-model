// Package mapping resolves each field's underlying database column and sqlc
// query result column, either from an explicit override or by exactly one
// unambiguous automatic match; ambiguous or missing matches are reported as
// diagnostics rather than guessed.
package mapping
