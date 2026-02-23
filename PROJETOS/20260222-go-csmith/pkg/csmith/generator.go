package csmith

import (
	"fmt"
	"strings"
)

type structTypeInfo struct {
	fieldNames []string
}

type unionTypeInfo struct {
	fieldNames []string
}

type compositeInfo struct {
	structs []structTypeInfo
	unions  []unionTypeInfo
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
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

func emitCompositeTypes(b *strings.Builder, r *rng, opts Options) compositeInfo {
	info := compositeInfo{}

	if opts.PackedStruct {
		writeLine(b, 0, "#if defined(__GNUC__) || defined(__clang__)")
		writeLine(b, 0, "#define CSMITH_GO_PACKED __attribute__((packed))")
		writeLine(b, 0, "#else")
		writeLine(b, 0, "#define CSMITH_GO_PACKED")
		writeLine(b, 0, "#endif")
		writeLine(b, 0, "")
	}

	if opts.Structs {
		structTypes := 1 + int(r.upto(uint32(min(opts.MaxStructFields, 2))))
		for sidx := 0; sidx < structTypes; sidx++ {
			fieldCount := 1 + int(r.upto(uint32(max(1, min(opts.MaxStructFields, 6)))))
			st := structTypeInfo{fieldNames: make([]string, 0, fieldCount)}
			if opts.PackedStruct {
				writeLine(b, 0, fmt.Sprintf("typedef struct CSMITH_GO_PACKED S_%d {", sidx))
			} else {
				writeLine(b, 0, fmt.Sprintf("typedef struct S_%d {", sidx))
			}
			for f := 0; f < fieldCount; f++ {
				if opts.Bitfields && r.upto(4) == 0 {
					name := fmt.Sprintf("bf_%d", f)
					width := 1 + int(r.upto(31))
					writeLine(b, 1, fmt.Sprintf("uint32_t %s : %d;", name, width))
					st.fieldNames = append(st.fieldNames, name)
					continue
				}
				name := fmt.Sprintf("f_%d", f)
				writeLine(b, 1, fmt.Sprintf("uint32_t %s;", name))
				st.fieldNames = append(st.fieldNames, name)
			}
			writeLine(b, 0, fmt.Sprintf("} S_%d;", sidx))
			writeLine(b, 0, "")
			info.structs = append(info.structs, st)
		}
	}

	if opts.Unions {
		unionTypes := 1 + int(r.upto(uint32(min(opts.MaxUnionFields, 2))))
		for uidx := 0; uidx < unionTypes; uidx++ {
			fieldCount := 2 + int(r.upto(uint32(max(1, min(opts.MaxUnionFields, 4)))))
			ut := unionTypeInfo{fieldNames: make([]string, 0, fieldCount)}
			writeLine(b, 0, fmt.Sprintf("typedef union U_%d {", uidx))
			for f := 0; f < fieldCount; f++ {
				name := fmt.Sprintf("u_%d", f)
				if f%2 == 0 {
					writeLine(b, 1, fmt.Sprintf("uint32_t %s;", name))
				} else {
					writeLine(b, 1, fmt.Sprintf("int32_t %s;", name))
				}
				ut.fieldNames = append(ut.fieldNames, name)
			}
			writeLine(b, 0, fmt.Sprintf("} U_%d;", uidx))
			writeLine(b, 0, "")
			info.unions = append(info.unions, ut)
		}
	}

	return info
}

func emitGlobals(b *strings.Builder, r *rng, globals int, info compositeInfo) {
	for i := 0; i < globals; i++ {
		writeLine(b, 0, fmt.Sprintf("static uint32_t g_%d = 0x%08Xu;", i, r.next31()))
	}
	for i := range info.structs {
		writeLine(b, 0, fmt.Sprintf("static S_%d gs_%d;", i, i))
	}
	for i := range info.unions {
		writeLine(b, 0, fmt.Sprintf("static U_%d gu_%d;", i, i))
	}
	writeLine(b, 0, "")
}

func emitFuncDecls(b *strings.Builder, funcs int) {
	for i := 0; i < funcs; i++ {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(uint32_t p0, uint32_t p1);", i))
	}
	writeLine(b, 0, "")
}

func emitFuncDefs(b *strings.Builder, r *rng, funcs, globals, maxBlock int, info compositeInfo) {
	extraCases := 0
	if len(info.structs) > 0 {
		extraCases++
	}
	if len(info.unions) > 0 {
		extraCases++
	}
	baseCases := 4
	maxChoice := uint32(baseCases + extraCases)

	for i := funcs - 1; i >= 0; i-- {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(uint32_t p0, uint32_t p1) {", i))
		writeLine(b, 1, fmt.Sprintf("uint32_t x = (p0 ^ 0x%08Xu) + (p1 + 0x%08Xu);", r.next31(), r.next31()))
		stmtCount := 2 + int(r.upto(uint32(maxBlock)))
		for s := 0; s < stmtCount; s++ {
			choice := r.upto(maxChoice)
			switch {
			case choice == 0:
				g := int(r.upto(uint32(globals)))
				writeLine(b, 1, fmt.Sprintf("g_%d = g_%d + (x ^ 0x%08Xu);", g, g, r.next31()))
			case choice == 1:
				writeLine(b, 1, fmt.Sprintf("x = (x << 1) ^ (x >> 1) ^ 0x%08Xu;", r.next31()))
			case choice == 2:
				writeLine(b, 1, fmt.Sprintf("if ((x & 1u) == 0u) { x += 0x%08Xu; } else { x ^= 0x%08Xu; }", r.next31(), r.next31()))
			case choice == 3:
				bound := 1 + int(r.upto(5))
				writeLine(b, 1, fmt.Sprintf("for (uint32_t i = 0; i < %du; ++i) { x += (i ^ 0x%08Xu); }", bound, r.next31()))
			case len(info.structs) > 0 && choice == 4:
				si := int(r.upto(uint32(len(info.structs))))
				fields := info.structs[si].fieldNames
				fi := int(r.upto(uint32(len(fields))))
				writeLine(b, 1, fmt.Sprintf("gs_%d.%s ^= x + 0x%08Xu;", si, fields[fi], r.next31()))
			case len(info.unions) > 0:
				ui := int(r.upto(uint32(len(info.unions))))
				fields := info.unions[ui].fieldNames
				fi := int(r.upto(uint32(len(fields))))
				writeLine(b, 1, fmt.Sprintf("gu_%d.%s = (uint32_t)(x ^ 0x%08Xu);", ui, fields[fi], r.next31()))
				writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)gu_%d.%s;", ui, fields[fi]))
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

func emitMain(b *strings.Builder, opts Options, globals int, info compositeInfo) {
	if opts.AcceptArgc {
		writeLine(b, 0, "int main(int argc, char *argv[]) {")
		writeLine(b, 1, "(void)argc;")
		writeLine(b, 1, "(void)argv;")
	} else {
		writeLine(b, 0, "int main(void) {")
	}

	writeLine(b, 1, "uint32_t checksum = 0u;")
	writeLine(b, 1, "uint32_t x = func_0(g_0, g_1);")
	if opts.ComputeHash {
		for i := 0; i < globals; i++ {
			writeLine(b, 1, fmt.Sprintf("checksum ^= g_%d + 0x9E3779B9u + (checksum << 6) + (checksum >> 2);", i))
		}
		for i, st := range info.structs {
			for _, f := range st.fieldNames {
				writeLine(b, 1, fmt.Sprintf("checksum ^= gs_%d.%s + 0x9E3779B9u;", i, f))
			}
		}
		for i, ut := range info.unions {
			if len(ut.fieldNames) > 0 {
				writeLine(b, 1, fmt.Sprintf("checksum ^= (uint32_t)gu_%d.%s + 0x9E3779B9u;", i, ut.fieldNames[0]))
			}
		}
		writeLine(b, 1, "checksum ^= x;")
		writeLine(b, 1, "printf(\"checksum = %u\\n\", checksum);")
	}
	writeLine(b, 1, "return 0;")
	writeLine(b, 0, "}")
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

	info := emitCompositeTypes(&b, r, opts)
	emitGlobals(&b, r, globals, info)
	emitFuncDecls(&b, funcs)
	emitFuncDefs(&b, r, funcs, globals, opts.MaxBlockSize, info)

	if !opts.NoMain {
		emitMain(&b, opts, globals, info)
	}

	return b.String(), nil
}
