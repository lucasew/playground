package herdr

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lewtec/lewkit/x/driver"
	"github.com/lucasew/env-mcp/internal/herdr"
	"github.com/lucasew/env-mcp/internal/runner"
)

// StateDir is where this executor keeps its layout record. Set it before
// the first use; the zero value keeps the record in the working directory.
var StateDir = "."

// factory registers the Herdr executor when a session is reachable.
type factory struct{}

func (factory) ID() string   { return "runner_herdr" }
func (factory) Name() string { return "herdr" }
func (factory) Weight() int  { return 90 }

// CheckCompatibility reports incompatible when no Herdr session answers, so
// a host with no multiplexer falls back to the plain-process executor.
func (factory) CheckCompatibility(ctx context.Context) error {
	sock, err := (&herdr.Client{}).Socket(ctx)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Dir(sock)); err != nil {
		return err
	}
	return nil
}

func (factory) New(context.Context) (runner.Executor, error) {
	return New(StateDir), nil
}

func init() {
	driver.Register[runner.Executor](factory{})
}
