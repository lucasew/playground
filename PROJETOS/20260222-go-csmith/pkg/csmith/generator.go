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

type functionFlowState struct {
	funcs    []funcInfo
	maxFuncs int
	nextIdx  int
	pool     []CType
	opts     Options
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

type exprVarCandidate struct {
	expr       string
	ctype      CType
	assignable bool
}

type genContext struct {
	mustUse *exprVarCandidate
}

type stmtDecision struct {
	vals [12]uint32
}

type exprRand struct {
	vals     []uint32
	idx      int
	fallback *rng
}

type funcDecision struct {
	vals [16]uint32
}

func nextFuncDecision(r *rng) funcDecision {
	d := funcDecision{}
	for i := 0; i < len(d.vals); i++ {
		d.vals[i] = r.next31()
	}
	return d
}

func (d funcDecision) pick(i int, n uint32) uint32 {
	if n == 0 || i < 0 || i >= len(d.vals) {
		return 0
	}
	return d.vals[i] % n
}

func newExprRand(r *rng, budget int) *exprRand {
	if budget < 1 {
		budget = 1
	}
	vals := make([]uint32, budget)
	for i := 0; i < budget; i++ {
		vals[i] = r.next31()
	}
	return &exprRand{vals: vals, fallback: r}
}

func (e *exprRand) next() uint32 {
	if e.idx < len(e.vals) {
		v := e.vals[e.idx]
		e.idx++
		return v
	}
	return e.fallback.next31()
}

func (e *exprRand) pick(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	return e.next() % n
}

func nextStmtDecision(r *rng) stmtDecision {
	d := stmtDecision{}
	for i := 0; i < len(d.vals); i++ {
		d.vals[i] = r.next31()
	}
	return d
}

func (d stmtDecision) pick(i int, n uint32) uint32 {
	if n == 0 || i < 0 || i >= len(d.vals) {
		return 0
	}
	return d.vals[i] % n
}

type compositeInfo struct {
	structs []structTypeInfo
	unions  []unionTypeInfo
}

type envInfo struct {
	globals  []globalInfo
	arrays   []arrayInfo
	pointers []pointerInfo
	chains   []string
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

func randomConstantExprFromER(t CType, er *exprRand, opts Options) string {
	if t.Bits <= 8 {
		return castLiteral(t, fmt.Sprintf("0x%02X", er.next()&0xFF))
	}
	if t.Bits <= 16 {
		return castLiteral(t, fmt.Sprintf("0x%04X", er.next()&0xFFFF))
	}
	if t.Bits <= 32 {
		suffix := "U"
		if t.Signed {
			suffix = "L"
		}
		return castLiteral(t, fmt.Sprintf("0x%08X%s", er.next(), suffix))
	}
	if opts.LongLong {
		if t.Signed {
			return castLiteral(t, fmt.Sprintf("0x%08X%08XLL", er.next(), er.next()))
		}
		return castLiteral(t, fmt.Sprintf("0x%08X%08XULL", er.next(), er.next()))
	}
	return castLiteral(t, fmt.Sprintf("0x%08X%08X", er.next(), er.next()))
}

func sameBaseType(a, b CType) bool {
	return a.Bits == b.Bits && a.Signed == b.Signed
}

func buildExprCandidates(r *rng, env envInfo, scope scopeInfo) []exprVarCandidate {
	candidates := make([]exprVarCandidate, 0, len(env.globals)+len(scope.params)+len(scope.locals)+len(env.pointers)+len(env.arrays))
	for _, g := range env.globals {
		candidates = append(candidates, exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst})
	}
	for _, p := range scope.params {
		candidates = append(candidates, exprVarCandidate{expr: p.name, ctype: p.ctype, assignable: true})
	}
	for _, l := range scope.locals {
		if l.name == "x" {
			candidates = append(candidates, exprVarCandidate{expr: l.name, ctype: l.ctype, assignable: true})
			continue
		}
		candidates = append(candidates, exprVarCandidate{expr: l.name, ctype: l.ctype, assignable: true})
	}
	for _, p := range env.pointers {
		candidates = append(candidates, exprVarCandidate{expr: "*" + p.name, ctype: p.targetTy, assignable: !p.constTarget})
	}
	for _, arr := range env.arrays {
		candidates = append(candidates, exprVarCandidate{
			expr:       fmt.Sprintf("%s[%d]", arr.name, int(r.upto(uint32(arr.len)))),
			ctype:      arr.ctype,
			assignable: true,
		})
	}
	return candidates
}

func buildExprCandidatesFromER(er *exprRand, env envInfo, scope scopeInfo) []exprVarCandidate {
	candidates := make([]exprVarCandidate, 0, len(env.globals)+len(scope.params)+len(scope.locals)+len(env.pointers)+len(env.arrays))
	for _, g := range env.globals {
		candidates = append(candidates, exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst})
	}
	for _, p := range scope.params {
		candidates = append(candidates, exprVarCandidate{expr: p.name, ctype: p.ctype, assignable: true})
	}
	for _, l := range scope.locals {
		candidates = append(candidates, exprVarCandidate{expr: l.name, ctype: l.ctype, assignable: true})
	}
	for _, p := range env.pointers {
		candidates = append(candidates, exprVarCandidate{expr: "*" + p.name, ctype: p.targetTy, assignable: !p.constTarget})
	}
	for _, arr := range env.arrays {
		candidates = append(candidates, exprVarCandidate{
			expr:       fmt.Sprintf("%s[%d]", arr.name, int(er.pick(uint32(arr.len)))),
			ctype:      arr.ctype,
			assignable: true,
		})
	}
	return candidates
}

func selectExprVariable(t CType, r *rng, candidates []exprVarCandidate, forAssign bool) (exprVarCandidate, bool) {
	filtered := make([]exprVarCandidate, 0, len(candidates))
	for _, c := range candidates {
		if forAssign && !c.assignable {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return exprVarCandidate{}, false
	}

	exact := make([]exprVarCandidate, 0, len(filtered))
	sameWidth := make([]exprVarCandidate, 0, len(filtered))
	for _, c := range filtered {
		if sameBaseType(c.ctype, t) {
			exact = append(exact, c)
			continue
		}
		if c.ctype.Bits == t.Bits {
			sameWidth = append(sameWidth, c)
		}
	}
	if len(exact) > 0 {
		return exact[int(r.upto(uint32(len(exact))))], true
	}
	if len(sameWidth) > 0 {
		return sameWidth[int(r.upto(uint32(len(sameWidth))))], true
	}
	return filtered[int(r.upto(uint32(len(filtered))))], true
}

func selectExprVariableFromER(t CType, er *exprRand, candidates []exprVarCandidate, forAssign bool) (exprVarCandidate, bool) {
	filtered := make([]exprVarCandidate, 0, len(candidates))
	for _, c := range candidates {
		if forAssign && !c.assignable {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return exprVarCandidate{}, false
	}

	exact := make([]exprVarCandidate, 0, len(filtered))
	sameWidth := make([]exprVarCandidate, 0, len(filtered))
	for _, c := range filtered {
		if sameBaseType(c.ctype, t) {
			exact = append(exact, c)
			continue
		}
		if c.ctype.Bits == t.Bits {
			sameWidth = append(sameWidth, c)
		}
	}
	if len(exact) > 0 {
		return exact[int(er.pick(uint32(len(exact))))], true
	}
	if len(sameWidth) > 0 {
		return sameWidth[int(er.pick(uint32(len(sameWidth))))], true
	}
	return filtered[int(er.pick(uint32(len(filtered))))], true
}

func randomLeafExpr(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, ctx *genContext) string {
	if ctx != nil && ctx.mustUse != nil && er.pick(100) < 60 {
		return castLiteral(t, ctx.mustUse.expr)
	}
	candidates := buildExprCandidatesFromER(er, env, scope)
	if len(candidates) > 0 && er.pick(100) < 80 {
		c, ok := selectExprVariableFromER(t, er, candidates, false)
		if ok {
			return castLiteral(t, c.expr)
		}
	}

	return randomConstantExprFromER(t, er, opts)
}

func maxExprDepth(opts Options) int {
	if opts.MaxExprComplexity <= 2 {
		return 1
	}
	if opts.MaxExprComplexity <= 6 {
		return 2
	}
	if opts.MaxExprComplexity <= 12 {
		return 3
	}
	return 4
}

func randomTypedExprDepth(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext) string {
	limit := maxExprDepth(opts)
	// Bias toward leaves to avoid giant expressions while still creating nested trees.
	if depth >= limit || er.pick(100) < uint32(40+depth*15) {
		return randomLeafExpr(t, er, opts, env, scope, ctx)
	}

	lhs := randomTypedExprDepth(t, er, opts, env, scope, depth+1, ctx)
	rhs := randomTypedExprDepth(t, er, opts, env, scope, depth+1, ctx)

	ops := []string{"+", "-", "^", "|", "&"}
	if opts.Muls {
		ops = append(ops, "*")
	}
	if opts.Divs {
		ops = append(ops, "/")
	}
	if opts.CommaOperators {
		ops = append(ops, ",")
	}

	op := ops[int(er.pick(uint32(len(ops))))]
	switch op {
	case "/":
		den := randomTypedExprDepth(t, er, opts, env, scope, depth+1, ctx)
		return castLiteral(t, fmt.Sprintf("((%s) / ((%s) | 1))", lhs, den))
	case ",":
		return castLiteral(t, fmt.Sprintf("((%s), (%s))", lhs, rhs))
	default:
		return castLiteral(t, fmt.Sprintf("((%s) %s (%s))", lhs, op, rhs))
	}
}

func exprDecisionBudget(opts Options) int {
	depth := maxExprDepth(opts)
	return 16 + (depth * 12)
}

func randomTypedExpr(t CType, r *rng, opts Options, env envInfo, scope scopeInfo, ctx *genContext) string {
	er := newExprRand(r, exprDecisionBudget(opts))
	return randomTypedExprDepth(t, er, opts, env, scope, 0, ctx)
}

func chooseLValue(r *rng, target CType, env envInfo, scope scopeInfo) (lvalueInfo, bool) {
	c := buildExprCandidates(r, env, scope)
	pick, ok := selectExprVariable(target, r, c, true)
	if !ok {
		return lvalueInfo{}, false
	}
	return lvalueInfo{expr: pick.expr, ctype: pick.ctype}, true
}

func emitLValueAssignment(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo, ctx *genContext) bool {
	targetType := CType{Name: "uint32_t", Signed: false, Bits: 32}
	if len(scope.locals) > 0 && r.upto(100) < 70 {
		targetType = scope.locals[int(r.upto(uint32(len(scope.locals))))].ctype
	} else if len(env.globals) > 0 {
		targetType = env.globals[int(r.upto(uint32(len(env.globals))))].ctype
	}

	lv, ok := chooseLValue(r, targetType, env, scope)
	if !ok {
		return false
	}
	rhs := randomTypedExpr(lv.ctype, r, opts, env, scope, ctx)
	if opts.CompoundAssignment && r.upto(2) == 0 {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", lv.expr, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s;", lv.expr, rhs))
	}
	writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", lv.expr))
	if ctx != nil {
		c := exprVarCandidate{expr: lv.expr, ctype: lv.ctype, assignable: true}
		ctx.mustUse = &c
	}
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
		globalCap := min(max(opts.MaxGlobals, 2), 64)
		globalCount := 2 + int(r.upto(uint32(globalCap)))
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
			ptrCount := min(max(len(env.globals)-start, 0), 6)
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

			// Extra pointer chains: closer to upstream global shape (T*, T**, T*** ...).
			chainCount := min(max(len(env.pointers), 1), 3)
			env.chains = make([]string, 0, chainCount)
			for i := 0; i < chainCount; i++ {
				chainBases := make([]pointerInfo, 0, len(env.pointers))
				for _, pb := range env.pointers {
					if pb.constTarget || pb.volatileTarget || pb.volatilePointer {
						continue
					}
					chainBases = append(chainBases, pb)
				}
				if len(chainBases) == 0 {
					break
				}
				base := chainBases[int(r.upto(uint32(len(chainBases))))]
				depth := 2 + int(r.upto(uint32(max(1, min(opts.MaxPointerDepth, 4)-1))))

				prevName := base.name
				baseType := base.targetTy.Name
				for d := 2; d <= depth; d++ {
					name := fmt.Sprintf("gpp_%d_%d", i, d)
					stars := strings.Repeat("*", d)
					writeLine(b, 0, fmt.Sprintf("static %s %s%s = &%s;", baseType, stars, name, prevName))
					prevName = name
					env.chains = append(env.chains, name)
				}
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
	if opts.Muls && r.upto(4) == 0 {
		writeLine(b, 1, fmt.Sprintf("x = x * (0x%08Xu | 1u);", r.next31()))
		return
	}
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

func emitGlobalMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo, ctx *genContext) {
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
	rhs := randomTypedExpr(g.ctype, r, opts, env, scope, ctx)
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", g.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s;", g.name, safeAddExpr(g.ctype, g.name, rhs, opts)))
	}
	if ctx != nil {
		c := exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst}
		ctx.mustUse = &c
	}
}

func emitArrayMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo, ctx *genContext) {
	if !opts.Arrays || len(env.arrays) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	ai := env.arrays[int(r.upto(uint32(len(env.arrays))))]
	idxMask := max(1, min(opts.MaxArrayLenPerDim, 8)-1)
	rhs := randomTypedExpr(ai.ctype, r, opts, env, scope, ctx)
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
	if ctx != nil {
		c := exprVarCandidate{expr: fmt.Sprintf("%s[x & %du]", ai.name, idxMask), ctype: ai.ctype, assignable: true}
		ctx.mustUse = &c
	}
}

func emitPointerMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo, ctx *genContext) {
	if !opts.Pointers || len(env.pointers) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	pi := env.pointers[int(r.upto(uint32(len(env.pointers))))]
	rhs := randomTypedExpr(pi.targetTy, r, opts, env, scope, ctx)
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("*%s ^= %s;", pi.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("*%s = *%s ^ %s;", pi.name, pi.name, rhs))
	}
	if ctx != nil {
		c := exprVarCandidate{expr: "*" + pi.name, ctype: pi.targetTy, assignable: !pi.constTarget}
		ctx.mustUse = &c
	}
}

func emitLocalMutation(b *strings.Builder, r *rng, opts Options, env envInfo, scope scopeInfo, ctx *genContext) {
	if len(scope.locals) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	li := scope.locals[int(r.upto(uint32(len(scope.locals))))]
	rhs := randomTypedExpr(li.ctype, r, opts, env, scope, ctx)
	if opts.CompoundAssignment {
		writeLine(b, 1, fmt.Sprintf("%s += %s;", li.name, rhs))
	} else {
		writeLine(b, 1, fmt.Sprintf("%s = %s;", li.name, safeAddExpr(li.ctype, li.name, rhs, opts)))
	}
	writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", li.name))
	if ctx != nil {
		c := exprVarCandidate{expr: li.name, ctype: li.ctype, assignable: true}
		ctx.mustUse = &c
	}
}

func (s *functionFlowState) appendNewFunction(r *rng) (funcInfo, bool) {
	if len(s.funcs) >= s.maxFuncs {
		return funcInfo{}, false
	}
	fn := makeFuncSignature(r, s.opts, s.pool, s.nextIdx)
	s.nextIdx++
	s.funcs = append(s.funcs, fn)
	return fn, true
}

func emitFunctionCallMutation(
	b *strings.Builder,
	r *rng,
	opts Options,
	env envInfo,
	scope scopeInfo,
	state *functionFlowState,
	from int,
	ctx *genContext,
) bool {
	if state == nil {
		return false
	}

	candidates := make([]int, 0, len(state.funcs))
	for i := 0; i < len(state.funcs); i++ {
		// Keep acyclic call graph in generated order to avoid runaway recursion.
		if i <= from {
			continue
		}
		candidates = append(candidates, i)
	}

	// Upstream-like function invocation strategy:
	// 1) with a coin flip, try existing function first;
	// 2) if none available (or the coin says no), create one if limit allows.
	useExisting := len(candidates) > 0 && r.upto(2) == 0
	var callee funcInfo
	if useExisting {
		calleeIdx := candidates[int(r.upto(uint32(len(candidates))))]
		callee = state.funcs[calleeIdx]
	} else {
		created, ok := state.appendNewFunction(r)
		if ok {
			callee = created
		} else if len(candidates) > 0 {
			calleeIdx := candidates[int(r.upto(uint32(len(candidates))))]
			callee = state.funcs[calleeIdx]
		} else {
			return false
		}
	}
	args := "void"
	if len(callee.params) > 0 {
		argExprs := make([]string, 0, len(callee.params))
		for _, p := range callee.params {
			argExprs = append(argExprs, randomTypedExpr(p.ctype, r, opts, env, scope, ctx))
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

func emitStatement(
	b *strings.Builder,
	r *rng,
	opts Options,
	env envInfo,
	scope scopeInfo,
	state *functionFlowState,
	info compositeInfo,
	from int,
	maxChoice uint32,
	stmtBudget *int,
	ctx *genContext,
	dec stmtDecision,
) {
	if stmtBudget != nil && *stmtBudget == 0 {
		return
	}
	if stmtBudget != nil && *stmtBudget > 0 {
		*stmtBudget = *stmtBudget - 1
	}
	choice := dec.pick(0, maxChoice)
	switch {
	case choice == 0:
		if !emitLValueAssignment(b, r, opts, env, scope, ctx) {
			emitGlobalMutation(b, r, opts, env, scope, ctx)
		}
	case choice == 1:
		emitArithmeticMutation(b, r, opts)
	case choice == 2:
		if !emitIncDecMutation(b, r, opts) {
			writeLine(
				b,
				1,
				fmt.Sprintf(
					"if ((x & 1u) == 0u) { x += 0x%08Xu; } else { x ^= 0x%08Xu; }",
					dec.vals[1],
					dec.vals[2],
				),
			)
		}
	case choice == 3:
		if opts.Jumps {
			bound := 1 + int(dec.pick(3, 5))
			writeLine(b, 1, fmt.Sprintf("for (uint32_t i = 0; i < %du; ++i) { x += (i ^ 0x%08Xu); }", bound, dec.vals[4]))
		} else {
			writeLine(b, 1, fmt.Sprintf("x += 0x%08Xu;", dec.vals[5]))
		}
	case choice == 4:
		emitLocalMutation(b, r, opts, env, scope, ctx)
	case choice == 5 && len(info.structs) > 0:
		si := int(r.upto(uint32(len(info.structs))))
		fields := info.structs[si].fields
		fi := int(r.upto(uint32(len(fields))))
		f := fields[fi]
		writeLine(b, 1, fmt.Sprintf("gs_%d.%s ^= %s;", si, f.name, randomTypedExpr(f.ctype, r, opts, env, scope, ctx)))
		if ctx != nil {
			c := exprVarCandidate{expr: fmt.Sprintf("gs_%d.%s", si, f.name), ctype: f.ctype, assignable: true}
			ctx.mustUse = &c
		}
	case len(info.unions) > 0 && (choice == 6 || (len(info.structs) == 0 && choice == 5)):
		ui := int(r.upto(uint32(len(info.unions))))
		fields := info.unions[ui].fields
		fi := int(r.upto(uint32(len(fields))))
		f := fields[fi]
		writeLine(b, 1, fmt.Sprintf("gu_%d.%s = %s;", ui, f.name, randomTypedExpr(f.ctype, r, opts, env, scope, ctx)))
		writeLine(b, 1, fmt.Sprintf("x ^= (uint32_t)gu_%d.%s;", ui, f.name))
		if ctx != nil {
			c := exprVarCandidate{expr: fmt.Sprintf("gu_%d.%s", ui, f.name), ctype: f.ctype, assignable: true}
			ctx.mustUse = &c
		}
	case state != nil && len(state.funcs) > 1:
		if !emitFunctionCallMutation(b, r, opts, env, scope, state, from, ctx) {
			emitArithmeticMutation(b, r, opts)
		}
	case opts.Arrays && len(env.arrays) > 0:
		emitArrayMutation(b, r, opts, env, scope, ctx)
	default:
		emitPointerMutation(b, r, opts, env, scope, ctx)
	}
}

func emitStatements(
	b *strings.Builder,
	r *rng,
	opts Options,
	env envInfo,
	scope scopeInfo,
	state *functionFlowState,
	info compositeInfo,
	from int,
	maxChoice uint32,
	depth int,
	stmtBudget *int,
	ctx *genContext,
) {
	if stmtBudget != nil && *stmtBudget == 0 {
		return
	}
	stmtCount := 2 + int(r.upto(uint32(max(1, opts.MaxBlockSize))))
	for s := 0; s < stmtCount; s++ {
		dec := nextStmtDecision(r)
		if stmtBudget != nil && *stmtBudget == 0 {
			break
		}
		// Introduce nested blocks to better match Csmith statement shape.
		if depth+1 < max(1, opts.MaxBlockDepth) && opts.Jumps && dec.pick(6, 100) < 25 {
			mask := 1 + dec.pick(7, 7)
			cond := fmt.Sprintf("((x & %du) != 0u)", mask)
			if opts.ConstAsCondition && dec.pick(8, 5) == 0 {
				cond = fmt.Sprintf("(%s != 0u)", randomConstantExpr(CType{Name: "uint32_t", Signed: false, Bits: 32}, r, opts))
			} else if dec.pick(9, 3) == 0 {
				cond = fmt.Sprintf("((uint32_t)%s != 0u)", randomTypedExpr(CType{Name: "uint32_t", Signed: false, Bits: 32}, r, opts, env, scope, ctx))
			}
			writeLine(b, 1, fmt.Sprintf("if %s {", cond))
			emitStatements(b, r, opts, env, scope, state, info, from, maxChoice, depth+1, stmtBudget, ctx)
			writeLine(b, 1, "} else {")
			emitStatement(b, r, opts, env, scope, state, info, from, maxChoice, stmtBudget, ctx, dec)
			writeLine(b, 1, "}")
			continue
		}
		emitStatement(b, r, opts, env, scope, state, info, from, maxChoice, stmtBudget, ctx, dec)
	}
}

func statementChoiceCount(opts Options, env envInfo, info compositeInfo, funcs []funcInfo) uint32 {
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
	return uint32(baseCases + extraCases)
}

func emitSingleFuncDef(
	r *rng,
	opts Options,
	fn funcInfo,
	state *functionFlowState,
	idx int,
	maxBlock int,
	env envInfo,
	info compositeInfo,
	stmtBudget *int,
) string {
	fdec := nextFuncDecision(r)
	maxChoice := statementChoiceCount(opts, env, info, state.funcs)
	var b strings.Builder

	params := "void"
	if len(fn.params) > 0 {
		pp := make([]string, 0, len(fn.params))
		for _, p := range fn.params {
			pp = append(pp, fmt.Sprintf("%s %s", p.ctype.Name, p.name))
		}
		params = strings.Join(pp, ", ")
	}
	writeLine(&b, 0, fmt.Sprintf("static %s %s(%s) {", fn.ret.Name, fn.name, params))
	writeLine(&b, 1, fmt.Sprintf("%s l_0 = %s;", fn.ret.Name, castLiteral(fn.ret, fmt.Sprintf("0x%08Xu", r.next31()))))
	if len(env.globals) >= 2 {
		writeLine(
			&b,
			1,
			fmt.Sprintf(
				"uint32_t x = ((uint32_t)%s ^ 0x%08Xu) + ((uint32_t)%s + 0x%08Xu);",
				env.globals[0].name, r.next31(), env.globals[1].name, r.next31(),
			),
		)
	} else {
		writeLine(&b, 1, fmt.Sprintf("uint32_t x = 0x%08Xu;", r.next31()))
	}

	localCount := 1 + int(fdec.pick(0, uint32(max(1, min(maxBlock, 3)))))
	locals := make([]localInfo, 0, localCount+1)
	tpool := typePool(opts)
	for l := 0; l < localCount; l++ {
		lt := pickType(r, tpool)
		name := fmt.Sprintf("l_%d", l+1)
		writeLine(&b, 1, fmt.Sprintf("%s %s = %s;", lt.Name, name, randomConstantExpr(lt, r, opts)))
		locals = append(locals, localInfo{name: name, ctype: lt})
	}
	locals = append(locals, localInfo{name: "x", ctype: CType{Name: "uint32_t", Signed: false, Bits: 32}})
	scope := scopeInfo{params: fn.params, locals: locals}
	ctx := &genContext{}

	for _, p := range fn.params {
		writeLine(&b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", p.name))
	}
	emitStatements(&b, r, opts, env, scope, state, info, idx, maxChoice, 0, stmtBudget, ctx)
	_ = emitFunctionCallMutation(&b, r, opts, env, scope, state, idx, ctx)
	if len(env.globals) > 0 {
		start := 0
		if opts.Consts && len(env.globals) > 1 {
			start = 1
		}
		gix := start + int(fdec.pick(2, uint32(max(len(env.globals)-start, 1))))
		g := env.globals[gix]
		writeLine(&b, 1, fmt.Sprintf("%s ^= %s;", g.name, randomTypedExpr(g.ctype, r, opts, env, scope, ctx)))
	}
	writeLine(&b, 1, fmt.Sprintf("l_0 ^= %s;", castLiteral(fn.ret, "x")))
	writeLine(&b, 1, "return l_0;")
	writeLine(&b, 0, "}")
	writeLine(&b, 0, "")
	return b.String()
}

func makeFuncSignature(r *rng, opts Options, pool []CType, idx int) funcInfo {
	fn := funcInfo{
		name: fmt.Sprintf("func_%d", idx),
		ret:  pickType(r, pool),
	}
	if idx == 1 {
		// Upstream's entry function is func_1(void), returning a fixed integral type.
		fn.ret = CType{Name: "uint32_t", Signed: false, Bits: 32}
	}
	maxParams := min(opts.MaxParams, 4)
	if idx == 1 {
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
	return fn
}

func emitFunctionsUpstreamFlow(b *strings.Builder, r *rng, opts Options, pool []CType, maxBlock int, env envInfo, info compositeInfo) []funcInfo {
	maxFuncs := min(max(opts.MaxFuncs, 1), 8)
	state := &functionFlowState{
		funcs:    []funcInfo{makeFuncSignature(r, opts, pool, 1)},
		maxFuncs: maxFuncs,
		nextIdx:  2,
		pool:     pool,
		opts:     opts,
	}
	built := []bool{false}
	defs := []string{""}
	stmtBudget := opts.StopByStmt
	if stmtBudget < 0 {
		stmtBudget = -1
	}

	for cur := 0; cur < len(state.funcs); cur++ {
		if built[cur] {
			continue
		}
		defs[cur] = emitSingleFuncDef(r, opts, state.funcs[cur], state, cur, maxBlock, env, info, &stmtBudget)
		built[cur] = true
		for len(built) < len(state.funcs) {
			built = append(built, false)
			defs = append(defs, "")
		}
	}

	emitFuncDecls(b, state.funcs)
	for i := 0; i < len(defs); i++ {
		b.WriteString(defs[i])
	}
	return state.funcs
}

func emitMain(b *strings.Builder, opts Options, env envInfo, info compositeInfo, entry string) {
	useRuntime := opts.SafeMath || opts.ComputeHash
	useHashPrintf := opts.HashValuePrintf
	if opts.AcceptArgc {
		writeLine(b, 0, "int main(int argc, char *argv[]) {")
		writeLine(b, 1, "int print_hash_value = 0;")
		if useRuntime && useHashPrintf {
			writeLine(b, 1, "if (argc == 2 && strcmp(argv[1], \"1\") == 0) print_hash_value = 1;")
		}
	} else {
		writeLine(b, 0, "int main(void) {")
		if useRuntime {
			writeLine(b, 1, "int print_hash_value = 0;")
		}
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

	var b strings.Builder
	// Phase 1 (upstream-like): output header/comments/includes.
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

	// Phase 2 (upstream-like GenerateAllTypes): synthesize types.
	info := emitCompositeTypes(&b, r, opts, pool)

	// Phase 3 (upstream-like GenerateFunctions pre-state): globals/state.
	env := emitGlobals(&b, r, opts, info, pool)

	// Phase 4 (upstream-like GenerateFunctions): first function + FuncList walk.
	funcs := emitFunctionsUpstreamFlow(&b, r, opts, pool, opts.MaxBlockSize, env, info)

	// Phase 5 (upstream Output): main / checksums.
	if !opts.NoMain {
		emitMain(&b, opts, env, info, funcs[0].name)
	}

	return b.String(), nil
}
