package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func minimalValidRequest() *pb.GenerateRequest {
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Widget": {
					"row": "Widget",
					"operations": {"insert": "CreateWidget"},
					"fields": {"id": {"readable": true}}
				}
			}
		}]
	}`
	return &pb.GenerateRequest{
		PluginOptions: []byte(options),
		Queries: []*pb.Query{
			{
				Name: "CreateWidget",
				Cmd:  ":one",
				Text: "INSERT INTO widgets DEFAULT VALUES RETURNING id;",
				Columns: []*pb.Column{
					{Name: "id", OriginalName: "id", NotNull: true, Type: &pb.Identifier{Name: "int8"}},
				},
			},
		},
	}
}

func TestHandle_Success(t *testing.T) {
	resp, err := handle(context.Background(), minimalValidRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || len(resp.Files) == 0 {
		t.Fatalf("expected a non-empty response, got %+v", resp)
	}
}

func TestHandle_ConfigError(t *testing.T) {
	req := &pb.GenerateRequest{PluginOptions: []byte(`not json`)}
	resp, err := handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected an error for invalid configuration")
	}
	if resp != nil {
		t.Fatalf("expected a nil response, got %+v", resp)
	}
	if !strings.Contains(err.Error(), "invalid configuration") {
		t.Fatalf("expected the diagnostic text in the error, got %q", err.Error())
	}
}

// TestHandle_SuccessWithWarnings covers handle's "print warnings to stderr,
// still return the response" branch: a `generated: save` field with no
// configured update operation succeeds but produces a warning diagnostic.
func TestHandle_SuccessWithWarnings(t *testing.T) {
	options := `{
		"version": 1,
		"sqlc": {"package": "p", "import": "i", "driver": "pgx/v5"},
		"contexts": [{
			"name": "content", "package": "content", "directory": "content",
			"models": {
				"Widget": {
					"row": "Widget",
					"operations": {"insert": "CreateWidget"},
					"fields": {
						"id": {"readable": true},
						"stamp": {"readable": true, "generated": "save"}
					}
				}
			}
		}]
	}`
	req := &pb.GenerateRequest{
		PluginOptions: []byte(options),
		Queries: []*pb.Query{
			{
				Name: "CreateWidget",
				Cmd:  ":one",
				Text: "INSERT INTO widgets DEFAULT VALUES RETURNING id, stamp;",
				Columns: []*pb.Column{
					{Name: "id", OriginalName: "id", NotNull: true, Type: &pb.Identifier{Name: "int8"}},
					{Name: "stamp", OriginalName: "stamp", NotNull: true, Type: &pb.Identifier{Name: "timestamptz"}},
				},
			},
		},
	}

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	resp, herr := handle(context.Background(), req)
	if err := w.Close(); err != nil {
		t.Fatalf("close stderr pipe: %v", err)
	}
	captured, _ := io.ReadAll(r)

	if herr != nil {
		t.Fatalf("unexpected error: %v", herr)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response despite the warning")
	}
	if !strings.Contains(string(captured), "generated: save has no effect") {
		t.Fatalf("expected the warning diagnostic on stderr, got %q", captured)
	}
}

// TestMain_FullRoundTrip drives the real main() entrypoint end to end
// through the plugin SDK's stdio RPC protocol: main() blocks reading
// os.Stdin for a marshaled plugin.GenerateRequest and writes a marshaled
// plugin.GenerateResponse to os.Stdout, so both are redirected to pipes
// this test controls. os.Args is trimmed to just the program name so the
// SDK's argument-based method routing falls back to its documented
// backwards-compatible default (the single Generate method) rather than
// trying to interpret `go test`'s own flags as an RPC method name. The
// request is engineered to fully succeed: on any failure the SDK calls
// os.Exit itself, which would tear down this test process.
func TestMain_FullRoundTrip(t *testing.T) {
	reqBytes, err := proto.Marshal(minimalValidRequest())
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe (stdin): %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe (stdout): %v", err)
	}

	origStdin, origStdout, origArgs := os.Stdin, os.Stdout, os.Args
	os.Stdin = stdinR
	os.Stdout = stdoutW
	os.Args = []string{origArgs[0]}
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
		os.Args = origArgs
	}()

	go func() {
		if _, err := stdinW.Write(reqBytes); err != nil {
			t.Errorf("write request: %v", err)
		}
		if err := stdinW.Close(); err != nil {
			t.Errorf("close stdin pipe: %v", err)
		}
	}()

	var captured bytes.Buffer
	done := make(chan struct{})
	go func() {
		if _, err := io.Copy(&captured, stdoutR); err != nil {
			t.Errorf("copy response: %v", err)
		}
		close(done)
	}()

	main()

	if err := stdoutW.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}
	<-done

	var resp pb.GenerateResponse
	if err := proto.Unmarshal(captured.Bytes(), &resp); err != nil {
		t.Fatalf("proto.Unmarshal response: %v", err)
	}
	if len(resp.Files) == 0 {
		t.Fatalf("expected main() to produce generated files, got %+v", &resp)
	}
}
