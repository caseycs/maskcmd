package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := buildCmdMask()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildCmdMask() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "maskcmd [--flags] [command] [args...]",
		Short:   "Run a command while masking specified sensitive data from its stdout and stderr",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: cmdMaskPreRun,
		RunE:    cmdMask,
	}

	cmd.Flags().String("secrets-dir", "", "Path to directory containing files with secret values")
	cmd.Flags().String("env-vars", "", "Mask values of certain environment variables")
	cmd.Flags().Bool("all-env-vars", false, "Mask all environment variables values")
	return cmd
}
