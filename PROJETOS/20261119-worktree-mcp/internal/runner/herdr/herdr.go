// Package herdr runs commands inside a live Herdr session.
//
// One env is one Herdr workspace bound to the checkout. Execs are tabs in
// that workspace, so the user sees each command beside the agent shell. The
// runner stores tab identifiers, never returns them, and only touches tabs
// it created.
package herdr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/herdr"
	"github.com/lucasew/env-mcp/internal/runner"
)

// Executor drives a Herdr session over its socket.
type Executor struct {
	// Client talks to the session. A nil client uses the default socket.
	Client *herdr.Client
	// State is where the runner keeps the place of each env.
	State string

	mu     sync.Mutex
	places map[string]runner.Place
}

// New returns an executor that keeps its layout record under state.
func New(state string) *Executor {
	return &Executor{State: state, places: make(map[string]runner.Place)}
}

// ID is the registry name.
func (e *Executor) ID() string { return "herdr" }

func (e *Executor) client() *herdr.Client {
	if e.Client != nil {
		return e.Client
	}
	return &herdr.Client{}
}

// Open ensures the workspace for an env and returns its place.
//
// The workspace is the execution surface: the agent shell lives in it and
// every exec becomes a tab. Each env's workspace nests under the workspace
// of its repository, so the sidebar groups one repo's envs together.
// Reopening an env returns the same workspace.
func (e *Executor) Open(ctx context.Context, env runner.Env) (runner.Place, error) {
	if place, err := e.storedPlace(env.Checkout); err == nil && place != "" {
		if alive, err := e.workspaceAlive(ctx, string(place)); err == nil && alive {
			return place, nil
		}
	}
	parent, err := e.repoWorkspace(ctx, env)
	if err != nil {
		return "", err
	}
	opened, err := e.client().OpenWorktreeUnder(ctx, parent, env.Checkout, env.Label, env.Source)
	if err != nil {
		return "", err
	}
	tabs, err := e.client().ListTabs(ctx, opened.Workspace.ID)
	if err != nil {
		return "", err
	}
	if len(tabs) == 0 {
		if _, err := e.client().CreateTab(ctx, opened.Workspace.ID, env.Checkout, env.Label); err != nil {
			return "", err
		}
	}
	place := runner.Place(opened.Workspace.ID)
	if err := e.storePlace(env.Checkout, place); err != nil {
		return "", err
	}
	return place, nil
}

// repoWorkspace returns the workspace that groups the envs of one repo.
//
// The repository's main checkout gets a workspace of its own when it has
// none; every env then nests under it. The lookup reads Herdr's own state,
// so a workspace the user already has for that repo is reused and no
// second grouping is invented.
func (e *Executor) repoWorkspace(ctx context.Context, env runner.Env) (string, error) {
	if env.Source == "" {
		return "", nil
	}
	if id, ok := e.workspaceForCheckout(ctx, env.Source); ok {
		return id, nil
	}
	opened, err := e.client().OpenWorktree(ctx, env.Source, repoLabel(env.Source), env.Source)
	if err != nil {
		return "", err
	}
	return opened.Workspace.ID, nil
}

// workspaceForCheckout reports the workspace already bound to a checkout.
func (e *Executor) workspaceForCheckout(ctx context.Context, checkout string) (string, bool) {
	spaces, err := e.client().ListWorkspaces(ctx)
	if err != nil {
		return "", false
	}
	want := git.Resolve(checkout)
	for _, space := range spaces {
		if space.Worktree == nil || space.Worktree.Checkout == "" {
			continue
		}
		if git.Resolve(space.Worktree.Checkout) == want {
			return space.ID, true
		}
	}
	return "", false
}

// repoLabel names a repository's workspace for a human watching the sidebar.
func repoLabel(source string) string {
	return filepath.Base(source)
}

// Start opens a tab for one command in the env workspace, sends the command
// with a sentinel that marks the end, and returns the tab as the place.
func (e *Executor) Start(ctx context.Context, place runner.Place, start runner.Start) (runner.Place, error) {
	dir := start.Dir
	if dir == "" {
		dir = "."
	}
	tab, err := e.client().CreateTab(ctx, string(place), dir, start.Label)
	if err != nil {
		return "", err
	}
	panes, err := e.client().ListPanes(ctx, string(place))
	if err != nil {
		return "", err
	}
	paneID := ""
	for _, p := range panes {
		if p.TabID == tab.ID {
			paneID = p.ID
			break
		}
	}
	if paneID == "" {
		return "", fmt.Errorf("herdr: tab %s has no pane", tab.ID)
	}
	if err := e.typeCommand(ctx, paneID, commandLine(dir, start.Argv)); err != nil {
		return "", err
	}
	return runner.Place(paneID), nil
}

// commandLine builds the line typed into a tab. cd is explicit because a tab
// may be created before its cwd exists, and the sentinel carries the status
// so a later process can report it without owning the child.
func commandLine(dir string, argv []string) string {
	return fmt.Sprintf("cd %s && %s; echo %s$?", shellQuote(dir), shellJoin(argv), Done)
}

// promptSettle is how long a new shell needs to answer its own terminal
// queries. Until those replies land, anything typed joins them as junk.
const promptSettle = 1500 * time.Millisecond

// typeCommand runs one command in a pane.
//
// A fresh shell asks the terminal for reports (cursor position, kitty
// keyboard state) and reads the answers back as input. Typing too early
// mixes those answers into the command line. So: let the prompt settle,
// flush whatever it caught with Enter, clear the line, then type and submit.
func (e *Executor) typeCommand(ctx context.Context, paneID, line string) error {
	c := e.client()
	if err := sleep(ctx, promptSettle); err != nil {
		return err
	}
	if err := c.SendKeys(ctx, paneID, "Enter"); err != nil {
		return err
	}
	if err := sleep(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	if err := c.SendKeys(ctx, paneID, "ctrl+u"); err != nil {
		return err
	}
	if err := sleep(ctx, 200*time.Millisecond); err != nil {
		return err
	}
	if err := c.SendText(ctx, paneID, line); err != nil {
		return err
	}
	if err := sleep(ctx, 400*time.Millisecond); err != nil {
		return err
	}
	return c.SendKeys(ctx, paneID, "Enter")
}

// sleep waits, but gives up when the caller does.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Status asks Herdr whether the command has printed its exit sentinel.
func (e *Executor) Status(ctx context.Context, place runner.Place) (runner.Status, error) {
	text, err := e.client().Read(ctx, string(place), 2000)
	if err != nil {
		return runner.Status{}, err
	}
	if code, ok := exitFrom(text); ok {
		return runner.Status{State: runner.Exited, ExitCode: code}, nil
	}
	return runner.Status{State: runner.Running}, nil
}

// Read returns the output of one command.
//
// Herdr owns the scrollback, so the runner reads the whole recent window and
// addresses it by line offset. The cursor advances past every scanned line.
func (e *Executor) Read(ctx context.Context, place runner.Place, out runner.Output) (runner.Page, error) {
	var re *regexp.Regexp
	if out.Pattern != "" {
		var err error
		re, err = regexp.Compile(out.Pattern)
		if err != nil {
			return runner.Page{}, fmt.Errorf("%w: %w", runner.ErrPattern, err)
		}
	}
	text, err := e.client().Read(ctx, string(place), 5000)
	if err != nil {
		return runner.Page{}, err
	}
	lines := outputLines(text)
	if out.Offset > len(lines) {
		return runner.Page{}, fmt.Errorf("%w: %d > %d", runner.ErrOffset, out.Offset, len(lines))
	}
	page := runner.Page{Lines: []string{}, NextOffset: out.Offset}
	for n := out.Offset; n < len(lines); n++ {
		line := lines[n]
		if re == nil || re.MatchString(line) {
			page.Lines = append(page.Lines, line)
		}
		page.NextOffset = n + 1
		if out.Limit > 0 && len(page.Lines) >= out.Limit {
			break
		}
	}
	return page, nil
}

// Kill closes the tab holding the command.
func (e *Executor) Kill(ctx context.Context, place runner.Place) error {
	tab, err := e.client().TabOfPane(ctx, string(place))
	if err != nil {
		return nil
	}
	return e.client().CloseTab(ctx, tab)
}

// Close closes the workspace the runner opened for the env.
func (e *Executor) Close(ctx context.Context, place runner.Place) error {
	return e.client().CloseWorkspace(ctx, string(place))
}

// Done marks the line that carries a command's exit status.
const Done = "__env_mcp_done__"

var doneRE = regexp.MustCompile(Done + `(\d+)`)

func exitFrom(text string) (int, bool) {
	m := doneRE.FindAllStringSubmatch(text, -1)
	if len(m) == 0 {
		return 0, false
	}
	var code int
	_, err := fmt.Sscanf(m[len(m)-1][1], "%d", &code)
	return code, err == nil
}

// dcsRE drops terminal reports such as a cursor-position or kitty-keyboard
// reply. They are answers to queries the shell asked, not command output.
var dcsRE = regexp.MustCompile(`\x1b(?:\[[0-9;?]*[$y~u]|\][^\x07\x1b]*(?:\x07|\x1b\\)|[()][0-9A-B])`)

// junkRE matches the commands the shell runs when it eats its own terminal
// replies, e.g. "bash: command not found: 2026" or a bare "2026;2$y".
var (
	notFoundRE = regexp.MustCompile(`^(?:\w+: )?command not found: [0-9;$y ]+$`)
	bareJunkRE = regexp.MustCompile(`^[0-9;]+[$y]?[0-9;]*$`)
)

// promptRE matches the shell prompt that ends a command, e.g.
// "user@host /path 0$".
var promptRE = regexp.MustCompile(`\S+@\S+ .*[#$]\s*$`)

// outputLines reduces a pane's scrollback to the lines a caller asked for.
//
// The pane holds more than a command's output: the prompt's terminal
// replies, the flush the runner sent to clear them, the echoed command line,
// and the prompt that follows. Those become empty strings instead of being
// removed, so a line offset keeps meaning the same line as the pane grows.
func outputLines(text string) []string {
	text = dcsRE.ReplaceAllString(text, "")
	raw := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	// The region before the echoed command is preamble, not output, so it
	// is dropped outright; offsets then count from the first real line.
	lines := make([]string, 0, len(raw))
	started := false
	for _, line := range raw {
		line = strings.TrimRight(line, " \t")
		switch {
		case doneRE.MatchString(line):
			// The sentinel ends the command; later lines are a new prompt.
			started = false
			continue
		case !started:
			// Skip everything up to the echoed command line.
			if isEcho(line) {
				started = true
			}
			continue
		case notFoundRE.MatchString(line), bareJunkRE.MatchString(line), promptRE.MatchString(line):
			continue
		}
		lines = append(lines, line)
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// isEcho reports whether a line is the command the runner typed. The shell
// prints it after its prompt, so the prompt is still in front of it; the
// sentinel is what identifies it, since the marker line is matched earlier.
func isEcho(line string) bool {
	return strings.Contains(line, Done)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func shellJoin(argv []string) string {
	quoted := make([]string, len(argv))
	for i, a := range argv {
		quoted[i] = shellQuote(a)
	}
	return strings.Join(quoted, " ")
}

func (e *Executor) placeFile(envRoot string) string {
	return filepath.Join(e.State, "herdr", key(envRoot)+".json")
}

func (e *Executor) storedPlace(envRoot string) (runner.Place, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if p, ok := e.places[envRoot]; ok {
		return p, nil
	}
	b, err := os.ReadFile(e.placeFile(envRoot))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var rec struct {
		Place runner.Place `json:"place"`
	}
	if err := json.Unmarshal(b, &rec); err != nil {
		return "", err
	}
	e.places[envRoot] = rec.Place
	return rec.Place, nil
}

func (e *Executor) storePlace(envRoot string, place runner.Place) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.places[envRoot] = place
	if err := os.MkdirAll(filepath.Join(e.State, "herdr"), 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(map[string]any{"root": envRoot, "place": place})
	if err != nil {
		return err
	}
	return os.WriteFile(e.placeFile(envRoot), b, 0o600)
}

func (e *Executor) workspaceAlive(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}
	_, err := e.client().WorkspaceByID(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// key turns an absolute path into a file name.
func key(path string) string {
	sum := uint64(14695981039346656037)
	for i := 0; i < len(path); i++ {
		sum ^= uint64(path[i])
		sum *= 1099511628211
	}
	return fmt.Sprintf("%016x", sum)
}

// Peek reports the place already stored for a checkout without opening the
// surface. It is what lets a close tear down an env that is not running.
func (e *Executor) Peek(_ context.Context, envRoot string) (runner.Place, bool) {
	place, err := e.storedPlace(envRoot)
	return place, err == nil && place != ""
}
