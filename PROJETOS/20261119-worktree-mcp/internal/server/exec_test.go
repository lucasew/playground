package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lucasew/env-mcp/internal/backend/local"
	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner/host"
	"github.com/stretchr/testify/require"
)

// hostRoots remembers where each test's host executor keeps its runs, so a
// test can read the pid of a command it started.
var hostRoots = map[*Server]string{}

func hostServer(t *testing.T, state string) *Server {
	t.Helper()
	registry, err := env.NewStore(filepath.Join(state, "envs"))
	require.NoError(t, err)
	b := local.New(filepath.Join(state, "checkouts"))
	b.State = state
	b.Executors = host.New(filepath.Join(state, "host"))
	b.Store.Declared["o-r"] = gitRepo(t)
	s := NewServer(registry, b)
	hostRoots[s] = filepath.Join(state, "host")
	return s
}

func TestExecHandleCrossesProcesses(t *testing.T) {
	state := t.TempDir()
	// The first server owns the shared state; the second is what a new CLI
	// process builds over the same directory.
	first := hostServer(t, state)

	view, err := first.EnvGet(t.Context(), "github:o/r@main", "x")
	require.NoError(t, err)

	started, err := first.Exec(t.Context(), string(view.Handle), "", []string{"sh", "-c", "echo one; echo two; exit 3"})
	require.NoError(t, err)

	// A fresh server, as a new CLI invocation would build.
	second := hostServer(t, state)
	var page env.Page
	var state2 ExecState
	for range 100 {
		state2, err = second.Poll(t.Context(), started.Handle.String())
		require.NoError(t, err)
		if state2.State == env.Exited {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, env.Exited, state2.State)
	require.Equal(t, 3, state2.ExitCode)

	page, err = second.Logs(t.Context(), started.Handle.String(), 0, 10, "")
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two"}, page.Lines)
	require.Equal(t, 2, page.NextOffset)
}

func TestExecLimitIsEight(t *testing.T) {
	s := hostServer(t, t.TempDir())
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "limit")
	require.NoError(t, err)
	for i := range env.MaxExecs {
		_, err := s.Exec(t.Context(), string(view.Handle), "", []string{"sleep", "30"})
		require.NoError(t, err, "exec %d", i+1)
	}
	_, err = s.Exec(t.Context(), string(view.Handle), "", []string{"sleep", "30"})
	require.ErrorIs(t, err, env.ErrExecLimit)

	require.NoError(t, s.KillEnv(t.Context(), view.Handle))
	settle(t, s, view.Handle)
}

func TestCloseKillsLiveExecs(t *testing.T) {
	s := hostServer(t, t.TempDir())
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "close")
	require.NoError(t, err)
	started, err := s.Exec(t.Context(), string(view.Handle), "", []string{"sleep", "30"})
	require.NoError(t, err)
	_ = started
	pid := livePID(t, s)
	require.True(t, processAlive(pid))

	require.NoError(t, s.EnvClose(t.Context(), string(view.Handle)))
	require.Eventually(t, func() bool { return !processAlive(pid) }, 2*time.Second, 10*time.Millisecond, "exec survived env close")
}

func TestCwdEscapeRejected(t *testing.T) {
	s := hostServer(t, t.TempDir())
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "escape")
	require.NoError(t, err)
	_, err = s.Exec(t.Context(), string(view.Handle), "../..", []string{"true"})
	require.ErrorIs(t, err, env.ErrPath)
}

func TestEmptyArgvRejected(t *testing.T) {
	s := hostServer(t, t.TempDir())
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "empty")
	require.NoError(t, err)
	_, err = s.Exec(t.Context(), string(view.Handle), "", nil)
	require.ErrorIs(t, err, env.ErrArgv)
}

// settle waits until no host process of handle is live, so no reaper writes
// after the test returns.
func settle(t *testing.T, s *Server, handle env.Handle) {
	t.Helper()
	require.Eventually(t, func() bool {
		for _, pid := range envPIDs(t, s) {
			if processAlive(pid) {
				return false
			}
		}
		return true
	}, 2*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
}

// livePID returns a pid the host executor started that is still running.
func livePID(t *testing.T, s *Server) int {
	t.Helper()
	for _, pid := range envPIDs(t, s) {
		if processAlive(pid) {
			return pid
		}
	}
	require.FailNow(t, "no live host pid")
	return 0
}

// envPIDs reads every pid the host executor recorded for this server. The
// test built the backend, so it knows where the executor keeps its runs.
func envPIDs(t *testing.T, s *Server) []int {
	t.Helper()
	root := filepath.Join(hostRoots[s], "runs")
	dirs, err := os.ReadDir(root)
	require.NoError(t, err)
	var pids []int
	for _, dir := range dirs {
		path := filepath.Join(root, dir.Name())
		files, err := os.ReadDir(path)
		require.NoError(t, err)
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".pid") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(path, f.Name()))
			require.NoError(t, err)
			pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
			require.NoError(t, err)
			pids = append(pids, pid)
		}
	}
	return pids
}
