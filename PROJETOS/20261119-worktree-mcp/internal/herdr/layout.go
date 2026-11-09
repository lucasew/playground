package herdr

import (
	"context"
	"fmt"
	"strings"
)

// Workspace is one row of the Herdr workspace list.
type Workspace struct {
	ID     string `json:"workspace_id"`
	Label  string `json:"label"`
	Number int    `json:"number"`
}

// Opened is the workspace worktree.open returned.
type Opened struct {
	Workspace   Workspace `json:"workspace"`
	AlreadyOpen bool      `json:"already_open"`
}

// OpenWorktree shows an existing checkout as a workspace. source is the
// repo the checkout belongs to; Herdr requires it for a linked worktree.
func (c *Client) OpenWorktree(ctx context.Context, path, label, source string) (Opened, error) {
	params := map[string]any{
		"path":  path,
		"label": label,
		"focus": false,
	}
	if source != "" {
		params["cwd"] = source
	}
	var opened Opened
	err := c.call(ctx, "worktree.open", params, &opened)
	return opened, err
}

// Tab is one row of the Herdr tab list.
type Tab struct {
	ID          string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
}

// Pane is one row of the Herdr pane list.
type Pane struct {
	ID          string `json:"pane_id"`
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
}

// CreateTab opens a tab in workspaceID with a shell at cwd.
func (c *Client) CreateTab(ctx context.Context, workspaceID, cwd, label string) (Tab, error) {
	var body struct {
		Tab  Tab  `json:"tab"`
		Pane Pane `json:"root_pane"`
	}
	if err := c.call(ctx, "tab.create", map[string]any{
		"workspace_id": workspaceID,
		"cwd":          cwd,
		"label":        label,
		"focus":        false,
	}, &body); err != nil {
		return Tab{}, err
	}
	if body.Tab.ID == "" {
		return Tab{}, fmt.Errorf("herdr tab.create: %w", ErrEmpty)
	}
	return body.Tab, nil
}

// ListTabs returns the tabs of a workspace.
func (c *Client) ListTabs(ctx context.Context, workspaceID string) ([]Tab, error) {
	var body struct {
		Tabs []Tab `json:"tabs"`
	}
	if err := c.call(ctx, "tab.list", map[string]any{"workspace_id": workspaceID}, &body); err != nil {
		return nil, err
	}
	return body.Tabs, nil
}

// ListPanes returns the panes of a workspace.
func (c *Client) ListPanes(ctx context.Context, workspaceID string) ([]Pane, error) {
	var body struct {
		Panes []Pane `json:"panes"`
	}
	if err := c.call(ctx, "pane.list", map[string]any{"workspace_id": workspaceID}, &body); err != nil {
		return nil, err
	}
	return body.Panes, nil
}

// SendKeys sends key presses such as Enter to a pane.
func (c *Client) SendKeys(ctx context.Context, paneID string, keys ...string) error {
	return c.call(ctx, "pane.send_keys", map[string]any{
		"pane_id": paneID,
		"keys":    keys,
	}, nil)
}

// SubmitText types a command and runs it.
//
// The carriage return rides in the same send_text call. A separate Enter
// keystroke arrives after the next prompt and re-runs terminal replies as
// shell input, so it is deliberately not used.
func (c *Client) SubmitText(ctx context.Context, paneID, text string) error {
	return c.SendText(ctx, paneID, strings.TrimRight(text, "\n")+"\r")
}

// SendText writes text to a pane without a trailing Enter.
func (c *Client) SendText(ctx context.Context, paneID, text string) error {
	return c.call(ctx, "pane.send_text", map[string]any{
		"pane_id": paneID,
		"text":    text,
	}, nil)
}

// Read returns recent unwrapped output of a pane. The server nests the text
// under read, not at the top level.
func (c *Client) Read(ctx context.Context, paneID string, lines int) (string, error) {
	var body struct {
		Read struct {
			Text string `json:"text"`
		} `json:"read"`
	}
	if err := c.call(ctx, "pane.read", map[string]any{
		"pane_id": paneID,
		"source":  "recent_unwrapped",
		"lines":   lines,
	}, &body); err != nil {
		return "", err
	}
	return body.Read.Text, nil
}

// WorkspaceByID returns one workspace, or an error when it is gone.
func (c *Client) WorkspaceByID(ctx context.Context, id string) (Workspace, error) {
	var body struct {
		Workspace Workspace `json:"workspace"`
	}
	if err := c.call(ctx, "workspace.get", map[string]any{"workspace_id": id}, &body); err != nil {
		return Workspace{}, err
	}
	return body.Workspace, nil
}

// TabOfPane returns the tab that holds a pane.
func (c *Client) TabOfPane(ctx context.Context, paneID string) (string, error) {
	var body struct {
		Pane struct {
			TabID string `json:"tab_id"`
		} `json:"pane"`
	}
	if err := c.call(ctx, "pane.get", map[string]any{"pane_id": paneID}, &body); err != nil {
		return "", err
	}
	if body.Pane.TabID == "" {
		return "", fmt.Errorf("herdr pane.get: %w", ErrEmpty)
	}
	return body.Pane.TabID, nil
}

// CloseTab closes a tab the caller created.
func (c *Client) CloseTab(ctx context.Context, tabID string) error {
	return c.call(ctx, "tab.close", map[string]any{"tab_id": tabID}, nil)
}

// CloseWorkspace closes a workspace the caller created.
func (c *Client) CloseWorkspace(ctx context.Context, workspaceID string) error {
	return c.call(ctx, "workspace.close", map[string]any{"workspace_id": workspaceID}, nil)
}

// MovePaneNewTab moves paneID into workspaceID as a new tab.
func (c *Client) MovePaneNewTab(ctx context.Context, paneID, workspaceID, label string) error {
	dest := map[string]any{
		"type":         "new_tab",
		"workspace_id": workspaceID,
	}
	if label != "" {
		dest["label"] = label
	}
	return c.call(ctx, "pane.move", map[string]any{
		"pane_id":     paneID,
		"focus":       false,
		"destination": dest,
	}, nil)
}
