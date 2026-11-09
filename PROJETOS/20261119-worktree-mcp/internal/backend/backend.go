// Package backend is the provisioner contract the server owns.
//
// The server asks a backend for an env and receives a handle. Every later
// interaction — running a command, reading its output, stopping it, closing
// the env — goes through that handle. The server never learns where the env
// lives, what runs its commands, or which backend provisioned it.
//
// A local backend keeps clones and worktrees under one directory and runs
// commands in a terminal multiplexer. A sandbox backend would provision the
// same env somewhere else. Both answer the same calls.
package backend

import (
	"context"

	"github.com/lewtec/lewkit/x/driver"
	"github.com/lewtec/lewkit/x/errs"
)

var (
	// ErrEnv means the handle does not name an env this backend knows.
	ErrEnv = errs.Wrap(errs.ErrInvalid, "env")
	// ErrExec means the handle does not name a command this env started.
	ErrExec = errs.Wrap(errs.ErrNotFound, "exec")
	// ErrPattern means the RE2 pattern does not compile.
	ErrPattern = errs.Wrap(errs.ErrSyntax, "pattern")
	// ErrOffset means the offset is past what the backend can still reach.
	ErrOffset = errs.Wrap(errs.ErrOutOfRange, "offset")
	// ErrClosed means the handle names an env that was closed.
	ErrClosed = errs.Wrap(errs.ErrClosed, "env")
	// ErrLimit means the env already holds its maximum live commands.
	ErrLimit = errs.Wrap(errs.ErrUnavailable, "exec limit")
)

// Request is one env the server asks a backend to provision.
type Request struct {
	// Owner, Name, and Base identify the repository.
	Owner string
	Name  string
	Base  string
	// Slug is the unit of work, and the branch to check out.
	Slug string
	// Label names the env for a human watching the backend's surface.
	Label string
}

// Handle addresses one env inside a backend.
//
// It is a value the server can persist and hand back in a later process. Its
// contents are the backend's business: the server stores it verbatim and
// never interprets it.
type Handle struct {
	// Backend is the provisioner ID, so a later process can reach the same
	// implementation.
	Backend string `json:"backend"`
	// Token is the backend's own address for this env.
	Token string `json:"token"`
}

// Provisioned is what a backend returns for one request.
type Provisioned struct {
	Handle Handle
	// Created reports whether provisioning made something new.
	Created bool
}

// State is the lifecycle of one command.
type State string

const (
	Running State = "running"
	Exited  State = "exited"
)

// Start describes one command to start in an env.
type Start struct {
	// Dir is the working directory relative to the env root.
	Dir string
	// Argv is the command and its arguments. Argv[0] is non-empty.
	Argv []string
	// Label names the command for a human watching the surface.
	Label string
}

// Exec addresses one command inside one env.
//
// An op handle is scoped to the env that started it: the value carries the
// env's handle, so the same handle can never be replayed against another
// env. The server stores the value verbatim and never builds one itself.
//
// The two levels are explicit in the type: Env names an env, Op names a
// command inside that env. A handle is short because an agent repeats it in
// every call.
type Exec struct {
	// Env is the handle of the env that started this command.
	Env string `json:"env"`
	// Op is the backend's own name for this command.
	Op string `json:"op"`
}

// String is the stable form the server prints and matches errors against.
func (e Exec) String() string { return e.Env + "." + e.Op }

// OwnedBy reports whether this op belongs to the env with the given handle.
// The backend refuses any op handle from another env.
func (e Exec) OwnedBy(handle string) bool { return e.Env == handle }

// Status is what a backend reports about one command.
type Status struct {
	State    State
	ExitCode int
}

// Page is one read of a command's output, starting at a line offset.
type Page struct {
	Lines      []string
	NextOffset int
}

// Read asks for whole lines from an offset. Pattern is an RE2 expression;
// empty matches every line. The cursor advances past every scanned line,
// whether it matched.
type Read struct {
	Offset  int
	Limit   int
	Pattern string
}

// Env is one provisioned env as the server addresses it.
//
// Everything a caller does with an env goes through this value. It is the
// whole public surface of a provisioner after provisioning, so no checkout
// path and no terminal identifier ever reaches an agent.
type Env interface {
	// Exec starts one command and returns its handle immediately.
	Exec(ctx context.Context, start Start) (Exec, error)
	// Status reports one command's lifecycle. A handle for another env, or
	// for a command this env never started, fails with ErrExec.
	Status(ctx context.Context, id Exec) (Status, error)
	// Read returns lines of one command's output.
	Read(ctx context.Context, id Exec, read Read) (Page, error)
	// Kill stops one command. Killing an exited command succeeds.
	Kill(ctx context.Context, id Exec) error
	// KillAll stops every live command. It is what closing does first.
	KillAll(ctx context.Context) error
	// Close tears down the env and frees its handle.
	Close(ctx context.Context) error
}

// Provisioner makes envs real and hands back a handle for each one.
//
// Implementations must be safe to use from several processes: the server may
// be a short-lived CLI process each time, so nothing may live only in
// memory. Open must be callable with a handle a previous process stored.
type Provisioner interface {
	// ID is the registry name stored with each handle.
	ID() string
	// Provision ensures the env for one request and returns its handle.
	Provision(ctx context.Context, req Request) (Provisioned, error)
	// Open returns the env behind a handle this backend issued. It fails
	// with ErrEnv when the handle is unknown and ErrClosed when it was
	// closed.
	//
	// name is the short handle the server shows an agent. Op handles are
	// minted from it, so an agent never sees the backend's own token.
	Open(ctx context.Context, h Handle, name string) (Env, error)
}

// Open builds the provisioner for this host.
//
// It reads the lewkit driver registry, so a backend registered with a higher
// weight wins. A build that lacks a backend fails here rather than halfway
// through provisioning an env.
func Open(ctx context.Context) (Provisioner, error) {
	p, err := driver.Get[Provisioner](ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}
