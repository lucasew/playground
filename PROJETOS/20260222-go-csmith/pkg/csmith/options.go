package csmith

// Options is the canonical Go API contract for generation settings.
type Options struct {
	Seed uint64
}

// Defaults returns default generation settings.
func Defaults() Options {
	return Options{}
}
