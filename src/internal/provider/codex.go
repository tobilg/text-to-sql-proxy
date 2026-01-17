package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

const codexPromptTemplate = `You are a %s expert. Generate ONLY a raw SQL query with no markdown, no explanations, no code blocks. Format the SQL nicely with 2-space indentation.

DDL: %s
Question: %s

Respond with ONLY the SQL query.`

// CodexClient implements SQLGenerator using the Codex CLI.
type CodexClient struct {
	database string
}

// NewCodexClient creates a new Codex CLI client.
func NewCodexClient(database string) *CodexClient {
	return &CodexClient{database: database}
}

// GenerateSQL calls the Codex CLI to generate SQL from DDL and a question.
func (c *CodexClient) GenerateSQL(ddl, question string) (string, error) {
	prompt := FormatPrompt(codexPromptTemplate, c.database, ddl, question)

	cmd := exec.Command("codex", "exec",
		prompt,
		"--json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Join(ErrCLIExecution, errors.New(stderr.String()))
	}

	sql, err := parseCodexResponse(stdout.Bytes())
	if err != nil {
		return "", err
	}

	return sql, nil
}

// codexEvent represents a single NDJSON event from Codex.
type codexEvent struct {
	Type    string `json:"type"`
	Message *struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message,omitempty"`
	Item *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item,omitempty"`
	Response string `json:"response,omitempty"`
}

// parseCodexResponse extracts the SQL from Codex's NDJSON response.
func parseCodexResponse(data []byte) (string, error) {
	// Codex outputs NDJSON - one JSON object per line
	// We need to find the last assistant message or response

	var lastContent string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Check for item.completed events with agent_message type (actual Codex CLI format)
		if event.Type == "item.completed" && event.Item != nil && event.Item.Type == "agent_message" && event.Item.Text != "" {
			lastContent = event.Item.Text
		}

		// Check for message events with assistant role (alternative format)
		if event.Message != nil && event.Message.Role == "assistant" && event.Message.Content != "" {
			lastContent = event.Message.Content
		}

		// Check for direct response field
		if event.Response != "" {
			lastContent = event.Response
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
