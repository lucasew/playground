package server

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner"
	"github.com/lucasew/env-mcp/internal/runner/host"
	"github.com/stretchr/testify/require"
)

// newServerAt builds a server whose store and exec state both live in state.
func newServerAt(t *testing.T, state string, r runner.Executor) *Server {
	t.Helper()
	store, err := env.NewStore(state + "/envs")
	require.NoError(t, err)
	checkout := gitRepo(t)
	s := NewServer(store, &env.Backend{
		Git:       &git.Git{},
		Workdir:   state + "/wt",
		Checkouts: map[string]string{"github:o/r": checkout},
	}, state)
	s.Runner = r
	return s
}

func TestExecHandleCrossesProcesses(t *testing.T) {
	state := t.TempDir()
	// The first server owns the shared state; the second is what a new CLI
	// process builds over the same directory.
	first := newServerAt(t, state, host.New(state))

	view, err := first.EnvGet(t.Context(), "github:o/r@main", "x")
	require.NoError(t, err)

	started, err := first.Exec(t.Context(), string(view.Handle), "", []string{"sh", "-c", "echo one; echo two; exit 3"})
	require.NoError(t, err)

	// A fresh server, as a new CLI invocation would build.
	second := newServerAt(t, state, host.New(state))
	var page env.Page
	var state2 ExecState
	for range 100 {
		state2, err = second.Poll(t.Context(), string(started.Handle))
		require.NoError(t, err)
		if state2.State == env.Exited {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, env.Exited, state2.State)
	require.Equal(t, 3, state2.ExitCode)

	page, err = second.Logs(t.Context(), string(started.Handle), 0, 10, "")
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two"}, page.Lines)
	require.Equal(t, 2, page.NextOffset)
}

func TestExecLimitIsEight(t *testing.T) {
	s := newServer(t, host.New(t.TempDir()))
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "limit")
	require.NoError(t, err)
	for i := range env.MaxExecs {
		_, err := s.Exec(t.Context(), string(view.Handle), "", []string{"sleep", "30"})
		require.NoError(t, err, "exec %d", i+1)
	}
	_, err = s.Exec(t.Context(), string(view.Handle), "", []string{"sleep", "30"})
	require.ErrorIs(t, err, env.ErrExecLimit)

	// Kill the children and let their reapers finish writing before the
	// temporary directory is removed.
	require.NoError(t, s.KillEnv(t.Context(), view.Handle))
	settle(t, s, view.Handle)
}

func TestCloseKillsLiveExecs(t *testing.T) {
	s := newServer(t, host.New(t.TempDir()))
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "close")
	require.NoError(t, err)
	started, err := s.Exec(t.Context(), string(view.Handle), "", []string{"sleep", "30"})
	require.NoError(t, err)
	pid := hostPID(t, s, started.Handle)
	require.True(t, processAlive(pid))

	require.NoError(t, s.EnvClose(t.Context(), string(view.Handle)))
	require.Eventually(t, func() bool { return !processAlive(pid) }, 2*time.Second, 10*time.Millisecond, "exec survived env close")
}

func TestCwdEscapeRejected(t *testing.T) {
	s := newServer(t, host.New(t.TempDir()))
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "escape")
	require.NoError(t, err)
	_, err = s.Exec(t.Context(), string(view.Handle), "../..", []string{"true"})
	require.ErrorIs(t, err, env.ErrPath)
}

func TestEmptyArgvRejected(t *testing.T) {
	s := newServer(t, host.New(t.TempDir()))
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "empty")
	require.NoError(t, err)
	_, err = s.Exec(t.Context(), string(view.Handle), "", nil)
	require.ErrorIs(t, err, env.ErrArgv)
}

// settle waits until no exec of handle is live, so no reaper writes after the
// test returns.
func settle(t *testing.T, s *Server, handle env.Handle) {
	t.Helper()
	require.Eventually(t, func() bool {
		recs, err := s.allExecs(handle)
		if err != nil {
			return false
		}
		for _, rec := range recs {
			if pid := pidFromPlace(t, rec.Place); processAlive(pid) {
				return false
			}
		}
		return true
	}, 2*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
}

// hostPID reads the process the host executor started for an exec.
func hostPID(t *testing.T, s *Server, handle env.Handle) int {
	t.Helper()
	rec, err := s.lookup(string(handle))
	require.NoError(t, err)
	return pidFromPlace(t, rec.Place)
}

// pidFromPlace reads the pid the host executor recorded beside a stream.
func pidFromPlace(t *testing.T, place runner.Place) int {
	t.Helper()
	b, err := os.ReadFile(string(place) + ".pid")
	require.NoError(t, err)
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	require.NoError(t, err)
	return pid
}
