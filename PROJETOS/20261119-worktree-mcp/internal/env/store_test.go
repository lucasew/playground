package env

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreRoundTrip(t *testing.T) {
	store, err := NewStore(t.TempDir())
	require.NoError(t, err)
	h := NewHandle(Repo{Scheme: "github", Owner: "lewtec", Name: "lewkit", Base: "main"}, "slug")
	rec := Record{Handle: h, Repo: "github:lewtec/lewkit@main", Slug: "slug", Root: "/x"}
	require.NoError(t, store.Put(rec))

	got, err := store.Get(h)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, rec, *got)

	all, err := store.List()
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestStoreMissingIsNil(t *testing.T) {
	store, err := NewStore(t.TempDir())
	require.NoError(t, err)
	got, err := store.Get(Handle("nope"))
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestHandleIsFilenameSafe(t *testing.T) {
	h := NewHandle(Repo{Scheme: "github", Owner: "o", Name: "n", Base: "main"}, "feat/x")
	require.NotContains(t, HashHandle(h), "/")
}
