package host

import (
	"context"

	"github.com/lewtec/lewkit/x/driver"
	"github.com/lucasew/env-mcp/internal/runner"
)

// StateDir is where this executor keeps its per-env directories. Set it
// before the first use.
var StateDir = "."

// factory registers the plain-process executor. It always works, so it is
// the fallback when no terminal multiplexer is running.
type factory struct{}

func (factory) ID() string   { return "runner_host" }
func (factory) Name() string { return "host" }
func (factory) Weight() int  { return 10 }

func (factory) CheckCompatibility(context.Context) error { return nil }

func (factory) New(context.Context) (runner.Executor, error) {
	return New(StateDir), nil
}

func init() {
	driver.Register[runner.Executor](factory{})
}
