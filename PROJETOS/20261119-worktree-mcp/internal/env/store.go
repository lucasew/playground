package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Record is the persisted state of one env.
//
// It is the server's registry entry. Paths and terminal identifiers are not
// part of it: the backend's opaque token is the only thing recorded about
// where the env lives.
type Record struct {
	Handle Handle `json:"handle"`
	Repo   string `json:"repo"`
	Slug   string `json:"slug"`
	Base   string `json:"base"`
	// Backend is the provisioner that issued Token.
	Backend string `json:"backend"`
	// Token is the backend's handle for this env.
	Token   string `json:"token"`
	Closed  bool   `json:"closed"`
	Created bool   `json:"created"`
}

// Store keeps one JSON record per env in a directory the agent never sees.
type Store struct {
	dir string
}

// NewStore opens the record directory, creating it when absent.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) file(h Handle) string {
	return filepath.Join(s.dir, HashHandle(h)+".json")
}

// Dir is the directory holding the registry records.
func (s *Store) Dir() string { return s.dir }

// HashHandle turns a handle into a filename-safe key.
func HashHandle(h Handle) string {
	sum := uint64(14695981039346656037)
	for i := 0; i < len(h); i++ {
		sum ^= uint64(h[i])
		sum *= 1099511628211
	}
	return fmt.Sprintf("%016x", sum)
}

// Get reads one record. A missing record is nil and no error.
func (s *Store) Get(h Handle) (*Record, error) {
	b, err := os.ReadFile(s.file(h))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec Record
	if err := json.Unmarshal(b, &rec); err != nil {
		return nil, fmt.Errorf("store: %s: %w", s.file(h), err)
	}
	return &rec, nil
}

// Put writes one record.
func (s *Store) Put(rec Record) error {
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file(rec.Handle), b, 0o600)
}

// List returns every record.
func (s *Store) List() ([]Record, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []Record
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var rec Record
		if err := json.Unmarshal(b, &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}
