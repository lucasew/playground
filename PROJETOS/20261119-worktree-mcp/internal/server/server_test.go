package server

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lucasew/env-mcp/internal/backend"
	"github.com/lucasew/env-mcp/internal/env"
	"github.com/stretchr/testify/require"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "init", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a"), []byte("a\n"), 0o644))
	run(t, dir, "add", "a")
	run(t, dir, "commit", "-m", "init")
	return dir
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func newServer(t *testing.T) (*Server, *fakeProvisioner) {
	t.Helper()
	store, err := env.NewStore(t.TempDir())
	require.NoError(t, err)
	fake := newFake()
	return NewServer(store, fake), fake
}

func TestEnvGetTwiceProvisionsIdempotently(t *testing.T) {
	s, _ := newServer(t)

	first, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.Equal(t, "created", first.Status)

	second, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.Equal(t, "resumed", second.Status)
	require.Equal(t, first.Handle, second.Handle)

	all, err := s.Store.List()
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Equal(t, first.Handle, all[0].Handle)
}

func TestEnvViewHasNoPath(t *testing.T) {
	s, _ := newServer(t)
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.Len(t, string(view.Handle), env.HandleLen)
	_, err = env.ParseHandle(string(view.Handle))
	require.NoError(t, err)
	b, err := json.Marshal(view)
	require.NoError(t, err)
	require.NotContains(t, string(b), "checkouts")
	require.NotContains(t, string(b), "worktrees")
}

func TestRegistryHoldsOnlyBackendHandle(t *testing.T) {
	s, _ := newServer(t)
	_, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	all, err := s.Store.List()
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Equal(t, "fake", all[0].Backend)
	require.NotEmpty(t, all[0].Token)
	require.NotContains(t, all[0].Token, "/tmp")
}

func TestEnvGetReprovisionsRecordedHandle(t *testing.T) {
	s, fake := newServer(t)
	_, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	_, err = s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	provisions, _ := fake.counts()
	require.Equal(t, 2, provisions)
}

func TestEnvCloseSpendsHandle(t *testing.T) {
	s, _ := newServer(t)
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.NoError(t, s.EnvClose(t.Context(), string(view.Handle)))
	_, err = s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.ErrorIs(t, err, env.ErrClosed)
}

func TestBadRepoRejected(t *testing.T) {
	s, _ := newServer(t)
	for _, bad := range []string{"o/r@main", "gitlab:o/r@main", "github:o/r", "github:o@main", "github:o/r@"} {
		_, err := s.EnvGet(t.Context(), bad, "feat")
		require.ErrorIs(t, err, env.ErrRepo, bad)
	}
}

func TestExecReturnsOpHandleScopedToEnv(t *testing.T) {
	s, _ := newServer(t)
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	started, err := s.Exec(t.Context(), string(view.Handle), "", []string{"echo", "hi"})
	require.NoError(t, err)
	require.Equal(t, view.Handle, env.Handle(started.Env))
	require.True(t, started.Handle.OwnedBy(string(view.Handle)))
	require.NotContains(t, started.Handle.String(), "/")
}

func TestOpHandleIsShort(t *testing.T) {
	s, _ := newServer(t)
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	started, err := s.Exec(t.Context(), string(view.Handle), "", []string{"echo", "hi"})
	require.NoError(t, err)
	// "<handle>.<op>" with a single-digit op.
	require.Equal(t, env.HandleLen+2, len(started.Handle.String()))
}

func TestEnvGetMintsAShortHandleOnce(t *testing.T) {
	s, _ := newServer(t)
	first, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	second, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.Equal(t, first.Handle, second.Handle)
	require.Len(t, string(first.Handle), env.HandleLen)
}

func TestDifferentSlugsGetDifferentHandles(t *testing.T) {
	s, _ := newServer(t)
	one, err := s.EnvGet(t.Context(), "github:o/r@main", "one")
	require.NoError(t, err)
	two, err := s.EnvGet(t.Context(), "github:o/r@main", "two")
	require.NoError(t, err)
	require.NotEqual(t, one.Handle, two.Handle)
}

func TestOpHandleFromAnotherEnvIsRefused(t *testing.T) {
	s, _ := newServer(t)
	one, err := s.EnvGet(t.Context(), "github:o/r@main", "one")
	require.NoError(t, err)
	two, err := s.EnvGet(t.Context(), "github:o/r@main", "two")
	require.NoError(t, err)
	started, err := s.Exec(t.Context(), string(one.Handle), "", []string{"echo", "hi"})
	require.NoError(t, err)
	// The op belongs to env one; asking env two's backend env for it fails.
	e, _, err := s.open(t.Context(), two.Handle)
	require.NoError(t, err)
	_, err = e.Status(t.Context(), started.Handle)
	require.ErrorIs(t, err, backend.ErrExec)
}
