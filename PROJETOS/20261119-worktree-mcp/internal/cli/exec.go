package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/lewtec/lewkit/x/cmd"
)

// ExecCmd runs one command in an env and prints its handle.
type ExecCmd struct {
	Env  cmd.StringArg   `arg:"" help:"env handle"`
	Cwd  cmd.StringArg   `long:"cwd" default:"" help:"working directory relative to the env root"`
	Argv []cmd.StringArg `arg:"" help:"command and arguments"`
}

// Run starts the command and returns immediately.
func (c ExecCmd) Run(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	argv := make([]string, len(c.Argv))
	for i, a := range c.Argv {
		argv[i] = a.Value()
	}
	handle, err := srv.Exec(ctx, c.Env.Value(), c.Cwd.Value(), argv)
	if err != nil {
		return err
	}
	return printJSON(os.Stdout, handle)
}

// LogsCmd reads an event stream from a line offset.
type LogsCmd struct {
	Exec    cmd.StringArg   `arg:"" help:"exec handle"`
	Offset  cmd.IntArg[int] `long:"offset" default:"0" help:"first line to read"`
	Limit   cmd.IntArg[int] `long:"limit" default:"200" help:"most lines to return"`
	Pattern cmd.StringArg   `long:"grep" default:"" help:"RE2 pattern"`
}

// Run prints the page.
func (c LogsCmd) Run(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	page, err := srv.Logs(ctx, c.Exec.Value(), c.Offset.Value(), c.Limit.Value(), c.Pattern.Value())
	if err != nil {
		return err
	}
	for _, line := range page.Lines {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			return err
		}
	}
	return nil
}

// PollCmd reports one exec state without reading output.
type PollCmd struct {
	Exec cmd.StringArg `arg:"" help:"exec handle"`
}

// Run prints the state.
func (c PollCmd) Run(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	state, err := srv.Poll(ctx, c.Exec.Value())
	if err != nil {
		return err
	}
	return printJSON(os.Stdout, state)
}

// KillCmd stops one exec.
type KillCmd struct {
	Exec cmd.StringArg `arg:"" help:"exec handle"`
}

// Run stops the exec. Killing an exited exec succeeds.
func (c KillCmd) Run(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	if err := srv.Kill(ctx, c.Exec.Value()); err != nil {
		return err
	}
	_, err = fmt.Println("killed", c.Exec.Value())
	return err
}
