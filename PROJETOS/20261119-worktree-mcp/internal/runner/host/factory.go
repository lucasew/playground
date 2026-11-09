package host

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/lewtec/lewkit/x/driver"
	"github.com/lucasew/env-mcp/internal/runner"
)

// EnableVar opts this executor in. Without it the executor is incompatible,
// so a host with no Herdr session fails instead of quietly running commands
// as loose processes.
const EnableVar = "ENV_MCP_ALLOW_HOST_RUNNER"

// StateDir is where this executor keeps its per-env directories. Set it
// before the first use.
var StateDir = "."

// factory registers the plain-process executor. It is opt-in, so Herdr is
// the only default.
type factory struct{}

func (factory) ID() string   { return "runner_host" }
func (factory) Name() string { return "host" }
func (factory) Weight() int  { return 10 }

func (factory) CheckCompatibility(ctx context.Context) error { return CheckCompatibility(ctx) }

// CheckCompatibility reports incompatible unless EnableVar is set to a true
// value. The variable is read on every check, so a caller can opt in for one
// run without rebuilding.
func CheckCompatibility(context.Context) error {
	on, _ := strconv.ParseBool(os.Getenv(EnableVar))
	if !on {
		return fmt.Errorf("%w: set %s=1 to use it", driver.ErrIncompatible, EnableVar)
	}
	return nil
}

func (factory) New(context.Context) (runner.Executor, error) {
	return New(StateDir), nil
}

func init() {
	driver.Register[runner.Executor](factory{})
}
