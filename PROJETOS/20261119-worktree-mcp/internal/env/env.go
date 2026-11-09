// Package env is one named unit of work: a repository checkout, the places
// commands run in it, and the handles that address both.
//
// An [Env] is addressed by the opaque handle <scheme>:<owner>/<name>/<slug>.
// Callers never receive a checkout path. Every file path is relative to the
// env root and resolves inside it.
package env

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lewtec/lewkit/x/errs"
)

var (
	ErrRepo      = errs.Wrap(errs.ErrInvalid, "repo")
	ErrSlug      = errs.Wrap(errs.ErrInvalid, "slug")
	ErrHandle    = errs.Wrap(errs.ErrInvalid, "env handle")
	ErrPath      = errs.Wrap(errs.ErrInvalid, "path")
	ErrClosed    = errs.Wrap(errs.ErrClosed, "env")
	ErrExecLimit = errs.Wrap(errs.ErrUnavailable, "exec limit")
	ErrExec      = errs.Wrap(errs.ErrNotFound, "exec")
	ErrOffset    = errs.Wrap(errs.ErrOutOfRange, "offset")
	ErrPattern   = errs.Wrap(errs.ErrSyntax, "pattern")
	ErrArgv      = errs.Wrap(errs.ErrEmpty, "argv")
)

// MaxExecs is the most live execs one env holds at a time.
const MaxExecs = 8

// Repo is the identity github:owner/repo@base.
type Repo struct {
	Scheme string
	Owner  string
	Name   string
	Base   string
}

// ParseRepo reads github:owner/repo@base. Scheme, owner, name, and base are
// each non-empty, and name contains no slash.
func ParseRepo(s string) (Repo, error) {
	scheme, rest, ok := strings.Cut(s, ":")
	if !ok {
		return Repo{}, fmt.Errorf("%w: %q: missing scheme", ErrRepo, s)
	}
	if scheme != "github" {
		return Repo{}, fmt.Errorf("%w: %q: unknown scheme %q", ErrRepo, s, scheme)
	}
	path, base, ok := strings.Cut(rest, "@")
	if !ok {
		return Repo{}, fmt.Errorf("%w: %q: missing base", ErrRepo, s)
	}
	owner, name, ok := strings.Cut(path, "/")
	if !ok || owner == "" || name == "" || strings.Contains(name, "/") {
		return Repo{}, fmt.Errorf("%w: %q: want owner/name", ErrRepo, s)
	}
	if base == "" {
		return Repo{}, fmt.Errorf("%w: %q: empty base", ErrRepo, s)
	}
	return Repo{Scheme: scheme, Owner: owner, Name: name, Base: base}, nil
}

func (r Repo) String() string {
	return r.Scheme + ":" + r.Owner + "/" + r.Name + "@" + r.Base
}

// Slug is one unit of work inside a repo. It is also the branch name.
type Slug string

// Validate accepts the slug as one branch name.
func (s Slug) Validate() error {
	text := string(s)
	if text == "" {
		return fmt.Errorf("%w: empty", ErrSlug)
	}
	if strings.HasPrefix(text, "-") || strings.HasPrefix(text, ".") {
		return fmt.Errorf("%w: %q: bad start", ErrSlug, text)
	}
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "/") {
		return fmt.Errorf("%w: %q: bad end", ErrSlug, text)
	}
	if strings.Contains(text, "..") {
		return fmt.Errorf("%w: %q: contains ..", ErrSlug, text)
	}
	for _, part := range strings.Split(text, "/") {
		if part == "" {
			return fmt.Errorf("%w: %q: empty segment", ErrSlug, text)
		}
	}
	for _, r := range text {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '/', r == '.', r == '_', r == '-':
		default:
			return fmt.Errorf("%w: %q: character %q", ErrSlug, text, r)
		}
	}
	return nil
}

// Handle names one env. It is opaque to the agent.
type Handle string

// NewHandle builds the handle for a repo and slug.
func NewHandle(repo Repo, slug Slug) Handle {
	return Handle(repo.Scheme + ":" + repo.Owner + "/" + repo.Name + "/" + string(slug))
}

// IsClosed reports whether err means a spent handle.
func IsClosed(err error) bool {
	return errors.Is(err, ErrClosed)
}
