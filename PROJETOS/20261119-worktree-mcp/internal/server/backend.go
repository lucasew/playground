package server

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lewtec/lewkit/x/git"
	"github.com/lucasew/env-mcp/internal/env"
)

// Config is the server's declaration of local checkouts.
type Config struct {
	// Repos maps github:owner/repo to a local checkout path.
	Repos map[string]string `json:"repos"`
	// Workdir is the directory new worktrees are created under.
	Workdir string `json:"workdir"`
}

// LoadConfig reads config.json from stateDir. A missing file is an empty
// config, so a fresh install answers with "no local checkout declared".
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

// DefaultBackend builds the filesystem backend and the Herdr layout from
// the state directory config.
func DefaultBackend(stateDir string) *env.Backend {
	cfg, err := LoadConfig(stateDir)
	if err != nil {
		cfg = Config{}
	}
	workdir := cfg.Workdir
	if workdir == "" {
		workdir = filepath.Join(stateDir, "worktrees")
	}
	return &env.Backend{
		Git:       &git.Git{},
		Workdir:   workdir,
		Checkouts: cfg.Repos,
	}
}
