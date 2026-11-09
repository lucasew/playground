// Package cli is the terminal surface of the operation table.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lewtec/lewkit/x/cmd"
	"github.com/lucasew/env-mcp/internal/env"
	herdr_runner "github.com/lucasew/env-mcp/internal/runner/herdr"
	"github.com/lucasew/env-mcp/internal/runner/host"
	"github.com/lucasew/env-mcp/internal/server"
)

// Root is the env-mcp command line.
type Root struct {
	Env   *EnvCmd   `cmd:"" help:"manage envs"`
	Exec  *ExecCmd  `cmd:"" help:"run a command in an env"`
	Logs  *LogsCmd  `cmd:"" help:"read an exec event stream"`
	Poll  *PollCmd  `cmd:"" help:"report one exec state"`
	Kill  *KillCmd  `cmd:"" help:"stop one exec"`
	Serve *ServeCmd `cmd:"" help:"run the MCP transport"`
	cmd.Output
}

// Description is the program summary.
func (Root) Description() string { return "manage worktree envs and their commands" }

// Run prints usage when no subcommand is selected.
func (r *Root) Run(context.Context) error { return cmd.ErrUsage }

// EnvCmd groups the env subcommands.
type EnvCmd struct {
	Get   *EnvGetCmd   `cmd:"" help:"open or resume an env"`
	List  *EnvListCmd  `cmd:"" help:"list envs"`
	Close *EnvCloseCmd `cmd:"" help:"close an env"`
}

// EnvGetCmd opens or resumes one env.
type EnvGetCmd struct {
	Repo cmd.StringArg `arg:"" help:"github:owner/repo@base"`
	Slug cmd.StringArg `arg:"" help:"unit of work and branch name"`
}

// Run ensures the env.
func (c EnvGetCmd) Run(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	view, err := srv.EnvGet(ctx, c.Repo.Value(), c.Slug.Value())
	if err != nil {
		return err
	}
	return printJSON(os.Stdout, view)
}

// EnvListCmd lists envs.
type EnvListCmd struct{}

// Run prints every env.
func (c EnvListCmd) Run(context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	views, err := srv.EnvList()
	if err != nil {
		return err
	}
	return printJSON(os.Stdout, views)
}

// EnvCloseCmd closes one env.
type EnvCloseCmd struct {
	Handle cmd.StringArg `arg:"" help:"env handle"`
}

// Run closes the env.
func (c EnvCloseCmd) Run(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	return srv.EnvClose(ctx, c.Handle.Value())
}

func printJSON(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// StateDir is the directory the server owns. ENV_MCP_STATE overrides it.
func StateDir() (string, error) {
	if dir := os.Getenv("ENV_MCP_STATE"); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "env-mcp"), nil
}

// Open builds the server from the state directory.
func Open() (*server.Server, error) {
	state, err := StateDir()
	if err != nil {
		return nil, err
	}
	store, err := env.NewStore(filepath.Join(state, "envs"))
	if err != nil {
		return nil, err
	}
	ConfigureRunners(state)
	return server.NewServer(store, server.DefaultBackend(state)), nil
}

// ConfigureRunners points every registered executor at the state directory.
// Executors keep their own records there, so a new CLI process resumes the
// same surface.
func ConfigureRunners(state string) {
	host.StateDir = state
	herdr_runner.StateDir = state
}
