package content

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestFrameworkErrorIdentities(t *testing.T) {
	for _, err := range []error{
		ErrModelDetached,
		ErrModelDeleted,
		ErrNotFound,
		ErrInvalidModelState,
		ErrUnsupportedQueryContract,
	} {
		if !errors.Is(err, err) {
			t.Fatalf("%v should be usable with errors.Is", err)
		}
	}
}

func TestDatabaseErrorClassification(t *testing.T) {
	cases := map[string]DatabaseErrorKind{
		"23505": DatabaseErrorUniqueViolation,
		"23503": DatabaseErrorForeignKeyViolation,
		"23502": DatabaseErrorNotNullViolation,
		"23514": DatabaseErrorCheckViolation,
		"40001": DatabaseErrorSerializationFailure,
		"40P01": DatabaseErrorDeadlock,
		"99999": DatabaseErrorUnknown,
	}
	for code, want := range cases {
		driverErr := &pgconn.PgError{Code: code}
		err := classifyDatabaseError(driverErr)
		var dbErr *DatabaseError
		if !errors.As(err, &dbErr) {
			t.Fatalf("classifyDatabaseError(%s) = %T, want *DatabaseError", code, err)
		}
		if dbErr.Kind != want {
			t.Fatalf("classifyDatabaseError(%s) kind = %v, want %v", code, dbErr.Kind, want)
		}
		if !errors.Is(err, driverErr) {
			t.Fatalf("classified error should unwrap original driver error")
		}
	}
}
