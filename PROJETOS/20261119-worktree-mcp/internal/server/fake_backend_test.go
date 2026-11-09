package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/lucasew/env-mcp/internal/backend"
)

// fakeProvisioner is an in-memory backend for tests. It records what the
// server asked for and hands back opaque handles, never paths.
type fakeProvisioner struct {
	mu         sync.Mutex
	provisions int
	envs       map[string]*fakeEnv
	closed     int
}

func newFake() *fakeProvisioner {
	return &fakeProvisioner{envs: map[string]*fakeEnv{}}
}

func (f *fakeProvisioner) ID() string { return "fake" }

func (f *fakeProvisioner) Provision(_ context.Context, req backend.Request) (backend.Provisioned, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.provisions++
	token := fmt.Sprintf("%s-%s-%s-%s", req.Owner, req.Name, req.Base, req.Slug)
	_, seen := f.envs[token]
	if !seen {
		f.envs[token] = &fakeEnv{id: "fake", token: token, ops: map[string]*fakeOp{}}
	}
	return backend.Provisioned{
		Handle:  backend.Handle{Backend: "fake", Token: token},
		Created: !seen,
	}, nil
}

func (f *fakeProvisioner) Open(_ context.Context, h backend.Handle, name string) (backend.Env, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.envs[h.Token]
	if !ok {
		return nil, fmt.Errorf("%w: %s", backend.ErrEnv, h.Token)
	}
	if e.closed {
		return nil, fmt.Errorf("%w: %s", backend.ErrClosed, h.Token)
	}
	e.name = name
	return e, nil
}

func (f *fakeProvisioner) counts() (provisions, closed int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.provisions, f.closed
}

// fakeEnv is one env the fake provisioner handed out.
type fakeEnv struct {
	mu     sync.Mutex
	id     string
	name   string
	token  string
	ops    map[string]*fakeOp
	next   int
	closed bool
}

// fakeOp is one command started in a fake env.
type fakeOp struct {
	argv   []string
	dir    string
	lines  []string
	killed bool
}

func (e *fakeEnv) Exec(_ context.Context, start backend.Start) (backend.Exec, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.next++
	op := fmt.Sprintf("%d", e.next)
	e.ops[op] = &fakeOp{argv: start.Argv, dir: start.Dir}
	return backend.Exec{Env: e.name, Op: op}, nil
}

func (e *fakeEnv) note(id backend.Exec) (*fakeOp, error) {
	if !id.OwnedBy(e.name) {
		return nil, fmt.Errorf("%w: %s: not an op in this env", backend.ErrExec, id)
	}
	op, ok := e.ops[id.Op]
	if !ok {
		return nil, fmt.Errorf("%w: %s", backend.ErrExec, id)
	}
	return op, nil
}

func (e *fakeEnv) Status(_ context.Context, id backend.Exec) (backend.Status, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	op, err := e.note(id)
	if err != nil {
		return backend.Status{}, err
	}
	if op.killed {
		return backend.Status{State: backend.Exited}, nil
	}
	return backend.Status{State: backend.Running}, nil
}

func (e *fakeEnv) Read(_ context.Context, id backend.Exec, read backend.Read) (backend.Page, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	op, err := e.note(id)
	if err != nil {
		return backend.Page{}, err
	}
	page := backend.Page{Lines: []string{}, NextOffset: read.Offset}
	for n := read.Offset; n < len(op.lines); n++ {
		page.Lines = append(page.Lines, op.lines[n])
		page.NextOffset = n + 1
	}
	return page, nil
}

func (e *fakeEnv) Kill(_ context.Context, id backend.Exec) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	op, err := e.note(id)
	if err != nil {
		return err
	}
	op.killed = true
	return nil
}

func (e *fakeEnv) KillAll(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, op := range e.ops {
		op.killed = true
	}
	return nil
}

func (e *fakeEnv) Close(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}

// started returns the commands this env started, in order.
func (e *fakeEnv) started() []*fakeOp {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]*fakeOp, 0, len(e.ops))
	for n := 1; n <= len(e.ops); n++ {
		out = append(out, e.ops[fmt.Sprintf("%d", n)])
	}
	return out
}
