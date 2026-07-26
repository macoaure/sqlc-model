package unit

import (
	"testing"

	"github.com/macoaure/sqlc-model/internal/contract"
	"github.com/macoaure/sqlc-model/internal/diagnostics"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestValidateOperation_UnknownKind(t *testing.T) {
	q, diags := contract.ValidateOperation(contract.OperationKind("bogus"), "AnyQuery", nil, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query for unknown operation kind, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error for unknown operation kind, got %+v", diags)
	}
}

func TestValidateOperation_QueryNotFound(t *testing.T) {
	q, diags := contract.ValidateOperation(contract.Find, "GetUser", nil, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error for missing query, got %+v", diags)
	}
}

func TestValidateOperation_InsertRequiresOne(t *testing.T) {
	queries := []*pb.Query{{Name: "CreateUser", Cmd: ":exec"}}
	q, diags := contract.ValidateOperation(contract.Insert, "CreateUser", queries, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query for :exec insert, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected FR-004 error for insert without RETURNING (:exec), got %+v", diags)
	}
}

func TestValidateOperation_InsertOneSucceeds(t *testing.T) {
	queries := []*pb.Query{{Name: "CreateUser", Cmd: ":one"}}
	q, diags := contract.ValidateOperation(contract.Insert, "CreateUser", queries, "path", "ctx", "Model")
	if diagnostics.HasError(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if q == nil || q.Name != "CreateUser" {
		t.Fatalf("expected matched query, got %+v", q)
	}
}

func TestValidateOperation_UpdateRequiresOne(t *testing.T) {
	queries := []*pb.Query{{Name: "UpdateUser", Cmd: ":execrows"}}
	q, diags := contract.ValidateOperation(contract.Update, "UpdateUser", queries, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected FR-004 error for update without RETURNING, got %+v", diags)
	}
}

func TestValidateOperation_DeleteExecFallbackIsWarning(t *testing.T) {
	queries := []*pb.Query{{Name: "DeleteUser", Cmd: ":exec"}}
	q, diags := contract.ValidateOperation(contract.Delete, "DeleteUser", queries, "path", "ctx", "Model")
	if q == nil {
		t.Fatalf("expected :exec to be accepted as a fallback for delete")
	}
	if diagnostics.HasError(diags) {
		t.Fatalf("expected only a warning for :exec delete fallback, got error: %+v", diags)
	}
	if len(diags) == 0 {
		t.Fatalf("expected a warning diagnostic for the :exec fallback")
	}
}

func TestValidateOperation_DeleteExecrowsNoWarning(t *testing.T) {
	queries := []*pb.Query{{Name: "DeleteUser", Cmd: ":execrows"}}
	q, diags := contract.ValidateOperation(contract.Delete, "DeleteUser", queries, "path", "ctx", "Model")
	if q == nil {
		t.Fatalf("expected :execrows to be accepted for delete")
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for :execrows delete, got %+v", diags)
	}
}

func TestValidateOperation_FindRequiresOne(t *testing.T) {
	queries := []*pb.Query{{Name: "GetUser", Cmd: ":many"}}
	q, diags := contract.ValidateOperation(contract.Find, "GetUser", queries, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query for :many find, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error for find requiring :one, got %+v", diags)
	}
}

func TestValidateOperation_ListRequiresMany(t *testing.T) {
	queries := []*pb.Query{{Name: "ListUsers", Cmd: ":one"}}
	q, diags := contract.ValidateOperation(contract.List, "ListUsers", queries, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query for :one list, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error for list requiring :many, got %+v", diags)
	}
}

func TestValidateOperation_SingleRequiresOne(t *testing.T) {
	queries := []*pb.Query{{Name: "FindByEmail", Cmd: ":many"}}
	q, diags := contract.ValidateOperation(contract.Single, "FindByEmail", queries, "path", "ctx", "Model")
	if q != nil {
		t.Fatalf("expected nil query for :many single, got %+v", q)
	}
	if !diagnostics.HasError(diags) {
		t.Fatalf("expected an error for single requiring :one, got %+v", diags)
	}
}
