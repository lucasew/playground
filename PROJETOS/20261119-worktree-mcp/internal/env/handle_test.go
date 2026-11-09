package env

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMintHandleIsShortAndValid(t *testing.T) {
	seen := map[Handle]bool{}
	for range 500 {
		h, err := MintHandle()
		require.NoError(t, err)
		require.Len(t, string(h), HandleLen)
		_, err = ParseHandle(string(h))
		require.NoError(t, err)
		require.False(t, seen[h], "minted a duplicate handle")
		seen[h] = true
	}
}

func TestParseHandleRejectsWrongShape(t *testing.T) {
	for _, bad := range []string{"", "short", "github:o/r/feat", "k3m9xq2p1aextra", "k3m9xq2p1A", "k3m9xq2p1!"} {
		_, err := ParseHandle(bad)
		require.ErrorIs(t, err, ErrHandle, bad)
	}
}
