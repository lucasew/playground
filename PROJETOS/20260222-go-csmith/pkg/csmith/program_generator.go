package csmith

// absProgramGenerator mirrors the minimal upstream abstraction
// used by AbsProgramGenerator::CreateInstance + goGenerator.
type absProgramGenerator interface {
	initialize()
	goGenerator() string
}

func createProgramGenerator(opts Options) absProgramGenerator {
	// Upstream picks DFSProgramGenerator when dfs-exhaustive is enabled.
	// This port still routes both modes through the same concrete generator.
	// The mode-specific behavior is handled by option normalization/filters.
	return newDefaultProgramGenerator(opts)
}
