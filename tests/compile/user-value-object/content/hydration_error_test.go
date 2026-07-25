package content

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type invalidEmailExecutor struct{}

func (invalidEmailExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (invalidEmailExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (invalidEmailExecutor) QueryRow(context.Context, string, ...any) pgx.Row {
	return invalidEmailRow{}
}

type invalidEmailRow struct{}

func (invalidEmailRow) Scan(dest ...any) error {
	*dest[0].(*pgtype.UUID) = pgtype.UUID{}
	*dest[1].(*string) = "not-an-email"
	return nil
}

func TestHydrationErrorNamesFieldAndWrapsConstructorError(t *testing.T) {
	_, err := newUserStore(invalidEmailExecutor{}).find(context.Background(), userRecord{})
	if err == nil {
		t.Fatal("expected invalid email hydration to fail")
	}
	if !strings.Contains(err.Error(), "User.Email") {
		t.Fatalf("error = %q, want model and field context", err.Error())
	}
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("error = %v, want wrapped ErrInvalidEmail", err)
	}
}
