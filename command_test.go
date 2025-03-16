package main

import (
	"bytes"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestMaskCommand_MaskingAllEnvs(t *testing.T) {
	os.Setenv("MYPASSWORD", "mypassword")
	defer os.Unsetenv("MYPASSWORD")

	output, err := executeCommand(buildCmdMask(), "--all-env-vars", "--", "echo", "Password is mypassword")
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	expected := "Password is *****"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestMaskCommand_MaskingCertainEnvs(t *testing.T) {
	os.Setenv("MYP1", "p1")
	os.Setenv("MYP2", "p2")
	defer os.Unsetenv("MYP1")
	defer os.Unsetenv("MYP2")

	output, err := executeCommand(buildCmdMask(), "--env-vars", "MYP1,MYP2", "--", "echo", "Passwords are p1 and p2")
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	expected := "Passwords are ***** and *****"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestMaskCommand_MaskingDir(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "secrets_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Create secret files
	createTempSecretFile(tempDir, "password.txt", "mypassword")

	output, err := executeCommand(buildCmdMask(), "--secrets-dir", tempDir, "--", "echo", "Password is mypassword")
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	expected := "Password is *****"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}
