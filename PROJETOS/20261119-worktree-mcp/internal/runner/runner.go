// Package runner starts long-lived commands for an env and reports their
// output. It is the seam between the operation table and whatever actually
// owns a terminal: Herdr in the default build, a plain process elsewhere.
//
// A runner never returns a terminal identifier to the agent. It returns a
// [Place], an opaque token it alone interprets on later calls.
package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/lewtec/lewkit/x/driver"
)

var (
	// ErrPattern means the RE2 pattern does not compile.
	ErrPattern = errors.New("invalid pattern")
	// ErrOffset means the offset is past what the runner can still reach.
	ErrOffset = errors.New("offset out of range")
)

// State is the lifecycle of one command.
type State string

const (
	Running State = "running"
	Exited  State = "exited"
)

// Env is one env as a runner sees it. Checkout is the worktree; Source is
// the main checkout it was cut from, which a terminal multiplexer may need
// to attach a linked worktree to its parent.
type Env struct {
	Checkout string
	Source   string
	Label    string
}

// Place addresses one started command inside a runner. The value is opaque
// outside the runner; the server stores it and hands it back unchanged.
type Place string

// Start describes one command to start.
type Start struct {
	// Dir is the absolute working directory, already resolved inside the env.
	Dir string
	// Argv is the command and its arguments. Argv[0] is non-empty.
	Argv []string
	// Label names the place for a human watching the terminal surface.
	Label string
}

// Status is what a runner reports about one place.
type Status struct {
	State    State
	ExitCode int
}

// Page is one read of a place's output, starting at a line offset.
type Page struct {
	Lines      []string
	NextOffset int
	// Truncated is set when the runner can no longer reach the requested
	// offset because its scrollback dropped those lines.
	Truncated bool
}

// Output asks for whole lines from an offset. Pattern is an RE2 expression;
// empty matches every line. The cursor advances past every scanned line,
// whether it matched.
type Output struct {
	Offset  int
	Limit   int
	Pattern string
}

// Executor starts commands and reads their output.
//
// Implementations must be safe to use from several processes: the server
// may be a short-lived CLI process each time, so nothing may live only in
// memory.
type Executor interface {
	// ID is the registry name stored with each exec.
	ID() string
	// Open ensures the surface for one env exists and returns its place.
	Open(ctx context.Context, env Env) (Place, error)
	// Start runs a command on that surface and returns the place holding it.
	Start(ctx context.Context, place Place, start Start) (Place, error)
	// Status reports one started command.
	Status(ctx context.Context, place Place) (Status, error)
	// Read returns lines from one started command.
	Read(ctx context.Context, place Place, out Output) (Page, error)
	// Kill stops one started command. Killing an exited command succeeds.
	Kill(ctx context.Context, place Place) error
	// Close tears down everything the runner created for one env.
	Close(ctx context.Context, place Place) error
}

// Open builds the executor for this host.
//
// It reads the lewkit driver registry, so a registered executor wins by
// weight and by compatibility. A build that imports only the host executor
// still works; a build that imports the Herdr executor uses it whenever a
// session is running.
func Open(ctx context.Context) (Executor, error) {
	e, err := driver.Get[Executor](ctx)
	if err != nil {
		return nil, fmt.Errorf("runner: %w", err)
	}
	return e, nil
}
