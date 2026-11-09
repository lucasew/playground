package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/env-mcp/internal/backend/local"
	"github.com/stretchr/testify/require"
)

func TestDefaultBackendClonesUnderTheWorkdir(t *testing.T) {
	state := t.TempDir()
	b := DefaultBackend(state).(*local.Backend)
	require.NotNil(t, b.Store)
	require.Equal(t, state+"/checkouts/clones", b.Store.Clones)
	require.Equal(t, state+"/checkouts/worktrees", b.Store.Checkouts)
	require.Equal(t, "local", b.ID())
}

func TestDefaultBackendReadsDeclaredRepos(t *testing.T) {
	state := t.TempDir()
	cfg := `{"repos": {"github:o/r": "/somewhere/r"}, "workdir": "/w"}`
	require.NoError(t, os.WriteFile(filepath.Join(state, "config.json"), []byte(cfg), 0o600))
	b := DefaultBackend(state).(*local.Backend)
	require.Equal(t, "/w/clones", b.Store.Clones)
	require.Equal(t, "/somewhere/r", b.Store.Declared["o-r"])
}
