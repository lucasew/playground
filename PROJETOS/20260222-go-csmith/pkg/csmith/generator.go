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

type envInfo struct {
	globals  int
	arrays   int
	pointers int
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

func emitGlobals(b *strings.Builder, r *rng, opts Options, info compositeInfo) envInfo {
	env := envInfo{}
	if opts.GlobalVariables {
		env.globals = 2 + int(r.upto(uint32(min(opts.MaxGlobals, 12))))
		for i := 0; i < env.globals; i++ {
			if opts.Consts && i == 0 {
				writeLine(b, 0, fmt.Sprintf("static const uint32_t g_%d = 0x%08Xu;", i, r.next31()))
			} else {
				writeLine(b, 0, fmt.Sprintf("static uint32_t g_%d = 0x%08Xu;", i, r.next31()))
			}
		}

		if opts.Arrays {
			env.arrays = 1 + int(r.upto(uint32(min(opts.MaxArrayDim, 2))))
			arrLen := max(2, min(opts.MaxArrayLenPerDim, 8))
			for i := 0; i < env.arrays; i++ {
				writeLine(b, 0, fmt.Sprintf("static uint32_t ga_%d[%d] = {0u};", i, arrLen))
			}
		}

		if opts.Pointers {
			start := 0
			if opts.Consts && env.globals > 1 {
				start = 1
			}
			env.pointers = min(max(env.globals-start, 0), 2)
			for i := 0; i < env.pointers; i++ {
				writeLine(b, 0, fmt.Sprintf("static uint32_t *gp_%d = &g_%d;", i, start+i))
			}
		}
	}

	for i := range info.structs {
		writeLine(b, 0, fmt.Sprintf("static S_%d gs_%d;", i, i))
	}
	for i := range info.unions {
		writeLine(b, 0, fmt.Sprintf("static U_%d gu_%d;", i, i))
	}
	writeLine(b, 0, "")
	return env
}

func emitFuncDecls(b *strings.Builder, funcs int) {
	for i := 0; i < funcs; i++ {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(void);", i))
	}
	writeLine(b, 0, "")
}

func emitArithmeticMutation(b *strings.Builder, r *rng, opts Options) {
	if opts.Divs && r.upto(3) == 0 {
		writeLine(b, 1, "x = x / ((x & 255u) + 1u);")
		return
	}
	if opts.UnaryPlusOperator && r.upto(3) == 0 {
		writeLine(b, 1, fmt.Sprintf("x = (+x) ^ 0x%08Xu;", r.next31()))
		return
	}
	writeLine(b, 1, fmt.Sprintf("x = (x << 1) ^ (x >> 1) ^ 0x%08Xu;", r.next31()))
}

func emitIncDecMutation(b *strings.Builder, r *rng, opts Options) bool {
	if !(opts.PreIncrOperator || opts.PreDecrOperator || opts.PostIncrOperator || opts.PostDecrOperator) {
		return false
	}
	switch r.upto(4) {
	case 0:
		if opts.PreIncrOperator {
			writeLine(b, 1, "++x;")
			return true
		}
	case 1:
		if opts.PreDecrOperator {
			writeLine(b, 1, "--x;")
			return true
		}
	case 2:
		if opts.PostIncrOperator {
			writeLine(b, 1, "x++;")
			return true
		}
	case 3:
		if opts.PostDecrOperator {
			writeLine(b, 1, "x--;")
			return true
		}
	}
	return false
}

func emitGlobalMutation(b *strings.Builder, r *rng, opts Options, env envInfo) {
	if env.globals == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	start := 0
	if opts.Consts && env.globals > 1 {
		start = 1
	}
	g := start + int(r.upto(uint32(max(env.globals-start, 1))))
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("g_%d += (x ^ 0x%08Xu);", g, r.next31()))
	} else {
		writeLine(b, 1, fmt.Sprintf("g_%d = g_%d + (x ^ 0x%08Xu);", g, g, r.next31()))
	}
}

func emitArrayMutation(b *strings.Builder, r *rng, opts Options, env envInfo) {
	if !opts.Arrays || env.arrays == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	ai := int(r.upto(uint32(env.arrays)))
	idxMask := max(1, min(opts.MaxArrayLenPerDim, 8)-1)
	writeLine(b, 1, fmt.Sprintf("ga_%d[x & %du] ^= x + 0x%08Xu;", ai, idxMask, r.next31()))
	if opts.EmbeddedAssigns {
		writeLine(b, 1, fmt.Sprintf("x = (ga_%d[x & %du] = ga_%d[x & %du] + 1u);", ai, idxMask, ai, idxMask))
	}
}

func emitPointerMutation(b *strings.Builder, r *rng, opts Options, env envInfo) {
	if !opts.Pointers || env.pointers == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	pi := int(r.upto(uint32(env.pointers)))
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("*gp_%d ^= x + 0x%08Xu;", pi, r.next31()))
	} else {
		writeLine(b, 1, fmt.Sprintf("*gp_%d = *gp_%d ^ (x + 0x%08Xu);", pi, pi, r.next31()))
	}
}

func emitFuncDefs(b *strings.Builder, r *rng, opts Options, funcs, maxBlock int, env envInfo, info compositeInfo) {
	extraCases := 0
	if len(info.structs) > 0 {
		extraCases++
	}
	if len(info.unions) > 0 {
		extraCases++
	}
	if opts.Arrays && env.arrays > 0 {
		extraCases++
	}
	if opts.Pointers && env.pointers > 0 {
		extraCases++
	}
	baseCases := 5
	maxChoice := uint32(baseCases + extraCases)

	for i := funcs - 1; i >= 0; i-- {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(void) {", i))
		if env.globals >= 2 {
			writeLine(b, 1, fmt.Sprintf("uint32_t x = (g_0 ^ 0x%08Xu) + (g_1 + 0x%08Xu);", r.next31(), r.next31()))
		} else {
			writeLine(b, 1, fmt.Sprintf("uint32_t x = 0x%08Xu;", r.next31()))
		}
		stmtCount := 2 + int(r.upto(uint32(maxBlock)))
		for s := 0; s < stmtCount; s++ {
			choice := r.upto(maxChoice)
			switch {
			case choice == 0:
				emitGlobalMutation(b, r, opts, env)
			case choice == 1:
				emitArithmeticMutation(b, r, opts)
			case choice == 2:
				if !emitIncDecMutation(b, r, opts) {
					writeLine(b, 1, fmt.Sprintf("if ((x & 1u) == 0u) { x += 0x%08Xu; } else { x ^= 0x%08Xu; }", r.next31(), r.next31()))
				}
			case choice == 3:
				if opts.Jumps {
					bound := 1 + int(r.upto(5))
					writeLine(b, 1, fmt.Sprintf("for (uint32_t i = 0; i < %du; ++i) { x += (i ^ 0x%08Xu); }", bound, r.next31()))
				} else {
					writeLine(b, 1, fmt.Sprintf("x += 0x%08Xu;", r.next31()))
				}
			case choice == 4 && len(info.structs) > 0:
				si := int(r.upto(uint32(len(info.structs))))
				fields := info.structs[si].fieldNames
				fi := int(r.upto(uint32(len(fields))))
				writeLine(b, 1, fmt.Sprintf("gs_%d.%s ^= x + 0x%08Xu;", si, fields[fi], r.next31()))
			case len(info.unions) > 0 && (choice == 5 || (len(info.structs) == 0 && choice == 4)):
				ui := int(r.upto(uint32(len(info.unions))))
				fields := info.unions[ui].fieldNames
				fi := int(r.upto(uint32(len(fields))))
				writeLine(b, 1, fmt.Sprintf("gu_%d.%s = (uint32_t)(x ^ 0x%08Xu);", ui, fields[fi], r.next31()))
				writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)gu_%d.%s;", ui, fields[fi]))
			case opts.Arrays && env.arrays > 0:
				emitArrayMutation(b, r, opts, env)
			default:
				emitPointerMutation(b, r, opts, env)
			}
		}
		if i+1 < funcs {
			writeLine(b, 1, fmt.Sprintf("x ^= func_%d();", i+1))
		}
		if env.globals > 0 {
			start := 0
			if opts.Consts && env.globals > 1 {
				start = 1
			}
			g := start + int(r.upto(uint32(max(env.globals-start, 1))))
			writeLine(b, 1, fmt.Sprintf("g_%d ^= x + 0x%08Xu;", g, r.next31()))
		}
		writeLine(b, 1, "return x;")
		writeLine(b, 0, "}")
		writeLine(b, 0, "")
	}
}

func emitMain(b *strings.Builder, opts Options, env envInfo, info compositeInfo) {
	useRuntime := opts.SafeMath || opts.ComputeHash
	if opts.AcceptArgc {
		writeLine(b, 0, "int main(int argc, char *argv[]) {")
		writeLine(b, 1, "int print_hash_value = 0;")
		if useRuntime {
			writeLine(b, 1, "if (argc == 2 && strcmp(argv[1], \"1\") == 0) print_hash_value = 1;")
		}
	} else {
		writeLine(b, 0, "int main(void) {")
	}

	if useRuntime {
		writeLine(b, 1, "platform_main_begin();")
		if opts.ComputeHash {
			writeLine(b, 1, "crc32_gentab();")
		}
	}
	writeLine(b, 1, "uint32_t checksum = 0u;")
	writeLine(b, 1, "uint32_t x = func_0();")
	if opts.ComputeHash {
		for i := 0; i < env.globals; i++ {
			if useRuntime {
				writeLine(b, 1, fmt.Sprintf("transparent_crc(g_%d, \"g_%d\", print_hash_value);", i, i))
			}
			writeLine(b, 1, fmt.Sprintf("checksum ^= g_%d + 0x9E3779B9u + (checksum << 6) + (checksum >> 2);", i))
		}
		for i := 0; i < env.arrays; i++ {
			if useRuntime {
				writeLine(b, 1, fmt.Sprintf("for (int ai = 0; ai < %d; ++ai) transparent_crc(ga_%d[ai], \"ga\", print_hash_value);", max(2, min(opts.MaxArrayLenPerDim, 8)), i))
			}
			writeLine(b, 1, fmt.Sprintf("checksum ^= ga_%d[0] + 0x9E3779B9u;", i))
		}
		for i, st := range info.structs {
			for _, f := range st.fieldNames {
				if useRuntime {
					writeLine(b, 1, fmt.Sprintf("transparent_crc(gs_%d.%s, \"gs\", print_hash_value);", i, f))
				}
				writeLine(b, 1, fmt.Sprintf("checksum ^= gs_%d.%s + 0x9E3779B9u;", i, f))
			}
		}
		for i, ut := range info.unions {
			if len(ut.fieldNames) > 0 {
				if useRuntime {
					writeLine(b, 1, fmt.Sprintf("transparent_crc((uint32_t)gu_%d.%s, \"gu\", print_hash_value);", i, ut.fieldNames[0]))
				}
				writeLine(b, 1, fmt.Sprintf("checksum ^= (uint32_t)gu_%d.%s + 0x9E3779B9u;", i, ut.fieldNames[0]))
			}
		}
		writeLine(b, 1, "checksum ^= x;")
		if useRuntime {
			writeLine(b, 1, "platform_main_end(crc32_context ^ 0xFFFFFFFFUL, print_hash_value);")
		} else {
			writeLine(b, 1, "printf(\"checksum = %u\\n\", checksum);")
		}
	}
	if !opts.ComputeHash && useRuntime {
		writeLine(b, 1, "platform_main_end(0u, 0);")
	}
	writeLine(b, 1, "return 0;")
	writeLine(b, 0, "}")
}

// Generate emits deterministic C code from options and seed.
func Generate(opts Options) (string, error) {
	var err error
	opts, err = opts.resolvePlatformInfo()
	if err != nil {
		return "", err
	}

	if err := opts.validate(); err != nil {
		return "", err
	}

	r := newRNG(opts.Seed)
	funcs := 1 + int(r.upto(uint32(min(opts.MaxFuncs, 8))))

	var b strings.Builder
	b.WriteString("/* csmith-go: seed = ")
	b.WriteString(fmt.Sprintf("%d", opts.Seed))
	b.WriteString(" */\n")
	b.WriteString("/* int-size = ")
	b.WriteString(fmt.Sprintf("%d", opts.IntSize))
	b.WriteString(", ptr-size = ")
	b.WriteString(fmt.Sprintf("%d", opts.PointerSize))
	b.WriteString(" */\n")
	if opts.SafeMath || opts.ComputeHash {
		b.WriteString("#include \"csmith.h\"\n")
	} else {
		b.WriteString("#include <stdint.h>\n")
		b.WriteString("#include <stdio.h>\n")
	}
	b.WriteString("\n")

	info := emitCompositeTypes(&b, r, opts)
	env := emitGlobals(&b, r, opts, info)
	emitFuncDecls(&b, funcs)
	emitFuncDefs(&b, r, opts, funcs, opts.MaxBlockSize, env, info)

	if !opts.NoMain {
		emitMain(&b, opts, env, info)
	}

	return b.String(), nil
}
