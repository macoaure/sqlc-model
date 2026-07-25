package content

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type recordingExecutor struct {
	called bool
}

func (e *recordingExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	e.called = true
	return pgconn.CommandTag{}, errors.New("query should not execute")
}

func (e *recordingExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	e.called = true
	return nil, errors.New("query should not execute")
}

func (e *recordingExecutor) QueryRow(context.Context, string, ...any) pgx.Row {
	e.called = true
	return failingRow{}
}

type failingRow struct{}

func (failingRow) Scan(...any) error { return errors.New("query should not execute") }

func TestSaveReturnsValidationErrorBeforeQuery(t *testing.T) {
	executor := &recordingExecutor{}
	sess := &Session{executor: executor, identity: &sessionIdentity{}}
	sess.initCollections()
	user := sess.Users.New()
	validationErr := errors.New("name is required")

	user.setFieldError(UserFieldName, validationErr)
	err := user.Save(context.Background())

	if !errors.Is(err, validationErr) {
		t.Fatalf("Save() error = %v, want validation error", err)
	}
	if executor.called {
		t.Fatal("Save() should not execute a query when validation fails")
	}
}
