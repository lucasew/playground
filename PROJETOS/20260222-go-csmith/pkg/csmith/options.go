package csmith

import "fmt"

// Options is the canonical Go API contract for generation settings.
type Options struct {
	Seed         uint64
	MaxFuncs     int
	MaxBlockSize int
	MaxGlobals   int
}

// Defaults returns default generation settings.
func Defaults() Options {
	return Options{
		MaxFuncs:     4,
		MaxBlockSize: 6,
		MaxGlobals:   6,
	}
}

func (o Options) validate() error {
	if o.MaxFuncs < 1 {
		return fmt.Errorf("max-funcs must be at least 1")
	}
	if o.MaxBlockSize < 1 {
		return fmt.Errorf("max-block-size must be at least 1")
	}
	if o.MaxGlobals < 1 {
		return fmt.Errorf("max-globals must be at least 1")
	}
	return nil
}
