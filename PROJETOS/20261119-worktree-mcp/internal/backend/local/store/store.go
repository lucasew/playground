// Package store is the local backend's disk layout: the clones it keeps and
// the worktrees it cuts from them.
//
// It is an implementation detail of the local backend. The server owns which
// envs exist; this package only answers where the local ones live on disk.
//
// The store is safe to use from several processes: every question it answers
// is read back from git, and the per-repo lock serializes concurrent clones.
package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lewtec/lewkit/x/errs"
	"github.com/lewtec/lewkit/x/git"
)

var (
	// ErrRepo means the repo identity cannot be used on this host.
	ErrRepo = errs.Wrap(errs.ErrInvalid, "repo")
	// ErrCheckout means the worktree could not be ensured.
	ErrCheckout = errs.Wrap(errs.ErrUnavailable, "checkout")
)

// Repo is a repository the store can resolve, independent of how the server
// spells it. Owner and Name make the dirname and the clone URL; Base picks
// the branch a new worktree starts from.
type Repo struct {
	Owner string
	Name  string
	Base  string
}

// Remote is the clone URL for a repo. It is a field, not a constant, so a
// test or an enterprise install can point the store at another host.
func (r Repo) Remote(host string) string {
	return "https://" + host + "/" + r.Owner + "/" + r.Name + ".git"
}

// Dir is the directory name a repo's clone lives under.
func (r Repo) Dir() string {
	return r.Owner + "-" + r.Name
}

// Store keeps clones under Clones and worktrees under Checkouts.
type Store struct {
	Git *git.Git
	// Clones holds one clone per repo.
	Clones string
	// Checkouts holds one linked worktree per unit of work.
	Checkouts string
	// Declared maps owner/name to a clone the user already has. A declared
	// repo is never cloned; the store uses it as-is.
	Declared map[string]string
	// Host is the git host a clone URL names.
	Host string
	// Remote overrides the clone URL for every repo. Tests and local
	// mirrors set it; production leaves it empty.
	Remote func(Repo) string

	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// New returns a store rooted at root, cloning from host.
func New(root, host string) *Store {
	s := &Store{
		Clones:    filepath.Join(root, "clones"),
		Checkouts: filepath.Join(root, "worktrees"),
		Declared:  make(map[string]string),
		Host:      host,
		locks:     make(map[string]*sync.Mutex),
	}
	if s.Host == "" {
		s.Host = "github.com"
	}
	return s
}

func (s *Store) git() *git.Git {
	if s.Git != nil {
		return s.Git
	}
	return &git.Git{}
}

// Clone returns the main checkout for repo, cloning it when the user did
// not declare one. It reports whether the clone was created now.
func (s *Store) Clone(ctx context.Context, repo Repo) (string, bool, error) {
	if declared, ok := s.Declared[repo.Dir()]; ok {
		if declared == "" {
			return "", false, fmt.Errorf("%w: %s: declared checkout is empty", ErrRepo, repo.Dir())
		}
		if ok, err := s.isCheckout(ctx, declared); err != nil {
			return "", false, err
		} else if !ok {
			return "", false, fmt.Errorf("%w: %s: declared checkout is not a repo: %s", ErrRepo, repo.Dir(), declared)
		}
		return git.Resolve(declared), false, nil
	}
	path := filepath.Join(s.Clones, repo.Dir())
	unlock := s.lock(repo.Dir())
	defer unlock()
	if ok, err := s.isCheckout(ctx, path); err != nil {
		return "", false, err
	} else if ok {
		return path, false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", false, fmt.Errorf("%w: %s: %w", ErrRepo, repo.Dir(), err)
	}
	// A half-finished clone from an interrupted run would fail the next one.
	if _, err := os.Stat(path); err == nil {
		if err := os.RemoveAll(path); err != nil {
			return "", false, fmt.Errorf("%w: %s: %w", ErrRepo, repo.Dir(), err)
		}
	}
	if err := s.git().Run(ctx, "", "clone", "--quiet", s.remote(repo), path); err != nil {
		return "", false, fmt.Errorf("%w: %s: %w", ErrRepo, repo.Dir(), err)
	}
	return path, true, nil
}

func (s *Store) remote(repo Repo) string {
	if s.Remote != nil {
		return s.Remote(repo)
	}
	return repo.Remote(s.Host)
}

// Checkout ensures one linked worktree for branch on main, cut from repo.Base
// when it is new, and returns its path. A worktree already on that branch is
// reused, so a second call creates nothing.
func (s *Store) Checkout(ctx context.Context, repo Repo, main, branch string) (string, bool, error) {
	g := s.git()
	if existing, ok := g.LinkedWorktree(ctx, main, branch); ok {
		return git.Resolve(existing), false, nil
	}
	// A main checkout already standing on the branch is the worktree.
	head, ok := g.Branch(ctx, main)
	if !ok {
		return "", false, fmt.Errorf("%w: %s: not a git checkout", ErrRepo, main)
	}
	if head == branch {
		return git.Resolve(main), false, nil
	}
	path := filepath.Join(s.Checkouts, filepath.Base(main), SafeName(branch))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", false, fmt.Errorf("%w: %s: %w", ErrCheckout, branch, err)
	}
	base := repo.Base
	if base == "" {
		base = "HEAD"
	}
	if err := g.Run(ctx, main, "worktree", "add", "-B", branch, path, base); err != nil {
		return "", false, fmt.Errorf("%w: %s: %w", ErrCheckout, branch, err)
	}
	g.Clear()
	return git.Resolve(path), true, nil
}

// Release removes one linked worktree. It never touches a main checkout, so
// a branch that is already on the main checkout is left alone.
func (s *Store) Release(ctx context.Context, main, path string) error {
	if git.Resolve(path) == git.Resolve(main) {
		return nil
	}
	return s.git().Run(ctx, main, "worktree", "remove", "--force", path)
}

func (s *Store) isCheckout(ctx context.Context, path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	_, ok := s.git().Branch(ctx, path)
	return ok, nil
}

// lock returns an unlock func for one repo name. Two clones of the same repo
// never race inside one process.
func (s *Store) lock(name string) func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.locks[name]
	if !ok {
		m = &sync.Mutex{}
		s.locks[name] = m
	}
	m.Lock()
	return m.Unlock
}

// SafeName turns a branch into one path segment.
func SafeName(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}
