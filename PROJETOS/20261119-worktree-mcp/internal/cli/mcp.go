package cli

import (
	"context"
	"os"

	"github.com/lucasew/env-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Serve runs the MCP server over stdio until the client disconnects.
func Serve(ctx context.Context) error {
	srv, err := Open()
	if err != nil {
		return err
	}
	m := mcp.NewServer(&mcp.Implementation{Name: "env-mcp", Version: "0.0.0"}, nil)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "env_get",
		Description: "Open or resume the env for a repo and slug. Idempotent. Returns the env handle.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in envGetIn) (*mcp.CallToolResult, server.EnvView, error) {
		view, err := srv.EnvGet(ctx, in.Repo, in.Slug)
		return nil, view, err
	})
	mcp.AddTool(m, &mcp.Tool{
		Name:        "env_list",
		Description: "List every env this server knows.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, []server.EnvView, error) {
		views, err := srv.EnvList()
		return nil, views, err
	})
	mcp.AddTool(m, &mcp.Tool{
		Name:        "env_close",
		Description: "Close an env and free its handle.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in envCloseIn) (*mcp.CallToolResult, struct{}, error) {
		return nil, struct{}{}, srv.EnvClose(ctx, in.Env)
	})
	mcp.AddTool(m, &mcp.Tool{
		Name:        "exec",
		Description: "Start a command in an env and return its handle immediately.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in execIn) (*mcp.CallToolResult, server.ExecView, error) {
		view, err := srv.Exec(ctx, in.Env, in.Cwd, in.Argv)
		return nil, view, err
	})
	mcp.AddTool(m, &mcp.Tool{
		Name:        "logs",
		Description: "Read lines of an exec event stream from a line offset, filtered by RE2.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in logsIn) (*mcp.CallToolResult, pageOut, error) {
		page, err := srv.Logs(ctx, in.Exec, in.Offset, in.Limit, in.Grep)
		return nil, pageOut{Page: page}, err
	})
	mcp.AddTool(m, &mcp.Tool{
		Name:        "poll",
		Description: "Report one exec state and its stream length without reading output.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in pollIn) (*mcp.CallToolResult, server.ExecState, error) {
		state, err := srv.Poll(ctx, in.Exec)
		return nil, state, err
	})
	mcp.AddTool(m, &mcp.Tool{
		Name:        "kill",
		Description: "Stop one exec. Killing an exited exec succeeds.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in killIn) (*mcp.CallToolResult, struct{}, error) {
		return nil, struct{}{}, srv.Kill(ctx, in.Exec)
	})

	return m.Run(ctx, &mcp.StdioTransport{})
}

type envGetIn struct {
	Repo string `json:"repo" jsonschema:"repo identity as github:owner/repo@base"`
	Slug string `json:"slug" jsonschema:"unit of work and branch name"`
}

type envCloseIn struct {
	Env string `json:"env" jsonschema:"env handle"`
}

type execIn struct {
	Env  string   `json:"env" jsonschema:"env handle"`
	Argv []string `json:"argv" jsonschema:"command and arguments"`
	Cwd  string   `json:"cwd,omitempty" jsonschema:"working directory relative to the env root"`
}

type killIn struct {
	Exec string `json:"exec" jsonschema:"exec handle"`
}

type logsIn struct {
	Exec   string `json:"exec" jsonschema:"exec handle"`
	Offset int    `json:"offset" jsonschema:"first line to read"`
	Limit  int    `json:"limit,omitempty" jsonschema:"most lines to return"`
	Grep   string `json:"grep,omitempty" jsonschema:"RE2 pattern"`
}

type pollIn struct {
	Exec string `json:"exec" jsonschema:"exec handle"`
}

type pageOut struct {
	Page any `json:"page"`
}

var _ = os.Stdout
