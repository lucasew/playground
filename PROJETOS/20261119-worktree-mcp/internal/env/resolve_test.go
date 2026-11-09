package env

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveRejects(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		"absolute":       "/etc/passwd",
		"parent":         "../outside",
		"nested parent":  "a/../../outside",
		"bare parent":    "..",
		"empty name nul": "a\x00b",
	}
	for name, rel := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Resolve(root, rel)
			require.ErrorIs(t, err, ErrPath)
		})
	}
}

func TestResolveAccepts(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		"empty":        root,
		"dot":          root,
		"file":         filepath.Join(root, "a.txt"),
		"nested":       filepath.Join(root, "a", "b.txt"),
		"dot nested":   filepath.Join(root, "a", "b.txt"),
		"same dir dot": filepath.Join(root, "a"),
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			rel := map[string]string{
				"empty":        "",
				"dot":          ".",
				"file":         "a.txt",
				"nested":       "a/b.txt",
				"dot nested":   "a/./b.txt",
				"same dir dot": "a/.",
			}[name]
			got, err := Resolve(root, rel)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

func TestWithinRejectsSiblingPrefix(t *testing.T) {
	root := filepath.Join(t.TempDir(), "a", "b")
	require.True(t, within(root, filepath.Join(root, "x")))
	require.False(t, within(root, filepath.Join(root, "..", "bc", "x")))
}
