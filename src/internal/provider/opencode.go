package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

const opencodePromptTemplate = `You are a %s expert. Generate ONLY a raw SQL query with no markdown, no explanations, no code blocks. Format the SQL nicely with 2-space indentation.

DDL: %s
Question: %s

Respond with ONLY the SQL query.`

// OpenCodeClient implements SQLGenerator using the OpenCode CLI.
type OpenCodeClient struct {
	database string
}

// NewOpenCodeClient creates a new OpenCode CLI client.
func NewOpenCodeClient(database string) *OpenCodeClient {
	return &OpenCodeClient{database: database}
}

// GenerateSQL calls the OpenCode CLI to generate SQL from DDL and a question.
func (c *OpenCodeClient) GenerateSQL(ddl, question string) (string, error) {
	prompt := FormatPrompt(opencodePromptTemplate, c.database, ddl, question)

	cmd := exec.Command("opencode", "run",
		prompt,
		"--format", "json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Join(ErrCLIExecution, errors.New(stderr.String()))
	}

	sql, err := parseOpenCodeResponse(stdout.Bytes())
	if err != nil {
		return "", err
	}

	return sql, nil
}

// opencodeEvent represents a single NDJSON event from OpenCode.
type opencodeEvent struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	SessionID string `json:"sessionID,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// parseOpenCodeResponse extracts the SQL from OpenCode's NDJSON response.
func parseOpenCodeResponse(data []byte) (string, error) {
	// OpenCode outputs NDJSON - one JSON object per line
	// We look for "text" type events which contain the model output

	var lastContent string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event opencodeEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Check for text events containing model output
		if event.Type == "text" && event.Content != "" {
			lastContent = event.Content
		}
	}

	if lastContent != "" {
		return CleanSQL(lastContent), nil
	}

	// Fallback: try the whole output as raw text
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "" {
		return CleanSQL(trimmed), nil
	}

	return "", ErrParsing
}
