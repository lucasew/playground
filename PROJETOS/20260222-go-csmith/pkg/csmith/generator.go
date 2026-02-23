package csmith

import (
	"fmt"
	"strings"
)

type structTypeInfo struct {
	fields []fieldInfo
}

type unionTypeInfo struct {
	fields []fieldInfo
}

type fieldInfo struct {
	name     string
	ctype    CType
	bitfield bool
	bitWidth int
}

type globalInfo struct {
	name    string
	ctype   CType
	isConst bool
}

type arrayInfo struct {
	name  string
	ctype CType
	len   int
}

type pointerInfo struct {
	name     string
	target   string
	targetTy CType
}

type compositeInfo struct {
	structs []structTypeInfo
	unions  []unionTypeInfo
}

type envInfo struct {
	globals  []globalInfo
	arrays   []arrayInfo
	pointers []pointerInfo
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

func emitCompositeTypes(b *strings.Builder, r *rng, opts Options, pool []CType) compositeInfo {
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
			st := structTypeInfo{fields: make([]fieldInfo, 0, fieldCount)}
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
					st.fields = append(st.fields, fieldInfo{
						name: name, ctype: CType{Name: "uint32_t", Bits: 32}, bitfield: true, bitWidth: width,
					})
					continue
				}
				name := fmt.Sprintf("f_%d", f)
				t := pickType(r, pool)
				writeLine(b, 1, fmt.Sprintf("%s %s;", t.Name, name))
				st.fields = append(st.fields, fieldInfo{name: name, ctype: t})
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
			ut := unionTypeInfo{fields: make([]fieldInfo, 0, fieldCount)}
			writeLine(b, 0, fmt.Sprintf("typedef union U_%d {", uidx))
			for f := 0; f < fieldCount; f++ {
				name := fmt.Sprintf("u_%d", f)
				t := pickType(r, pool)
				writeLine(b, 1, fmt.Sprintf("%s %s;", t.Name, name))
				ut.fields = append(ut.fields, fieldInfo{name: name, ctype: t})
			}
			writeLine(b, 0, fmt.Sprintf("} U_%d;", uidx))
			writeLine(b, 0, "")
			info.unions = append(info.unions, ut)
		}
	}

	return info
}

func emitGlobals(b *strings.Builder, r *rng, opts Options, info compositeInfo, pool []CType) envInfo {
	env := envInfo{}
	if opts.GlobalVariables {
		globalCount := 2 + int(r.upto(uint32(min(opts.MaxGlobals, 12))))
		env.globals = make([]globalInfo, 0, globalCount)
		for i := 0; i < globalCount; i++ {
			g := globalInfo{
				name:    fmt.Sprintf("g_%d", i),
				ctype:   pickType(r, pool),
				isConst: opts.Consts && i == 0,
			}
			lit := castLiteral(g.ctype, fmt.Sprintf("0x%08Xu", r.next31()))
			if g.isConst {
				writeLine(b, 0, fmt.Sprintf("static const %s %s = %s;", g.ctype.Name, g.name, lit))
			} else {
				writeLine(b, 0, fmt.Sprintf("static %s %s = %s;", g.ctype.Name, g.name, lit))
			}
			env.globals = append(env.globals, g)
		}

		if opts.Arrays {
			arrayCount := 1 + int(r.upto(uint32(min(opts.MaxArrayDim, 2))))
			arrLen := max(2, min(opts.MaxArrayLenPerDim, 8))
			env.arrays = make([]arrayInfo, 0, arrayCount)
			for i := 0; i < arrayCount; i++ {
				ai := arrayInfo{
					name:  fmt.Sprintf("ga_%d", i),
					ctype: pickType(r, pool),
					len:   arrLen,
				}
				writeLine(b, 0, fmt.Sprintf("static %s %s[%d] = {0};", ai.ctype.Name, ai.name, arrLen))
				env.arrays = append(env.arrays, ai)
			}
		}

		if opts.Pointers {
			start := 0
			if opts.Consts && len(env.globals) > 1 {
				start = 1
			}
			ptrCount := min(max(len(env.globals)-start, 0), 2)
			env.pointers = make([]pointerInfo, 0, ptrCount)
			for i := 0; i < ptrCount; i++ {
				target := env.globals[start+i]
				p := pointerInfo{
					name:     fmt.Sprintf("gp_%d", i),
					target:   target.name,
					targetTy: target.ctype,
				}
				writeLine(b, 0, fmt.Sprintf("static %s *%s = &%s;", target.ctype.Name, p.name, p.target))
				env.pointers = append(env.pointers, p)
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
	if len(env.globals) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	start := 0
	if opts.Consts && len(env.globals) > 1 {
		start = 1
	}
	gix := start + int(r.upto(uint32(max(len(env.globals)-start, 1))))
	g := env.globals[gix]
	rhs := castLiteral(g.ctype, fmt.Sprintf("(x ^ 0x%08Xu)", r.next31()))
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", g.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s + %s;", g.name, g.name, rhs))
	}
}

func emitArrayMutation(b *strings.Builder, r *rng, opts Options, env envInfo) {
	if !opts.Arrays || len(env.arrays) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	ai := env.arrays[int(r.upto(uint32(len(env.arrays))))]
	idxMask := max(1, min(opts.MaxArrayLenPerDim, 8)-1)
	rhs := castLiteral(ai.ctype, fmt.Sprintf("(x + 0x%08Xu)", r.next31()))
	writeLine(b, 1, fmt.Sprintf("%s[x & %du] ^= %s;", ai.name, idxMask, rhs))
	if opts.EmbeddedAssigns {
		one := castLiteral(ai.ctype, "1u")
		writeLine(b, 1, fmt.Sprintf("x = (%s[x & %du] = %s[x & %du] + %s);", ai.name, idxMask, ai.name, idxMask, one))
	}
}

func emitPointerMutation(b *strings.Builder, r *rng, opts Options, env envInfo) {
	if !opts.Pointers || len(env.pointers) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	pi := env.pointers[int(r.upto(uint32(len(env.pointers))))]
	rhs := castLiteral(pi.targetTy, fmt.Sprintf("(x + 0x%08Xu)", r.next31()))
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("*%s ^= %s;", pi.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("*%s = *%s ^ %s;", pi.name, pi.name, rhs))
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
	if opts.Arrays && len(env.arrays) > 0 {
		extraCases++
	}
	if opts.Pointers && len(env.pointers) > 0 {
		extraCases++
	}
	baseCases := 5
	maxChoice := uint32(baseCases + extraCases)

	for i := funcs - 1; i >= 0; i-- {
		writeLine(b, 0, fmt.Sprintf("static uint32_t func_%d(void) {", i))
		if len(env.globals) >= 2 {
			writeLine(
				b,
				1,
				fmt.Sprintf(
					"uint32_t x = ((uint32_t)%s ^ 0x%08Xu) + ((uint32_t)%s + 0x%08Xu);",
					env.globals[0].name, r.next31(), env.globals[1].name, r.next31(),
				),
			)
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
				fields := info.structs[si].fields
				fi := int(r.upto(uint32(len(fields))))
				f := fields[fi]
				writeLine(b, 1, fmt.Sprintf("gs_%d.%s ^= %s;", si, f.name, castLiteral(f.ctype, fmt.Sprintf("(x + 0x%08Xu)", r.next31()))))
			case len(info.unions) > 0 && (choice == 5 || (len(info.structs) == 0 && choice == 4)):
				ui := int(r.upto(uint32(len(info.unions))))
				fields := info.unions[ui].fields
				fi := int(r.upto(uint32(len(fields))))
				f := fields[fi]
				writeLine(b, 1, fmt.Sprintf("gu_%d.%s = %s;", ui, f.name, castLiteral(f.ctype, fmt.Sprintf("(x ^ 0x%08Xu)", r.next31()))))
				writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)gu_%d.%s;", ui, f.name))
			case opts.Arrays && len(env.arrays) > 0:
				emitArrayMutation(b, r, opts, env)
			default:
				emitPointerMutation(b, r, opts, env)
			}
		}
		if i+1 < funcs {
			writeLine(b, 1, fmt.Sprintf("x ^= func_%d();", i+1))
		}
		if len(env.globals) > 0 {
			start := 0
			if opts.Consts && len(env.globals) > 1 {
				start = 1
			}
			gix := start + int(r.upto(uint32(max(len(env.globals)-start, 1))))
			g := env.globals[gix]
			writeLine(b, 1, fmt.Sprintf("%s ^= %s;", g.name, castLiteral(g.ctype, fmt.Sprintf("(x + 0x%08Xu)", r.next31()))))
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
		for _, g := range env.globals {
			if useRuntime {
				writeLine(b, 1, fmt.Sprintf("transparent_crc((uint64_t)%s, \"%s\", print_hash_value);", g.name, g.name))
			}
			writeLine(b, 1, fmt.Sprintf("checksum ^= (uint32_t)%s + 0x9E3779B9u + (checksum << 6) + (checksum >> 2);", g.name))
		}
		for _, arr := range env.arrays {
			if useRuntime {
				writeLine(b, 1, fmt.Sprintf("for (int ai = 0; ai < %d; ++ai) transparent_crc((uint64_t)%s[ai], \"%s\", print_hash_value);", arr.len, arr.name, arr.name))
			}
			writeLine(b, 1, fmt.Sprintf("checksum ^= (uint32_t)%s[0] + 0x9E3779B9u;", arr.name))
		}
		for i, st := range info.structs {
			for _, f := range st.fields {
				if useRuntime {
					writeLine(b, 1, fmt.Sprintf("transparent_crc((uint64_t)gs_%d.%s, \"gs\", print_hash_value);", i, f.name))
				}
				writeLine(b, 1, fmt.Sprintf("checksum ^= (uint32_t)gs_%d.%s + 0x9E3779B9u;", i, f.name))
			}
		}
		for i, ut := range info.unions {
			if len(ut.fields) > 0 {
				f := ut.fields[0]
				if useRuntime {
					writeLine(b, 1, fmt.Sprintf("transparent_crc((uint64_t)gu_%d.%s, \"gu\", print_hash_value);", i, f.name))
				}
				writeLine(b, 1, fmt.Sprintf("checksum ^= (uint32_t)gu_%d.%s + 0x9E3779B9u;", i, f.name))
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
	pool := typePool(opts)
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

	info := emitCompositeTypes(&b, r, opts, pool)
	env := emitGlobals(&b, r, opts, info, pool)
	emitFuncDecls(&b, funcs)
	emitFuncDefs(&b, r, opts, funcs, opts.MaxBlockSize, env, info)

	if !opts.NoMain {
		emitMain(&b, opts, env, info)
	}

	return b.String(), nil
}
