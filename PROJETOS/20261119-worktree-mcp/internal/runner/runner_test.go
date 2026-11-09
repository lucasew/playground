package runner_test

import (
	"context"
	"testing"

	"github.com/lucasew/env-mcp/internal/runner"
	_ "github.com/lucasew/env-mcp/internal/runner/herdr"
	_ "github.com/lucasew/env-mcp/internal/runner/host"
)

func TestHostIsTheFallbackWhenHerdrIsDown(t *testing.T) {
	t.Setenv("HERDR_SOCKET_PATH", "/nonexistent/herdr.sock")
	e, err := runner.Open(context.Background())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if e.ID() != "host" {
		t.Fatalf("want the host executor, got %q", e.ID())
	}
}
