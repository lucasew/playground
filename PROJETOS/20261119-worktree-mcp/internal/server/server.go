// Package server runs the operation table against envs.
//
// The server owns the registry: which envs exist, what repository and slug
// each one stands for, and the handle its backend issued. It never learns
// where an env lives. Provisioning, running commands, and teardown all go
// through [backend.Handle] values that the backend interprets.
package server

import (
	"context"
	"fmt"

	"github.com/lucasew/env-mcp/internal/backend"
	"github.com/lucasew/env-mcp/internal/env"
)

// Server owns the env registry and the provisioner it delegates to.
type Server struct {
	// Store keeps the registry records. It never holds a checkout path.
	Store *env.Store
	// Backend provisions envs and runs their commands.
	Backend backend.Provisioner
}

// NewServer builds a server over a registry and a provisioner.
func NewServer(store *env.Store, provisioner backend.Provisioner) *Server {
	return &Server{Store: store, Backend: provisioner}
}

// EnvGet ensures the env for repo and slug and returns its view.
//
// The handle names the env, not its location. A handle the server already
// recorded is re-provisioned, so an env that disappeared underneath the
// backend comes back on the same handle.
func (s *Server) EnvGet(ctx context.Context, repoText, slugText string) (EnvView, error) {
	repo, err := env.ParseRepo(repoText)
	if err != nil {
		return EnvView{}, err
	}
	slug := env.Slug(slugText)
	if err := slug.Validate(); err != nil {
		return EnvView{}, err
	}
	rec, err := s.Store.Find(repo.String(), string(slug))
	if err != nil {
		return EnvView{}, err
	}
	handle := env.Handle("")
	if rec != nil {
		handle = rec.Handle
	}
	if rec != nil && rec.Closed {
		return EnvView{}, fmt.Errorf("%w: %s", env.ErrClosed, handle)
	}
	p, err := s.Backend.Provision(ctx, backend.Request{
		Owner: repo.Owner,
		Name:  repo.Name,
		Base:  repo.Base,
		Slug:  string(slug),
		Label: label(repo, slug),
	})
	if err != nil {
		return EnvView{}, err
	}
	// Provisioning is idempotent, so the backend always answers with the
	// env's handle. A record whose token changes was re-provisioned.
	if rec == nil {
		rec = &env.Record{}
		if handle, err = env.MintHandle(); err != nil {
			return EnvView{}, err
		}
	}
	plugged := rec.Token == p.Handle.Token
	rec.Handle = handle
	rec.Repo = repo.String()
	rec.Slug = string(slug)
	rec.Base = repo.Base
	rec.Token = p.Handle.Token
	rec.Backend = p.Handle.Backend
	if err := s.Store.Put(*rec); err != nil {
		return EnvView{}, err
	}
	status := "resumed"
	if !plugged || p.Created {
		status = "created"
	}
	return EnvView{Handle: handle, Repo: rec.Repo, Slug: rec.Slug, Base: rec.Base, Status: status}, nil
}

// label names an env for a human watching the backend's surface. The handle
// is short on purpose, so it says nothing about the work; the label does.
func label(repo env.Repo, slug env.Slug) string {
	return repo.Owner + "/" + repo.Name + ":" + string(slug)
}

// EnvView is what a caller sees of one env. It has no checkout path.
type EnvView struct {
	Handle env.Handle `json:"handle"`
	Repo   string     `json:"repo"`
	Slug   string     `json:"slug"`
	Base   string     `json:"base"`
	Status string     `json:"status"`
}

// EnvList returns every env the registry knows.
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

// EnvClose closes one env: every live command stops and the backend tears
// down everything it provisioned for it.
func (s *Server) EnvClose(ctx context.Context, handleText string) error {
	handle, err := env.ParseHandle(handleText)
	if err != nil {
		return err
	}
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
	e, err := s.Backend.Open(ctx, backend.Handle{Backend: rec.Backend, Token: rec.Token}, string(rec.Handle))
	if err != nil {
		return err
	}
	if err := e.Close(ctx); err != nil {
		return err
	}
	rec.Closed = true
	return s.Store.Put(*rec)
}

// open returns the backend env behind a registry record.
func (s *Server) open(ctx context.Context, handle env.Handle) (backend.Env, *env.Record, error) {
	rec, err := s.Store.Get(handle)
	if err != nil {
		return nil, nil, err
	}
	if rec == nil {
		return nil, nil, fmt.Errorf("%w: %s", env.ErrHandle, handle)
	}
	if rec.Closed {
		return nil, nil, fmt.Errorf("%w: %s", env.ErrClosed, handle)
	}
	e, err := s.Backend.Open(ctx, backend.Handle{Backend: rec.Backend, Token: rec.Token}, string(rec.Handle))
	if err != nil {
		return nil, nil, err
	}
	return e, rec, nil
}
