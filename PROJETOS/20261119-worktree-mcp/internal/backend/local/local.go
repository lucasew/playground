// Package local provisions envs on this machine.
//
// It keeps one clone per repository and one linked worktree per env under a
// single store directory, and runs each env's commands on a terminal
// surface: Herdr when a session is reachable, plain processes otherwise.
// Herdr is an implementation detail here; nothing above this package knows
// about it, and no path leaves it.
package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/env-mcp/internal/backend"
	"github.com/lucasew/env-mcp/internal/backend/local/store"
	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner"
)

// Backend provisions envs on the local filesystem.
type Backend struct {
	// Store holds the clones and worktrees. Only this backend reads it.
	Store *store.Store
	// State is where the backend remembers the envs it provisioned.
	State string
	// Executors overrides the registry. Tests set it; production leaves it
	// nil so the driver registry picks the terminal surface.
	Executors runner.Executor
}

// New returns a local backend whose clones, worktrees, and records all live
// under base.
func New(base string) *Backend {
	return &Backend{Store: store.New(base, ""), State: base}
}

// ID is the registry name stored with each handle.
func (b *Backend) ID() string { return "local" }

// note is the backend's private record of one env. It holds the checkout and
// the surface, neither of which the server ever sees.
type note struct {
	Repo     string       `json:"repo"`
	Slug     string       `json:"slug"`
	Checkout string       `json:"checkout"`
	Source   string       `json:"source"`
	Surface  runner.Place `json:"surface"`
	Executor string       `json:"executor"`
	Closed   bool         `json:"closed"`
}

// Provision ensures the clone, the worktree, and the surface for one env.
func (b *Backend) Provision(ctx context.Context, req backend.Request) (backend.Provisioned, error) {
	repo := store.Repo{Owner: req.Owner, Name: req.Name, Base: req.Base}
	main, cloned, err := b.Store.Clone(ctx, repo)
	if err != nil {
		return backend.Provisioned{}, err
	}
	// Provisioning one env again resumes it; it must not adopt a worktree
	// that already belongs to another env of the same branch.
	token := key(req)
	if n, ok, err := b.read(token); err != nil {
		return backend.Provisioned{}, err
	} else if ok && !n.Closed {
		if _, err := os.Stat(n.Checkout); err == nil {
			return backend.Provisioned{Handle: backend.Handle{Backend: b.ID(), Token: token}, Created: false}, nil
		}
	}
	checkout, made, err := b.Store.Checkout(ctx, repo, main, req.Slug)
	if err != nil {
		return backend.Provisioned{}, err
	}
	exec, err := b.executor(ctx)
	if err != nil {
		return backend.Provisioned{}, err
	}
	surface, err := exec.Open(ctx, runner.Env{Checkout: checkout, Source: main, Label: req.Label})
	if err != nil {
		return backend.Provisioned{}, err
	}
	n := note{
		Repo:     req.Owner + "/" + req.Name,
		Slug:     req.Slug,
		Checkout: checkout,
		Source:   main,
		Surface:  surface,
		Executor: exec.ID(),
	}
	if err := b.write(token, n); err != nil {
		return backend.Provisioned{}, err
	}
	return backend.Provisioned{
		Handle:  backend.Handle{Backend: b.ID(), Token: token},
		Created: cloned || made,
	}, nil
}

// Open returns the env behind a handle this backend issued.
func (b *Backend) Open(ctx context.Context, h backend.Handle) (backend.Env, error) {
	if h.Backend != b.ID() {
		return nil, fmt.Errorf("%w: %s: not a local handle", backend.ErrEnv, h.Token)
	}
	n, ok, err := b.read(h.Token)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%w: %s", backend.ErrEnv, h.Token)
	}
	if n.Closed {
		return nil, fmt.Errorf("%w: %s", backend.ErrClosed, h.Token)
	}
	exec, err := b.executor(ctx)
	if err != nil {
		return nil, err
	}
	return &envHandle{backend: b, note: n, exec: exec, token: h.Token}, nil
}

// envHandle is the addressable env the server interacts with.
//
// Everything it does goes through the surface the backend opened at
// provision time. The server holds this value only for the life of a request.
type envHandle struct {
	backend *Backend
	note    note
	exec    runner.Executor
	token   string
}

func (e *envHandle) Exec(ctx context.Context, start backend.Start) (backend.Exec, error) {
	dir, err := env.Resolve(e.note.Checkout, start.Dir)
	if err != nil {
		return backend.Exec{}, err
	}
	live, err := e.live(ctx)
	if err != nil {
		return backend.Exec{}, err
	}
	if live >= env.MaxExecs {
		return backend.Exec{}, fmt.Errorf("%w: %d", backend.ErrLimit, env.MaxExecs)
	}
	id, err := e.nextExec()
	if err != nil {
		return backend.Exec{}, err
	}
	place, err := e.exec.Start(ctx, e.note.Surface, runner.Start{
		Dir:   dir,
		Argv:  start.Argv,
		Label: start.Label,
	})
	if err != nil {
		return backend.Exec{}, err
	}
	if err := e.putExec(execNote{ID: id, Place: place}); err != nil {
		return backend.Exec{}, err
	}
	return backend.Exec{Env: e.token, Op: id}, nil
}

func (e *envHandle) Status(ctx context.Context, id backend.Exec) (backend.Status, error) {
	rec, err := e.execNote(id)
	if err != nil {
		return backend.Status{}, err
	}
	if rec.Killed {
		return backend.Status{State: backend.Exited}, nil
	}
	status, err := e.exec.Status(ctx, rec.Place)
	if err != nil {
		return backend.Status{}, err
	}
	return backend.Status{State: backend.State(status.State), ExitCode: status.ExitCode}, nil
}

func (e *envHandle) Read(ctx context.Context, id backend.Exec, read backend.Read) (backend.Page, error) {
	rec, err := e.execNote(id)
	if err != nil {
		return backend.Page{}, err
	}
	if rec.Killed {
		return backend.Page{Lines: []string{}, NextOffset: read.Offset}, nil
	}
	page, err := e.exec.Read(ctx, rec.Place, runner.Output{
		Offset:  read.Offset,
		Limit:   read.Limit,
		Pattern: read.Pattern,
	})
	if err != nil {
		return backend.Page{}, mapErr(err)
	}
	return backend.Page{Lines: page.Lines, NextOffset: page.NextOffset}, nil
}

func (e *envHandle) Kill(ctx context.Context, id backend.Exec) error {
	rec, err := e.execNote(id)
	if err != nil {
		return err
	}
	if rec.Killed {
		return nil
	}
	if status, err := e.exec.Status(ctx, rec.Place); err == nil && status.State == runner.Exited {
		return nil
	}
	if err := e.exec.Kill(ctx, rec.Place); err != nil {
		return err
	}
	rec.Killed = true
	return e.putExec(rec)
}

func (e *envHandle) KillAll(ctx context.Context) error {
	ids, err := e.execIDs()
	if err != nil {
		return err
	}
	for _, id := range ids {
		_ = e.Kill(ctx, backend.Exec{Env: e.token, Op: id})
	}
	return nil
}

func (e *envHandle) Close(ctx context.Context) error {
	if err := e.KillAll(ctx); err != nil {
		return err
	}
	_ = e.exec.Close(ctx, e.note.Surface)
	// The worktree goes; the clone stays, since other envs may use it.
	if err := e.backend.Store.Release(ctx, e.note.Source, e.note.Checkout); err != nil {
		return err
	}
	closed := e.note
	closed.Closed = true
	return e.backend.write(e.token, closed)
}

// mapErr maps a runner error onto the backend sentinels the server matches.
func mapErr(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, runner.ErrPattern.Error()):
		return fmt.Errorf("%w: %w", backend.ErrPattern, err)
	case strings.Contains(msg, runner.ErrOffset.Error()):
		return fmt.Errorf("%w: %w", backend.ErrOffset, err)
	default:
		return err
	}
}

func (b *Backend) executor(ctx context.Context) (runner.Executor, error) {
	if b.Executors != nil {
		return b.Executors, nil
	}
	return runner.Open(ctx)
}

// key names one env inside this backend: the repository, its base, and the
// slug. The base is part of the name, so two bases of one repo are two envs.
func key(req backend.Request) string {
	return req.Owner + "-" + req.Name + "-" + store.SafeName(req.Base) + "-" + store.SafeName(req.Slug)
}

func (b *Backend) noteFile(token string) string {
	return filepath.Join(b.State, "local", "envs", token+".json")
}

// read returns the note for one env and whether it exists.
func (b *Backend) read(token string) (note, bool, error) {
	buf, err := os.ReadFile(b.noteFile(token))
	if os.IsNotExist(err) {
		return note{}, false, nil
	}
	if err != nil {
		return note{}, false, err
	}
	var n note
	if err := json.Unmarshal(buf, &n); err != nil {
		return note{}, false, fmt.Errorf("local: %s: %w", token, err)
	}
	return n, true, nil
}

func (b *Backend) write(token string, n note) error {
	if err := os.MkdirAll(filepath.Dir(b.noteFile(token)), 0o700); err != nil {
		return err
	}
	buf, err := json.Marshal(n)
	if err != nil {
		return err
	}
	return os.WriteFile(b.noteFile(token), buf, 0o600)
}

// execNote is the backend's record of one command it started.
type execNote struct {
	ID     string       `json:"id"`
	Place  runner.Place `json:"place"`
	Killed bool         `json:"killed"`
}

func (e *envHandle) execDir() string {
	return filepath.Join(e.backend.State, "local", "execs", e.token)
}

func (e *envHandle) execFile(id string) string {
	return filepath.Join(e.execDir(), id+".json")
}

func (e *envHandle) execNote(id backend.Exec) (execNote, error) {
	if !id.OwnedBy(e.token) {
		return execNote{}, fmt.Errorf("%w: %s: not an op in this env", backend.ErrExec, id)
	}
	buf, err := os.ReadFile(e.execFile(id.Op))
	if os.IsNotExist(err) {
		return execNote{}, fmt.Errorf("%w: %s", backend.ErrExec, id)
	}
	if err != nil {
		return execNote{}, err
	}
	var rec execNote
	if err := json.Unmarshal(buf, &rec); err != nil {
		return execNote{}, err
	}
	return rec, nil
}

func (e *envHandle) putExec(rec execNote) error {
	if err := os.MkdirAll(e.execDir(), 0o700); err != nil {
		return err
	}
	buf, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(e.execFile(rec.ID), buf, 0o600)
}

// nextExec mints the next command id in this env. The id is claimed before
// the command starts, so two processes never mint the same one.
func (e *envHandle) nextExec() (string, error) {
	if err := os.MkdirAll(e.execDir(), 0o700); err != nil {
		return "", err
	}
	for n := 1; ; n++ {
		id := fmt.Sprintf("%d", n)
		if _, err := os.Stat(e.execFile(id)); err == nil {
			continue
		}
		return id, nil
	}
}

func (e *envHandle) execIDs() ([]string, error) {
	entries, err := os.ReadDir(e.execDir())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			ids = append(ids, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	return ids, nil
}

// live counts the commands still running in this env.
func (e *envHandle) live(ctx context.Context) (int, error) {
	ids, err := e.execIDs()
	if err != nil {
		return 0, err
	}
	live := 0
	for _, id := range ids {
		rec, err := e.execNote(backend.Exec{Env: e.token, Op: id})
		if err != nil || rec.Killed {
			continue
		}
		if status, err := e.exec.Status(ctx, rec.Place); err == nil && status.State == runner.Running {
			live++
		}
	}
	return live, nil
}
