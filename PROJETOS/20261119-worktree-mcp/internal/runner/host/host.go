// Package host is the plain-process executor.
//
// It starts a command under a small POSIX shell wrapper that records the
// exit status beside the output, so a later process can report the result
// without owning the child. It needs no daemon and no terminal multiplexer.
package host

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/env-mcp/internal/runner"
)

// Executor runs commands as direct children of the host, detached into their
// own process group.
type Executor struct {
	// Root is the directory holding one subdirectory per env.
	Root string
}

// New returns an executor that keeps its state under root.
func New(root string) *Executor { return &Executor{Root: root} }

// ID is the registry name.
func (e *Executor) ID() string { return "host" }

// wrapper runs the command and writes its exit status to "$1".
const wrapper = `p=$1; shift; { "$@"; printf '%s' "$?" > "$p"; }`

// Open returns the place for one env. The place is the env's state
// directory; it holds no live process until Start runs.
func (e *Executor) Open(_ context.Context, env runner.Env) (runner.Place, error) {
	dir := filepath.Join(e.Root, "runs", key(env.Checkout))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return runner.Place(dir), nil
}

// Start runs one command and returns the place holding its output.
func (e *Executor) Start(_ context.Context, place runner.Place, start runner.Start) (runner.Place, error) {
	dir := string(place)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	for n := 1; ; n++ {
		stream := filepath.Join(dir, fmt.Sprintf("%03d.log", n))
		status := stream + ".exit"
		if _, err := os.Stat(status); err == nil {
			continue
		}
		f, err := os.Create(stream)
		if err != nil {
			return "", err
		}
		args := append([]string{"-c", wrapper, "env-mcp-exec", status}, start.Argv...)
		cmd := exec.Command("sh", args...)
		cmd.Dir = start.Dir
		cmd.Stdout = f
		cmd.Stderr = f
		cmd.Env = append(os.Environ(), "TERM=dumb", "NO_COLOR=1", "GIT_PAGER=cat", "PAGER=cat")
		cmd.SysProcAttr = detached()
		if err := cmd.Start(); err != nil {
			f.Close()
			return "", fmt.Errorf("exec %s: %w", start.Argv[0], err)
		}
		if err := os.WriteFile(stream+".pid", []byte(strconv.Itoa(cmd.Process.Pid)), 0o600); err != nil {
			_ = cmd.Process.Kill()
			f.Close()
			return "", err
		}
		go func() {
			_ = cmd.Wait()
			f.Close()
		}()
		return runner.Place(stream), nil
	}
}

// Status reports whether the command behind place is still running and how
// it ended.
func (e *Executor) Status(_ context.Context, place runner.Place) (runner.Status, error) {
	stream := string(place)
	if b, err := os.ReadFile(stream + ".exit"); err == nil {
		code, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		return runner.Status{State: runner.Exited, ExitCode: code}, nil
	}
	return runner.Status{State: runner.Running}, nil
}

// Read returns whole lines from an offset. The status is not part of the
// page; the caller asks Status for it.
func (e *Executor) Read(_ context.Context, place runner.Place, out runner.Output) (runner.Page, error) {
	var re *regexp.Regexp
	if out.Pattern != "" {
		var err error
		re, err = regexp.Compile(out.Pattern)
		if err != nil {
			return runner.Page{}, fmt.Errorf("%w: %w", runner.ErrPattern, err)
		}
	}
	f, err := os.Open(string(place))
	if err != nil {
		return runner.Page{}, err
	}
	defer f.Close()
	page := runner.Page{Lines: []string{}, NextOffset: out.Offset}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for n := 0; scanner.Scan(); n++ {
		if n < out.Offset {
			continue
		}
		line := scanner.Text()
		if re == nil || re.MatchString(line) {
			page.Lines = append(page.Lines, line)
		}
		page.NextOffset = n + 1
		if out.Limit > 0 && len(page.Lines) >= out.Limit {
			break
		}
	}
	return page, scanner.Err()
}

// Kill stops the process group behind place. The wrapper is the group
// leader, so its children die with it.
func (e *Executor) Kill(_ context.Context, place runner.Place) error {
	pid, err := e.pid(place)
	if err != nil {
		return err
	}
	signalGroup(pid)
	return nil
}

// Close stops nothing on its own: the server calls Kill for each exec before
// it closes an env.
func (e *Executor) Close(context.Context, runner.Place) error { return nil }

func (e *Executor) pid(place runner.Place) (int, error) {
	b, err := os.ReadFile(string(place) + ".pid")
	if err != nil {
		return 0, nil
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

// key turns an absolute path into a directory name.
func key(path string) string {
	sum := uint64(14695981039346656037)
	for i := 0; i < len(path); i++ {
		sum ^= uint64(path[i])
		sum *= 1099511628211
	}
	return fmt.Sprintf("%016x", sum)
}
