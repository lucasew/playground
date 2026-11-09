package herdr

import (
	"reflect"
	"testing"
)

func TestOutputLinesKeepsOnlyOutput(t *testing.T) {
	// A real pane holds the prompt's terminal replies, the flush the runner
	// sent to clear them, the echoed command, the output, the sentinel, and
	// the next prompt.
	pane := "^[[?2026;2$y^[[?2027;1$yuser@host /wt/x 0$ 2026;2$y2027;1$y\n" +
		"bash: command not found: 2026\n" +
		"bash: command not found: 2\n" +
		"bash: command not found: 1\n" +
		"user@host /wt/x 127$ cd '/wt/x' && 'sh' '-c' 'echo a; echo b'; echo " + Done + "$?\n" +
		"a\n" +
		"b\n" +
		Done + "5\n" +
		"user@host /wt/x 1$"
	want := []string{"a", "b"}
	if got := outputLines(pane); !reflect.DeepEqual(got, want) {
		t.Fatalf("outputLines = %q, want %q", got, want)
	}
}

func TestOutputLinesKeepsInteriorBlank(t *testing.T) {
	pane := "x 0$ cd '/x' && 'cmd'; echo " + Done + "$?\n" +
		"one\n" +
		"\n" +
		"two\n" +
		Done + "0\n" +
		"x 0$"
	want := []string{"one", "", "two"}
	if got := outputLines(pane); !reflect.DeepEqual(got, want) {
		t.Fatalf("outputLines = %q, want %q", got, want)
	}
}

func TestExitFromReadsLastSentinel(t *testing.T) {
	text := Done + "1\nmore\n" + Done + "7"
	if code, ok := exitFrom(text); !ok || code != 7 {
		t.Fatalf("exitFrom = (%d, %v), want (7, true)", code, ok)
	}
	if _, ok := exitFrom("no sentinel here"); ok {
		t.Fatal("exitFrom found a sentinel that was not there")
	}
}

func TestCommandLineCarriesExitSentinel(t *testing.T) {
	line := commandLine("/tmp/a b", []string{"echo", "hi there"})
	want := `cd '/tmp/a b' && 'echo' 'hi there'; echo ` + Done + `$?`
	if line != want {
		t.Fatalf("commandLine = %q, want %q", line, want)
	}
}
