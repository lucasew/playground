package server

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/runner"
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

func newServer(t *testing.T, r runner.Executor) *Server {
	t.Helper()
	store, err := env.NewStore(t.TempDir())
	require.NoError(t, err)
	checkout := gitRepo(t)
	s := NewServer(store, &env.Backend{
		Git:       &git.Git{},
		Workdir:   t.TempDir(),
		Checkouts: map[string]string{"github:o/r": checkout},
	}, t.TempDir())
	// Always install a runner: falling through to the registry would write
	// into the package directory.
	s.Runner = r
	return s
}

func TestEnvGetTwiceMakesOneWorktree(t *testing.T) {
	s := newServer(t, &fakeRunner{})

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
}

func TestEnvViewHasNoPath(t *testing.T) {
	s := newServer(t, &fakeRunner{})
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.Equal(t, "github:o/r/feat", string(view.Handle))
	b, err := json.Marshal(view)
	require.NoError(t, err)
	require.NotContains(t, string(b), viewCheckoutRoot(t, s))
}

func TestEnvCloseSpendsHandle(t *testing.T) {
	s := newServer(t, &fakeRunner{})
	view, err := s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.NoError(t, err)
	require.NoError(t, s.EnvClose(t.Context(), string(view.Handle)))
	_, err = s.EnvGet(t.Context(), "github:o/r@main", "feat")
	require.ErrorIs(t, err, env.ErrClosed)
}

func TestBadRepoRejected(t *testing.T) {
	s := newServer(t, &fakeRunner{})
	for _, bad := range []string{"o/r@main", "gitlab:o/r@main", "github:o/r", "github:o@main", "github:o/r@"} {
		_, err := s.EnvGet(t.Context(), bad, "feat")
		require.ErrorIs(t, err, env.ErrRepo, bad)
	}
}

// viewCheckoutRoot reads the stored root for the only env in the store.
func viewCheckoutRoot(t *testing.T, s *Server) string {
	t.Helper()
	all, err := s.Store.List()
	require.NoError(t, err)
	require.Len(t, all, 1)
	return all[0].Root
}
