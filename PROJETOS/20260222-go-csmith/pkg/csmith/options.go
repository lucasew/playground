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
	MaxArrayNumInLoop    int
	MaxExhaustiveDepth   int
	InlineFunctionProb   int
	BuiltinFunctionProb  int
	ArrayOOBProb         int
	NullPtrDerefProb     int
	DanglingPtrDerefProb int

	// Extension/mode switches
	RandomBased   bool
	DFSExhaustive bool
	LangCPP       bool
	CPP11         bool
	FastExecution bool

	// Core generation features
	ComputeHash        bool
	AcceptArgc         bool
	Arrays             bool
	Bitfields          bool
	CompoundAssignment bool
	Consts             bool
	Divs               bool
	EmbeddedAssigns    bool
	PreIncrOperator    bool
	PreDecrOperator    bool
	PostIncrOperator   bool
	PostDecrOperator   bool
	UnaryPlusOperator  bool
	Jumps              bool
	LongLong           bool
	Int8               bool
	UInt8              bool
	EnableFloat        bool
	Math64             bool
	InlineFunction     bool
	Pointers           bool
	Structs            bool
	Unions             bool
	Volatiles          bool
	VolatilePointers   bool
	ConstPointers      bool
	GlobalVariables    bool
	SafeMath           bool
	PackedStruct       bool
	Paranoid           bool
	Quiet              bool
	Concise            bool

	// Keep an escape hatch for the current simplified generator shape.
	MaxGlobals int
}

func Defaults() Options {
	return Options{
		OutputPath:       "",
		MaxSplitFiles:    0,
		SplitFilesDir:    "./output",
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
		MaxPointerDepth:      5,
		MaxArrayDim:          3,
		MaxArrayLenPerDim:    10,
		MaxArrayNumInLoop:    4,
		MaxExhaustiveDepth:   -1,
		InlineFunctionProb:   50,
		BuiltinFunctionProb:  50,
		ArrayOOBProb:         0,
		NullPtrDerefProb:     0,
		DanglingPtrDerefProb: 0,

		RandomBased:   true,
		DFSExhaustive: false,
		LangCPP:       false,
		CPP11:         false,
		FastExecution: false,

		ComputeHash:        true,
		AcceptArgc:         true,
		Arrays:             true,
		Bitfields:          true,
		CompoundAssignment: true,
		Consts:             true,
		Divs:               true,
		EmbeddedAssigns:    true,
		PreIncrOperator:    true,
		PreDecrOperator:    true,
		PostIncrOperator:   true,
		PostDecrOperator:   true,
		UnaryPlusOperator:  true,
		Jumps:              true,
		LongLong:           true,
		Int8:               true,
		UInt8:              true,
		EnableFloat:        false,
		Math64:             true,
		InlineFunction:     false,
		Pointers:           true,
		Structs:            true,
		Unions:             true,
		Volatiles:          true,
		VolatilePointers:   true,
		ConstPointers:      true,
		GlobalVariables:    true,
		SafeMath:           true,
		PackedStruct:       true,
		Paranoid:           false,
		Quiet:              false,
		Concise:            false,

		MaxGlobals: 6,
	}
}

func (o Options) resolvePlatformInfo() (Options, error) {
	if o.IntSize > 0 && o.PointerSize > 0 {
		path := strings.TrimSpace(o.PlatformInfoPath)
		if path == "" {
			path = defaultPlatformInfoPath
		}
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return o, nil
			}
			return o, err
		}
		defer f.Close()

		seenInt := false
		seenPtr := false

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "integer size =") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "integer size ="))
				n, err := strconv.Atoi(v)
				if err != nil {
					return o, fmt.Errorf("invalid integer size in %s", path)
				}
				o.IntSize = n
				seenInt = true
			}
			if strings.HasPrefix(line, "pointer size =") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "pointer size ="))
				n, err := strconv.Atoi(v)
				if err != nil {
					return o, fmt.Errorf("invalid pointer size in %s", path)
				}
				o.PointerSize = n
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
	if !o.LangCPP && o.CPP11 {
		return fmt.Errorf("--cpp11 option makes sense only with --lang-cpp option enabled")
	}
	if o.DFSExhaustive {
		if o.MaxExhaustiveDepth <= 0 {
			return fmt.Errorf("max-exhaustive-depth must be at least 0")
		}
		if !o.Structs && o.PackedStruct {
			return fmt.Errorf("expand-struct/struct-specific options cannot be used with --no-structs")
		}
	}
	if o.RandomBased {
		if o.DFSExhaustive {
			return fmt.Errorf("random-based and dfs-exhaustive modes cannot both be enabled")
		}
	}
	if !o.RandomBased {
		if o.MaxSplitFiles > 0 {
			return fmt.Errorf("max_split_files can only be applied to random mode")
		}
		if o.SplitFilesDir != "" && o.SplitFilesDir != "./output" {
			return fmt.Errorf("split_files_dir can only be applied to random mode")
		}
	}
	if o.FastExecution && !o.LangCPP {
		return fmt.Errorf("fast-execution requires C++ mode semantics (lang-cpp)")
	}
	return nil
}

func (o Options) validate() error {
	return o.Validate()
}
