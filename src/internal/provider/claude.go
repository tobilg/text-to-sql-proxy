package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

const (
	claudeSystemPrompt = "You are a DuckDB expert. Generate ONLY raw SQL queries. No markdown, no explanations."
	claudeJSONSchema   = `{"type":"object","properties":{"sql":{"type":"string"}},"required":["sql"]}`
)

// ClaudeClient implements SQLGenerator using the Claude CLI.
type ClaudeClient struct{}

// NewClaudeClient creates a new Claude CLI client.
func NewClaudeClient() *ClaudeClient {
	return &ClaudeClient{}
}

// GenerateSQL calls the Claude CLI to generate SQL from DDL and a question.
func (c *ClaudeClient) GenerateSQL(ddl, question string) (string, error) {
	userPrompt := "DDL: " + ddl + "\nQuestion: " + question

	cmd := exec.Command("claude",
		"-p", userPrompt,
		"--append-system-prompt", claudeSystemPrompt,
		"--output-format", "json",
		"--json-schema", claudeJSONSchema,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Join(ErrCLIExecution, errors.New(stderr.String()))
	}

	sql, err := parseClaudeResponse(stdout.Bytes())
	if err != nil {
		return "", err
	}

	return sql, nil
}

// parseClaudeResponse extracts the SQL from Claude's JSON response.
func parseClaudeResponse(data []byte) (string, error) {
	// Claude returns: {"structured_output": {"sql": "..."}, ...}
	var response struct {
		StructuredOutput struct {
			SQL string `json:"sql"`
		} `json:"structured_output"`
	}

	if err := json.Unmarshal(data, &response); err == nil && response.StructuredOutput.SQL != "" {
		return response.StructuredOutput.SQL, nil
	}

	// Fallback: try to extract raw content if structured parsing fails
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "" {
		return trimmed, nil
	}

	return "", ErrParsing
}
