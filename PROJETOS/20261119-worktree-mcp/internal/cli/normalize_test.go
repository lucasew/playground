package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertDashes(t *testing.T) {
	cases := map[string][]string{
		"plain":       {"exec", "env", "sh", "-c", "x"},
		"explicit":    {"exec", "env", "--", "sh", "-c", "x"},
		"with cwd":    {"exec", "--cwd", "sub", "env", "sh", "-c", "x"},
		"root format": {"--format", "jsonl", "exec", "env", "sh", "-c", "x"},
	}
	want := map[string][]string{
		"plain":       {"exec", "env", "--", "sh", "-c", "x"},
		"explicit":    {"exec", "env", "--", "sh", "-c", "x"},
		"with cwd":    {"exec", "--cwd", "sub", "env", "--", "sh", "-c", "x"},
		"root format": {"--format", "jsonl", "exec", "env", "--", "sh", "-c", "x"},
	}
	for name, in := range cases {
		require.Equal(t, want[name], InsertDashes(in), name)
	}
}
