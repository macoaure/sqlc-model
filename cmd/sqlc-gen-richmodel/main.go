// Command sqlc-gen-richmodel is a sqlc codegen plugin that generates an
// Eloquent-inspired rich model layer over sqlc-generated queries.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/generate"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func main() {
	codegen.Run(handle)
}

// handle adapts internal/generate.Generate to the plugin SDK's Handler
// signature. Per contracts/plugin-io.md, a run with any error-severity
// diagnostic returns zero files — here, that means returning a non-nil
// error, which the SDK's Run surfaces on stderr and a non-zero exit code
// rather than a partial GenerateResponse. Warning-severity diagnostics
// don't block output; they're still printed to stderr so they're visible
// without corrupting the protobuf response on stdout.
func handle(_ context.Context, req *pb.GenerateRequest) (*pb.GenerateResponse, error) {
	resp, diags := generate.Generate(req)
	if resp == nil {
		return nil, errors.New(diagnostics.FormatAll(diags))
	}
	if len(diags) > 0 {
		fmt.Fprintln(os.Stderr, diagnostics.FormatAll(diags))
	}
	return resp, nil
}
