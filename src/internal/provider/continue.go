package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

const continuePromptTemplate = `You are a DuckDB expert. Generate ONLY a raw SQL query with no markdown, no explanations, no code blocks.

DDL: %s
Question: %s

Respond with ONLY the SQL query.`

// ContinueClient implements SQLGenerator using the Continue CLI (cn).
type ContinueClient struct{}

// NewContinueClient creates a new Continue CLI client.
func NewContinueClient() *ContinueClient {
	return &ContinueClient{}
}

// GenerateSQL calls the Continue CLI to generate SQL from DDL and a question.
func (c *ContinueClient) GenerateSQL(ddl, question string) (string, error) {
	prompt := FormatPrompt(continuePromptTemplate, ddl, question)

	cmd := exec.Command("cn",
		"-p", prompt,
		"--format", "json",
		"--silent",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Join(ErrCLIExecution, errors.New(stderr.String()))
	}

	sql, err := parseContinueResponse(stdout.Bytes())
	if err != nil {
		return "", err
	}

	return sql, nil
}

// continueResponse represents the JSON response from Continue CLI.
type continueResponse struct {
	Response string `json:"response"`
	Status   string `json:"status"`
}

// parseContinueResponse extracts the SQL from Continue CLI's JSON response.
func parseContinueResponse(data []byte) (string, error) {
	// Continue CLI wraps plain text responses in:
	// {"response": "...", "status": "success", "note": "..."}

	var response continueResponse
	if err := json.Unmarshal(data, &response); err == nil && response.Response != "" {
		return CleanSQL(response.Response), nil
	}

	// Fallback: try to extract raw content if JSON parsing fails
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "" {
		return CleanSQL(trimmed), nil
	}

	return "", ErrParsing
}
