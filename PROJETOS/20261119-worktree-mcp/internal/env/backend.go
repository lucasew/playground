// Package env's backend ensures one linked worktree per env.
package env

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lewtec/lewkit/x/git"
)

// Backend places envs on the local filesystem and Herdr's session.
type Backend struct {
	Git     *git.Git
	Workdir string
	// Checkouts maps github:owner/repo to an existing local checkout.
	Checkouts map[string]string
	// Layout places a new env somewhere visible. Nil keeps it on disk only.
	Layout Layout
}

// Layout is the standard layout the server derives for one env.
type Layout interface {
	// Open opens or resumes the env and returns its place numbers.
	Open(ctx context.Context, root, label string) ([]int, error)
	// Close tears down the places the server created.
	Close(ctx context.Context, root string, places []int) error
}

// Checkout returns the declared main checkout for repo.
func (b *Backend) Checkout(repo Repo) (string, bool) {
	path, ok := b.Checkouts[repo.Scheme+":"+repo.Owner+"/"+repo.Name]
	return path, ok
}

// Ensure returns the worktree root for repo and slug, creating the worktree
// when absent. A second call for the same pair returns the same root.
func (b *Backend) Ensure(ctx context.Context, repo Repo, slug Slug) (string, bool, error) {
	checkout, ok := b.Checkouts[repo.Scheme+":"+repo.Owner+"/"+repo.Name]
	if !ok {
		return "", false, fmt.Errorf("%w: %s: no local checkout declared", ErrRepo, repo)
	}
	g := b.Git
	if g == nil {
		g = &git.Git{}
	}
	if existing, ok := g.LinkedWorktree(ctx, checkout, string(slug)); ok {
		return git.Resolve(existing), false, nil
	}
	head, ok := g.Branch(ctx, checkout)
	if !ok {
		return "", false, fmt.Errorf("%w: %s: not a git checkout", ErrRepo, checkout)
	}
	if head == string(slug) {
		return git.Resolve(checkout), false, nil
	}
	path := filepath.Join(b.Workdir, repo.Owner+"-"+repo.Name, strings.ReplaceAll(string(slug), "/", "-"))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", false, err
	}
	if err := g.Run(ctx, checkout, "worktree", "add", "-B", string(slug), path, repo.Base); err != nil {
		return "", false, err
	}
	g.Clear()
	return git.Resolve(path), true, nil
}
