// Package server runs the operation table against envs.
package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner"
)

// Server owns the env store, the backend that checks out envs, and the
// executor that runs their commands.
type Server struct {
	Store    *env.Store
	Backend  *env.Backend
	StateDir string
	// Runner overrides the executor the registry would pick. Tests set it;
	// production leaves it nil so the driver registry decides.
	Runner runner.Executor

	mu sync.Mutex
}

// NewServer builds a server that keeps its exec state in stateDir.
func NewServer(store *env.Store, backend *env.Backend, stateDir string) *Server {
	return &Server{
		Store:    store,
		Backend:  backend,
		StateDir: stateDir,
	}
}

// EnvGet ensures the env for repo and slug and returns its view.
func (s *Server) EnvGet(ctx context.Context, repoText, slugText string) (EnvView, error) {
	repo, err := env.ParseRepo(repoText)
	if err != nil {
		return EnvView{}, err
	}
	slug := env.Slug(slugText)
	if err := slug.Validate(); err != nil {
		return EnvView{}, err
	}
	handle := env.NewHandle(repo, slug)
	rec, err := s.Store.Get(handle)
	if err != nil {
		return EnvView{}, err
	}
	if rec != nil && rec.Closed {
		return EnvView{}, fmt.Errorf("%w: %s", env.ErrClosed, handle)
	}
	created := false
	if rec == nil {
		root, made, err := s.Backend.Ensure(ctx, repo, slug)
		if err != nil {
			return EnvView{}, err
		}
		created = made
		source, _ := s.Backend.Checkout(repo)
		rec = &env.Record{
			Handle:  handle,
			Repo:    repo.String(),
			Slug:    string(slug),
			Base:    repo.Base,
			Root:    root,
			Source:  source,
			Created: made,
		}
		if err := s.Store.Put(*rec); err != nil {
			return EnvView{}, err
		}
	}
	// Opening the env opens its surface: one workspace, one agent shell,
	// and the place every exec of this env will use.
	if _, err := s.openSurface(ctx, handle, rec.Root, rec.Source); err != nil {
		return EnvView{}, err
	}
	status := "resumed"
	if created {
		status = "created"
	}
	return EnvView{Handle: handle, Repo: rec.Repo, Slug: rec.Slug, Base: rec.Base, Status: status}, nil
}

// EnvView is what a caller sees of one env. It has no checkout path.
type EnvView struct {
	Handle env.Handle `json:"handle"`
	Repo   string     `json:"repo"`
	Slug   string     `json:"slug"`
	Base   string     `json:"base"`
	Status string     `json:"status"`
}

// EnvList returns every env the store knows.
func (s *Server) EnvList() ([]EnvView, error) {
	records, err := s.Store.List()
	if err != nil {
		return nil, err
	}
	out := make([]EnvView, 0, len(records))
	for _, rec := range records {
		status := "open"
		if rec.Closed {
			status = "closed"
		}
		out = append(out, EnvView{
			Handle: rec.Handle,
			Repo:   rec.Repo,
			Slug:   rec.Slug,
			Base:   rec.Base,
			Status: status,
		})
	}
	return out, nil
}

// EnvClose closes one env and frees its handle.
func (s *Server) EnvClose(ctx context.Context, handleText string) error {
	handle := env.Handle(handleText)
	rec, err := s.Store.Get(handle)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("%w: %s", env.ErrHandle, handle)
	}
	if rec.Closed {
		return fmt.Errorf("%w: %s", env.ErrClosed, handle)
	}
	if err := s.KillEnv(ctx, handle); err != nil {
		return err
	}
	if surface, ok := s.peekSurface(ctx, rec.Root); ok && surface != "" {
		if r, err := s.runner(); err == nil {
			_ = r.Close(ctx, surface)
		}
	}
	rec.Closed = true
	return s.Store.Put(*rec)
}
