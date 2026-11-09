package local_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lucasew/env-mcp/internal/backend"
	"github.com/lucasew/env-mcp/internal/backend/local"
	"github.com/lucasew/env-mcp/internal/runner/host"
	"github.com/stretchr/testify/require"
)

// repo makes a git repo with one commit.
func repo(t *testing.T) string {
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

func backendAt(t *testing.T) *local.Backend {
	t.Helper()
	base := t.TempDir()
	b := local.New(filepath.Join(base, "store"))
	b.State = base
	b.Executors = host.New(filepath.Join(base, "host"))
	b.Store.Declared["o-r"] = repo(t)
	return b
}

func provision(t *testing.T, b *local.Backend, slug, name string) (backend.Env, backend.Handle) {
	t.Helper()
	ctx := context.Background()
	p, err := b.Provision(ctx, backend.Request{Owner: "o", Name: "r", Base: "main", Slug: slug, Label: slug})
	require.NoError(t, err)
	e, err := b.Open(ctx, p.Handle, name)
	require.NoError(t, err)
	return e, p.Handle
}

func TestOpHandleFromAnotherEnvIsRefused(t *testing.T) {
	b := backendAt(t)
	one, _ := provision(t, b, "one", "aaaaaaa111")
	two, _ := provision(t, b, "two", "bbbbbbb222")
	ctx := context.Background()

	op, err := one.Exec(ctx, backend.Start{Argv: []string{"echo", "hi"}})
	require.NoError(t, err)
	require.Equal(t, "aaaaaaa111", op.Env)

	// The op names env one; env two must refuse it rather than run it.
	_, err = two.Status(ctx, op)
	require.ErrorIs(t, err, backend.ErrExec)
	_, err = two.Read(ctx, op, backend.Read{})
	require.ErrorIs(t, err, backend.ErrExec)
	require.ErrorIs(t, two.Kill(ctx, op), backend.ErrExec)

	// Its own env accepts it.
	_, err = one.Status(ctx, op)
	require.NoError(t, err)
}

func TestProvisionIsIdempotentOnTheSameHandle(t *testing.T) {
	b := backendAt(t)
	ctx := context.Background()
	req := backend.Request{Owner: "o", Name: "r", Base: "main", Slug: "same", Label: "same"}
	first, err := b.Provision(ctx, req)
	require.NoError(t, err)
	second, err := b.Provision(ctx, req)
	require.NoError(t, err)
	require.Equal(t, first.Handle, second.Handle)
	require.False(t, second.Created)
}

func TestDifferentBasesAreDifferentEnvs(t *testing.T) {
	b := backendAt(t)
	ctx := context.Background()
	one, err := b.Provision(ctx, backend.Request{Owner: "o", Name: "r", Base: "main", Slug: "s"})
	require.NoError(t, err)
	two, err := b.Provision(ctx, backend.Request{Owner: "o", Name: "r", Base: "dev", Slug: "s"})
	require.NoError(t, err)
	require.NotEqual(t, one.Handle, two.Handle)
}
