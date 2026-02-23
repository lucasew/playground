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

			if !seedSet {
				opts.Seed = uint64(time.Now().UnixNano())
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
	cmd.Flags().IntVar(&opts.MaxUnionFields, "max-union-fields", opts.MaxUnionFields, "limit union field count")
	cmd.Flags().IntVar(&opts.MaxPointerDepth, "max-pointer-depth", opts.MaxPointerDepth, "limit pointer indirection depth")
	cmd.Flags().IntVar(&opts.MaxArrayDim, "max-array-dim", opts.MaxArrayDim, "limit array dimensions")
	cmd.Flags().IntVar(&opts.MaxArrayLenPerDim, "max-array-len-per-dim", opts.MaxArrayLenPerDim, "limit array length per dimension")
	cmd.Flags().IntVar(&opts.MaxExhaustiveDepth, "max-exhaustive-depth", opts.MaxExhaustiveDepth, "maximum exhaustive depth")
	cmd.Flags().IntVar(&opts.InlineFunctionProb, "inline-function-prob", opts.InlineFunctionProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.BuiltinFunctionProb, "builtin-function-prob", opts.BuiltinFunctionProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.ArrayOOBProb, "array-oob-prob", opts.ArrayOOBProb, "probability [0,100]")
	cmd.Flags().IntVar(&opts.MaxGlobals, "max-globals", opts.MaxGlobals, "maximum number of generated globals")

	addBoolPair(cmd, &negBindings, &opts.AcceptArgc, "argc", "generate argc/argv in main")
	addBoolPair(cmd, &negBindings, &opts.Arrays, "arrays", "enable arrays")
	addBoolPair(cmd, &negBindings, &opts.Bitfields, "bitfields", "enable bitfields")
	addBoolPair(cmd, &negBindings, &opts.ComputeHash, "checksum", "enable checksum calculation")
	addBoolPair(cmd, &negBindings, &opts.CompoundAssignment, "compound-assignment", "enable compound assignment")
	addBoolPair(cmd, &negBindings, &opts.Consts, "consts", "enable const qualifiers")
	addBoolPair(cmd, &negBindings, &opts.Divs, "divs", "enable division operators")
	addBoolPair(cmd, &negBindings, &opts.EmbeddedAssigns, "embedded-assigns", "enable embedded assignments")
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
	addBoolPair(cmd, &negBindings, &opts.Unions, "unions", "enable unions")
	addBoolPair(cmd, &negBindings, &opts.Volatiles, "volatiles", "enable volatiles")
	addBoolPair(cmd, &negBindings, &opts.VolatilePointers, "volatile-pointers", "enable volatile pointers")
	addBoolPair(cmd, &negBindings, &opts.ConstPointers, "const-pointers", "enable const pointers")
	addBoolPair(cmd, &negBindings, &opts.GlobalVariables, "global-variables", "enable global variables")
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
		for _, b := range negBindings {
			if *b.neg {
				*b.target = false
			}
		}
	}

	return cmd
}
