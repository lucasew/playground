package runner_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/lewtec/lewkit/x/driver"
	"github.com/lucasew/env-mcp/internal/runner"
	_ "github.com/lucasew/env-mcp/internal/runner/herdr"
	"github.com/lucasew/env-mcp/internal/runner/host"
)

// TestHostCompatibilityFollowsTheVariable checks the factory directly. The
// registry caches compatibility per driver for the life of the process, so
// the on-and-off cases cannot share one process through runner.Open.
func TestHostCompatibilityFollowsTheVariable(t *testing.T) {
	cases := map[string]bool{
		"":     false,
		"0":    false,
		"no":   false,
		"1":    true,
		"true": true,
	}
	for value, wantOn := range cases {
		t.Run(value, func(t *testing.T) {
			t.Setenv(host.EnableVar, value)
			err := host.CheckCompatibility(context.Background())
			gotOn := err == nil
			if gotOn != wantOn {
				t.Fatalf("%s=%q compatible=%v, want %v (err %v)", host.EnableVar, value, gotOn, wantOn, err)
			}
			if err != nil && !errors.Is(err, driver.ErrIncompatible) {
				t.Fatalf("error should wrap ErrIncompatible, got %v", err)
			}
		})
	}
}

// TestOpenWithoutOptInFails runs in a child process so the registry's
// per-process compatibility cache starts empty.
func TestOpenWithoutOptInFails(t *testing.T) {
	if os.Getenv("RUNNER_TEST_CHILD") == "1" {
		t.Setenv("HERDR_SOCKET_PATH", "/nonexistent/herdr.sock")
		t.Setenv(host.EnableVar, "")
		if _, err := runner.Open(context.Background()); err == nil {
			t.Fatal("Open succeeded with no Herdr session and no opt-in")
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestOpenWithoutOptInFails$")
	cmd.Env = append(os.Environ(), "RUNNER_TEST_CHILD=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}
}

// TestOpenUsesHostWhenOptedIn is the mirror case, also in a fresh process.
func TestOpenUsesHostWhenOptedIn(t *testing.T) {
	if os.Getenv("RUNNER_TEST_CHILD") == "2" {
		e, err := runner.Open(context.Background())
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		if e.ID() != "host" {
			t.Fatalf("want the host executor, got %q", e.ID())
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestOpenUsesHostWhenOptedIn$")
	cmd.Env = append(os.Environ(),
		"RUNNER_TEST_CHILD=2",
		"HERDR_SOCKET_PATH=/nonexistent/herdr.sock",
		host.EnableVar+"=1",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}
}
