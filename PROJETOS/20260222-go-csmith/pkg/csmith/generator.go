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
	funcs        []funcInfo
	built        []bool
	defs         []string
	maxFuncs     int
	nextIdx      int
	nextParamID  int
	nextLocalID  int
	pool         []CType
	info         compositeInfo
	opts         Options
	dynGlobals   []globalInfo
	lateGlobals  strings.Builder
	nextGlobalID int
	stmtBudget   int
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
)

type localInfo struct {
	name  string
	ctype CType
}

type scopeInfo struct {
	params    []paramInfo
	locals    []localInfo
	returnVar string
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
	info    compositeInfo
}

type genSnapshot struct {
	dynLocLen      int
	funcsLen       int
	builtLen       int
	defsLen        int
	nextIdx        int
	nextParamID    int
	nextLocalID    int
	dynGlobalsLen  int
	nextGlobalID   int
	stmtBudget     int
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
		s.builtLen = len(ctx.state.built)
		s.defsLen = len(ctx.state.defs)
		s.nextIdx = ctx.state.nextIdx
		s.nextParamID = ctx.state.nextParamID
		s.nextLocalID = ctx.state.nextLocalID
		s.dynGlobalsLen = len(ctx.state.dynGlobals)
		s.nextGlobalID = ctx.state.nextGlobalID
		s.stmtBudget = ctx.state.stmtBudget
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
		if len(ctx.state.built) >= s.builtLen {
			ctx.state.built = ctx.state.built[:s.builtLen]
		}
		if len(ctx.state.defs) >= s.defsLen {
			ctx.state.defs = ctx.state.defs[:s.defsLen]
		}
		if len(ctx.state.dynGlobals) >= s.dynGlobalsLen {
			ctx.state.dynGlobals = ctx.state.dynGlobals[:s.dynGlobalsLen]
		}
		ctx.state.nextIdx = s.nextIdx
		ctx.state.nextParamID = s.nextParamID
		ctx.state.nextLocalID = s.nextLocalID
		ctx.state.nextGlobalID = s.nextGlobalID
		ctx.state.stmtBudget = s.stmtBudget
		ctx.state.lateGlobals.Reset()
		ctx.state.lateGlobals.WriteString(s.lateGlobalsBuf)
	}
}

type stmtDecision struct {
	r    *rng
	vals [12]uint32
}

type exprRand struct {
	vals     []uint32
	idx      int
	fallback *rng
}

type funcDecision struct {
	r    *rng
	vals [16]uint32
}

func nextFuncDecision(r *rng) funcDecision {
	return funcDecision{r: r}
}

func (d funcDecision) pick(i int, n uint32) uint32 {
	if n == 0 || i < 0 || i >= len(d.vals) {
		return 0
	}
	if d.r != nil {
		return d.r.upto(n)
	}
	return d.vals[i] % n
}

func newExprRand(r *rng, budget int) *exprRand {
	if budget < 1 {
		budget = 1
	}
	_ = budget
	// Consume expression choices on-demand via rnd_upto in pick(),
	// closer to upstream random wrappers.
	return &exprRand{vals: nil, fallback: r}
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
	if e.fallback != nil {
		return e.fallback.upto(n)
	}
	return e.next() % n
}

func nextStmtDecision(r *rng) stmtDecision {
	return stmtDecision{r: r}
}

func (d stmtDecision) pick(i int, n uint32) uint32 {
	if n == 0 || i < 0 || i >= len(d.vals) {
		return 0
	}
	if d.r != nil {
		return d.r.upto(n)
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
	nextID   int
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
		// Synthetic accumulator local does not exist in upstream variable pools.
		if l.name == "x" {
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
		if l.name == "x" {
			continue
		}
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
			if l.name == "x" {
				continue
			}
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
	// Upstream defaults: regular const 10%, regular volatile 50% (when enabled).
	isConst := opts.Consts && er.pick(100) < 10
	isVolatile := opts.Volatiles && er.pick(100) < 50
	if isConst && isVolatile && er.pick(2) == 0 {
		isConst = false
	}
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

func createOnDemandFromParentLocalPathER(er *exprRand, opts Options, t CType, ctx *genContext) (exprVarCandidate, bool) {
	if er == nil || er.fallback == nil || ctx == nil || ctx.state == nil {
		return exprVarCandidate{}, false
	}
	// Type::random_type_from_type(type, true, false) for simple types maps to
	// choose_random_simple() and consumes rnd_upto(MAX_SIMPLE_TYPES=14).
	chosen := t
	if t.Bits > 0 && len(ctx.state.pool) > 0 {
		i := int(er.fallback.upto(14))
		if i >= 0 && i < len(ctx.state.pool) {
			chosen = ctx.state.pool[i]
		} else {
			chosen = ctx.state.pool[i%len(ctx.state.pool)]
		}
	}

	// create_and_initialize + qualifier path consume these decisions in upstream.
	_ = er.fallback.flipcoin(50) // NewArrayVariableProb
	_ = er.fallback.flipcoin(10) // const prob
	_ = er.fallback.flipcoin(20) // volatile prob
	_ = er.fallback.flipcoin(50)
	_ = er.fallback.flipcoin(50)
	_ = er.fallback.upto(20)

	// Materialize as a generated global in this simplified backend.
	return createOnDemandGlobalFromER(er, opts, chosen, ctx)
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

	useExisting := false
	if er != nil && er.fallback != nil {
		useExisting = er.fallback.flipcoin(50)
	} else {
		useExisting = er.pick(2) == 0
	}
	if er != nil && er.fallback != nil {
		p := uint32(0)
		if opts.Builtins {
			p = uint32(opts.BuiltinFunctionProb)
		}
		_ = er.fallback.flipcoin(p)
	}
	var callee funcInfo
	calleeIdx := -1
	if useExisting && len(candidates) > 0 {
		calleeIdx = candidates[int(er.pick(uint32(len(candidates))))]
		callee = state.funcs[calleeIdx]
	} else {
		if len(candidates) == 0 {
			// Upstream often fails this path early when no callable function is
			// available in context; caller retries another expression form.
			return "", false
		}
		created, newIdx, ok := state.appendNewFunction(er.fallback, &t)
		if ok {
			calleeIdx = newIdx
			callee = created
			if !state.built[calleeIdx] {
				state.defs[calleeIdx] = emitSingleFuncDef(
					er.fallback,
					opts,
					callee,
					state,
					calleeIdx,
					opts.MaxBlockSize,
					env,
					ctx.info,
					&state.stmtBudget,
				)
				state.built[calleeIdx] = true
			}
		} else if len(candidates) > 0 {
			calleeIdx = candidates[int(er.pick(uint32(len(candidates))))]
			callee = state.funcs[calleeIdx]
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

	type termEntry struct {
		term termChoice
		prob int
	}
	entries := make([]termEntry, 0, 5)
	funcW := 70
	varW := 20
	constW := 10
	assignW := 10
	commaW := 10
	if isParam {
		// Upstream make_random_param baseline:
		// function 40, variable 40, constant 0 (+ optional assign/comma).
		funcW = 40
		varW = 40
		constW = 0
	}
	if !opts.EmbeddedAssigns {
		assignW = 0
	}
	if !opts.CommaOperators {
		commaW = 0
	}
	entries = append(entries, termEntry{term: termFunction, prob: funcW})
	entries = append(entries, termEntry{term: termVariable, prob: varW})
	entries = append(entries, termEntry{term: termConstant, prob: constW})
	if assignW > 0 {
		entries = append(entries, termEntry{term: termAssign, prob: assignW})
	}
	if commaW > 0 {
		entries = append(entries, termEntry{term: termComma, prob: commaW})
	}
	maxProb := 0
	for _, e := range entries {
		maxProb += e.prob
	}
	if maxProb <= 0 {
		return randomConstantExprFromER(t, er, opts)
	}
	decode := func(v int) termChoice {
		for _, e := range entries {
			if e.prob <= 0 {
				continue
			}
			if v < e.prob {
				return e.term
			}
			v -= e.prob
		}
		return termVariable
	}
	disallowed := func(tc termChoice) bool {
		if (tc == termFunction && (noFunc || depth+2 > maxExprDepth(opts))) ||
			(tc == termConstant && noConst) ||
			((tc == termAssign || tc == termComma) && depth+2 > maxExprDepth(opts)) {
			return true
		}
		return false
	}

	for tries := 0; tries < 6; tries++ {
		snap := takeGenSnapshot(ctx)
		var choice termChoice
		if er != nil && er.fallback != nil {
			raw := int(er.fallback.uptoWithFilter(uint32(maxProb), func(x uint32) bool {
				return disallowed(decode(int(x)))
			}))
			choice = decode(raw)
		} else {
			raw := int(er.pick(uint32(maxProb)))
			choice = decode(raw)
			if disallowed(choice) {
				restoreGenSnapshot(ctx, snap)
				continue
			}
		}
		switch choice {
		case termFunction:
			if depth > 0 && er != nil && er.fallback != nil {
				// ExpressionFuncall::ExpressionFunctionProbability + stdfunc path.
				if er.fallback.flipcoin(80) {
					if er.fallback.flipcoin(5) {
						_ = er.fallback.flipcoin(50)
						_ = er.fallback.upto(4)
						// In upstream make_random_unary(), operand expression is
						// generated under the current expr_depth context.
						operand := randomTypedExprDepthFlags(t, er, opts, env, scope, depth, ctx, false, false)
						return castLiteral(t, fmt.Sprintf("(~(%s))", operand))
					}
					ptrCmpProb := uint32(0)
					if opts.Pointers {
						ptrCmpProb = 10
					}
					_ = er.fallback.flipcoin(ptrCmpProb)
					_ = er.fallback.upto(18)
					_ = er.fallback.flipcoin(50)
					_ = er.fallback.flipcoin(50)
					_ = er.fallback.upto(4)
					// In upstream make_random_binary(), lhs/rhs are generated
					// from child contexts carrying current expr_depth.
					lhs := randomTypedExprDepthFlags(t, er, opts, env, scope, depth, ctx, false, false)
					rhs := randomTypedExprDepthFlags(t, er, opts, env, scope, depth, ctx, false, false)
					return castLiteral(t, fmt.Sprintf("((%s) ^ (%s))", lhs, rhs))
				}
			}
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
			if scopePick == 1 && er != nil && er.fallback != nil {
				// Parent-local selection starts by choosing a parent stack block.
				_ = er.pick(1)
			}
			candidates := buildScopedCandidatesFromER(er, env, scope, scopePick, ctx)
			if len(candidates) == 0 {
				if scopePick == 0 {
					if g, ok := createOnDemandGlobalFromER(er, opts, t, ctx); ok {
						return castLiteral(t, g.expr)
					}
				}
				if scopePick == 1 {
					if g, ok := createOnDemandFromParentLocalPathER(er, opts, t, ctx); ok {
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
			lhsType := t
			if er != nil && er.fallback != nil && ctx != nil && ctx.state != nil {
				allCount := len(ctx.state.pool) + len(ctx.state.info.structs) + len(ctx.state.info.unions)
				if allCount > 0 {
					pick := int(er.fallback.upto(uint32(allCount)))
					switch {
					case pick < len(ctx.state.pool):
						lhsType = ctx.state.pool[pick]
					case pick < len(ctx.state.pool)+len(ctx.state.info.structs):
						lhsType = CType{Name: fmt.Sprintf("struct S%d", pick-len(ctx.state.pool)), Bits: 32}
					default:
						lhsType = CType{Name: fmt.Sprintf("union U%d", pick-len(ctx.state.pool)-len(ctx.state.info.structs)), Bits: 32}
					}
				}
			}
			// Upstream ExpressionComma::make_random:
			// lhs = make_random(..., type=nil, no_const=true), rhs = make_random(..., type=t)
			lhs := randomTypedExprDepthFlags(lhsType, er, opts, env, scope, depth+1, ctx, false, true)
			rhs := randomTypedExprDepthFlags(t, er, opts, env, scope, depth+1, ctx, false, false)
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
	_ = depth
	return randomLeafExpr(t, er, opts, env, scope, depth, ctx)
}

func randomTypedExprDepthFlags(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext, noFunc bool, noConst bool) string {
	_ = depth
	return randomLeafExprWithMode(t, er, opts, env, scope, depth, ctx, false, noFunc, noConst)
}

func randomParamExprDepth(t CType, er *exprRand, opts Options, env envInfo, scope scopeInfo, depth int, ctx *genContext) string {
	_ = depth
	return randomParamLeafExpr(t, er, opts, env, scope, depth, ctx)
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
			if l.name == "x" {
				continue
			}
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
	for _, l := range scope.locals {
		if l.name == "x" {
			targetType = l.ctype
			break
		}
	}

	lv := lvalueInfo{expr: "x", ctype: targetType}
	if len(scope.locals) == 0 {
		if picked, ok := chooseLValue(r, opts, targetType, env, scope, ctx); ok {
			lv = picked
		} else {
			// Upstream-like fallback: create a new value when selection fails.
			name := fmt.Sprintf("lv_%d", r.next31()&0xFFFF)
			writeLine(b, 1, fmt.Sprintf("%s %s = %s;", targetType.Name, name, randomConstantExpr(targetType, r, opts)))
			lv = lvalueInfo{expr: name, ctype: targetType}
		}
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
	// Disabled until parent-scope local placement mirrors upstream block scoping.
	_ = b
	_ = r
	_ = opts
	_ = ctx
}

func emitCompositeTypes(b *strings.Builder, r *rng, opts Options, pool []CType) compositeInfo {
	info := compositeInfo{}
	writeLine(b, 0, "/* --- Struct/Union Declarations --- */")
	// Upstream Probabilities defaults (Probabilities.cpp).
	const (
		moreStructUnionTypeProb       = 50
		bitfieldsCreationProb         = 50
		bitfieldInNormalStructProb    = 10
		scalarFieldInFullBitfieldProb = 10
		bitfieldsSignedProb           = 50
		fieldVolatileProb             = 30
		fieldConstProb                = 20
	)
	moreTypesProbability := func(existingTypeCount int) bool {
		// Type::MoreTypesProbability: keep adding while <10 total types,
		// then 50% chance for each additional aggregate type.
		if existingTypeCount < 10 {
			return true
		}
		return r.flipcoin(moreStructUnionTypeProb)
	}
	// Upstream Type::GenerateSimpleTypes pushes eChar..eUInt128, i.e. 13
	// simple types before aggregate generation starts.
	typeCount := 13
	fieldQual := func() string {
		// Mirrors CVQualifiers::random_qualifiers(..., FieldConstProb, FieldVolatileProb):
		// volatile draw first, then const draw.
		isVolatile := opts.VolStructUnionFields && r.flipcoin(fieldVolatileProb)
		isConst := opts.ConstStructUnionFields && r.flipcoin(fieldConstProb)
		if isConst && isVolatile && !opts.AllowConstVolatile {
			isConst = false
		}
		q := ""
		if isConst {
			q += "const "
		}
		if isVolatile {
			q += "volatile "
		}
		return q
	}
	bitfieldLength := func(maxLength int, prior []fieldInfo) int {
		if maxLength < 1 {
			maxLength = 1
		}
		length := int(r.upto(uint32(maxLength)))
		noZeroLen := len(prior) == 0 || (prior[len(prior)-1].bitfield && prior[len(prior)-1].bitWidth == 0)
		if length == 0 && noZeroLen {
			if maxLength <= 2 {
				length = 1
			} else {
				length = int(r.upto(uint32(maxLength-1))) + 1
			}
		}
		return length
	}

	if opts.Structs {
		sidx := 0
		maxStructs := min(max(opts.MaxStructFields, 1), 32)
		for sidx < maxStructs && moreTypesProbability(typeCount) {
			fieldCount := 1 + int(r.upto(uint32(max(1, opts.MaxStructFields))))
			st := structTypeInfo{fields: make([]fieldInfo, 0, fieldCount)}
			writeLine(b, 0, fmt.Sprintf("struct S%d {", sidx))
			fullBitfields := opts.Bitfields && r.flipcoin(bitfieldsCreationProb)
			for f := 0; f < fieldCount; f++ {
				if fullBitfields {
					if r.flipcoin(scalarFieldInFullBitfieldProb) {
						name := fmt.Sprintf("f%d", f)
						t := pickType(r, pool)
						writeLine(b, 1, fmt.Sprintf("%s%s %s;", fieldQual(), t.Name, name))
						st.fields = append(st.fields, fieldInfo{name: name, ctype: t})
						continue
					}
					name := fmt.Sprintf("f%d", f)
					base := "unsigned"
					if r.flipcoin(bitfieldsSignedProb) {
						base = "signed"
					}
					qual := fieldQual()
					width := bitfieldLength(opts.IntSize*8, st.fields)
					writeLine(b, 1, fmt.Sprintf("%s%s %s : %d;", qual, base, name, width))
					st.fields = append(st.fields, fieldInfo{
						name: name, ctype: CType{Name: "uint32_t", Bits: 32}, bitfield: true, bitWidth: width,
					})
					continue
				}
				if opts.Bitfields && r.flipcoin(bitfieldInNormalStructProb) {
					name := fmt.Sprintf("f%d", f)
					base := "unsigned"
					if r.flipcoin(bitfieldsSignedProb) {
						base = "signed"
					}
					qual := fieldQual()
					width := bitfieldLength(opts.IntSize*8, st.fields)
					writeLine(b, 1, fmt.Sprintf("%s%s %s : %d;", qual, base, name, width))
					st.fields = append(st.fields, fieldInfo{
						name: name, ctype: CType{Name: "uint32_t", Bits: 32}, bitfield: true, bitWidth: width,
					})
					continue
				}
				name := fmt.Sprintf("f%d", f)
				t := pickType(r, pool)
				writeLine(b, 1, fmt.Sprintf("%s%s %s;", fieldQual(), t.Name, name))
				st.fields = append(st.fields, fieldInfo{name: name, ctype: t})
			}
			if opts.PackedStruct {
				// Type::make_random_struct_type consumes rnd_flipcoin(50) when
				// packed-struct is enabled (default upstream behavior).
				_ = r.flipcoin(50)
			}
			writeLine(b, 0, "};")
			writeLine(b, 0, "")
			info.structs = append(info.structs, st)
			sidx++
			typeCount++
		}
	}

	if opts.Unions {
		uidx := 0
		maxUnions := min(max(opts.MaxUnionFields, 1), 32)
		for uidx < maxUnions && moreTypesProbability(typeCount) {
			fieldCount := 1 + int(r.upto(uint32(max(1, opts.MaxUnionFields))))
			ut := unionTypeInfo{fields: make([]fieldInfo, 0, fieldCount)}
			writeLine(b, 0, fmt.Sprintf("union U%d {", uidx))
			for f := 0; f < fieldCount; f++ {
				name := fmt.Sprintf("f%d", f)
				if opts.Bitfields && r.flipcoin(bitfieldInNormalStructProb) {
					base := "unsigned"
					if r.flipcoin(bitfieldsSignedProb) {
						base = "signed"
					}
					qual := fieldQual()
					width := bitfieldLength(opts.IntSize*8, ut.fields)
					writeLine(b, 1, fmt.Sprintf("%s%s %s : %d;", qual, base, name, width))
					ut.fields = append(ut.fields, fieldInfo{
						name: name, ctype: CType{Name: "uint32_t", Bits: 32}, bitfield: true, bitWidth: width,
					})
					continue
				}
				t := pickType(r, pool)
				writeLine(b, 1, fmt.Sprintf("%s%s %s;", fieldQual(), t.Name, name))
				ut.fields = append(ut.fields, fieldInfo{name: name, ctype: t})
			}
			writeLine(b, 0, "};")
			writeLine(b, 0, "")
			info.unions = append(info.unions, ut)
			uidx++
			typeCount++
		}
	}
	writeLine(b, 0, "")

	return info
}

func emitGlobals(b *strings.Builder, r *rng, opts Options, info compositeInfo, pool []CType) envInfo {
	env := envInfo{}
	nextGlobalID := 0
	writeLine(b, 0, "/* --- GLOBAL VARIABLES --- */")
	if opts.GlobalVariables {
		globalCap := max(opts.MaxGlobals, 2)
		env.globals = make([]globalInfo, 0, globalCap)
		env.arrays = make([]arrayInfo, 0, globalCap/2+1)
		env.pointers = make([]pointerInfo, 0, globalCap/2+1)

		newGlobalName := func() string {
			name := fmt.Sprintf("g_%d", nextGlobalID)
			nextGlobalID++
			return name
		}

		// Keep scalar globals as the primary pool used by expressions.
		scalarTarget := min(globalCap, 26+int(r.upto(18)))
		for i := 0; i < scalarTarget; i++ {
			// Upstream regular qualifiers (Probabilities.cpp defaults).
			isConst := false
			if opts.Consts {
				isConst = r.upto(100) < 10
			}
			isVolatile := false
			if opts.Volatiles {
				isVolatile = r.upto(100) < 50
			}
			// Avoid degenerate const+volatile saturation.
			if isConst && isVolatile && r.upto(2) == 0 {
				isConst = false
			}
			g := globalInfo{
				name:       newGlobalName(),
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
			writeLine(b, 0, fmt.Sprintf("static %s%s %s = %s;", qual, g.ctype.Name, g.name, lit))
			env.globals = append(env.globals, g)
		}

		if opts.Arrays {
			// Generate a richer array set with canonical g_N naming.
			arrayTarget := min(max(12, len(env.globals)), 40)
			for i := 0; i < arrayTarget; i++ {
				arrLen := 2 + int(r.upto(uint32(max(2, min(opts.MaxArrayLenPerDim, 10)))))
				ai := arrayInfo{
					name:  newGlobalName(),
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
			ptrTarget := min(max(4, len(env.globals)/4), 12)
			for i := 0; i < ptrTarget; i++ {
				target := env.globals[start+int(r.upto(uint32(max(1, len(env.globals)-start))))]
				p := pointerInfo{
					name:            newGlobalName(),
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

			// Extra pointer chains: mimic global pointer ladders seen in upstream output.
			chainCount := min(max(len(env.pointers)/2, 1), 4)
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
					name := newGlobalName()
					stars := strings.Repeat("*", d)
					writeLine(b, 0, fmt.Sprintf("static %s %s%s = &%s;", baseType, stars, name, prevName))
					prevName = name
					env.chains = append(env.chains, name)
				}
			}
		}
	}

	for i := range info.structs {
		writeLine(b, 0, fmt.Sprintf("static struct S%d gs_%d;", i, i))
	}
	for i := range info.unions {
		writeLine(b, 0, fmt.Sprintf("static union U%d gu_%d;", i, i))
	}
	env.nextID = nextGlobalID
	writeLine(b, 0, "")
	return env
}

func emitFuncDecls(b *strings.Builder, funcs []funcInfo) {
	writeLine(b, 0, "/* --- FORWARD DECLARATIONS --- */")
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
	writable := make([]globalInfo, 0, len(env.globals))
	for _, g := range env.globals {
		if g.isConst {
			continue
		}
		writable = append(writable, g)
	}
	if len(writable) == 0 {
		emitArithmeticMutation(b, r, opts)
		return
	}
	g := writable[int(r.upto(uint32(len(writable))))]
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

func (s *functionFlowState) appendNewFunction(r *rng, forceRet *CType) (funcInfo, int, bool) {
	if len(s.funcs) >= s.maxFuncs {
		return funcInfo{}, -1, false
	}
	fn := funcInfo{
		name: fmt.Sprintf("func_%d", s.nextIdx),
	}
	if forceRet != nil {
		fn.ret = *forceRet
		maxParams := s.opts.MaxParams
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
				name:  s.allocParamName(),
				ctype: pickType(r, s.pool),
			})
		}
	} else {
		fn = s.makeFuncSignature(r, s.nextIdx)
	}
	s.nextIdx++
	s.funcs = append(s.funcs, fn)
	s.built = append(s.built, false)
	s.defs = append(s.defs, "")
	return fn, len(s.funcs) - 1, true
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
		created, _, ok := state.appendNewFunction(r, nil)
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
		toKind := func(v int) stmtKind {
			switch {
			case v < 15:
				return stmtIfElse
			case v < 30:
				return stmtFor
			case v < 35:
				return stmtReturn
			case v < 40:
				return stmtContinue
			case v < 45:
				return stmtBreak
			case opts.Jumps && opts.Arrays && v < 50:
				return stmtGoto
			case opts.Jumps && opts.Arrays && v < 60:
				return stmtArrayOp
			case opts.Jumps && !opts.Arrays && v < 50:
				return stmtGoto
			case !opts.Jumps && opts.Arrays && v < 55:
				return stmtArrayOp
			default:
				return stmtAssign
			}
		}
		// Mimics upstream rnd_upto(..., StatementFilter):
		// retries happen inside one RNG API call.
		if dec.r != nil {
			v := int(dec.r.uptoWithFilter(100, func(x uint32) bool {
				k := toKind(int(x))
				if (k == stmtBreak || k == stmtContinue) && !inLoop {
					return true
				}
				if depth >= max(1, opts.MaxBlockDepth) && (k == stmtIfElse || k == stmtFor) {
					return true
				}
				return false
			}))
			return toKind(v)
		}
		return toKind(int(dec.pick(0, 100)))
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
		emitStatements(b, r, opts, env, scope, state, info, from, depth+1, false, stmtBudget, ctx)
		writeLine(b, 1, "}")
	case stmtFor:
		bound := 1 + int(dec.pick(3, 5))
		writeLine(b, 1, fmt.Sprintf("for (uint32_t i = 0; i < %du; ++i) {", bound))
		writeLine(b, 2, fmt.Sprintf("x += (i ^ 0x%08Xu);", dec.vals[4]))
		emitStatements(b, r, opts, env, scope, state, info, from, depth+1, true, stmtBudget, ctx)
		writeLine(b, 1, "}")
	case stmtReturn:
		ret := scope.returnVar
		if ret == "" {
			ret = "l_0"
		}
		writeLine(b, 1, fmt.Sprintf("%s ^= (uint32_t)x;", ret))
		writeLine(b, 1, fmt.Sprintf("return %s;", ret))
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
	stmtLimit := max(1, opts.MaxBlockSize)
	base := 2
	if depth > 0 {
		base = 1
	}
	stmtCount := base + int(r.upto(uint32(stmtLimit)))
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
	return emitSingleFuncDefOnce(r, opts, fn, state, idx, maxBlock, env, info, stmtBudget)
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
	retName := "l_0"
	if state != nil {
		retName = state.allocLocalName()
	}
	writeLine(&b, 1, fmt.Sprintf("%s %s = %s;", fn.ret.Name, retName, castLiteral(fn.ret, "0u")))
	if len(env.globals) >= 2 {
		writeLine(&b, 1, fmt.Sprintf("uint32_t x = ((uint32_t)%s) + ((uint32_t)%s);", env.globals[0].name, env.globals[1].name))
	} else {
		writeLine(&b, 1, "uint32_t x = 0u;")
	}

	locals := make([]localInfo, 0, 1)
	locals = append(locals, localInfo{name: "x", ctype: CType{Name: "uint32_t", Signed: false, Bits: 32}})
	scope := scopeInfo{params: fn.params, locals: locals, returnVar: retName}
	ctx := &genContext{
		state: state,
		from:  idx,
		info:  info,
	}

	for _, p := range fn.params {
		writeLine(&b, 1, fmt.Sprintf("x ^= (uint32_t)%s;", p.name))
	}
	emitStatements(&b, r, opts, env, scope, state, info, idx, 0, false, stmtBudget, ctx)
	if len(env.globals) > 0 {
		writable := make([]globalInfo, 0, len(env.globals))
		for _, g := range env.globals {
			if g.isConst {
				continue
			}
			writable = append(writable, g)
		}
		if len(writable) > 0 {
			g := writable[int(fdec.pick(2, uint32(len(writable))))]
			writeLine(&b, 1, fmt.Sprintf("%s ^= %s;", g.name, randomTypedExpr(g.ctype, r, opts, env, scope, ctx)))
		}
	}
	writeLine(&b, 1, fmt.Sprintf("%s ^= %s;", retName, castLiteral(fn.ret, "x")))
	writeLine(&b, 1, fmt.Sprintf("return %s;", retName))
	writeLine(&b, 0, "}")
	writeLine(&b, 0, "")
	return b.String()
}

func (s *functionFlowState) allocParamName() string {
	name := fmt.Sprintf("p_%d", s.nextParamID)
	s.nextParamID++
	return name
}

func (s *functionFlowState) allocLocalName() string {
	name := fmt.Sprintf("l_%d", s.nextLocalID)
	s.nextLocalID++
	return name
}

func (s *functionFlowState) makeFuncSignature(r *rng, idx int) funcInfo {
	fn := funcInfo{
		name: fmt.Sprintf("func_%d", idx),
	}
	// Function::make_first uses RandomReturnType() over AllTypes (simple + aggregates),
	// while later signatures are chosen from random types as they are created.
	if idx == 1 {
		allCount := len(s.pool) + len(s.info.structs) + len(s.info.unions)
		if allCount <= 0 {
			fn.ret = CType{Name: "uint32_t", Signed: false, Bits: 32}
		} else {
			pick := int(r.upto(uint32(allCount)))
			switch {
			case pick < len(s.pool):
				fn.ret = s.pool[pick]
			case pick < len(s.pool)+len(s.info.structs):
				fn.ret = CType{Name: fmt.Sprintf("struct S%d", pick-len(s.pool)), Bits: 32}
			default:
				fn.ret = CType{Name: fmt.Sprintf("union U%d", pick-len(s.pool)-len(s.info.structs)), Bits: 32}
			}
		}
	} else {
		fn.ret = pickType(r, s.pool)
	}
	if idx == 1 {
		// make_first() creates return variable qualifiers via
		// CVQualifiers::random_qualifiers(type) (no_volatile=true), which still
		// consumes volatile/const draws on the object itself.
		if s.opts.Volatiles {
			_ = r.flipcoin(50)
		} else {
			_ = r.flipcoin(0)
		}
		if s.opts.Consts {
			_ = r.flipcoin(10)
		} else {
			_ = r.flipcoin(0)
		}
	}
	maxParams := s.opts.MaxParams
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
			name:  s.allocParamName(),
			ctype: pickType(r, s.pool),
		})
	}
	return fn
}

func emitFunctionsUpstreamFlow(b *strings.Builder, r *rng, opts Options, pool []CType, maxBlock int, env envInfo, info compositeInfo) ([]funcInfo, []globalInfo) {
	maxFuncs := max(opts.MaxFuncs, 1)
	state := &functionFlowState{
		funcs:        []funcInfo{},
		built:        []bool{},
		defs:         []string{},
		maxFuncs:     maxFuncs,
		nextIdx:      2,
		nextParamID:  1,
		nextLocalID:  0,
		pool:         pool,
		info:         info,
		opts:         opts,
		dynGlobals:   []globalInfo{},
		nextGlobalID: env.nextID,
		stmtBudget:   opts.StopByStmt,
	}
	state.funcs = append(state.funcs, state.makeFuncSignature(r, 1))
	state.built = append(state.built, false)
	state.defs = append(state.defs, "")
	if state.stmtBudget < 0 {
		state.stmtBudget = -1
	}

	for cur := 0; cur < len(state.funcs); cur++ {
		if state.built[cur] {
			continue
		}
		state.defs[cur] = emitSingleFuncDef(r, opts, state.funcs[cur], state, cur, maxBlock, env, info, &state.stmtBudget)
		state.built[cur] = true
	}

	if state.lateGlobals.Len() > 0 {
		b.WriteString(state.lateGlobals.String())
		writeLine(b, 0, "")
	}
	emitFuncDecls(b, state.funcs)
	writeLine(b, 0, "/* --- FUNCTIONS --- */")
	writeLine(b, 0, "/* ------------------------------------------ */")
	for i := 0; i < len(state.defs); i++ {
		b.WriteString(state.defs[i])
	}
	return state.funcs, state.dynGlobals
}

func emitComputeHashFunc(b *strings.Builder, env envInfo, info compositeInfo) {
	writeLine(b, 0, "void csmith_compute_hash(int print_hash_value)")
	writeLine(b, 0, "{")
	for _, g := range env.globals {
		writeLine(b, 1, fmt.Sprintf("transparent_crc((uint64_t)%s, \"%s\", print_hash_value);", g.name, g.name))
	}
	for _, arr := range env.arrays {
		writeLine(b, 1, fmt.Sprintf("for (int i = 0; i < %d; i++)", arr.len))
		writeLine(b, 2, fmt.Sprintf("transparent_crc((uint64_t)%s[i], \"%s[i]\", print_hash_value);", arr.name, arr.name))
	}
	_ = info
	writeLine(b, 0, "}")
	writeLine(b, 0, "")
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
	writeLine(b, 1, fmt.Sprintf("(void)%s();", entry))
	if opts.ComputeHash {
		writeLine(b, 1, "csmith_compute_hash(print_hash_value);")
		if useRuntime {
			writeLine(b, 1, "platform_main_end(crc32_context ^ 0xFFFFFFFFUL, print_hash_value);")
		} else {
			writeLine(b, 1, "platform_main_end(0,0);")
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
	gen := createProgramGenerator(opts)
	gen.initialize()
	return gen.goGenerator(), nil
}
