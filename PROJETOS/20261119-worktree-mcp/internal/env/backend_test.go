package env

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lewtec/lewkit/x/git"
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
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestBackendIsIdempotent(t *testing.T) {
	checkout := gitRepo(t)
	b := &Backend{Git: &git.Git{}, Workdir: t.TempDir(), Checkouts: map[string]string{"github:o/r": checkout}}
	repo := Repo{Scheme: "github", Owner: "o", Name: "r", Base: "main"}

	first, created, err := b.Ensure(t.Context(), repo, "feat-x")
	require.NoError(t, err)
	require.True(t, created)

	second, created, err := b.Ensure(t.Context(), repo, "feat-x")
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, first, second)
}

func TestBackendRequiresDeclaredCheckout(t *testing.T) {
	b := &Backend{Git: &git.Git{}, Workdir: t.TempDir(), Checkouts: map[string]string{}}
	_, _, err := b.Ensure(t.Context(), Repo{Scheme: "github", Owner: "o", Name: "r", Base: "main"}, "s")
	require.ErrorIs(t, err, ErrRepo)
}
