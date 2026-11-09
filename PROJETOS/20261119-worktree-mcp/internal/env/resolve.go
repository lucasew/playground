package env

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Resolve turns a caller-supplied relative path into one path inside root.
// An absolute path, a path with a ".." segment, or a path that escapes root
// is ErrPath. The empty path is the root itself.
func Resolve(root, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: %q: absolute", ErrPath, rel)
	}
	if strings.ContainsRune(rel, 0) {
		return "", fmt.Errorf("%w: %q: NUL", ErrPath, rel)
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if part == ".." {
			return "", fmt.Errorf("%w: %q: parent segment", ErrPath, rel)
		}
	}
	joined := filepath.Join(root, filepath.FromSlash(rel))
	if !within(root, joined) {
		return "", fmt.Errorf("%w: %q: outside the env root", ErrPath, rel)
	}
	return joined, nil
}

// Within reports whether path is root or a descendant of root after both are
// cleaned. It compares whole path segments, so /a/bc is not inside /a/b.
func within(root, path string) bool {
	r := filepath.Clean(root)
	p := filepath.Clean(path)
	if p == r {
		return true
	}
	return strings.HasPrefix(p, r+string(filepath.Separator))
}
