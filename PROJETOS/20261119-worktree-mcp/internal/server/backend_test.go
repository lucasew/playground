package server

import (
	"testing"

	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/env"
	"github.com/lucasew/env-mcp/internal/store"
	"github.com/stretchr/testify/require"
)

// backendAt builds a backend whose store keeps its checkouts under root and
// uses clone for repo o/r.
func backendAt(t *testing.T, root, clone string) *env.Backend {
	t.Helper()
	st := store.New(root, "github.com")
	st.Git = &git.Git{}
	if clone != "" {
		st.Declared["o-r"] = clone
	}
	return &env.Backend{Store: st}
}

func TestDefaultBackendClonesUnderTheStateDir(t *testing.T) {
	state := t.TempDir()
	b := DefaultBackend(state)
	require.NotNil(t, b.Store)
	require.Equal(t, state+"/checkouts/clones", b.Store.Clones)
	require.Equal(t, state+"/checkouts/worktrees", b.Store.Checkouts)
}
