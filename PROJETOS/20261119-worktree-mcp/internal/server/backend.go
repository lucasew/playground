package server

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/backend"
	"github.com/lucasew/env-mcp/internal/backend/local"
	"github.com/lucasew/env-mcp/internal/backend/local/store"
	"github.com/lucasew/env-mcp/internal/env"
)

// Config is the server's declaration of local checkouts.
type Config struct {
	// Repos maps github:owner/repo to a local checkout path. A repo named
	// here is never cloned; the backend uses the checkout as-is.
	Repos map[string]string `json:"repos"`
	// Workdir is the directory clones and worktrees live under.
	Workdir string `json:"workdir"`
}

// LoadConfig reads config.json from stateDir. A missing file is an empty
// config, which means every repo is cloned on demand.
func LoadConfig(stateDir string) (Config, error) {
	b, err := os.ReadFile(filepath.Join(stateDir, "config.json"))
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// DefaultBackend builds the local backend over the server's state directory.
//
// The server keeps the registry; this backend owns the mechanics — clones,
// worktrees, and the terminal surface commands run on. Nothing in its internals
// reaches the caller.
func DefaultBackend(stateDir string) backend.Provisioner {
	cfg, err := LoadConfig(stateDir)
	if err != nil {
		cfg = Config{}
	}
	root := cfg.Workdir
	if root == "" {
		root = filepath.Join(stateDir, "checkouts")
	}
	b := local.New(root)
	b.State = stateDir
	b.Store.Git = &git.Git{}
	// Config keys are github:owner/repo, the store keys on owner-name.
	for key, path := range cfg.Repos {
		repo, err := env.ParseRepo(key + "@main")
		if err != nil {
			continue
		}
		r := store.Repo{Owner: repo.Owner, Name: repo.Name, Base: repo.Base}
		b.Store.Declared[r.Dir()] = path
	}
	return b
}
