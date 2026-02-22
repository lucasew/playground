package options

// Options stores CLI options for the generator.
type Options struct {
	Seed       uint64
	SeedSet    bool
	OutputPath string
}

func Defaults() Options {
	return Options{}
}
