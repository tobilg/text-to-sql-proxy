package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

const geminiPromptTemplate = `You are a DuckDB expert. Generate ONLY a raw SQL query with no markdown, no explanations, no code blocks. Format the SQL nicely with 2-space indentation.

DDL: %s
Question: %s

Respond with ONLY the SQL query.`

// GeminiClient implements SQLGenerator using the Gemini CLI.
type GeminiClient struct{}

// NewGeminiClient creates a new Gemini CLI client.
func NewGeminiClient() *GeminiClient {
	return &GeminiClient{}
}

// GenerateSQL calls the Gemini CLI to generate SQL from DDL and a question.
func (g *GeminiClient) GenerateSQL(ddl, question string) (string, error) {
	prompt := FormatPrompt(geminiPromptTemplate, ddl, question)

	cmd := exec.Command("gemini",
		"-p", prompt,
		"--output-format", "json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Join(ErrCLIExecution, errors.New(stderr.String()))
	}

	sql, err := parseGeminiResponse(stdout.Bytes())
	if err != nil {
		return "", err
	}

	return sql, nil
}

// parseGeminiResponse extracts the SQL from Gemini's JSON response.
func parseGeminiResponse(data []byte) (string, error) {
	// Gemini returns: {"response": "...", ...}
	var response struct {
		Response string `json:"response"`
	}

	if err := json.Unmarshal(data, &response); err == nil && response.Response != "" {
		return CleanSQL(response.Response), nil
	}

	// Fallback: try to extract raw content if structured parsing fails
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "" {
		return CleanSQL(trimmed), nil
	}

	return "", ErrParsing
}
