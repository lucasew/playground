package cli

import "context"

// ServeCmd runs the MCP transport on stdin and stdout.
type ServeCmd struct{}

// Run starts the MCP server.
func (c ServeCmd) Run(ctx context.Context) error {
	return Serve(ctx)
}
