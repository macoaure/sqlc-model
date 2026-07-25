package contract

// OperationKind identifies one of a model's lifecycle operations
// (data-model.md "Lifecycle Operation").
type OperationKind string

const (
	Find    OperationKind = "find"
	Insert  OperationKind = "insert"
	Update  OperationKind = "update"
	Delete  OperationKind = "delete"
	Refresh OperationKind = "refresh"
)

// requirement describes what sqlc query command a given operation kind
// requires, per data-model.md's Lifecycle Operation table. For insert and
// update, sqlc only accepts the `:one` annotation on a query whose
// statement actually has a RETURNING clause returning exactly one row — so
// requiring `:one` here also enforces FR-004's "full persisted row via
// RETURNING" without this plugin needing to parse SQL text itself.
type requirement struct {
	cmd         string
	fallbackCmd string // non-empty: an alternative command allowed with a warning, not an error
	altCmd      string // non-empty: an alternative command equally acceptable, no warning
}

var requirements = map[OperationKind]requirement{
	Find:    {cmd: ":one"},
	Insert:  {cmd: ":one"},
	Update:  {cmd: ":one"},
	Delete:  {cmd: ":execrows", fallbackCmd: ":exec"},
	Refresh: {cmd: ":one"},
}
