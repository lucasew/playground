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

func NewRootCmd() *cobra.Command {
	opts := csmith.Defaults()
	seedSet := false
	outputPath := ""
	showVersion := false

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

			if !seedSet {
				opts.Seed = uint64(time.Now().UnixNano())
			}

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

	_ = cmd.MarkFlagFilename("output", "c")

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		seedSet = cmd.Flags().Changed("seed")
	}

	return cmd
}
