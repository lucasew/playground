package csmith

import (
	"fmt"
	"strings"
)

// defaultProgramGenerator mirrors the high-level upstream flow:
// initialize -> OutputHeader -> GenerateAllTypes -> GenerateFunctions -> Output.
type defaultProgramGenerator struct {
	opts       Options
	r          *rng
	pool       []CType
	b          strings.Builder
	info       compositeInfo
	env        envInfo
	funcs      []funcInfo
	dynGlobals []globalInfo
}

func newDefaultProgramGenerator(opts Options) *defaultProgramGenerator {
	return &defaultProgramGenerator{opts: opts}
}

func (g *defaultProgramGenerator) initialize() {
	g.r = newRNG(g.opts.Seed)
	g.pool = typePool(g.opts)
}

func (g *defaultProgramGenerator) outputHeader() {
	g.b.WriteString("/*\n")
	g.b.WriteString(" * This is a RANDOMLY GENERATED PROGRAM.\n")
	g.b.WriteString(" *\n")
	g.b.WriteString(" * Generator: csmith 2.3.0\n")
	g.b.WriteString(" * Git version: 30dccd7\n")
	g.b.WriteString(" * Options:   --seed ")
	g.b.WriteString(fmt.Sprintf("%d", g.opts.Seed))
	g.b.WriteString("\n")
	g.b.WriteString(" * Seed:      ")
	g.b.WriteString(fmt.Sprintf("%d", g.opts.Seed))
	g.b.WriteString("\n")
	g.b.WriteString(" */\n\n")
	g.b.WriteString("#include \"csmith.h\"\n\n")
	g.b.WriteString("static long __undefined;\n\n")
}

func (g *defaultProgramGenerator) generateAllTypes() {
	g.info = emitCompositeTypes(&g.b, g.r, g.opts, g.pool)
}

func (g *defaultProgramGenerator) generateFunctions() {
	// Upstream does not pre-generate a random global pool before Function::make_first.
	// Globals are introduced while function bodies are generated.
	g.env = envInfo{}
	g.funcs, g.dynGlobals = emitFunctionsUpstreamFlow(&g.b, g.r, g.opts, g.pool, g.opts.MaxBlockSize, g.env, g.info)
	if len(g.dynGlobals) > 0 {
		g.env.globals = append(g.env.globals, g.dynGlobals...)
	}
}

func (g *defaultProgramGenerator) output() {
	if g.opts.ComputeHash {
		emitComputeHashFunc(&g.b, g.env, g.info)
	}
	if g.opts.NoMain || len(g.funcs) == 0 {
		return
	}
	emitMain(&g.b, g.opts, g.env, g.info, g.funcs[0].name)
}

func (g *defaultProgramGenerator) goGenerator() string {
	g.outputHeader()
	g.generateAllTypes()
	g.generateFunctions()
	g.output()
	return g.b.String()
}
