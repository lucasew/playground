// Package herdr is the Unix-socket client for a running Herdr session.
//
// It wraps the subset the server needs: open a worktree as a workspace, add
// a tab, send text, read output, and move a pane. Callers address places by
// the numbers this package returns and never by tab, pane, or workspace name.
package herdr

import (
	"bufio"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

var (
	ErrServerDown = errors.New("herdr server is not running")
	ErrNotSocket  = errors.New("herdr socket path is not a socket")
	ErrEmpty      = errors.New("herdr returned an empty result")
)

// ServerError is a JSON error from the Herdr server.
type ServerError struct {
	Method  string `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ServerError) Error() string {
	return "herdr " + e.Method + ": " + e.Code + ": " + e.Message
}

// Client calls the Herdr server over its Unix socket.
// Path overrides HERDR_SOCKET_PATH and the default config socket.
type Client struct {
	Path string
}

// Socket returns the socket path. A missing path means the server is down.
func (c *Client) Socket(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	path := cmp.Or(c.Path, os.Getenv("HERDR_SOCKET_PATH"))
	if path == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(dir, "herdr", "herdr.sock")
	}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrServerDown
		}
		return "", fmt.Errorf("herdr socket: %w", err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		return "", fmt.Errorf("%w: %s", ErrNotSocket, path)
	}
	return path, nil
}

func (c *Client) call(ctx context.Context, method string, params, dest any) error {
	if params == nil {
		params = map[string]any{}
	}
	sock, err := c.Socket(ctx)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{
		"id":     method,
		"method": method,
		"params": params,
	})
	if err != nil {
		return err
	}
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unix", sock)
	if err != nil {
		return fmt.Errorf("herdr socket: %w", err)
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("herdr %s: %w", method, err)
	}
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("herdr %s: %w", method, err)
	}
	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(line, &resp); err != nil {
		return fmt.Errorf("herdr %s: %w", method, err)
	}
	if resp.Error != nil {
		return ServerError{Method: method, Code: resp.Error.Code, Message: resp.Error.Message}
	}
	if dest == nil {
		return nil
	}
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return ErrEmpty
	}
	return json.Unmarshal(resp.Result, dest)
}
