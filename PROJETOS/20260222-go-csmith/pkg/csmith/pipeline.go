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
	g.b.WriteString("/* csmith-go: seed = ")
	g.b.WriteString(fmt.Sprintf("%d", g.opts.Seed))
	g.b.WriteString(" */\n")
	g.b.WriteString("/* int-size = ")
	g.b.WriteString(fmt.Sprintf("%d", g.opts.IntSize))
	g.b.WriteString(", ptr-size = ")
	g.b.WriteString(fmt.Sprintf("%d", g.opts.PointerSize))
	g.b.WriteString(" */\n")
	if g.opts.SafeMath || g.opts.ComputeHash {
		g.b.WriteString("#include \"csmith.h\"\n")
	} else {
		g.b.WriteString("#include <stdint.h>\n")
		g.b.WriteString("#include <stdio.h>\n")
	}
	g.b.WriteString("\n")
}

func (g *defaultProgramGenerator) generateAllTypes() {
	g.info = emitCompositeTypes(&g.b, g.r, g.opts, g.pool)
}

func (g *defaultProgramGenerator) generateFunctions() {
	g.env = emitGlobals(&g.b, g.r, g.opts, g.info, g.pool)
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
