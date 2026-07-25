package contract

// Relation operation kinds extend the lifecycle OperationKind table above
// with the command/cardinality requirements a relation's configured queries
// must satisfy (data-model.md "Relation Graph Validation" stage 7).
// ValidateOperation is reused as-is — it was already generic over any
// OperationKind present in the requirements table.
const (
	// LazyToOne is a belongs_to/has_one relation's lazy_query: exactly one
	// related row, or none.
	LazyToOne OperationKind = "relation_lazy_to_one"
	// LazyToMany is a has_many/many_to_many relation's lazy_query: any
	// number of related rows.
	LazyToMany OperationKind = "relation_lazy_to_many"
	// Eager is any relation kind's eager_query: a batch query returning
	// related rows for many parents at once.
	Eager OperationKind = "relation_eager"
	// Attach is a many-to-many relation's attach_query (standalone or via
	// sync_queries.attach): :exec and :execrows are equally acceptable —
	// there is no row-hydration need, only an affected-row-count option.
	Attach OperationKind = "relation_attach"
	// Detach is a many-to-many relation's detach_query (standalone or via
	// sync_queries.detach).
	Detach OperationKind = "relation_detach"
	// SyncList is a many-to-many relation's sync_queries.list: the
	// currently-attached related IDs for one parent.
	SyncList OperationKind = "relation_sync_list"
)

func init() {
	requirements[LazyToOne] = requirement{cmd: ":one"}
	requirements[LazyToMany] = requirement{cmd: ":many"}
	requirements[Eager] = requirement{cmd: ":many"}
	requirements[Attach] = requirement{cmd: ":exec", altCmd: ":execrows"}
	requirements[Detach] = requirement{cmd: ":exec", altCmd: ":execrows"}
	requirements[SyncList] = requirement{cmd: ":many"}
}
