package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

const maxSecretLength = 4096

func cmdMaskPreRun(cmd *cobra.Command, args []string) error {
	// Ensure at least one flag is set
	if !cmd.Flags().Changed("secrets-dir") && !cmd.Flags().Changed("all-env-vars") && !cmd.Flags().Changed("env-vars") {
		return errors.New("at least one flag is required")
	}
	return nil
}

func cmdMask(cmd *cobra.Command, args []string) error {
	// do not show hind after initial validation
	cmd.SilenceUsage = true

	// Collect secrets to mask
	var masks []string

	// Read secrets from files in directory if provided
	if secretsDir, _ := cmd.Flags().GetString("secrets-dir"); secretsDir != "" {
		secretsFromFiles, err := readSecretsFromDir(secretsDir)
		if err != nil {
			return fmt.Errorf("error reading secrets from directory: %v", err)
		}
		masks = append(masks, secretsFromFiles...)
	}

	// Treat all environment variables values as secrets
	if allEnv, _ := cmd.Flags().GetBool("all-env-vars"); allEnv != false {
		masks = append(masks, collectAllEnvValues()...)
	}

	// Treat certain environment variables values as secrets
	if env, _ := cmd.Flags().GetString("env-vars"); env != "" {
		masks = append(masks, collectEnvValues(env)...)
	}

	if len(masks) == 0 {
		return fmt.Errorf("no secrets defined")
	}

	// Sort masks by string length (longest first) - to tacke overlapping secrets
	sort.Slice(masks, func(i, j int) bool {
		return len(masks[i]) > len(masks[j])
	})

	// check for overlapping secrets
	checkForOverlappingSecrets(cmd, masks)

	// Run external command
	cmdToRun := exec.Command(args[0], args[1:]...)

	// Capture stdout and stderr
	stdout, err := cmdToRun.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error getting stdout: %v", err)
	}
	stderr, err := cmdToRun.StderrPipe()
	if err != nil {
		return fmt.Errorf("error getting stderr: %v", err)
	}

	// Start the command
	if err := cmdToRun.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	// Process output in real-time
	var wg sync.WaitGroup
	wg.Add(2)
	go streamOutput(stdout, masks, cmd.OutOrStdout(), &wg)
	go streamOutput(stderr, masks, cmd.OutOrStderr(), &wg)

	wg.Wait()

	// Wait for the command to complete and capture its exit code
	if err := cmdToRun.Wait(); err != nil {
		if cmdExit, ok := err.(*exec.ExitError); ok {
			return &ExitCodeError{Code: cmdExit.ExitCode()} // Return the same exit code as the original command
		}
		return fmt.Errorf("command execution failed: %v", err)
	}

	return nil
}

func checkForOverlappingSecrets(cmd *cobra.Command, masks []string) {
	var ignoreMask []int
	for i := 0; i < len(masks); i++ {
		if slices.Contains(ignoreMask, i) {
			continue
		}
		for j := i + 1; j < len(masks); j++ {
			if slices.Contains(ignoreMask, j) {
				continue
			}
			if strings.Contains(masks[i], masks[j]) || strings.Contains(masks[j], masks[i]) {
				m1 := masks[i]
				m2 := masks[j]
				if len(m1) > 2 {
					m1 = fmt.Sprintf("%c%s%c", m1[0], strings.Repeat("*", len(m1)-2), m1[len(m1)-1])
				}
				if len(m2) > 2 {
					m2 = fmt.Sprintf("%c%s%c", m2[0], strings.Repeat("*", len(m2)-2), m2[len(m2)-1])
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: overlapping secrets detected: %s and %s\n", m1, m2)
				ignoreMask = append(ignoreMask, i, j)
			}
		}
	}
}

func collectEnvValues(envVars string) []string {
	var secrets []string

	envVarsList := strings.Split(envVars, ",")
	for _, envVar := range envVarsList {
		envVar = strings.TrimSpace(envVar)
		if envVar == "" {
			continue
		}
		value := os.Getenv(envVar)
		if value != "" {
			secrets = append(secrets, value)
		}
	}
	return secrets
}

func collectAllEnvValues() []string {
	var secrets []string
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 && parts[1] != "" {
			secrets = append(secrets, parts[1])
		}
	}
	return secrets
}

// maskLine replaces sensitive substrings in a given line.
func maskLine(line string, masks []string) string {
	for _, mask := range masks {
		if len(mask) > len(line) {
			continue
		}
		line = strings.ReplaceAll(line, mask, "*****")
	}
	return line
}

// streamOutput reads from the reader, masks sensitive data, and writes to the output writer.
func streamOutput(reader io.ReadCloser, masks []string, writer io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		maskedLine := maskLine(scanner.Text(), masks)
		fmt.Fprintln(writer, maskedLine)
	}
}

// readSecretsFromDir recursively reads all files in a directory and returns their contents as a slice of strings.
func readSecretsFromDir(dirPath string) ([]string, error) {
	var secrets []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Resolve symlinks
		resolvedInfo, err := os.Stat(path)
		if err != nil {
			return err // skip broken symlinks or inaccessible files
		}

		if !resolvedInfo.Mode().IsRegular() {
			return nil
		}

		if info.Size() > maxSecretLength {
			return fmt.Errorf("secret file %s is too large (above %dkb)", path, maxSecretLength/1024)
		}

		secret, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		secrets = append(secrets, strings.TrimSpace(string(secret)))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return secrets, nil
}
