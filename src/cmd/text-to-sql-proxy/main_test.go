package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name string
		flag string
	}{
		{"long flag", "--version"},
		{"short flag", "-v"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".", tc.flag)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v", err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "text-to-sql-proxy") {
				t.Errorf("expected output to contain 'text-to-sql-proxy', got %q", outputStr)
			}
			if !strings.Contains(outputStr, "dev") {
				t.Errorf("expected output to contain 'dev' (default version), got %q", outputStr)
			}
		})
	}
}

func TestUnknownProvider(t *testing.T) {
	// Set an invalid provider
	os.Setenv("TEXT_TO_SQL_PROXY_PROVIDER", "invalid-provider")
	defer os.Unsetenv("TEXT_TO_SQL_PROXY_PROVIDER")

	cmd := exec.Command("go", "run", ".")
	output, err := cmd.CombinedOutput()

	// Should exit with error
	if err == nil {
		t.Fatal("expected command to fail with unknown provider")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Unknown provider") {
		t.Errorf("expected error about unknown provider, got %q", outputStr)
	}
	if !strings.Contains(outputStr, "invalid-provider") {
		t.Errorf("expected error to mention 'invalid-provider', got %q", outputStr)
	}
}
