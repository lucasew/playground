package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeStream(t *testing.T, lines []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "exec.log")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	require.NoError(t, os.WriteFile(path, []byte(b.String()), 0o644))
	return path
}

func TestWindowedReadConcatenates(t *testing.T) {
	var lines []string
	for i := range 10000 {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	stream := OpenStream(writeStream(t, lines))
	var got []string
	offset := 0
	for {
		page, err := stream.Read(offset, 777, "")
		require.NoError(t, err)
		got = append(got, page.Lines...)
		require.GreaterOrEqual(t, page.NextOffset, offset)
		if page.NextOffset == offset {
			break
		}
		offset = page.NextOffset
	}
	require.Equal(t, lines, got)
}

func TestRegexAdvancesPastUnmatchedLines(t *testing.T) {
	stream := OpenStream(writeStream(t, []string{"ok", "error: boom", "ok", "error: bang"}))
	page, err := stream.Read(0, 10, "^error: ")
	require.NoError(t, err)
	require.Equal(t, []string{"error: boom", "error: bang"}, page.Lines)
	require.Equal(t, 4, page.NextOffset)

	again, err := stream.Read(page.NextOffset, 10, "^error: ")
	require.NoError(t, err)
	require.Empty(t, again.Lines)
}

func TestBadPatternIsSyntax(t *testing.T) {
	stream := OpenStream(writeStream(t, []string{"x"}))
	_, err := stream.Read(0, 10, "(")
	require.ErrorIs(t, err, ErrPattern)
}

func TestNegativeOffsetRejected(t *testing.T) {
	stream := OpenStream(writeStream(t, []string{"x"}))
	_, err := stream.Read(-1, 10, "")
	require.ErrorIs(t, err, ErrOffset)
}
