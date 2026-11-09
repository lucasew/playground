// Package env's backend turns one repo identity and slug into one checkout.
package env

import (
	"context"
	"fmt"

	"github.com/lucasew/env-mcp/internal/store"
)

// Backend places envs on the local filesystem. It owns no path logic of its
// own: the store decides where clones and worktrees live, so every backend
// shares one rule for the layout.
type Backend struct {
	Store *store.Store
}

// Place is one env's checkout and the clone it was cut from.
type Place struct {
	// Root is the worktree for this env.
	Root string
	// Source is the main checkout the worktree belongs to.
	Source string
	// Created reports whether the worktree was made now.
	Created bool
}

// Checkout returns the main checkout for repo, cloning it when the user did
// not declare one.
func (b *Backend) Checkout(ctx context.Context, repo Repo) (string, bool, error) {
	if b.Store == nil {
		return "", false, fmt.Errorf("%w: no store configured", ErrRepo)
	}
	return b.Store.Clone(ctx, storeRepo(repo))
}

// Ensure returns the worktree for repo and slug, creating it when absent. A
// second call for the same pair returns the same root.
func (b *Backend) Ensure(ctx context.Context, repo Repo, slug Slug) (Place, error) {
	if b.Store == nil {
		return Place{}, fmt.Errorf("%w: no store configured", ErrRepo)
	}
	main, cloned, err := b.Store.Clone(ctx, storeRepo(repo))
	if err != nil {
		return Place{}, err
	}
	root, made, err := b.Store.Checkout(ctx, storeRepo(repo), main, string(slug))
	if err != nil {
		return Place{}, err
	}
	return Place{Root: root, Source: main, Created: cloned || made}, nil
}

// Release removes the worktree for one env unless it is the main checkout.
func (b *Backend) Release(ctx context.Context, place Place) error {
	if b.Store == nil {
		return nil
	}
	return b.Store.Release(ctx, place.Source, place.Root)
}

// storeRepo maps the env identity onto a store repo.
func storeRepo(repo Repo) store.Repo {
	return store.Repo{Owner: repo.Owner, Name: repo.Name, Base: repo.Base}
}
