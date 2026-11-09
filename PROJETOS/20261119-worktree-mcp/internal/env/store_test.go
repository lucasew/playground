package env

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreRoundTrip(t *testing.T) {
	store, err := NewStore(t.TempDir())
	require.NoError(t, err)
	h := Handle("k3m9xq2p1a")
	rec := Record{Handle: h, Repo: "github:lewtec/lewkit@main", Slug: "slug", Base: "main", Backend: "local", Token: "tok"}
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
	h := Handle("k3m9xq2p1a")
	require.NotContains(t, HashHandle(h), "/")
}
