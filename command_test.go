package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMaskLine(t *testing.T) {
	masks := []string{"mypassword", "mytoken"}
	tests := []struct {
		input    string
		expected string
	}{
		{"Password is mypassword", "Password is *****"},
		{"Error token is mytoken", "Error token is *****"},
		{"No secret here", "No secret here"},
	}

	for _, test := range tests {
		masked := maskLine(test.input, masks)
		if masked != test.expected {
			t.Errorf("Expected: %q, Got: %q", test.expected, masked)
		}
	}
}

// Helper function to create a temporary secret file
func createTempSecretFile(dir, filename, content string) error {
	filePath := filepath.Join(dir, filename)
	return os.WriteFile(filePath, []byte(content), 0600)
}

func TestReadSecretsFromDir(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "secrets_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Create secret files
	createTempSecretFile(tempDir, "password.txt", "mypassword")
	createTempSecretFile(tempDir, "token.txt", "mytoken")

	// Read secrets
	secrets, err := readSecretsFromDir(tempDir)
	if err != nil {
		t.Fatalf("Error reading secrets: %v", err)
	}

	// Expected secrets
	expected := []string{"mypassword", "mytoken"}
	for _, exp := range expected {
		found := false
		for _, sec := range secrets {
			if sec == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected secret %q not found in result", exp)
		}
	}
}

func TestCmdMask_MaskAllEnvVars(t *testing.T) {
	os.Setenv("MYPASSWORD", "mypassword")
	defer os.Unsetenv("MYPASSWORD")

	output, _, err := executeCommand(buildCmdMask(), "--all-env-vars", "--", "echo", "Password is mypassword")
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	expected := "Password is *****"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestCmdMask_MaskCertainEnvs(t *testing.T) {
	tests := map[string]struct {
		envVars        map[string]string
		echoString     string
		expectedStdOut string
		expectedStdErr string
	}{
		"simplest case": {
			map[string]string{"MYP1": "p1", "MYP2": "p2"},
			"Passwords are p1 and p2",
			"Passwords are ***** and *****",
			""},
		"overlapping secrets": {
			map[string]string{"MYP1": "p1", "MYP2": "p1p2"},
			"Password is p1p2",
			"Password is *****",
			"Warning: overlapping secrets detected: p**2 and p1"},
	}

	for _, test := range tests {
		for k, v := range test.envVars {
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}

		envKeys := make([]string, 0, len(test.envVars))
		for k := range test.envVars {
			envKeys = append(envKeys, k)
		}
		joinedKeys := strings.Join(envKeys, ",")
		stdout, stderr, err := executeCommand(buildCmdMask(), "--env-vars", joinedKeys, "--", "echo", test.echoString)

		if err != nil {
			t.Fatalf("Error executing command: %v", err)
		}
		if strings.TrimSpace(stdout) != test.expectedStdOut {
			t.Errorf("Expected stdout %q, got %q", test.expectedStdOut, stdout)
		}
		if strings.TrimSpace(stderr) != test.expectedStdErr {
			t.Errorf("Expected stderr %q, got %q", test.expectedStdErr, stderr)
		}
	}
}

func TestCmdMask_MaskSecretsFromDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "secrets_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Create secret files
	createTempSecretFile(tempDir, "password.txt", "mypassword")

	output, _, err := executeCommand(buildCmdMask(), "--secrets-dir", tempDir, "--", "echo", "Password is mypassword")
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	expected := "Password is *****"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestCmdMask_CustomExitCode(t *testing.T) {
	os.Setenv("SECRET", "mysecret")
	defer os.Unsetenv("SECRET")

	_, _, err := executeCommand(buildCmdMask(), "--env-vars", "SECRET", "--", "sh", "-c", "exit 5")
	if err == nil {
		t.Fatalf("Error expected")
	}

	e, ok := err.(*ExitCodeError)
	if !ok {
		t.Fatalf("Expected ExitCodeError")
	}

	if e.Code != 5 {
		t.Fatalf("Expected exit code 5, got %d", e.Code)
	}
}

func executeCommand(cmd *cobra.Command, args ...string) (string, string, error) {
	bufOut := new(bytes.Buffer)
	cmd.SetOut(bufOut)

	bufErr := new(bytes.Buffer)
	cmd.SetErr(bufErr)

	cmd.SetArgs(args)

	err := cmd.Execute()
	return bufOut.String(), bufErr.String(), err
}
