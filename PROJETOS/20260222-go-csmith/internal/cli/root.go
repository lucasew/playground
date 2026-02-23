package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"csmith/pkg/csmith"
)

const (
	appName    = "csmith-go"
	appVersion = "0.1.0"
)

type negBoolBinding struct {
	target *bool
	neg    *bool
}

func addBoolPair(cmd *cobra.Command, bindings *[]negBoolBinding, target *bool, name string, usage string) {
	neg := new(bool)
	cmd.Flags().BoolVar(target, name, *target, usage)
	cmd.Flags().BoolVar(neg, "no-"+name, false, "disable "+name)
	*bindings = append(*bindings, negBoolBinding{target: target, neg: neg})
}

func NewRootCmd() *cobra.Command {
	opts := csmith.Defaults()
	seedSet := false
	outputPath := ""
	showVersion := false
	mainFlag := false
	nomainFlag := false
	takeNoUnionFieldAddrFlag := false
	returnDeadPointerFlag := false
	noHashValuePrintfFlag := false
	noSignedCharIndexFlag := false
	compilerAttributesFlag := false
	noCompilerAttributesFlag := false
	negBindings := make([]negBoolBinding, 0, 32)

	cmd := &cobra.Command{
		Use:           appName,
		Short:         "Random C program generator (Csmith port in progress)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unexpected arguments: %v", args)
			}

			if showVersion {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", appName, appVersion)
				return err
			}

			if mainFlag && nomainFlag {
				return fmt.Errorf("options conflict: cannot use --main and --nomain together")
			}
			if mainFlag {
				opts.NoMain = false
			}
			if nomainFlag {
				opts.NoMain = true
			}
			if takeNoUnionFieldAddrFlag {
				opts.TakeUnionFieldAddr = false
			}
			if returnDeadPointerFlag {
				opts.NoReturnDeadPointer = false
			}
			if noHashValuePrintfFlag {
				opts.HashValuePrintf = false
			}
			if noSignedCharIndexFlag {
				opts.SignedCharIndex = false
			}
			if compilerAttributesFlag {
				opts.FunctionAttributes = true
				opts.TypeAttributes = true
				opts.LabelAttributes = true
				opts.VariableAttributes = true
			}
			if noCompilerAttributesFlag {
				opts.FunctionAttributes = false
				opts.TypeAttributes = false
				opts.LabelAttributes = false
				opts.VariableAttributes = false
			}

			if !seedSet {
				opts.Seed = uint64(time.Now().UnixNano())
			}
			if opts.DFSExhaustive {
				// Upstream parser flips random_based off when dfs-exhaustive is enabled.
				opts.RandomBased = false
			}
			opts.OutputPath = outputPath

			program, err := csmith.Generate(opts)
			if err != nil {
				return err
			}

			if outputPath == "" {
				_, err = fmt.Fprint(cmd.OutOrStdout(), program)
				return err
			}
			return os.WriteFile(outputPath, []byte(program), 0o644)
		},
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "print version")
	cmd.Flags().Uint64VarP(&opts.Seed, "seed", "s", 0, "seed for deterministic generation")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write generated C code to file")
	cmd.Flags().StringVar(&opts.PlatformInfoPath, "platform-info", opts.PlatformInfoPath, "path to platform.info")
	cmd.Flags().IntVar(&opts.IntSize, "int-size", opts.IntSize, "target integer size in bytes")
	cmd.Flags().IntVar(&opts.PointerSize, "ptr-size", opts.PointerSize, "target pointer size in bytes")

	cmd.Flags().IntVar(&opts.MaxFuncs, "max-funcs", opts.MaxFuncs, "limit number of functions besides main")
	cmd.Flags().IntVar(&opts.MaxParams, "max-params", opts.MaxParams, "limit number of function parameters")
	cmd.Flags().IntVar(&opts.Func1MaxParams, "func1_max_params", opts.Func1MaxParams, "number of symbolic parameters passed to func_1")
	cmd.Flags().IntVar(&opts.MaxBlockSize, "max-block-size", opts.MaxBlockSize, "limit statements per block")
	cmd.Flags().IntVar(&opts.MaxBlockDepth, "max-block-depth", opts.MaxBlockDepth, "limit depth of nested blocks")
	cmd.Flags().IntVar(&opts.MaxExprComplexity, "max-expr-complexity", opts.MaxExprComplexity, "limit expression complexity")
	cmd.Flags().IntVar(&opts.MaxStructFields, "max-struct-fields", opts.MaxStructFields, "limit struct field count")
	cmd.Flags().IntVar(&opts.MaxNestedStructLevel, "max-struct-nested-level", opts.MaxNestedStructLevel, "limit nested struct depth")
	cmd.Flags().IntVar(&opts.MaxNestedStructLevel, "max-nested-struct-level", opts.MaxNestedStructLevel, "limit nested struct depth")
	cmd.Flags().IntVar(&opts.MaxUnionFields, "max-union-fields", opts.MaxUnionFields, "limit union field count")
	cmd.Flags().IntVar(&opts.MaxPointerDepth, "max-pointer-depth", opts.MaxPointerDepth, "limit pointer indirection depth")
	cmd.Flags().IntVar(&opts.MaxArrayDim, "max-array-dim", opts.MaxArrayDim, "limit array dimensions")
	cmd.Flags().IntVar(&opts.MaxArrayLenPerDim, "max-array-len-per-dim", opts.MaxArrayLenPerDim, "limit array length per dimension")
	cmd.Flags().IntVar(&opts.MaxArrayLength, "max-array-length", opts.MaxArrayLength, "limit total array length")
	cmd.Flags().IntVar(&opts.MaxExhaustiveDepth, "max-exhaustive-depth", opts.MaxExhaustiveDepth, "maximum exhaustive depth")
	cmd.Flags().IntVar(&opts.InlineFunctionProb, "inline-function-prob", opts.InlineFunctionProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.BuiltinFunctionProb, "builtin-function-prob", opts.BuiltinFunctionProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.ArrayOOBProb, "array-oob-prob", opts.ArrayOOBProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.NullPtrDerefProb, "null-ptr-deref-prob", opts.NullPtrDerefProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.DanglingPtrDerefProb, "dangling-ptr-deref-prob", opts.DanglingPtrDerefProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.StopByStmt, "stop-by-stmt", opts.StopByStmt, "stop generation after N statements")
	cmd.Flags().IntVar(&opts.MaxGlobals, "max-globals", opts.MaxGlobals, "maximum number of generated globals")
	cmd.Flags().IntVar(&opts.MaxSplitFiles, "max-split-files", opts.MaxSplitFiles, "maximum number of split output files")
	cmd.Flags().IntVar(&opts.CoverageTestSize, "coverage-test-size", opts.CoverageTestSize, "coverage test size")
	cmd.Flags().StringVar(&opts.SplitFilesDir, "split-files-dir", opts.SplitFilesDir, "directory for split files output")
	cmd.Flags().StringVar(&opts.StructOutput, "struct-output", opts.StructOutput, "write generated struct declarations to file")
	cmd.Flags().StringVar(&opts.DFSDebugSequence, "dfs-debug-sequence", opts.DFSDebugSequence, "debug sequence for dfs-exhaustive mode")
	cmd.Flags().StringVar(&opts.PartialExpand, "partial-expand", opts.PartialExpand, "partial expansion strategy")
	cmd.Flags().StringVar(&opts.DeltaMonitor, "delta-monitor", opts.DeltaMonitor, "delta monitor executable")
	cmd.Flags().StringVar(&opts.DeltaOutput, "delta-output", opts.DeltaOutput, "delta output file")
	cmd.Flags().StringVar(&opts.GoDelta, "go-delta", opts.GoDelta, "run selected delta reduction monitor")
	cmd.Flags().StringVar(&opts.DeltaInput, "delta-input", opts.DeltaInput, "delta input file")
	cmd.Flags().StringVar(&opts.ProbabilityConfiguration, "probability-configuration", opts.ProbabilityConfiguration, "probability configuration file")
	cmd.Flags().StringVar(&opts.DumpDefaultProbabilities, "dump-default-probabilities", opts.DumpDefaultProbabilities, "dump default probabilities to file")
	cmd.Flags().StringVar(&opts.DumpRandomProbabilities, "dump-random-probabilities", opts.DumpRandomProbabilities, "dump randomized probabilities to file")
	cmd.Flags().StringVar(&opts.SafeMathWrappers, "safe-math-wrappers", opts.SafeMathWrappers, "comma-separated safe math wrapper IDs")
	cmd.Flags().StringVar(&opts.MonitorFuncs, "monitor-funcs", opts.MonitorFuncs, "comma-separated list of functions to monitor")
	cmd.Flags().StringVar(&opts.EnableBuiltinKinds, "enable-builtin-kinds", opts.EnableBuiltinKinds, "enable builtin kinds list")
	cmd.Flags().StringVar(&opts.DisableBuiltinKinds, "disable-builtin-kinds", opts.DisableBuiltinKinds, "disable builtin kinds list")

	addBoolPair(cmd, &negBindings, &opts.AcceptArgc, "argc", "generate argc/argv in main")
	addBoolPair(cmd, &negBindings, &opts.Arrays, "arrays", "enable arrays")
	addBoolPair(cmd, &negBindings, &opts.FixedStructFields, "fixed-struct-fields", "use fixed number of struct fields")
	cmd.Flags().BoolVar(&opts.ExpandStruct, "expand-struct", opts.ExpandStruct, "expand struct fields more aggressively")
	addBoolPair(cmd, &negBindings, &opts.Bitfields, "bitfields", "enable bitfields")
	addBoolPair(cmd, &negBindings, &opts.ComputeHash, "checksum", "enable checksum calculation")
	addBoolPair(cmd, &negBindings, &opts.CompoundAssignment, "compound-assignment", "enable compound assignment")
	addBoolPair(cmd, &negBindings, &opts.Consts, "consts", "enable const qualifiers")
	addBoolPair(cmd, &negBindings, &opts.Divs, "divs", "enable division operators")
	addBoolPair(cmd, &negBindings, &opts.Muls, "muls", "enable multiplication operators")
	addBoolPair(cmd, &negBindings, &opts.EmbeddedAssigns, "embedded-assigns", "enable embedded assignments")
	addBoolPair(cmd, &negBindings, &opts.CommaOperators, "comma-operators", "enable comma operators")
	addBoolPair(cmd, &negBindings, &opts.PreIncrOperator, "pre-incr-operator", "enable pre-increment")
	addBoolPair(cmd, &negBindings, &opts.PreDecrOperator, "pre-decr-operator", "enable pre-decrement")
	addBoolPair(cmd, &negBindings, &opts.PostIncrOperator, "post-incr-operator", "enable post-increment")
	addBoolPair(cmd, &negBindings, &opts.PostDecrOperator, "post-decr-operator", "enable post-decrement")
	addBoolPair(cmd, &negBindings, &opts.UnaryPlusOperator, "unary-plus-operator", "enable unary plus")
	addBoolPair(cmd, &negBindings, &opts.Jumps, "jumps", "enable jump statements")
	addBoolPair(cmd, &negBindings, &opts.LongLong, "longlong", "enable long long")
	addBoolPair(cmd, &negBindings, &opts.Int8, "int8", "enable int8_t")
	addBoolPair(cmd, &negBindings, &opts.UInt8, "uint8", "enable uint8_t")
	addBoolPair(cmd, &negBindings, &opts.EnableFloat, "float", "enable float")
	addBoolPair(cmd, &negBindings, &opts.Math64, "math64", "enable 64-bit math")
	addBoolPair(cmd, &negBindings, &opts.InlineFunction, "inline-function", "enable inline function attribute")
	addBoolPair(cmd, &negBindings, &opts.Pointers, "pointers", "enable pointers")
	addBoolPair(cmd, &negBindings, &opts.Structs, "structs", "enable structs")
	addBoolPair(cmd, &negBindings, &opts.ReturnStructs, "return-structs", "enable returning structs")
	addBoolPair(cmd, &negBindings, &opts.ArgStructs, "arg-structs", "enable struct arguments")
	addBoolPair(cmd, &negBindings, &opts.Unions, "unions", "enable unions")
	addBoolPair(cmd, &negBindings, &opts.ReturnUnions, "return-unions", "enable returning unions")
	addBoolPair(cmd, &negBindings, &opts.ArgUnions, "arg-unions", "enable union arguments")
	addBoolPair(cmd, &negBindings, &opts.TakeUnionFieldAddr, "take-union-field-addr", "allow taking address of union fields")
	cmd.Flags().BoolVar(&takeNoUnionFieldAddrFlag, "take-no-union-field-addr", false, "disallow taking address of union fields")
	addBoolPair(cmd, &negBindings, &opts.VolStructUnionFields, "vol-struct-union-fields", "enable volatile struct/union fields")
	addBoolPair(cmd, &negBindings, &opts.ConstStructUnionFields, "const-struct-union-fields", "enable const struct/union fields")
	addBoolPair(cmd, &negBindings, &opts.Volatiles, "volatiles", "enable volatiles")
	addBoolPair(cmd, &negBindings, &opts.VolatilePointers, "volatile-pointers", "enable volatile pointers")
	addBoolPair(cmd, &negBindings, &opts.ConstPointers, "const-pointers", "enable const pointers")
	addBoolPair(cmd, &negBindings, &opts.GlobalVariables, "global-variables", "enable global variables")
	cmd.Flags().BoolVar(&opts.AccessOnce, "enable-access-once", opts.AccessOnce, "use access_once wrappers for volatile reads")
	cmd.Flags().BoolVar(&opts.StrictVolatileRule, "strict-volatile-rule", opts.StrictVolatileRule, "enforce one volatile access per sequence region")
	addBoolPair(cmd, &negBindings, &opts.AddrTakenOfLocals, "addr-taken-of-locals", "allow address-taken local variables")
	addBoolPair(cmd, &negBindings, &opts.StrictConstArrays, "strict-const-arrays", "restrict array elements to constants")
	addBoolPair(cmd, &negBindings, &opts.DanglingGlobalPointers, "dangling-global-pointers", "reset dangling global pointers to null")
	addBoolPair(cmd, &negBindings, &opts.Builtins, "builtins", "enable compiler builtins")
	cmd.Flags().BoolVar(&opts.RandomRandom, "random-random", opts.RandomRandom, "enable randomized probability mode")
	cmd.Flags().BoolVar(&opts.BlindCheckGlobal, "check-global", opts.BlindCheckGlobal, "enable global checking")
	cmd.Flags().BoolVar(&opts.StepHashByStmt, "step-hash-by-stmt", opts.StepHashByStmt, "hash after each statement")
	cmd.Flags().BoolVar(&opts.ConstAsCondition, "const-as-condition", opts.ConstAsCondition, "allow constants in conditions")
	cmd.Flags().BoolVar(&opts.MatchExactQualifiers, "match-exact-qualifiers", opts.MatchExactQualifiers, "match exact qualifiers during selection")
	cmd.Flags().BoolVar(&opts.FreshArrayCtrlVarNames, "fresh-array-ctrl-var-names", opts.FreshArrayCtrlVarNames, "use fresh array control variable names")
	cmd.Flags().BoolVar(&opts.IdentifyWrappers, "identify-wrappers", opts.IdentifyWrappers, "annotate safe math wrappers")
	cmd.Flags().BoolVar(&opts.MarkMutableConst, "mark-mutable-const", opts.MarkMutableConst, "emit mutable const wrappers")
	cmd.Flags().BoolVar(&opts.NoReturnDeadPointer, "no-return-dead-pointer", opts.NoReturnDeadPointer, "forbid returning dead pointers")
	cmd.Flags().BoolVar(&returnDeadPointerFlag, "return-dead-pointer", false, "allow returning dead pointers")
	cmd.Flags().BoolVar(&noHashValuePrintfFlag, "no-hash-value-printf", false, "disable hash value print support")
	cmd.Flags().BoolVar(&noSignedCharIndexFlag, "no-signed-char-index", false, "disable signed char index behavior")
	addBoolPair(cmd, &negBindings, &opts.ForceGlobalsStatic, "force-globals-static", "force static storage for globals and functions")
	addBoolPair(cmd, &negBindings, &opts.ForceNonUniformArrayInit, "force-non-uniform-arrays", "force non-uniform array initializers")
	addBoolPair(cmd, &negBindings, &opts.Int128, "int128", "enable __int128 type")
	addBoolPair(cmd, &negBindings, &opts.UInt128, "uint128", "enable unsigned __int128 type")
	addBoolPair(cmd, &negBindings, &opts.BinaryConstant, "binary-constant", "enable binary constants")
	addBoolPair(cmd, &negBindings, &opts.MathNoTmp, "math-notmp", "use no-temp safe math wrappers")
	cmd.Flags().BoolVar(&opts.StrictFloat, "strict-float", opts.StrictFloat, "enforce strict floating-point semantics")
	cmd.Flags().BoolVar(&opts.DepthProtect, "depth-protect", opts.DepthProtect, "enable depth protection")
	cmd.Flags().BoolVar(&opts.CompactOutput, "compact-output", opts.CompactOutput, "emit compact output")
	cmd.Flags().BoolVar(&opts.PrefixName, "prefix-name", opts.PrefixName, "prefix generated symbol names")
	cmd.Flags().BoolVar(&opts.SequenceNamePrefix, "sequence-name-prefix", opts.SequenceNamePrefix, "prefix names based on DFS sequence")
	cmd.Flags().BoolVar(&opts.CompatibleCheck, "compatible-check", opts.CompatibleCheck, "enable compatibility checking")
	cmd.Flags().BoolVar(&opts.Klee, "klee", opts.Klee, "enable KLEE-compatible generation")
	cmd.Flags().BoolVar(&opts.Crest, "crest", opts.Crest, "enable CREST-compatible generation")
	cmd.Flags().BoolVar(&opts.CComp, "ccomp", opts.CComp, "enable CompCert-compatible generation")
	cmd.Flags().BoolVar(&opts.CoverageTest, "coverage-test", opts.CoverageTest, "enable coverage-test mode")
	cmd.Flags().BoolVar(&opts.NoDeltaReduction, "no-delta-reduction", opts.NoDeltaReduction, "disable delta reduction support")
	addBoolPair(cmd, &negBindings, &opts.FunctionAttributes, "function-attributes", "enable function attributes")
	addBoolPair(cmd, &negBindings, &opts.TypeAttributes, "type-attributes", "enable type attributes")
	addBoolPair(cmd, &negBindings, &opts.LabelAttributes, "label-attributes", "enable label attributes")
	addBoolPair(cmd, &negBindings, &opts.VariableAttributes, "variable-attributes", "enable variable attributes")
	cmd.Flags().BoolVar(&compilerAttributesFlag, "compiler-attributes", false, "enable all compiler attributes")
	cmd.Flags().BoolVar(&noCompilerAttributesFlag, "no-compiler-attributes", false, "disable all compiler attributes")
	addBoolPair(cmd, &negBindings, &opts.SafeMath, "safe-math", "emit safe math wrappers")
	addBoolPair(cmd, &negBindings, &opts.PackedStruct, "packed-struct", "enable packed structs")
	addBoolPair(cmd, &negBindings, &opts.Paranoid, "paranoid", "enable paranoid pointer checks")

	cmd.Flags().BoolVar(&opts.Concise, "concise", opts.Concise, "emit minimal comments")
	cmd.Flags().BoolVar(&opts.Quiet, "quiet", opts.Quiet, "emit fewer comments")
	cmd.Flags().BoolVar(&opts.RandomBased, "random-based", opts.RandomBased, "enable random-based generation mode")
	cmd.Flags().BoolVar(&opts.DFSExhaustive, "dfs-exhaustive", opts.DFSExhaustive, "enable DFS exhaustive generation mode")
	cmd.Flags().BoolVar(&opts.LangCPP, "lang-cpp", opts.LangCPP, "generate C++")
	cmd.Flags().BoolVar(&opts.CPP11, "cpp11", opts.CPP11, "generate C++11 (requires --lang-cpp)")
	cmd.Flags().BoolVar(&opts.FastExecution, "fast-execution", opts.FastExecution, "favor fast-running generated programs")
	cmd.Flags().BoolVar(&mainFlag, "main", false, "force generating main")
	cmd.Flags().BoolVar(&nomainFlag, "nomain", false, "disable generating main")

	_ = cmd.MarkFlagFilename("output", "c")

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		seedSet = cmd.Flags().Changed("seed")
		opts.IntSizeExplicit = cmd.Flags().Changed("int-size")
		opts.PointerExplicit = cmd.Flags().Changed("ptr-size")
		for _, b := range negBindings {
			if *b.neg {
				*b.target = false
			}
		}
	}

	return cmd
}
