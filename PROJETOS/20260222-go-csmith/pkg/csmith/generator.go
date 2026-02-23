package csmith

import (
	"fmt"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func writeLine(b *strings.Builder, indent int, s string) {
	for i := 0; i < indent; i++ {
		b.WriteString("    ")
	}
	b.WriteString(s)
	b.WriteByte('\n')
}

func emitFuncDecls(b *strings.Builder, funcs int) {
	for i := 0; i < funcs; i++ {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(uint32_t p0, uint32_t p1);", i))
	}
	writeLine(b, 0, "")
}

func emitFuncDefs(b *strings.Builder, r *rng, funcs, globals, maxBlock int) {
	for i := funcs - 1; i >= 0; i-- {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(uint32_t p0, uint32_t p1) {", i))
		writeLine(b, 1, fmt.Sprintf("uint32_t x = (p0 ^ 0x%08Xu) + (p1 + 0x%08Xu);", r.next31(), r.next31()))
		stmtCount := 2 + int(r.upto(uint32(maxBlock)))
		for s := 0; s < stmtCount; s++ {
			switch r.upto(4) {
			case 0:
				g := int(r.upto(uint32(globals)))
				writeLine(b, 1, fmt.Sprintf("g_%d = g_%d + (x ^ 0x%08Xu);", g, g, r.next31()))
			case 1:
				writeLine(b, 1, fmt.Sprintf("x = (x << 1) ^ (x >> 1) ^ 0x%08Xu;", r.next31()))
			case 2:
				writeLine(b, 1, fmt.Sprintf("if ((x & 1u) == 0u) { x += 0x%08Xu; } else { x ^= 0x%08Xu; }", r.next31(), r.next31()))
			default:
				bound := 1 + int(r.upto(5))
				writeLine(b, 1, fmt.Sprintf("for (uint32_t i = 0; i < %du; ++i) { x += (i ^ 0x%08Xu); }", bound, r.next31()))
			}
		}
		if i+1 < funcs {
			writeLine(b, 1, fmt.Sprintf("x ^= func_%d(x ^ 0x%08Xu, p0 + 0x%08Xu);", i+1, r.next31(), r.next31()))
		}
		g := int(r.upto(uint32(globals)))
		writeLine(b, 1, fmt.Sprintf("g_%d ^= x + 0x%08Xu;", g, r.next31()))
		writeLine(b, 1, "return x;")
		writeLine(b, 0, "}")
		writeLine(b, 0, "")
	}
}

// Generate emits deterministic C code from options and seed.
func Generate(opts Options) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}

	r := newRNG(opts.Seed)
	funcs := 1 + int(r.upto(uint32(min(opts.MaxFuncs, 8))))
	globals := 2 + int(r.upto(uint32(min(opts.MaxGlobals, 12))))

	var b strings.Builder
	b.WriteString("/* csmith-go: seed = ")
	b.WriteString(fmt.Sprintf("%d", opts.Seed))
	b.WriteString(" */\n")
	b.WriteString("#include <stdint.h>\n")
	b.WriteString("#include <stdio.h>\n")
	b.WriteString("\n")

	for i := 0; i < globals; i++ {
		writeLine(&b, 0, fmt.Sprintf("static uint32_t g_%d = 0x%08Xu;", i, r.next31()))
	}
	writeLine(&b, 0, "")

	emitFuncDecls(&b, funcs)
	emitFuncDefs(&b, r, funcs, globals, opts.MaxBlockSize)

	if !opts.NoMain {
		if opts.AcceptArgc {
			writeLine(&b, 0, "int main(int argc, char *argv[]) {")
			writeLine(&b, 1, "(void)argc;")
			writeLine(&b, 1, "(void)argv;")
		} else {
			writeLine(&b, 0, "int main(void) {")
		}
		writeLine(&b, 1, "uint32_t checksum = 0u;")
		writeLine(&b, 1, "uint32_t x = func_0(g_0, g_1);")
		if opts.ComputeHash {
			for i := 0; i < globals; i++ {
				writeLine(&b, 1, fmt.Sprintf("checksum ^= g_%d + 0x9E3779B9u + (checksum << 6) + (checksum >> 2);", i))
			}
			writeLine(&b, 1, "checksum ^= x;")
			writeLine(&b, 1, "printf(\"checksum = %u\\n\", checksum);")
		}
		writeLine(&b, 1, "return 0;")
		writeLine(&b, 0, "}")
	}

	return b.String(), nil
}
