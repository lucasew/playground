package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/lucasew/env-mcp/internal/runner"
)

// fakeRunner is an in-memory executor for tests. It records what the server
// asked for and reports every command as running with fixed output.
type fakeRunner struct {
	mu      sync.Mutex
	opened  int
	closed  int
	started []runner.Start
	lines   []string
	killed  int
}

func (f *fakeRunner) ID() string { return "fake" }

func (f *fakeRunner) Open(context.Context, runner.Env) (runner.Place, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.opened++
	return runner.Place(fmt.Sprintf("surface-%d", f.opened)), nil
}

func (f *fakeRunner) Peek(context.Context, string) (runner.Place, bool) {
	return runner.Place("surface-1"), true
}

func (f *fakeRunner) Start(_ context.Context, place runner.Place, start runner.Start) (runner.Place, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, start)
	return runner.Place(fmt.Sprintf("%s|%d", place, len(f.started))), nil
}

func (f *fakeRunner) Status(context.Context, runner.Place) (runner.Status, error) {
	return runner.Status{State: runner.Running}, nil
}

func (f *fakeRunner) Read(_ context.Context, _ runner.Place, out runner.Output) (runner.Page, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	page := runner.Page{Lines: []string{}, NextOffset: out.Offset}
	for n := out.Offset; n < len(f.lines); n++ {
		page.Lines = append(page.Lines, f.lines[n])
		page.NextOffset = n + 1
	}
	return page, nil
}

func (f *fakeRunner) Kill(context.Context, runner.Place) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.killed++
	return nil
}

func (f *fakeRunner) Close(context.Context, runner.Place) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed++
	return nil
}

func (f *fakeRunner) counts() (opened, closed, started, killed int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.opened, f.closed, len(f.started), f.killed
}
