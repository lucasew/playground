package cli

import "strings"

// InsertDashes normalizes the exec argv.
//
// x/cmd treats a leading "-" as one of its own flags, so "env-mcp env exec
// <env> sh -c …" would fail with "unknown flag: -c". This inserts the "--"
// separator before <argv> when the caller omitted it, so the CLI spelling
// works with or without "--".
//
// exec's own flags (--cwd) stay before the separator; everything after the
// env handle is passed through untouched. Both spellings are accepted:
// "env exec" and the older top-level "exec".
func InsertDashes(args []string) []string {
	start := -1
	for i, a := range args {
		if a == "--" {
			return args
		}
		if a == "exec" {
			start = i
			break
		}
	}
	if start < 0 {
		return args
	}
	tail := args[start+1:]
	prefix := []string{}
	envAt := -1
	for i := 0; i < len(tail); i++ {
		a := tail[i]
		switch {
		case a == "--":
			return args
		case a == "--cwd" && i+1 < len(tail):
			prefix = append(prefix, a, tail[i+1])
			i++
		case strings.HasPrefix(a, "--cwd="):
			prefix = append(prefix, a)
		case len(a) > 1 && a[0] == '-':
			prefix = append(prefix, a)
		default:
			envAt = i
			i = len(tail)
		}
	}
	if envAt < 0 {
		return args
	}
	if envAt+1 < len(tail) && tail[envAt+1] == "--" {
		return args
	}
	out := make([]string, 0, len(args)+1)
	out = append(out, args[:start+1]...)
	out = append(out, prefix...)
	out = append(out, tail[envAt], "--")
	out = append(out, tail[envAt+1:]...)
	return out
}
