package mapping

import (
	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// GoType is the resolved Go representation for a query result column: the
// type expression to use in generated code, plus any import path it needs
// beyond the driver package itself.
type GoType struct {
	Expr   string // e.g. "pgtype.Text", "bool", "[]byte"
	Import string // e.g. "github.com/jackc/pgx/v5/pgtype"; "" if none needed
	// Unmapped is true when the column's Postgres type had no specific
	// mapping and Expr fell back to "string" — surfaced as a warning by
	// the caller rather than failing generation outright.
	Unmapped bool
}

const pgtypeImport = "github.com/jackc/pgx/v5/pgtype"

// notNullGo maps a Postgres builtin type name (Column.Type.Name, as
// reported by the sqlc plugin protocol) to the Go type sqlc-gen-go's own
// pgx/v5 driver support generates for a NOT NULL column of that type. This
// plugin doesn't have access to sqlc-gen-go's actual generated struct
// during its own invocation (contracts/plugin-io.md), so store/record
// adapters can only assign field-by-field against real sqlc-gen-go output
// without conversion if this table matches sqlc-gen-go's own defaults:
// native Go types for the scalar types that have a safe zero value, pgtype
// wrapper types for everything else (uuid, temporal types, numeric,
// interval, inet), matching pgx/v5's own recommended type set.
var notNullGo = map[string]GoType{
	"uuid":        {Expr: "pgtype.UUID", Import: pgtypeImport},
	"text":        {Expr: "string"},
	"varchar":     {Expr: "string"},
	"bpchar":      {Expr: "string"},
	"citext":      {Expr: "string"},
	"name":        {Expr: "string"},
	"bool":        {Expr: "bool"},
	"int2":        {Expr: "int16"},
	"int4":        {Expr: "int32"},
	"int8":        {Expr: "int64"},
	"float4":      {Expr: "float32"},
	"float8":      {Expr: "float64"},
	"numeric":     {Expr: "pgtype.Numeric", Import: pgtypeImport},
	"timestamptz": {Expr: "pgtype.Timestamptz", Import: pgtypeImport},
	"timestamp":   {Expr: "pgtype.Timestamp", Import: pgtypeImport},
	"date":        {Expr: "pgtype.Date", Import: pgtypeImport},
	"time":        {Expr: "pgtype.Time", Import: pgtypeImport},
	"timetz":      {Expr: "pgtype.Time", Import: pgtypeImport},
	"interval":    {Expr: "pgtype.Interval", Import: pgtypeImport},
	"bytea":       {Expr: "[]byte"},
	"json":        {Expr: "[]byte"},
	"jsonb":       {Expr: "[]byte"},
	"inet":        {Expr: "pgtype.Inet", Import: pgtypeImport},
}

// nullableGo overrides notNullGo for columns where NOT NULL is false: any
// type without a safe zero-value-as-NULL representation switches to its
// pgtype wrapper (native Go types can't represent SQL NULL). bytea/json/
// jsonb are exempt — a nil []byte already represents NULL for those.
var nullableGo = map[string]GoType{
	"text":    {Expr: "pgtype.Text", Import: pgtypeImport},
	"varchar": {Expr: "pgtype.Text", Import: pgtypeImport},
	"bpchar":  {Expr: "pgtype.Text", Import: pgtypeImport},
	"citext":  {Expr: "pgtype.Text", Import: pgtypeImport},
	"name":    {Expr: "pgtype.Text", Import: pgtypeImport},
	"bool":    {Expr: "pgtype.Bool", Import: pgtypeImport},
	"int2":    {Expr: "pgtype.Int2", Import: pgtypeImport},
	"int4":    {Expr: "pgtype.Int4", Import: pgtypeImport},
	"int8":    {Expr: "pgtype.Int8", Import: pgtypeImport},
	"float4":  {Expr: "pgtype.Float4", Import: pgtypeImport},
	"float8":  {Expr: "pgtype.Float8", Import: pgtypeImport},
}

// ResolveGoType maps col to its generated Go field type. Unknown/unmapped
// Postgres types fall back to "string" with Unmapped set, rather than
// failing generation — an explicit column/row_field override or a future
// type-mapping extension can refine this later without blocking a working
// baseline model today.
func ResolveGoType(col *pb.Column) GoType {
	if col.Type == nil || col.Type.Name == "" {
		return GoType{Expr: "string", Unmapped: true}
	}

	gt, ok := notNullGo[col.Type.Name]
	if !ok {
		return GoType{Expr: "string", Unmapped: true}
	}
	if !col.NotNull {
		if override, ok := nullableGo[col.Type.Name]; ok {
			gt = override
		}
	}

	if col.IsArray || col.ArrayDims > 0 {
		return GoType{Expr: "[]" + gt.Expr, Import: gt.Import}
	}
	return gt
}
