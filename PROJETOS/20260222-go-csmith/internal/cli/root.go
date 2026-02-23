package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"csmith/pkg/csmith"
)

const (
	appName    = "csmith-go"
	appVersion = "0.1.0"
)

type forcedBoolValue struct {
	target *bool
	value  bool
}

func (v *forcedBoolValue) Set(_ string) error {
	*v.target = v.value
	return nil
}

func (v *forcedBoolValue) String() string {
	return strconv.FormatBool(*v.target)
}

func (v *forcedBoolValue) Type() string {
	return "bool"
}

func (v *forcedBoolValue) IsBoolFlag() bool {
	return true
}

func addBoolPair(cmd *cobra.Command, target *bool, name string, usage string) {
	cmd.Flags().BoolVar(target, name, *target, usage)
	cmd.Flags().Var(&forcedBoolValue{target: target, value: false}, "no-"+name, "disable "+name)
}

func NewRootCmd() *cobra.Command {
	opts := csmith.Defaults()
	seedSet := false
	outputPath := ""
	showVersion := false
	mainFlag := false
	nomainFlag := false

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

	addBoolPair(cmd, &opts.AcceptArgc, "argc", "generate argc/argv in main")
	addBoolPair(cmd, &opts.Arrays, "arrays", "enable arrays")
	addBoolPair(cmd, &opts.Bitfields, "bitfields", "enable bitfields")
	addBoolPair(cmd, &opts.ComputeHash, "checksum", "enable checksum calculation")
	addBoolPair(cmd, &opts.CompoundAssignment, "compound-assignment", "enable compound assignment")
	addBoolPair(cmd, &opts.Consts, "consts", "enable const qualifiers")
	addBoolPair(cmd, &opts.Divs, "divs", "enable division operators")
	addBoolPair(cmd, &opts.EmbeddedAssigns, "embedded-assigns", "enable embedded assignments")
	addBoolPair(cmd, &opts.PreIncrOperator, "pre-incr-operator", "enable pre-increment")
	addBoolPair(cmd, &opts.PreDecrOperator, "pre-decr-operator", "enable pre-decrement")
	addBoolPair(cmd, &opts.PostIncrOperator, "post-incr-operator", "enable post-increment")
	addBoolPair(cmd, &opts.PostDecrOperator, "post-decr-operator", "enable post-decrement")
	addBoolPair(cmd, &opts.UnaryPlusOperator, "unary-plus-operator", "enable unary plus")
	addBoolPair(cmd, &opts.Jumps, "jumps", "enable jump statements")
	addBoolPair(cmd, &opts.LongLong, "longlong", "enable long long")
	addBoolPair(cmd, &opts.Int8, "int8", "enable int8_t")
	addBoolPair(cmd, &opts.UInt8, "uint8", "enable uint8_t")
	addBoolPair(cmd, &opts.EnableFloat, "float", "enable float")
	addBoolPair(cmd, &opts.Math64, "math64", "enable 64-bit math")
	addBoolPair(cmd, &opts.InlineFunction, "inline-function", "enable inline function attribute")
	addBoolPair(cmd, &opts.Pointers, "pointers", "enable pointers")
	addBoolPair(cmd, &opts.Structs, "structs", "enable structs")
	addBoolPair(cmd, &opts.Unions, "unions", "enable unions")
	addBoolPair(cmd, &opts.Volatiles, "volatiles", "enable volatiles")
	addBoolPair(cmd, &opts.VolatilePointers, "volatile-pointers", "enable volatile pointers")
	addBoolPair(cmd, &opts.ConstPointers, "const-pointers", "enable const pointers")
	addBoolPair(cmd, &opts.GlobalVariables, "global-variables", "enable global variables")
	addBoolPair(cmd, &opts.SafeMath, "safe-math", "emit safe math wrappers")
	addBoolPair(cmd, &opts.PackedStruct, "packed-struct", "enable packed structs")
	addBoolPair(cmd, &opts.Paranoid, "paranoid", "enable paranoid pointer checks")

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
	}

	return cmd
}
