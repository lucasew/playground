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
	name       string
	ctype      CType
	isConst    bool
	isVolatile bool
}

type arrayInfo struct {
	name  string
	ctype CType
	len   int
}

type pointerInfo struct {
	name            string
	target          string
	targetTy        CType
	volatilePointer bool
	volatileTarget  bool
	constTarget     bool
}

type paramInfo struct {
	name  string
	ctype CType
}

type funcInfo struct {
	name   string
	ret    CType
	params []paramInfo
}

type localInfo struct {
	name  string
	ctype CType
}

type scopeInfo struct {
	params []paramInfo
	locals []localInfo
}

type lvalueInfo struct {
	expr  string
	ctype CType
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

func safeAddExpr(t CType, a, b string, opts Options) string {
	if !opts.SafeMath {
		return fmt.Sprintf("((%s) + (%s))", a, b)
	}
	sign := "u_u"
	if t.Signed {
		sign = "s_s"
	}
	bits := t.Bits
	if bits != 8 && bits != 16 && bits != 32 && bits != 64 {
		bits = 32
	}
	prefix := "uint"
	if t.Signed {
		prefix = "int"
	}
	return fmt.Sprintf("safe_add_func_%s%d_t_%s(%s, %s)", prefix, bits, sign, a, b)
}

func safeDivU32Expr(a, b string, opts Options) string {
	if !opts.SafeMath {
		return fmt.Sprintf("((%s) / (%s))", a, b)
	}
	return fmt.Sprintf("safe_div_func_uint32_t_u_u(%s, %s)", a, b)
}

func safeLShiftU32Expr(a, b string, opts Options) string {
	if !opts.SafeMath {
		return fmt.Sprintf("((%s) << (%s))", a, b)
	}
	return fmt.Sprintf("safe_lshift_func_uint32_t_u_s(%s, %s)", a, b)
}

func safeRShiftU32Expr(a, b string, opts Options) string {
	if !opts.SafeMath {
		return fmt.Sprintf("((%s) >> (%s))", a, b)
	}
	return fmt.Sprintf("safe_rshift_func_uint32_t_u_s(%s, %s)", a, b)
}

func randomRawLiteral(t CType, r *rng) string {
	switch {
	case t.Bits <= 8:
		return fmt.Sprintf("0x%02Xu", r.next31()&0xFF)
	case t.Bits <= 16:
		return fmt.Sprintf("0x%04Xu", r.next31()&0xFFFF)
	case t.Bits <= 32:
		return fmt.Sprintf("0x%08Xu", r.next31())
	default:
		return fmt.Sprintf("0x%08X%08XULL", r.next31(), r.next31())
	}
}

func randomConstantExpr(t CType, r *rng, opts Options) string {
	if t.Bits <= 8 {
		return castLiteral(t, fmt.Sprintf("0x%02X", r.next31()&0xFF))
	}
	if t.Bits <= 16 {
		return castLiteral(t, fmt.Sprintf("0x%04X", r.next31()&0xFFFF))
	}
	if t.Bits <= 32 {
		suffix := "U"
		if t.Signed {
			suffix = "L"
		}
		return castLiteral(t, fmt.Sprintf("0x%08X%s", r.next31(), suffix))
	}
	if opts.LongLong {
		if t.Signed {
			return castLiteral(t, fmt.Sprintf("0x%08X%08XLL", r.next31(), r.next31()))
		}
		return castLiteral(t, fmt.Sprintf("0x%08X%08XULL", r.next31(), r.next31()))
	}
	return castLiteral(t, fmt.Sprintf("0x%08X%08X", r.next31(), r.next31()))
}

func randomTypedExpr(t CType, r *rng, opts Options, env envInfo, scope scopeInfo) string {
	candidates := make([]string, 0, len(env.globals)+len(scope.params)+len(scope.locals))

	for _, g := range env.globals {
		candidates = append(candidates, castLiteral(t, g.name))
	}
	for _, p := range scope.params {
		candidates = append(candidates, castLiteral(t, p.name))
	}
	for _, l := range scope.locals {
		candidates = append(candidates, castLiteral(t, l.name))
	}

	for _, p := range env.pointers {
		candidates = append(candidates, castLiteral(t, "*"+p.name))
	}
	for _, arr := range env.arrays {
		candidates = append(candidates, castLiteral(t, fmt.Sprintf("%s[%d]", arr.name, int(r.upto(uint32(arr.len))))))
	}

	if len(candidates) > 0 && r.upto(100) < 70 {
		return candidates[int(r.upto(uint32(len(candidates))))]
	}
	return randomConstantExpr(t, r, opts)
}

func chooseLValue(r *rng, opts Options, env envInfo, scope scopeInfo) (lvalueInfo, bool) {
	candidates := make([]lvalueInfo, 0, len(scope.locals)+len(env.globals)+len(env.arrays)+len(env.pointers))
	for _, l := range scope.locals {
		if l.name == "x" {
			continue
		}
		candidates = append(candidates, lvalueInfo{expr: l.name, ctype: l.ctype})
	}
	for _, g := range env.globals {
		if g.isConst {
			continue
		}
		candidates = append(candidates, lvalueInfo{expr: g.name, ctype: g.ctype})
	}
	for _, a := range env.arrays {
		if a.len <= 0 {
			continue
		}
		idx := int(r.upto(uint32(a.len)))
		candidates = append(candidates, lvalueInfo{expr: fmt.Sprintf("%s[%d]", a.name, idx), ctype: a.ctype})
	}
	for _, p := range env.pointers {
		if p.constTarget {
			continue
		}
		candidates = append(candidates, lvalueInfo{expr: "*" + p.name, ctype: p.targetTy})
	}
	if len(candidates) == 0 {
		return lvalueInfo{}, false
	}
	return candidates[int(r.upto(uint32(len(candidates))))], true
}

func emitLValueAssignment(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo) bool {
	lv, ok := chooseLValue(r, opts, env, scope)
	if !ok {
		return false
	}
	rhs := randomTypedExpr(lv.ctype, r, opts, env, scope)
	if opts.CompoundAssignment && r.upto(2) == 0 {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", lv.expr, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s;", lv.expr, rhs))
	}
	writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", lv.expr))
	return true
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
				name:       fmt.Sprintf("g_%d", i),
				ctype:      pickType(r, pool),
				isConst:    opts.Consts && i == 0,
				isVolatile: opts.Volatiles && r.upto(4) == 0,
			}
			lit := castLiteral(g.ctype, fmt.Sprintf("0x%08Xu", r.next31()))
			qual := ""
			if g.isConst {
				qual += "const "
			}
			if g.isVolatile {
				qual += "volatile "
			}
			if g.isConst {
				writeLine(b, 0, fmt.Sprintf("static %s%s %s = %s;", qual, g.ctype.Name, g.name, lit))
			} else {
				writeLine(b, 0, fmt.Sprintf("static %s%s %s = %s;", qual, g.ctype.Name, g.name, lit))
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
					name:            fmt.Sprintf("gp_%d", i),
					target:          target.name,
					targetTy:        target.ctype,
					volatilePointer: opts.VolatilePointers && r.upto(3) == 0,
					volatileTarget:  opts.VolatilePointers && r.upto(4) == 0,
					constTarget:     opts.ConstPointers && r.upto(3) == 0,
				}
				targetQual := ""
				if p.constTarget {
					targetQual += "const "
				}
				if p.volatileTarget {
					targetQual += "volatile "
				}
				ptrQual := ""
				if p.volatilePointer {
					ptrQual = "volatile "
				}
				writeLine(b, 0, fmt.Sprintf("static %s%s *%s%s = &%s;", targetQual, target.ctype.Name, ptrQual, p.name, p.target))
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

func emitFuncDecls(b *strings.Builder, funcs []funcInfo) {
	for _, fn := range funcs {
		params := "void"
		if len(fn.params) > 0 {
			pp := make([]string, 0, len(fn.params))
			for _, p := range fn.params {
				pp = append(pp, fmt.Sprintf("%s %s", p.ctype.Name, p.name))
			}
			params = strings.Join(pp, ", ")
		}
		writeLine(b, 0, fmt.Sprintf("static %s %s(%s);", fn.ret.Name, fn.name, params))
	}
	writeLine(b, 0, "")
}

func emitArithmeticMutation(b *strings.Builder, r *rng, opts Options) {
	if opts.Divs && r.upto(3) == 0 {
		writeLine(b, 1, fmt.Sprintf("x = %s;", safeDivU32Expr("x", "((x & 255u) + 1u)", opts)))
		return
	}
	if opts.UnaryPlusOperator && r.upto(3) == 0 {
		writeLine(b, 1, fmt.Sprintf("x = (+x) ^ 0x%08Xu;", r.next31()))
		return
	}
	writeLine(
		b,
		1,
		fmt.Sprintf("x = (%s) ^ (%s) ^ 0x%08Xu;", safeLShiftU32Expr("x", "1", opts), safeRShiftU32Expr("x", "1", opts), r.next31()),
	)
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

func emitGlobalMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo) {
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
	rhs := randomTypedExpr(g.ctype, r, opts, env, scope)
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", g.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s;", g.name, safeAddExpr(g.ctype, g.name, rhs, opts)))
	}
}

func emitArrayMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo) {
	if !opts.Arrays || len(env.arrays) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	ai := env.arrays[int(r.upto(uint32(len(env.arrays))))]
	idxMask := max(1, min(opts.MaxArrayLenPerDim, 8)-1)
	rhs := randomTypedExpr(ai.ctype, r, opts, env, scope)
	writeLine(b, 1, fmt.Sprintf("%s[x & %du] ^= %s;", ai.name, idxMask, rhs))
	if opts.EmbeddedAssigns {
		one := castLiteral(ai.ctype, "1u")
		writeLine(
			b,
			1,
			fmt.Sprintf(
				"x = (%s[x & %du] = %s);",
				ai.name,
				idxMask,
				safeAddExpr(ai.ctype, fmt.Sprintf("%s[x & %du]", ai.name, idxMask), one, opts),
			),
		)
	}
}

func emitPointerMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo) {
	if !opts.Pointers || len(env.pointers) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	pi := env.pointers[int(r.upto(uint32(len(env.pointers))))]
	rhs := randomTypedExpr(pi.targetTy, r, opts, env, scope)
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("*%s ^= %s;", pi.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("*%s = *%s ^ %s;", pi.name, pi.name, rhs))
	}
}

func emitLocalMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo) {
	if len(scope.locals) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	li := scope.locals[int(r.upto(uint32(len(scope.locals))))]
	rhs := randomTypedExpr(li.ctype, r, opts, env, scope)
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", li.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s;", li.name, safeAddExpr(li.ctype, li.name, rhs, opts)))
	}
	writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", li.name))
}

func emitFunctionCallMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo, funcs []funcInfo, from int) bool {
	if from+1 >= len(funcs) {
		return false
	}
	calleeIdx := from + 1 + int(r.upto(uint32(len(funcs)-from-1)))
	callee := funcs[calleeIdx]
	args := "void"
	if len(callee.params) > 0 {
		argExprs := make([]string, 0, len(callee.params))
		for _, p := range callee.params {
			argExprs = append(argExprs, randomTypedExpr(p.ctype, r, opts, env, scope))
		}
		args = strings.Join(argExprs, ", ")
	}
	if args == "void" {
		writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s();", callee.name))
	} else {
		writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s(%s);", callee.name, args))
	}
	return true
}

func emitFuncDefs(b *strings.Builder, r *rng, opts Options, funcs []funcInfo, maxBlock int, env envInfo, info compositeInfo) {
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
	if len(funcs) > 1 {
		extraCases++
	}
	baseCases := 6
	maxChoice := uint32(baseCases + extraCases)

	for i := len(funcs) - 1; i >= 0; i-- {
		fn := funcs[i]
		params := "void"
		if len(fn.params) > 0 {
			pp := make([]string, 0, len(fn.params))
			for _, p := range fn.params {
				pp = append(pp, fmt.Sprintf("%s %s", p.ctype.Name, p.name))
			}
			params = strings.Join(pp, ", ")
		}
		writeLine(b, 0, fmt.Sprintf("static %s %s(%s) {", fn.ret.Name, fn.name, params))
		writeLine(b, 1, fmt.Sprintf("%s l_0 = %s;", fn.ret.Name, castLiteral(fn.ret, fmt.Sprintf("0x%08Xu", r.next31()))))
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

		localCount := 1 + int(r.upto(uint32(max(1, min(maxBlock, 3)))))
		locals := make([]localInfo, 0, localCount+1)
		tpool := typePool(opts)
		for l := 0; l < localCount; l++ {
			lt := pickType(r, tpool)
			name := fmt.Sprintf("l_%d", l+1)
			writeLine(b, 1, fmt.Sprintf("%s %s = %s;", lt.Name, name, randomConstantExpr(lt, r, opts)))
			locals = append(locals, localInfo{name: name, ctype: lt})
		}
		locals = append(locals, localInfo{name: "x", ctype: CType{Name: "uint32_t", Signed: false, Bits: 32}})
		scope := scopeInfo{params: fn.params, locals: locals}

		for _, p := range fn.params {
			writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", p.name))
		}
		stmtCount := 2 + int(r.upto(uint32(maxBlock)))
		for s := 0; s < stmtCount; s++ {
			choice := r.upto(maxChoice)
			switch {
			case choice == 0:
				if !emitLValueAssignment(b, r, opts, env, scope) {
					emitGlobalMutation(b, r, opts, env, scope)
				}
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
			case choice == 4:
				emitLocalMutation(b, r, opts, env, scope)
			case choice == 5 && len(info.structs) > 0:
				si := int(r.upto(uint32(len(info.structs))))
				fields := info.structs[si].fields
				fi := int(r.upto(uint32(len(fields))))
				f := fields[fi]
				writeLine(b, 1, fmt.Sprintf("gs_%d.%s ^= %s;", si, f.name, randomTypedExpr(f.ctype, r, opts, env, scope)))
			case len(info.unions) > 0 && (choice == 6 || (len(info.structs) == 0 && choice == 5)):
				ui := int(r.upto(uint32(len(info.unions))))
				fields := info.unions[ui].fields
				fi := int(r.upto(uint32(len(fields))))
				f := fields[fi]
				writeLine(b, 1, fmt.Sprintf("gu_%d.%s = %s;", ui, f.name, randomTypedExpr(f.ctype, r, opts, env, scope)))
				writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)gu_%d.%s;", ui, f.name))
			case len(funcs) > 1:
				if !emitFunctionCallMutation(b, r, opts, env, scope, funcs, i) {
					emitArithmeticMutation(b, r, opts)
				}
			case opts.Arrays && len(env.arrays) > 0:
				emitArrayMutation(b, r, opts, env, scope)
			default:
				emitPointerMutation(b, r, opts, env, scope)
			}
		}
		if i+1 < len(funcs) {
			_ = emitFunctionCallMutation(b, r, opts, env, scope, funcs, i)
		}
		if len(env.globals) > 0 {
			start := 0
			if opts.Consts && len(env.globals) > 1 {
				start = 1
			}
			gix := start + int(r.upto(uint32(max(len(env.globals)-start, 1))))
			g := env.globals[gix]
			writeLine(b, 1, fmt.Sprintf("%s ^= %s;", g.name, randomTypedExpr(g.ctype, r, opts, env, scope)))
		}
		writeLine(b, 1, fmt.Sprintf("l_0 ^= %s;", castLiteral(fn.ret, "x")))
		writeLine(b, 1, "return l_0;")
		writeLine(b, 0, "}")
		writeLine(b, 0, "")
	}
}

func emitMain(b *strings.Builder, opts Options, env envInfo, info compositeInfo, entry string) {
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
	writeLine(b, 1, fmt.Sprintf("uint32_t x = (uint32_t)%s();", entry))
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
	funcCount := 1 + int(r.upto(uint32(min(opts.MaxFuncs, 8))))
	funcs := make([]funcInfo, 0, funcCount)
	for i := 0; i < funcCount; i++ {
		fn := funcInfo{
			name: fmt.Sprintf("func_%d", i+1),
			ret:  pickType(r, pool),
		}
		if i == 0 {
			fn.ret = CType{Name: "uint32_t", Signed: false, Bits: 32}
		}
		maxParams := min(opts.MaxParams, 4)
		if i == 0 {
			// Csmith keeps func_1(void) as the top-level entry.
			maxParams = 0
		}
		if maxParams < 0 {
			maxParams = 0
		}
		pcount := 0
		if maxParams > 0 {
			pcount = int(r.upto(uint32(maxParams + 1)))
		}
		fn.params = make([]paramInfo, 0, pcount)
		for p := 0; p < pcount; p++ {
			fn.params = append(fn.params, paramInfo{
				name:  fmt.Sprintf("p_%d", p+1),
				ctype: pickType(r, pool),
			})
		}
		funcs = append(funcs, fn)
	}

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
		emitMain(&b, opts, env, info, funcs[0].name)
	}

	return b.String(), nil
}
