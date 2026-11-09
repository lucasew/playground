package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner"
)

// ExecView is what a caller sees of one started command.
type ExecView struct {
	Handle env.Handle `json:"handle"`
	Env    env.Handle `json:"env"`
	Argv   []string   `json:"argv"`
}

// ExecState is one exec's lifecycle without its output.
type ExecState struct {
	Handle     env.Handle `json:"handle"`
	State      env.State  `json:"state"`
	ExitCode   int        `json:"exit_code"`
	NextOffset int        `json:"next_offset"`
}

// execRecord is the persisted part of one exec. It survives the process that
// started it, so the CLI can address an exec across invocations. Runner and
// Place are the executor seam: the record never names a PID or a tab, only
// the executor that owns it and the token that executor returned.
type execRecord struct {
	Handle env.Handle   `json:"handle"`
	Env    env.Handle   `json:"env"`
	Argv   []string     `json:"argv"`
	Runner string       `json:"runner"`
	Place  runner.Place `json:"place"`
	Offset int          `json:"offset"`
	// Killed records that the caller stopped this exec. A runner may drop
	// the place on kill, so the record answers for it afterwards.
	Killed bool `json:"killed"`
}

// Exec starts argv in the env at cwd and returns its handle immediately.
func (s *Server) Exec(ctx context.Context, handle, cwd string, argv []string) (ExecView, error) {
	if len(argv) == 0 || argv[0] == "" {
		return ExecView{}, env.ErrArgv
	}
	envHandle := env.Handle(handle)
	envRec, err := s.Store.Get(envHandle)
	if err != nil {
		return ExecView{}, err
	}
	if envRec == nil {
		return ExecView{}, fmt.Errorf("%w: %s", env.ErrHandle, handle)
	}
	if envRec.Closed {
		return ExecView{}, fmt.Errorf("%w: %s", env.ErrClosed, handle)
	}
	live, err := s.liveExecs(envHandle)
	if err != nil {
		return ExecView{}, err
	}
	if len(live) >= env.MaxExecs {
		return ExecView{}, fmt.Errorf("%w: %d", env.ErrExecLimit, env.MaxExecs)
	}
	dir, err := env.Resolve(envRec.Root, cwd)
	if err != nil {
		return ExecView{}, err
	}
	execHandle, err := s.nextExecHandle(envHandle)
	if err != nil {
		return ExecView{}, err
	}
	r, err := s.runner()
	if err != nil {
		return ExecView{}, err
	}
	surface, err := r.Open(ctx, runner.Env{Checkout: envRec.Root, Source: envRec.Source, Label: string(envHandle)})
	if err != nil {
		return ExecView{}, err
	}
	place, err := r.Start(ctx, surface, runner.Start{
		Dir:   dir,
		Argv:  argv,
		Label: labelFor(execHandle),
	})
	if err != nil {
		return ExecView{}, err
	}
	rec := execRecord{
		Handle: execHandle,
		Env:    envHandle,
		Argv:   argv,
		Runner: r.ID(),
		Place:  place,
	}
	if err := s.putExec(rec); err != nil {
		return ExecView{}, err
	}
	return ExecView{Handle: execHandle, Env: envHandle, Argv: argv}, nil
}

// Logs reads one page of an exec stream and reports its state.
func (s *Server) Logs(ctx context.Context, handle string, offset, limit int, pattern string) (env.Page, error) {
	rec, r, err := s.execAndRunner(handle)
	if err != nil {
		return env.Page{}, err
	}
	if rec.Killed {
		return env.Page{Lines: []string{}, NextOffset: offset, State: env.Exited}, nil
	}
	page, err := r.Read(ctx, rec.Place, runner.Output{Offset: offset, Limit: limit, Pattern: pattern})
	if err != nil {
		return env.Page{}, execErr(err)
	}
	status, err := r.Status(ctx, rec.Place)
	if err != nil {
		return env.Page{}, err
	}
	return env.Page{
		Lines:         page.Lines,
		NextOffset:    page.NextOffset,
		State:         env.State(status.State),
		ExitCode:      status.ExitCode,
		TruncatedFrom: truncatedFrom(page),
	}, nil
}

// Poll reports one exec state and its stream length.
func (s *Server) Poll(ctx context.Context, handle string) (ExecState, error) {
	rec, r, err := s.execAndRunner(handle)
	if err != nil {
		return ExecState{}, err
	}
	if rec.Killed {
		return ExecState{Handle: rec.Handle, State: env.Exited}, nil
	}
	status, err := r.Status(ctx, rec.Place)
	if err != nil {
		return ExecState{}, err
	}
	page, err := r.Read(ctx, rec.Place, runner.Output{})
	if err != nil {
		return ExecState{}, execErr(err)
	}
	return ExecState{
		Handle:     rec.Handle,
		State:      env.State(status.State),
		ExitCode:   status.ExitCode,
		NextOffset: page.NextOffset,
	}, nil
}

// Kill stops an exec. Killing an exited exec succeeds.
func (s *Server) Kill(ctx context.Context, handle string) error {
	rec, r, err := s.execAndRunner(handle)
	if err != nil {
		return err
	}
	if rec.Killed {
		return nil
	}
	if status, err := r.Status(ctx, rec.Place); err == nil && status.State == runner.Exited {
		return nil
	}
	if err := r.Kill(ctx, rec.Place); err != nil {
		return err
	}
	rec.Killed = true
	return s.putExec(rec)
}

// KillEnv stops every live exec of an env before the env handle is spent.
func (s *Server) KillEnv(ctx context.Context, handle env.Handle) error {
	all, err := s.allExecs(handle)
	if err != nil {
		return err
	}
	for _, rec := range all {
		r, err := s.runnerByID(rec.Runner)
		if err != nil {
			continue
		}
		if rec.Killed {
			continue
		}
		if status, err := r.Status(ctx, rec.Place); err != nil || status.State == runner.Running {
			if r.Kill(ctx, rec.Place) == nil {
				rec.Killed = true
				_ = s.putExec(rec)
			}
		}
	}
	return nil
}

// runner returns the executor this server uses.
func (s *Server) runner() (runner.Executor, error) {
	if s.Runner != nil {
		return s.Runner, nil
	}
	return runner.Open(context.Background())
}

// runnerByID rebuilds the executor named in a record. The registry is the
// only place that knows how to construct one, so a record typed by another
// executor is reported as unavailable rather than misread.
func (s *Server) runnerByID(id string) (runner.Executor, error) {
	r, err := s.runner()
	if err != nil {
		return nil, err
	}
	if r.ID() != id {
		return nil, fmt.Errorf("%w: exec started by %q", env.ErrExec, id)
	}
	return r, nil
}

// execAndRunner loads one record and the executor that owns it.
func (s *Server) execAndRunner(handle string) (execRecord, runner.Executor, error) {
	rec, err := s.lookup(handle)
	if err != nil {
		return execRecord{}, nil, err
	}
	r, err := s.runnerByID(rec.Runner)
	if err != nil {
		return execRecord{}, nil, err
	}
	return rec, r, nil
}

// execErr maps a runner error onto the public sentinels.
func execErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, runner.ErrPattern.Error()):
		return fmt.Errorf("%w: %w", env.ErrPattern, err)
	case strings.Contains(msg, runner.ErrOffset.Error()):
		return fmt.Errorf("%w: %w", env.ErrOffset, err)
	default:
		return err
	}
}

// truncatedFrom reports the first line a runner can still reach, when it
// dropped lines the caller asked for.
func truncatedFrom(page runner.Page) int {
	if !page.Truncated {
		return 0
	}
	return page.NextOffset
}

func labelFor(handle env.Handle) string {
	_, after, ok := strings.Cut(string(handle), "@")
	if !ok {
		return string(handle)
	}
	return after
}

func (s *Server) liveExecs(handle env.Handle) ([]execRecord, error) {
	all, err := s.allExecs(handle)
	if err != nil {
		return nil, err
	}
	var live []execRecord
	for _, rec := range all {
		r, err := s.runnerByID(rec.Runner)
		if err != nil {
			continue
		}
		if rec.Killed {
			continue
		}
		if status, err := r.Status(context.Background(), rec.Place); err == nil && status.State == runner.Running {
			live = append(live, rec)
		}
	}
	return live, nil
}

func (s *Server) lookup(handle string) (execRecord, error) {
	b, err := os.ReadFile(s.execFile(env.Handle(handle)))
	if err != nil {
		if os.IsNotExist(err) {
			return execRecord{}, fmt.Errorf("%w: %s", env.ErrExec, handle)
		}
		return execRecord{}, err
	}
	var rec execRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return execRecord{}, err
	}
	return rec, nil
}

func (s *Server) putExec(rec execRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.execDir(rec.Env), 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(s.execFile(rec.Handle), b, 0o600)
}

func (s *Server) allExecs(handle env.Handle) ([]execRecord, error) {
	entries, err := os.ReadDir(s.execDir(handle))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []execRecord
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.execDir(handle), e.Name()))
		if err != nil {
			return nil, err
		}
		var rec execRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Handle < out[j].Handle })
	return out, nil
}

func (s *Server) execDir(handle env.Handle) string {
	return filepath.Join(s.StateDir, "execs", env.HashHandle(handle))
}

func (s *Server) execFile(handle env.Handle) string {
	parent, _, _ := strings.Cut(string(handle), "@")
	return filepath.Join(s.StateDir, "execs", env.HashHandle(env.Handle(parent)), env.HashHandle(handle)+".json")
}

// nextExecHandle mints the next exec handle in an env. The handle is
// persisted before the process starts, so two invocations never mint the
// same one.
func (s *Server) nextExecHandle(handle env.Handle) (env.Handle, error) {
	if err := os.MkdirAll(s.execDir(handle), 0o700); err != nil {
		return "", err
	}
	for n := 1; ; n++ {
		candidate := env.Handle(fmt.Sprintf("%s@%d", handle, n))
		if _, err := os.Stat(s.execFile(candidate)); err == nil {
			continue
		}
		return candidate, nil
	}
}
