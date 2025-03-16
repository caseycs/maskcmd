package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := buildCmdMask()

	if err := cmd.Execute(); err != nil {
		var exitCode int
		if e, ok := err.(*ExitCodeError); ok {
			exitCode = e.Code // Extract exit code
		} else {
			exitCode = 1 // Default error code
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCode)
	}
}

func buildCmdMask() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "maskcmd [--flags] [command] [args...]",
		Short:   "Run a command while masking sensitive data from its stdout and stderr",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: cmdMaskPreRun,
		RunE:    cmdMask,
	}

	cmd.Flags().String("secrets-dir", "", "Treat files content in certain directory as secrets")
	cmd.Flags().String("env-vars", "", "Mark certain environment variables as secrets")
	cmd.Flags().Bool("all-env-vars", false, "Treat all environment variables values as secrets")
	return cmd
}

// ExitCodeError is a custom error type with an exit code
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("exit code: %d", e.Code)
}
