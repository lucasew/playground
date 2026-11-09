package env

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/lewtec/lewkit/x/driver/exec/prelude"
	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/store"
	"github.com/stretchr/testify/require"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644))
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

// backendFor builds a backend whose store uses the given clone as repo o-r.
func backendFor(t *testing.T, clone string) *Backend {
	t.Helper()
	st := store.New(t.TempDir(), "github.com")
	st.Git = &git.Git{}
	if clone != "" {
		st.Declared["o-r"] = clone
	}
	return &Backend{Store: st}
}

func TestBackendIsIdempotent(t *testing.T) {
	b := backendFor(t, gitRepo(t))
	repo := Repo{Scheme: "github", Owner: "o", Name: "r", Base: "main"}

	first, err := b.Ensure(t.Context(), repo, "feat-x")
	require.NoError(t, err)
	require.True(t, first.Created)

	second, err := b.Ensure(t.Context(), repo, "feat-x")
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Equal(t, first.Root, second.Root)
	require.Equal(t, first.Source, second.Source)
}

func TestBackendNeedsAStore(t *testing.T) {
	b := &Backend{}
	_, err := b.Ensure(t.Context(), Repo{Scheme: "github", Owner: "o", Name: "r", Base: "main"}, "s")
	require.ErrorIs(t, err, ErrRepo)
}
