package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucasew/env-mcp/internal/backend"
	"github.com/lucasew/env-mcp/internal/env"
)

// ExecView is what a caller sees of one started command.
//
// Handle is the op handle: it names one command inside one env. Env is the
// env's own handle, so a caller always knows which env the op belongs to.
type ExecView struct {
	Handle backend.Exec `json:"handle"`
	Env    env.Handle   `json:"env"`
	Argv   []string     `json:"argv"`
}

// ExecState is one exec's lifecycle without its output.
type ExecState struct {
	Handle     backend.Exec `json:"handle"`
	State      env.State    `json:"state"`
	ExitCode   int          `json:"exit_code"`
	NextOffset int          `json:"next_offset"`
}

// execRecord is the persisted part of one exec. Env and Handle are the two
// levels: the env the command runs in and the op inside it. The record never
// names a PID, a tab, or a checkout — only the handles the backend issued.
type execRecord struct {
	Env    env.Handle   `json:"env"`
	Handle backend.Exec `json:"handle"`
	Argv   []string     `json:"argv"`
}

// Exec starts argv in the env at cwd and returns its handle immediately.
func (s *Server) Exec(ctx context.Context, handle, cwd string, argv []string) (ExecView, error) {
	if len(argv) == 0 || argv[0] == "" {
		return ExecView{}, env.ErrArgv
	}
	envHandle := env.Handle(handle)
	e, _, err := s.open(ctx, envHandle)
	if err != nil {
		return ExecView{}, err
	}
	op, err := e.Exec(ctx, backend.Start{Dir: cwd, Argv: argv, Label: labelFor(argv)})
	if err != nil {
		return ExecView{}, execErr(err)
	}
	rec := execRecord{Env: envHandle, Handle: op, Argv: argv}
	if err := s.putExec(rec); err != nil {
		return ExecView{}, err
	}
	return ExecView{Handle: op, Env: envHandle, Argv: argv}, nil
}

// Logs reads one page of an exec stream and reports its state.
func (s *Server) Logs(ctx context.Context, handle string, offset, limit int, pattern string) (env.Page, error) {
	e, id, err := s.execAndEnv(ctx, handle)
	if err != nil {
		return env.Page{}, err
	}
	page, err := e.Read(ctx, id, backend.Read{Offset: offset, Limit: limit, Pattern: pattern})
	if err != nil {
		return env.Page{}, execErr(err)
	}
	status, err := e.Status(ctx, id)
	if err != nil {
		return env.Page{}, execErr(err)
	}
	return env.Page{
		Lines:      page.Lines,
		NextOffset: page.NextOffset,
		State:      env.State(status.State),
		ExitCode:   status.ExitCode,
	}, nil
}

// Poll reports one exec state and its stream length.
func (s *Server) Poll(ctx context.Context, handle string) (ExecState, error) {
	e, id, err := s.execAndEnv(ctx, handle)
	if err != nil {
		return ExecState{}, err
	}
	status, err := e.Status(ctx, id)
	if err != nil {
		return ExecState{}, execErr(err)
	}
	page, err := e.Read(ctx, id, backend.Read{})
	if err != nil {
		return ExecState{}, execErr(err)
	}
	return ExecState{
		Handle:     id,
		State:      env.State(status.State),
		ExitCode:   status.ExitCode,
		NextOffset: page.NextOffset,
	}, nil
}

// Kill stops an exec. Killing an exited exec succeeds.
func (s *Server) Kill(ctx context.Context, handle string) error {
	e, id, err := s.execAndEnv(ctx, handle)
	if err != nil {
		return err
	}
	return execErr(e.Kill(ctx, id))
}

// KillEnv stops every live exec of an env before the env handle is spent.
func (s *Server) KillEnv(ctx context.Context, handle env.Handle) error {
	e, _, err := s.open(ctx, handle)
	if err != nil {
		return err
	}
	return e.KillAll(ctx)
}

// execErr maps a backend error onto the public sentinels.
func execErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, backend.ErrExec):
		return fmt.Errorf("%w: %w", env.ErrExec, err)
	case errors.Is(err, backend.ErrPattern):
		return fmt.Errorf("%w: %w", env.ErrPattern, err)
	case errors.Is(err, backend.ErrOffset):
		return fmt.Errorf("%w: %w", env.ErrOffset, err)
	case errors.Is(err, backend.ErrLimit):
		return fmt.Errorf("%w: %w", env.ErrExecLimit, err)
	case errors.Is(err, backend.ErrClosed):
		return fmt.Errorf("%w: %w", env.ErrClosed, err)
	default:
		return err
	}
}

// labelFor names a command for a human watching the backend's surface.
func labelFor(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	return filepath.Base(argv[0])
}

// execAndEnv loads one exec record and opens the env that owns it.
//
// The op handle carries its env's token, so a handle minted by another env
// is refused by the backend before it can reach any command.
func (s *Server) execAndEnv(ctx context.Context, handle string) (backend.Env, backend.Exec, error) {
	rec, err := s.lookup(handle)
	if err != nil {
		return nil, backend.Exec{}, err
	}
	e, _, err := s.open(ctx, rec.Env)
	if err != nil {
		return nil, backend.Exec{}, err
	}
	return e, rec.Handle, nil
}

// execKey is the filename part of one op handle as stored on disk.
func execKey(op string) string { return env.HashHandle(env.Handle(op)) }

// execDir holds one env's exec records, keyed by the backend token the op
// handles embed, so a lookup from an op handle finds it without a scan.
func (s *Server) execDir(token string) string {
	return filepath.Join(s.Store.Dir(), "execs", env.HashHandle(env.Handle(token)))
}

// lookup loads one exec record. The op handle names its own env, so the
// record is a direct read: no scan and no ambiguity across envs.
func (s *Server) lookup(handle string) (execRecord, error) {
	id, err := parseExec(handle)
	if err != nil {
		return execRecord{}, err
	}
	b, err := os.ReadFile(filepath.Join(s.execDir(id.Env), execKey(handle)+".json"))
	if os.IsNotExist(err) {
		return execRecord{}, fmt.Errorf("%w: %s", env.ErrExec, handle)
	}
	if err != nil {
		return execRecord{}, err
	}
	var rec execRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return execRecord{}, err
	}
	return rec, nil
}

// parseExec reads the "<env>/<op>" spelling of an op handle. The op handle
// is the backend's, so this only splits it apart far enough to find the
// record on disk; the backend validates it again when the env opens it.
func parseExec(text string) (backend.Exec, error) {
	i := strings.LastIndex(text, "/")
	if i <= 0 || i == len(text)-1 {
		return backend.Exec{}, fmt.Errorf("%w: %s", env.ErrExec, text)
	}
	return backend.Exec{Env: text[:i], Op: text[i+1:]}, nil
}

func (s *Server) putExec(rec execRecord) error {
	dir := s.execDir(rec.Handle.Env)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, execKey(rec.Handle.String())+".json"), b, 0o600)
}

// allExecs lists one env's exec records in the order they were started. The
// caller passes the env's backend token.
func (s *Server) allExecs(token string) ([]execRecord, error) {
	entries, err := os.ReadDir(s.execDir(token))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []execRecord
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(s.execDir(token), e.Name()))
		if err != nil {
			return nil, err
		}
		var rec execRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Handle.Op < out[j].Handle.Op })
	return out, nil
}
