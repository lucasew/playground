package csmith

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

const defaultPlatformInfoPath = "platform.info"

// Options is the canonical API-level configuration contract for generation.
// Defaults are aligned with Csmith's CGOptions::set_default_settings where possible.
type Options struct {
	Seed uint64

	// Output/layout
	OutputPath    string
	MaxSplitFiles int
	SplitFilesDir string
	NoMain        bool

	// Target sizing (from platform.info or explicit override)
	PlatformInfoPath string
	IntSize          int
	PointerSize      int
	IntSizeExplicit  bool
	PointerExplicit  bool

	// Size/depth controls
	MaxFuncs             int
	MaxParams            int
	Func1MaxParams       int
	MaxBlockSize         int
	MaxBlockDepth        int
	MaxExprComplexity    int
	MaxStructFields      int
	MaxUnionFields       int
	MaxNestedStructLevel int
	MaxPointerDepth      int
	MaxArrayDim          int
	MaxArrayLenPerDim    int
	MaxArrayLength       int
	MaxArrayNumInLoop    int
	MaxExhaustiveDepth   int
	InlineFunctionProb   int
	BuiltinFunctionProb  int
	ArrayOOBProb         int
	NullPtrDerefProb     int
	DanglingPtrDerefProb int
	StopByStmt           int
	CoverageTestSize     int

	// Extension/mode switches
	RandomBased   bool
	DFSExhaustive bool
	LangCPP       bool
	CPP11         bool
	FastExecution bool
	DepthProtect  bool

	// Core generation features
	ComputeHash              bool
	AcceptArgc               bool
	Arrays                   bool
	Bitfields                bool
	CompoundAssignment       bool
	Consts                   bool
	Divs                     bool
	Muls                     bool
	EmbeddedAssigns          bool
	CommaOperators           bool
	PreIncrOperator          bool
	PreDecrOperator          bool
	PostIncrOperator         bool
	PostDecrOperator         bool
	UnaryPlusOperator        bool
	Jumps                    bool
	LongLong                 bool
	Int8                     bool
	UInt8                    bool
	EnableFloat              bool
	Math64                   bool
	InlineFunction           bool
	Pointers                 bool
	Structs                  bool
	ReturnStructs            bool
	ArgStructs               bool
	Unions                   bool
	ReturnUnions             bool
	ArgUnions                bool
	TakeUnionFieldAddr       bool
	VolStructUnionFields     bool
	ConstStructUnionFields   bool
	Volatiles                bool
	VolatilePointers         bool
	ConstPointers            bool
	GlobalVariables          bool
	StrictConstArrays        bool
	AccessOnce               bool
	StrictVolatileRule       bool
	AddrTakenOfLocals        bool
	DanglingGlobalPointers   bool
	NoReturnDeadPointer      bool
	HashValuePrintf          bool
	SignedCharIndex          bool
	ForceGlobalsStatic       bool
	ForceNonUniformArrayInit bool
	Int128                   bool
	UInt128                  bool
	BinaryConstant           bool
	SafeMath                 bool
	PackedStruct             bool
	Paranoid                 bool
	Quiet                    bool
	Concise                  bool
	Builtins                 bool
	RandomRandom             bool
	StepHashByStmt           bool
	ConstAsCondition         bool
	MatchExactQualifiers     bool
	BlindCheckGlobal         bool
	FreshArrayCtrlVarNames   bool
	IdentifyWrappers         bool
	MarkMutableConst         bool
	Klee                     bool
	Crest                    bool
	CComp                    bool
	CoverageTest             bool
	FixedStructFields        bool
	ExpandStruct             bool
	CompactOutput            bool
	PrefixName               bool
	SequenceNamePrefix       bool
	CompatibleCheck          bool
	MathNoTmp                bool
	StrictFloat              bool
	WrapVolatiles            bool
	AllowConstVolatile       bool
	FunctionAttributes       bool
	TypeAttributes           bool
	LabelAttributes          bool
	VariableAttributes       bool

	StructOutput             string
	DFSDebugSequence         string
	PartialExpand            string
	DeltaMonitor             string
	DeltaOutput              string
	GoDelta                  string
	DeltaInput               string
	ProbabilityConfiguration string
	DumpDefaultProbabilities string
	DumpRandomProbabilities  string
	SafeMathWrappers         string
	MonitorFuncs             string
	EnableBuiltinKinds       string
	DisableBuiltinKinds      string
	NoDeltaReduction         bool

	// Keep an escape hatch for the current simplified generator shape.
	MaxGlobals int
}

func Defaults() Options {
	return Options{
		OutputPath:       "",
		MaxSplitFiles:    0,
		SplitFilesDir:    "",
		NoMain:           false,
		PlatformInfoPath: defaultPlatformInfoPath,
		IntSize:          int(unsafe.Sizeof(int(0))),
		PointerSize:      int(unsafe.Sizeof(uintptr(0))),

		MaxFuncs:             10,
		MaxParams:            5,
		Func1MaxParams:       3,
		MaxBlockSize:         4,
		MaxBlockDepth:        5,
		MaxExprComplexity:    10,
		MaxStructFields:      10,
		MaxUnionFields:       5,
		MaxNestedStructLevel: 3,
		MaxPointerDepth:      2,
		MaxArrayDim:          3,
		MaxArrayLenPerDim:    10,
		MaxArrayLength:       256,
		MaxArrayNumInLoop:    4,
		MaxExhaustiveDepth:   -1,
		InlineFunctionProb:   50,
		BuiltinFunctionProb:  50,
		ArrayOOBProb:         0,
		NullPtrDerefProb:     0,
		DanglingPtrDerefProb: 0,
		StopByStmt:           -1,
		CoverageTestSize:     500,

		RandomBased:   true,
		DFSExhaustive: false,
		LangCPP:       false,
		CPP11:         false,
		FastExecution: false,
		DepthProtect:  false,

		ComputeHash:              true,
		AcceptArgc:               true,
		Arrays:                   true,
		Bitfields:                true,
		CompoundAssignment:       true,
		Consts:                   true,
		Divs:                     true,
		Muls:                     true,
		EmbeddedAssigns:          true,
		CommaOperators:           true,
		PreIncrOperator:          true,
		PreDecrOperator:          true,
		PostIncrOperator:         true,
		PostDecrOperator:         true,
		UnaryPlusOperator:        true,
		Jumps:                    true,
		LongLong:                 true,
		Int8:                     true,
		UInt8:                    true,
		EnableFloat:              false,
		Math64:                   true,
		InlineFunction:           false,
		Pointers:                 true,
		Structs:                  true,
		ReturnStructs:            true,
		ArgStructs:               true,
		Unions:                   true,
		ReturnUnions:             true,
		ArgUnions:                true,
		TakeUnionFieldAddr:       true,
		VolStructUnionFields:     true,
		ConstStructUnionFields:   true,
		Volatiles:                true,
		VolatilePointers:         true,
		ConstPointers:            true,
		GlobalVariables:          true,
		StrictConstArrays:        false,
		AccessOnce:               false,
		StrictVolatileRule:       false,
		AddrTakenOfLocals:        true,
		DanglingGlobalPointers:   true,
		NoReturnDeadPointer:      true,
		HashValuePrintf:          true,
		SignedCharIndex:          true,
		ForceGlobalsStatic:       true,
		ForceNonUniformArrayInit: true,
		Int128:                   false,
		UInt128:                  false,
		BinaryConstant:           false,
		SafeMath:                 true,
		PackedStruct:             true,
		Paranoid:                 false,
		Quiet:                    false,
		Concise:                  false,
		Builtins:                 false,
		RandomRandom:             false,
		StepHashByStmt:           false,
		ConstAsCondition:         false,
		MatchExactQualifiers:     false,
		BlindCheckGlobal:         false,
		FreshArrayCtrlVarNames:   false,
		IdentifyWrappers:         false,
		MarkMutableConst:         false,
		Klee:                     false,
		Crest:                    false,
		CComp:                    false,
		CoverageTest:             false,
		FixedStructFields:        false,
		ExpandStruct:             false,
		CompactOutput:            false,
		PrefixName:               false,
		SequenceNamePrefix:       false,
		CompatibleCheck:          false,
		MathNoTmp:                false,
		StrictFloat:              false,
		WrapVolatiles:            false,
		AllowConstVolatile:       true,
		FunctionAttributes:       false,
		TypeAttributes:           false,
		LabelAttributes:          false,
		VariableAttributes:       false,
		StructOutput:             "",
		DFSDebugSequence:         "",
		PartialExpand:            "",
		DeltaMonitor:             "",
		DeltaOutput:              "",
		GoDelta:                  "",
		DeltaInput:               "",
		ProbabilityConfiguration: "",
		DumpDefaultProbabilities: "",
		DumpRandomProbabilities:  "",
		SafeMathWrappers:         "",
		MonitorFuncs:             "",
		EnableBuiltinKinds:       "",
		DisableBuiltinKinds:      "",
		NoDeltaReduction:         false,

		MaxGlobals: 80,
	}
}

func (o Options) resolvePlatformInfo() (Options, error) {
	path := strings.TrimSpace(o.PlatformInfoPath)
	if path == "" {
		path = defaultPlatformInfoPath
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if o.IntSize <= 0 {
				o.IntSize = int(unsafe.Sizeof(int(0)))
			}
			if o.PointerSize <= 0 {
				o.PointerSize = int(unsafe.Sizeof(uintptr(0)))
			}
			return o, nil
		}
		return o, err
	}
	defer f.Close()

	seenInt := false
	seenPtr := false
	fileInt := 0
	filePtr := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "integer size =") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "integer size ="))
			n, err := strconv.Atoi(v)
			if err != nil {
				return o, fmt.Errorf("invalid integer size in %s", path)
			}
			fileInt = n
			seenInt = true
		}
		if strings.HasPrefix(line, "pointer size =") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "pointer size ="))
			n, err := strconv.Atoi(v)
			if err != nil {
				return o, fmt.Errorf("invalid pointer size in %s", path)
			}
			filePtr = n
			seenPtr = true
		}
	}
	if err := scanner.Err(); err != nil {
		return o, err
	}
	if !seenInt {
		return o, fmt.Errorf("please specify integer size in %s", path)
	}
	if !seenPtr {
		return o, fmt.Errorf("please specify pointer size in %s", path)
	}
	if !o.IntSizeExplicit {
		o.IntSize = fileInt
	}
	if !o.PointerExplicit {
		o.PointerSize = filePtr
	}
	return o, nil
}

func (o Options) Validate() error {
	if o.IntSize <= 0 {
		return fmt.Errorf("int-size must be positive")
	}
	if o.PointerSize <= 0 {
		return fmt.Errorf("ptr-size must be positive")
	}
	if o.MaxFuncs < 1 {
		return fmt.Errorf("max-funcs must be at least 1")
	}
	if o.MaxBlockSize < 1 {
		return fmt.Errorf("max-block-size must be at least 1")
	}
	if o.MaxBlockDepth < 1 {
		return fmt.Errorf("max-stmt-depth must be at least 1")
	}
	if o.MaxGlobals < 1 {
		return fmt.Errorf("max-globals must be at least 1")
	}
	if o.Func1MaxParams > o.MaxParams {
		return fmt.Errorf("func1_max_params() cannot be larger than max_params()")
	}
	if o.InlineFunctionProb < 0 || o.InlineFunctionProb > 100 {
		return fmt.Errorf("inline-function-prob value must between [0,100]")
	}
	if o.BuiltinFunctionProb < 0 || o.BuiltinFunctionProb > 100 {
		return fmt.Errorf("builtin-function-prob value must between [0,100]")
	}
	if o.ArrayOOBProb < 0 || o.ArrayOOBProb > 100 {
		return fmt.Errorf("array-oob-prob value must between [0,100]")
	}
	if o.NullPtrDerefProb < 0 || o.NullPtrDerefProb > 100 {
		return fmt.Errorf("null-ptr-deref-prob value must between [0,100]")
	}
	if o.DanglingPtrDerefProb < 0 || o.DanglingPtrDerefProb > 100 {
		return fmt.Errorf("dangling-ptr-deref-prob value must between [0,100]")
	}
	if !o.LangCPP && o.CPP11 {
		return fmt.Errorf("--cpp11 option makes sense only with --lang-cpp option enabled")
	}
	if o.DFSExhaustive {
		if o.MaxExhaustiveDepth <= 0 {
			return fmt.Errorf("max-exhaustive-depth must be at least 0")
		}
		if !o.Structs && o.ExpandStruct {
			return fmt.Errorf("expand-struct/struct-specific options cannot be used with --no-structs")
		}
		if o.Klee || o.Crest || o.CoverageTest {
			return fmt.Errorf("exhaustive mode doesn't support klee|crest|coverage-test extension")
		}
	}
	if o.RandomBased {
		if o.DFSExhaustive {
			return fmt.Errorf("random-based and dfs-exhaustive modes cannot both be enabled")
		}
		if o.SequenceNamePrefix {
			return fmt.Errorf("--sequence-name-prefix option can only be used with --dfs-exhaustive")
		}
	}
	if !o.RandomBased {
		if o.MaxSplitFiles > 0 {
			return fmt.Errorf("max_split_files can only be applied to random mode")
		}
		if o.SplitFilesDir != "" {
			return fmt.Errorf("split_files_dir can only be applied to random mode")
		}
	}
	if o.DeltaMonitor != "" && o.GoDelta != "" {
		return fmt.Errorf("you cannot specify --delta-monitor and --go-delta monitor at the same time")
	}
	if o.MaxSplitFiles > 0 && o.SplitFilesDir == "" {
		o.SplitFilesDir = "./output"
		if err := os.MkdirAll(o.SplitFilesDir, 0o755); err != nil {
			return fmt.Errorf("cannot create dir for split files: %w", err)
		}
	}
	extCount := 0
	if o.Klee {
		extCount++
	}
	if o.Crest {
		extCount++
	}
	if o.CoverageTest {
		extCount++
	}
	if extCount > 1 {
		return fmt.Errorf("you could only specify --klee or --crest or --coverage-test")
	}
	return nil
}

func (o Options) normalizeUpstreamFlow() Options {
	// Upstream fast-execution turns on C++ mode and tightens options.
	if o.FastExecution {
		o.LangCPP = true
		o.Jumps = false
		o.MaxArrayLenPerDim = min(o.MaxArrayLenPerDim, 5)
	}
	// Upstream C++ normalization.
	if o.LangCPP {
		o.MatchExactQualifiers = true
		o.VolStructUnionFields = false
		o.ConstStructUnionFields = false
	}
	// Upstream DFS mode forces fixed struct fields.
	if o.DFSExhaustive {
		o.FixedStructFields = true
	}
	return o
}

func (o Options) validate() error {
	return o.normalizeUpstreamFlow().Validate()
}
