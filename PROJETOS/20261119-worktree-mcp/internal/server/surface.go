package server

import (
	"context"

	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner"
)

// openSurface ensures the executor surface for an env and returns its place.
func (s *Server) openSurface(ctx context.Context, handle env.Handle, root, source string) (runner.Place, error) {
	r, err := s.runner()
	if err != nil {
		return "", err
	}
	return r.Open(ctx, runner.Env{Checkout: root, Source: source, Label: string(handle)})
}

// peekSurface reports the place the executor already holds for a checkout.
// It never opens a new surface, so closing an env cannot create one.
func (s *Server) peekSurface(ctx context.Context, root string) (runner.Place, bool) {
	r, err := s.runner()
	if err != nil {
		return "", false
	}
	peek, ok := r.(interface {
		Peek(context.Context, string) (runner.Place, bool)
	})
	if !ok {
		return "", false
	}
	return peek.Peek(ctx, root)
}
