package csmith

import (
	"fmt"
	"regexp"
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
	funcs        []funcInfo
	maxFuncs     int
	nextIdx      int
	pool         []CType
	opts         Options
	dynGlobals   []globalInfo
	lateGlobals  strings.Builder
	nextGlobalID int
}

type stmtKind int

const (
	stmtAssign stmtKind = iota
	stmtIfElse
	stmtFor
	stmtReturn
	stmtContinue
	stmtBreak
	stmtGoto
	stmtArrayOp
	stmtInvoke
)

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
	state   *functionFlowState
	from    int
	dynLocs []localInfo
}

type genSnapshot struct {
	dynLocLen      int
	funcsLen       int
	nextIdx        int
	dynGlobalsLen  int
	nextGlobalID   int
	lateGlobalsBuf string
}

func takeGenSnapshot(ctx *genContext) *genSnapshot {
	if ctx == nil {
		return nil
	}
	s := &genSnapshot{
		dynLocLen: len(ctx.dynLocs),
	}
	if ctx.state != nil {
		s.funcsLen = len(ctx.state.funcs)
		s.nextIdx = ctx.state.nextIdx
		s.dynGlobalsLen = len(ctx.state.dynGlobals)
		s.nextGlobalID = ctx.state.nextGlobalID
		s.lateGlobalsBuf = ctx.state.lateGlobals.String()
	}
	return s
}

func restoreGenSnapshot(ctx *genContext, s *genSnapshot) {
	if ctx == nil || s == nil {
		return
	}
	if len(ctx.dynLocs) >= s.dynLocLen {
		ctx.dynLocs = ctx.dynLocs[:s.dynLocLen]
	}
	if ctx.state != nil {
		if len(ctx.state.funcs) >= s.funcsLen {
			ctx.state.funcs = ctx.state.funcs[:s.funcsLen]
		}
		if len(ctx.state.dynGlobals) >= s.dynGlobalsLen {
			ctx.state.dynGlobals = ctx.state.dynGlobals[:s.dynGlobalsLen]
		}
		ctx.state.nextIdx = s.nextIdx
		ctx.state.nextGlobalID = s.nextGlobalID
		ctx.state.lateGlobals.Reset()
		ctx.state.lateGlobals.WriteString(s.lateGlobalsBuf)
	}
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

func mergedLocals(scope scopeInfo, ctx *genContext) []localInfo {
	if ctx == nil || len(ctx.dynLocs) == 0 {
		return scope.locals
	}
	out := make([]localInfo, 0, len(scope.locals)+len(ctx.dynLocs))
	out = append(out, scope.locals...)
	out = append(out, ctx.dynLocs...)
	return out
}

func mergedGlobals(env envInfo, ctx *genContext) []globalInfo {
	if ctx == nil || ctx.state == nil || len(ctx.state.dynGlobals) == 0 {
		return env.globals
	}
	out := make([]globalInfo, 0, len(env.globals)+len(ctx.state.dynGlobals))
	out = append(out, env.globals...)
	out = append(out, ctx.state.dynGlobals...)
	return out
}

func buildExprCandidates(r *rng, env envInfo, scope scopeInfo, ctx *genContext) []exprVarCandidate {
	candidates := make([]exprVarCandidate, 0, len(env.globals)+len(scope.params)+len(scope.locals)+len(env.pointers)+len(env.arrays))
	for _, g := range mergedGlobals(env, ctx) {
		candidates = append(candidates, exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst})
	}
	for _, p := range scope.params {
		candidates = append(candidates, exprVarCandidate{expr: p.name, ctype: p.ctype, assignable: true})
	}
	for _, l := range mergedLocals(scope, ctx) {
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

func buildExprCandidatesFromER(er *exprRand, env envInfo, scope scopeInfo, ctx *genContext) []exprVarCandidate {
	candidates := make([]exprVarCandidate, 0, len(env.globals)+len(scope.params)+len(scope.locals)+len(env.pointers)+len(env.arrays))
	for _, g := range mergedGlobals(env, ctx) {
		candidates = append(candidates, exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst})
	}
	for _, p := range scope.params {
		candidates = append(candidates, exprVarCandidate{expr: p.name, ctype: p.ctype, assignable: true})
	}
	for _, l := range mergedLocals(scope, ctx) {
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

func variableScopePickFromER(er *exprRand, opts Options) int {
	// Upstream VariableSelector::InitScopeTable probabilities.
	// 0=global, 1=parent local, 2=parent param, 3=new value.
	v := int(er.pick(100))
	if opts.GlobalVariables {
		switch {
		case v < 35:
			return 0
		case v < 65:
			return 1
		case v < 95:
			return 2
		default:
			return 3
		}
	}
	switch {
	case v < 50:
		return 1
	case v < 95:
		return 2
	default:
		return 3
	}
}

func buildScopedCandidatesFromER(er *exprRand, env envInfo, scope scopeInfo, scopePick int, ctx *genContext) []exprVarCandidate {
	out := make([]exprVarCandidate, 0, 16)
	switch scopePick {
	case 0:
		for _, g := range mergedGlobals(env, ctx) {
			out = append(out, exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst})
		}
	case 1:
		for _, l := range mergedLocals(scope, ctx) {
			out = append(out, exprVarCandidate{expr: l.name, ctype: l.ctype, assignable: true})
		}
	case 2:
		for _, p := range scope.params {
			out = append(out, exprVarCandidate{expr: p.name, ctype: p.ctype, assignable: true})
		}
	}
	if scopePick != 2 {
		for _, ptr := range env.pointers {
			out = append(out, exprVarCandidate{expr: "*" + ptr.name, ctype: ptr.targetTy, assignable: !ptr.constTarget})
		}
		for _, arr := range env.arrays {
			out = append(out, exprVarCandidate{
				expr:       fmt.Sprintf("%s[%d]", arr.name, int(er.pick(uint32(arr.len)))),
				ctype:      arr.ctype,
				assignable: true,
			})
		}
	}
	return out
}

func createOnDemandGlobalFromER(er *exprRand, opts Options, t CType, ctx *genContext) (exprVarCandidate, bool) {
	if ctx == nil || ctx.state == nil {
		return exprVarCandidate{}, false
	}
	id := ctx.state.nextGlobalID
	ctx.state.nextGlobalID = id + 1
	name := fmt.Sprintf("g_%d", id)
	isConst := opts.Consts && er.pick(100) < 20
	isVolatile := opts.Volatiles && er.pick(100) < 20
	qual := ""
	if isConst {
		qual += "const "
	}
	if isVolatile {
		qual += "volatile "
	}
	lit := randomConstantExprFromER(t, er, opts)
	writeLine(&ctx.state.lateGlobals, 0, fmt.Sprintf("static %s%s %s = %s;", qual, t.Name, name, lit))
	g := globalInfo{name: name, ctype: t, isConst: isConst, isVolatile: isVolatile}
	ctx.state.dynGlobals = append(ctx.state.dynGlobals, g)
	return exprVarCandidate{expr: name, ctype: t, assignable: !isConst}, true
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

func buildFunctionCallExpr(
	t CType,
	er *exprRand,
	opts Options,
	env envInfo,
	scope scopeInfo,
	depth int,
	ctx *genContext,
) (string, bool) {
	if ctx == nil || ctx.state == nil {
		return "", false
	}
	state := ctx.state
	from := ctx.from

	candidates := make([]int, 0, len(state.funcs))
	for i := 0; i < len(state.funcs); i++ {
		if i <= from {
			continue
		}
		candidates = append(candidates, i)
	}

	useExisting := len(candidates) > 0 && er.pick(2) == 0
	var callee funcInfo
	if useExisting {
		callee = state.funcs[candidates[int(er.pick(uint32(len(candidates))))]]
	} else {
		created, ok := state.appendNewFunction(er.fallback)
		if ok {
			callee = created
		} else if len(candidates) > 0 {
			callee = state.funcs[candidates[int(er.pick(uint32(len(candidates))))]]
		} else {
			return "", false
		}
	}

	args := "void"
	if len(callee.params) > 0 {
		argExprs := make([]string, 0, len(callee.params))
		for _, p := range callee.params {
			argExprs = append(argExprs, randomParamExprDepth(p.ctype, er, opts, env, scope, depth+1, ctx))
		}
		args = strings.Join(argExprs, ", ")
	}
	if args == "void" {
		return castLiteral(t, fmt.Sprintf("%s()", callee.name)), true
	}
	return castLiteral(t, fmt.Sprintf("%s(%s)", callee.name, args)), true
}

func randomLeafExprWithMode(
	t CType,
	er *exprRand,
	opts Options,
	env envInfo,
	scope scopeInfo,
	depth int,
	ctx *genContext,
	isParam bool,
	noFunc bool,
	noConst bool,
) string {
	type termChoice int
	const (
		termFunction termChoice = iota
		termVariable
		termConstant
		termAssign
		termComma
	)

	weighted := make([]termChoice, 0, 140)
	funcW := 70
	varW := 20
	constW := 10
	assignW := 0
	commaW := 0
	if isParam {
		// Upstream make_random_param baseline:
		// function 40, variable 40, constant 0 (+ optional assign/comma).
		funcW = 40
		varW = 40
		constW = 0
	}
	if noFunc || depth+2 > maxExprDepth(opts) {
		funcW = 0
	}
	if noConst {
		constW = 0
	}
	if depth+2 > maxExprDepth(opts) {
		assignW = 0
		commaW = 0
	}
	for i := 0; i < funcW; i++ {
		weighted = append(weighted, termFunction)
	}
	for i := 0; i < varW; i++ {
		weighted = append(weighted, termVariable)
	}
	for i := 0; i < constW; i++ {
		weighted = append(weighted, termConstant)
	}
	if opts.EmbeddedAssigns && assignW == 0 {
		assignW = 10
	}
	if assignW > 0 {
		for i := 0; i < assignW; i++ {
			weighted = append(weighted, termAssign)
		}
	}
	if opts.CommaOperators && commaW == 0 {
		commaW = 10
	}
	if commaW > 0 {
		for i := 0; i < commaW; i++ {
			weighted = append(weighted, termComma)
		}
	}
	if len(weighted) == 0 {
		return randomConstantExprFromER(t, er, opts)
	}

	for tries := 0; tries < 6; tries++ {
		snap := takeGenSnapshot(ctx)
		choice := weighted[int(er.pick(uint32(len(weighted))))]
		switch choice {
		case termFunction:
			if depth < maxExprDepth(opts) {
				if call, ok := buildFunctionCallExpr(t, er, opts, env, scope, depth, ctx); ok {
					return call
				}
			}
			restoreGenSnapshot(ctx, snap)
		case termVariable:
			scopePick := variableScopePickFromER(er, opts)
			if scopePick == 3 {
				if g, ok := createOnDemandGlobalFromER(er, opts, t, ctx); ok {
					return castLiteral(t, g.expr)
				}
				restoreGenSnapshot(ctx, snap)
				continue
			}
			candidates := buildScopedCandidatesFromER(er, env, scope, scopePick, ctx)
			if len(candidates) == 0 {
				if scopePick == 0 {
					if g, ok := createOnDemandGlobalFromER(er, opts, t, ctx); ok {
						return castLiteral(t, g.expr)
					}
				}
				candidates = buildExprCandidatesFromER(er, env, scope, ctx)
			}
			if len(candidates) > 0 {
				if c, ok := selectExprVariableFromER(t, er, candidates, false); ok {
					return castLiteral(t, c.expr)
				}
			}
			restoreGenSnapshot(ctx, snap)
		case termConstant:
			return randomConstantExprFromER(t, er, opts)
		case termAssign:
			scopePick := variableScopePickFromER(er, opts)
			candidates := buildScopedCandidatesFromER(er, env, scope, scopePick, ctx)
			if len(candidates) == 0 {
				if scopePick == 0 || scopePick == 3 {
					if g, ok := createOnDemandGlobalFromER(er, opts, t, ctx); ok {
						rhs := randomConstantExprFromER(g.ctype, er, opts)
						return castLiteral(t, fmt.Sprintf("(%s = %s)", g.expr, rhs))
					}
				}
				candidates = buildExprCandidatesFromER(er, env, scope, ctx)
			}
			if len(candidates) > 0 {
				if lv, ok := selectExprVariableFromER(t, er, candidates, true); ok {
					rhs := randomConstantExprFromER(lv.ctype, er, opts)
					return castLiteral(t, fmt.Sprintf("(%s = %s)", lv.expr, rhs))
				}
			}
			restoreGenSnapshot(ctx, snap)
		case termComma:
			lhs := randomConstantExprFromER(t, er, opts)
			rhs := randomConstantExprFromER(t, er, opts)
			return castLiteral(t, fmt.Sprintf("((%s), (%s))", lhs, rhs))
		}
	}

	return randomConstantExprFromER(t, er, opts)
}

func randomLeafExpr(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext) string {
	return randomLeafExprWithMode(t, er, opts, env, scope, depth, ctx, false, false, false)
}

func randomParamLeafExpr(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext) string {
	return randomLeafExprWithMode(t, er, opts, env, scope, depth, ctx, true, false, false)
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
		return randomLeafExpr(t, er, opts, env, scope, depth, ctx)
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

func randomTypedExprDepthFlags(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext, noFunc bool, noConst bool) string {
	limit := maxExprDepth(opts)
	if depth >= limit || er.pick(100) < uint32(40+depth*15) {
		return randomLeafExprWithMode(t, er, opts, env, scope, depth, ctx, false, noFunc, noConst)
	}
	lhs := randomTypedExprDepthFlags(t, er, opts, env, scope, depth+1, ctx, noFunc, noConst)
	rhs := randomTypedExprDepthFlags(t, er, opts, env, scope, depth+1, ctx, noFunc, noConst)
	ops := []string{"+", "-", "^", "|", "&"}
	if opts.Muls {
		ops = append(ops, "*")
	}
	if opts.Divs {
		ops = append(ops, "/")
	}
	if opts.CommaOperators && !noConst {
		ops = append(ops, ",")
	}
	op := ops[int(er.pick(uint32(len(ops))))]
	switch op {
	case "/":
		den := randomTypedExprDepthFlags(t, er, opts, env, scope, depth+1, ctx, noFunc, noConst)
		return castLiteral(t, fmt.Sprintf("((%s) / ((%s) | 1))", lhs, den))
	case ",":
		return castLiteral(t, fmt.Sprintf("((%s), (%s))", lhs, rhs))
	default:
		return castLiteral(t, fmt.Sprintf("((%s) %s (%s))", lhs, op, rhs))
	}
}

func randomParamExprDepth(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext) string {
	limit := maxExprDepth(opts)
	if depth >= limit || er.pick(100) < uint32(45+depth*15) {
		return randomParamLeafExpr(t, er, opts, env, scope, depth, ctx)
	}
	lhs := randomParamExprDepth(t, er, opts, env, scope, depth+1, ctx)
	rhs := randomParamExprDepth(t, er, opts, env, scope, depth+1, ctx)
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
		den := randomParamExprDepth(t, er, opts, env, scope, depth+1, ctx)
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

func variableScopePick(r *rng, opts Options) int {
	v := int(r.upto(100))
	if opts.GlobalVariables {
		switch {
		case v < 35:
			return 0
		case v < 65:
			return 1
		case v < 95:
			return 2
		default:
			return 3
		}
	}
	switch {
	case v < 50:
		return 1
	case v < 95:
		return 2
	default:
		return 3
	}
}

func buildScopedCandidates(r *rng, env envInfo, scope scopeInfo, scopePick int, ctx *genContext) []exprVarCandidate {
	out := make([]exprVarCandidate, 0, 16)
	switch scopePick {
	case 0:
		for _, g := range mergedGlobals(env, ctx) {
			out = append(out, exprVarCandidate{expr: g.name, ctype: g.ctype, assignable: !g.isConst})
		}
	case 1:
		for _, l := range mergedLocals(scope, ctx) {
			out = append(out, exprVarCandidate{expr: l.name, ctype: l.ctype, assignable: true})
		}
	case 2:
		for _, p := range scope.params {
			out = append(out, exprVarCandidate{expr: p.name, ctype: p.ctype, assignable: true})
		}
	}
	if scopePick != 2 {
		for _, ptr := range env.pointers {
			out = append(out, exprVarCandidate{expr: "*" + ptr.name, ctype: ptr.targetTy, assignable: !ptr.constTarget})
		}
		for _, arr := range env.arrays {
			out = append(out, exprVarCandidate{
				expr:       fmt.Sprintf("%s[%d]", arr.name, int(r.upto(uint32(arr.len)))),
				ctype:      arr.ctype,
				assignable: true,
			})
		}
	}
	return out
}

func chooseLValue(r *rng, opts Options, target CType, env envInfo, scope scopeInfo, ctx *genContext) (lvalueInfo, bool) {
	scopePick := variableScopePick(r, opts)
	c := buildScopedCandidates(r, env, scope, scopePick, ctx)
	if len(c) == 0 {
		c = buildExprCandidates(r, env, scope, ctx)
	}
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

	lv, ok := chooseLValue(r, opts, targetType, env, scope, ctx)
	if !ok {
		// Upstream-like fallback: create a new value when selection fails.
		name := fmt.Sprintf("lv_%d", r.next31()&0xFFFF)
		writeLine(b, 1, fmt.Sprintf("%s %s = %s;", targetType.Name, name, randomConstantExpr(targetType, r, opts)))
		lv = lvalueInfo{expr: name, ctype: targetType}
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

func maybeDeclareOnDemandLocal(b *strings.Builder, r *rng, opts Options, ctx *genContext) {
	if ctx == nil {
		return
	}
	if variableScopePick(r, opts) != 3 {
		return
	}
	t := pickType(r, typePool(opts))
	name := fmt.Sprintf("ld_%d", r.next31()&0xFFFF)
	writeLine(b, 1, fmt.Sprintf("%s %s = %s;", t.Name, name, randomConstantExpr(t, r, opts)))
	ctx.dynLocs = append(ctx.dynLocs, localInfo{name: name, ctype: t})
}

func emitCompositeTypes(b *strings.Builder, r *rng, opts Options, pool []CType) compositeInfo {
	info := compositeInfo{}
	totalTypes := len(pool)

	// Upstream-like MoreTypesProbability:
	// keep adding aggregate types while type universe is small, then 50% chance.
	moreTypes := func() bool {
		if totalTypes < 10 {
			return true
		}
		return r.upto(100) < 50
	}

	if opts.PackedStruct {
		writeLine(b, 0, "#if defined(__GNUC__) || defined(__clang__)")
		writeLine(b, 0, "#define CSMITH_GO_PACKED __attribute__((packed))")
		writeLine(b, 0, "#else")
		writeLine(b, 0, "#define CSMITH_GO_PACKED")
		writeLine(b, 0, "#endif")
		writeLine(b, 0, "")
	}

	if opts.Structs {
		sidx := 0
		for moreTypes() && sidx < min(max(opts.MaxStructFields, 1), 32) {
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
			totalTypes++
			sidx++
		}
	}

	if opts.Unions {
		uidx := 0
		for moreTypes() && uidx < min(max(opts.MaxUnionFields, 1), 32) {
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
			totalTypes++
			uidx++
		}
	}

	return info
}

func emitGlobals(b *strings.Builder, r *rng, opts Options, info compositeInfo, pool []CType) envInfo {
	env := envInfo{}
	if opts.GlobalVariables {
		globalCap := min(max(opts.MaxGlobals, 2), 64)
		env.globals = make([]globalInfo, 0, globalCap)
		moreGlobals := func() bool {
			// Upstream creates globals on-demand through VariableSelector.
			// Keep a larger initial universe, then taper probabilistically.
			if len(env.globals) < min(70, globalCap) {
				return true
			}
			if len(env.globals) < min(75, globalCap) {
				return len(env.globals) < globalCap && r.upto(100) < 80
			}
			return len(env.globals) < globalCap && r.upto(100) < 45
		}
		for i := 0; moreGlobals(); i++ {
			isConst := false
			if opts.Consts {
				isConst = r.upto(100) < 22
			}
			isVolatile := false
			if opts.Volatiles {
				isVolatile = r.upto(100) < 20
			}
			// Avoid degenerate const+volatile saturation.
			if isConst && isVolatile && r.upto(2) == 0 {
				isConst = false
			}
			g := globalInfo{
				name:       fmt.Sprintf("g_%d", i),
				ctype:      pickType(r, pool),
				isConst:    isConst,
				isVolatile: isVolatile,
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
			arrayCount := 1
			for arrayCount < min(max(opts.MaxArrayDim, 1), 4) && r.upto(100) < 50 {
				arrayCount++
			}
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
			ptrCount := min(max(len(env.globals)-start, 0), 8)
			if ptrCount > 1 {
				// Avoid eagerly creating pointers for all globals; closer to progressive creation.
				ptrCount = 1 + int(r.upto(uint32(ptrCount)))
			}
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
		er := newExprRand(r, exprDecisionBudget(opts))
		for _, p := range callee.params {
			argExprs = append(argExprs, randomParamExprDepth(p.ctype, er, opts, env, scope, 0, ctx))
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
	depth int,
	inLoop bool,
	stmtBudget *int,
	ctx *genContext,
	dec stmtDecision,
) bool {
	if stmtBudget != nil && *stmtBudget == 0 {
		return true
	}
	maybeDeclareOnDemandLocal(b, r, opts, ctx)
	if stmtBudget != nil && *stmtBudget > 0 {
		*stmtBudget = *stmtBudget - 1
	}
	chooseStmt := func() stmtKind {
		// Upstream default statement probabilities from pStatementProb:
		// ifelse 15, for 30, return 35, continue 40, break 45,
		// goto 50 (if jumps), arrayop 60/55 (if arrays), assign 100.
		tryPick := func() (stmtKind, bool) {
			v := int(dec.pick(0, 100))
			k := stmtAssign
			switch {
			case v < 15:
				k = stmtIfElse
			case v < 30:
				k = stmtFor
			case v < 35:
				k = stmtReturn
			case v < 40:
				k = stmtContinue
			case v < 45:
				k = stmtBreak
			case opts.Jumps && opts.Arrays && v < 50:
				k = stmtGoto
			case opts.Jumps && opts.Arrays && v < 60:
				k = stmtArrayOp
			case opts.Jumps && !opts.Arrays && v < 50:
				k = stmtGoto
			case !opts.Jumps && opts.Arrays && v < 55:
				k = stmtArrayOp
			default:
				k = stmtAssign
			}

			// StatementFilter-like constraints.
			if (k == stmtBreak || k == stmtContinue) && !inLoop {
				return k, false
			}
			if depth >= max(1, opts.MaxBlockDepth) && (k == stmtIfElse || k == stmtFor) {
				return k, false
			}
			if state != nil && len(state.funcs) >= state.maxFuncs && k == stmtInvoke {
				// Upstream StatementFilter filters out invoke when max funcs is reached.
				return k, false
			}
			return k, true
		}
		for i := 0; i < 8; i++ {
			k, ok := tryPick()
			if ok {
				return k
			}
		}
		return stmtAssign
	}

	switch chooseStmt() {
	case stmtAssign:
		if !emitLValueAssignment(b, r, opts, env, scope, ctx) {
			return false
		}
	case stmtIfElse:
		cond := fmt.Sprintf("((x & %du) != 0u)", 1+dec.pick(1, 7))
		if opts.ConstAsCondition && dec.pick(2, 5) == 0 {
			cond = fmt.Sprintf("(%s != 0u)", randomConstantExpr(CType{Name: "uint32_t", Signed: false, Bits: 32}, r, opts))
		} else if dec.pick(3, 3) == 0 {
			er := newExprRand(r, exprDecisionBudget(opts))
			noConst := !opts.ConstAsCondition
			e := randomTypedExprDepthFlags(CType{Name: "uint32_t", Signed: false, Bits: 32}, er, opts, env, scope, 0, ctx, false, noConst)
			cond = fmt.Sprintf("((uint32_t)%s != 0u)", e)
		}
		writeLine(b, 1, fmt.Sprintf("if %s {", cond))
		emitStatements(b, r, opts, env, scope, state, info, from, depth+1, false, stmtBudget, ctx)
		writeLine(b, 1, "} else {")
		if !emitStatement(b, r, opts, env, scope, state, info, from, depth+1, false, stmtBudget, ctx, nextStmtDecision(r)) {
			return false
		}
		writeLine(b, 1, "}")
	case stmtFor:
		bound := 1 + int(dec.pick(3, 5))
		writeLine(b, 1, fmt.Sprintf("for (uint32_t i = 0; i < %du; ++i) {", bound))
		writeLine(b, 2, fmt.Sprintf("x += (i ^ 0x%08Xu);", dec.vals[4]))
		if !emitStatement(b, r, opts, env, scope, state, info, from, depth+1, true, stmtBudget, ctx, nextStmtDecision(r)) {
			return false
		}
		writeLine(b, 1, "}")
	case stmtReturn:
		writeLine(b, 1, "l_0 ^= (uint32_t)x;")
		writeLine(b, 1, "return l_0;")
		if stmtBudget != nil && *stmtBudget > 0 {
			*stmtBudget = 0
		}
	case stmtContinue:
		writeLine(b, 1, "continue;")
	case stmtBreak:
		writeLine(b, 1, "break;")
	case stmtGoto:
		return false
	case stmtArrayOp:
		if !opts.Arrays || len(env.arrays) == 0 {
			return false
		}
		emitArrayMutation(b, r, opts, env, scope, ctx)
	case stmtInvoke:
		if !emitFunctionCallMutation(b, r, opts, env, scope, state, from, ctx) {
			return false
		}
	default:
		return false
	}
	return true
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
	depth int,
	inLoop bool,
	stmtBudget *int,
	ctx *genContext,
) {
	if stmtBudget != nil && *stmtBudget == 0 {
		return
	}
	stmtCount := 2 + int(r.upto(uint32(max(1, opts.MaxBlockSize))))
	for s := 0; s < stmtCount; s++ {
		if stmtBudget != nil && *stmtBudget == 0 {
			break
		}
		const maxStmtAttempts = 8
		ok := false
		for attempt := 0; attempt < maxStmtAttempts; attempt++ {
			dec := nextStmtDecision(r)
			snapStmtBudget := -1
			if stmtBudget != nil {
				snapStmtBudget = *stmtBudget
			}
			snap := takeGenSnapshot(ctx)

			var tmp strings.Builder
			if emitStatement(&tmp, r, opts, env, scope, state, info, from, depth, inLoop, stmtBudget, ctx, dec) {
				b.WriteString(tmp.String())
				ok = true
				break
			}

			// Reject path: rollback to keep statement-local retries side-effect free.
			if stmtBudget != nil && snapStmtBudget >= 0 {
				*stmtBudget = snapStmtBudget
			}
			restoreGenSnapshot(ctx, snap)
		}
		if !ok {
			// Last-resort deterministic no-op-like mutation when all attempts fail.
			writeLine(b, 1, "x ^= 0u;")
		}
	}
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
	const maxBuildAttempts = 6
	last := ""
	for i := 0; i < maxBuildAttempts; i++ {
		snapFuncsLen := len(state.funcs)
		snapNextIdx := state.nextIdx
		snapDynGlobalsLen := len(state.dynGlobals)
		snapNextGlobalID := state.nextGlobalID
		snapLateGlobals := state.lateGlobals.String()
		snapStmtBudget := -1
		if stmtBudget != nil {
			snapStmtBudget = *stmtBudget
		}

		candidate := emitSingleFuncDefOnce(r, opts, fn, state, idx, maxBlock, env, info, stmtBudget)
		last = candidate
		allGlobals := make([]globalInfo, 0, len(env.globals)+len(state.dynGlobals))
		allGlobals = append(allGlobals, env.globals...)
		allGlobals = append(allGlobals, state.dynGlobals...)
		if validateFunctionBody(candidate, allGlobals) {
			return candidate
		}

		// Reject path: discard side effects introduced during this attempt.
		state.funcs = state.funcs[:snapFuncsLen]
		state.nextIdx = snapNextIdx
		state.dynGlobals = state.dynGlobals[:snapDynGlobalsLen]
		state.nextGlobalID = snapNextGlobalID
		state.lateGlobals.Reset()
		state.lateGlobals.WriteString(snapLateGlobals)
		if stmtBudget != nil && snapStmtBudget >= 0 {
			*stmtBudget = snapStmtBudget
		}
	}
	return last
}

func emitSingleFuncDefOnce(
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

	localCap := max(2, min(maxBlock*2, 12))
	localCount := 1 + int(fdec.pick(0, uint32(localCap)))
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
	ctx := &genContext{
		state: state,
		from:  idx,
	}

	for _, p := range fn.params {
		writeLine(&b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", p.name))
	}
	emitStatements(&b, r, opts, env, scope, state, info, idx, 0, false, stmtBudget, ctx)
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

var (
	reUnsafeDivZero  = regexp.MustCompile(`/\s*\(\s*0[uUlL]*\s*\)`)
	reUnsafeModZero  = regexp.MustCompile(`%\s*\(\s*0[uUlL]*\s*\)`)
	reUnsafeShiftNeg = regexp.MustCompile(`(<<|>>)\s*\(\s*-\d+`)
)

func validateFunctionBody(def string, globals []globalInfo) bool {
	if strings.Count(def, "{") != strings.Count(def, "}") {
		return false
	}
	if !strings.Contains(def, "return") {
		return false
	}
	if (strings.Contains(def, "continue;") || strings.Contains(def, "break;")) && !strings.Contains(def, "for (") {
		return false
	}
	if reUnsafeDivZero.MatchString(def) || reUnsafeModZero.MatchString(def) || reUnsafeShiftNeg.MatchString(def) {
		return false
	}
	for _, g := range globals {
		if !g.isConst {
			continue
		}
		id := regexp.QuoteMeta(g.name)
		writePat := `\b(` + id + `\s*(=|\+=|-=|\*=|/=|%=|<<=|>>=|&=|\|=|\^=)|(\+\+|--)\s*` + id + `|` + id + `\s*(\+\+|--))\b`
		if regexp.MustCompile(writePat).FindString(def) != "" {
			return false
		}
	}
	return true
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

func emitFunctionsUpstreamFlow(b *strings.Builder, r *rng, opts Options, pool []CType, maxBlock int, env envInfo, info compositeInfo) ([]funcInfo, []globalInfo) {
	maxFuncs := max(opts.MaxFuncs, 1)
	state := &functionFlowState{
		funcs:        []funcInfo{makeFuncSignature(r, opts, pool, 1)},
		maxFuncs:     maxFuncs,
		nextIdx:      2,
		pool:         pool,
		opts:         opts,
		dynGlobals:   []globalInfo{},
		nextGlobalID: len(env.globals),
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

	if state.lateGlobals.Len() > 0 {
		b.WriteString(state.lateGlobals.String())
		writeLine(b, 0, "")
	}
	emitFuncDecls(b, state.funcs)
	for i := 0; i < len(defs); i++ {
		b.WriteString(defs[i])
	}
	return state.funcs, state.dynGlobals
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
	opts = opts.normalizeUpstreamFlow()

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
	funcs, dynGlobals := emitFunctionsUpstreamFlow(&b, r, opts, pool, opts.MaxBlockSize, env, info)
	if len(dynGlobals) > 0 {
		env.globals = append(env.globals, dynGlobals...)
	}

	// Phase 5 (upstream Output): main / checksums.
	if !opts.NoMain {
		emitMain(&b, opts, env, info, funcs[0].name)
	}

	return b.String(), nil
}
