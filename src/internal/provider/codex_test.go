package provider

import (
	"testing"
)

func TestParseCodexResponse_ItemCompleted(t *testing.T) {
	// Actual Codex CLI format with item.completed events
	input := `{"type":"thread.started","thread_id":"thread_abc123"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_0","type":"agent_message","text":"SELECT SUM(total) AS total_sales FROM orders;"}}
{"type":"turn.completed","usage":{"input_tokens":100,"output_tokens":20}}`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT SUM(total) AS total_sales FROM orders;"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_AssistantMessage(t *testing.T) {
	// NDJSON with assistant message (alternative format)
	input := `{"type":"init","session_id":"123"}
{"type":"message","message":{"role":"user","content":"Generate SQL"}}
{"type":"message","message":{"role":"assistant","content":"SELECT * FROM users"}}`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_ResponseField(t *testing.T) {
	// NDJSON with response field
	input := `{"type":"init"}
{"type":"result","response":"SELECT COUNT(*) FROM orders"}`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT COUNT(*) FROM orders"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_MultipleAssistantMessages(t *testing.T) {
	// Should return the last assistant message
	input := `{"type":"message","message":{"role":"assistant","content":"Thinking..."}}
{"type":"message","message":{"role":"assistant","content":"SELECT * FROM users WHERE active = true"}}`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users WHERE active = true"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_WithMarkdownCodeBlock(t *testing.T) {
	input := `{"type":"message","message":{"role":"assistant","content":"` + "```sql\\nSELECT * FROM users\\n```" + `"}}`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_RawSQL(t *testing.T) {
	// Fallback to raw content
	input := `SELECT * FROM users`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_EmptyResponse(t *testing.T) {
	input := ``

	_, err := parseCodexResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseCodexResponse_EmptyLines(t *testing.T) {
	// Should handle empty lines in NDJSON
	input := `{"type":"init"}

{"type":"message","message":{"role":"assistant","content":"SELECT 1"}}
`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT 1"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseCodexResponse_InvalidJSON(t *testing.T) {
	// Should skip invalid JSON lines and continue
	input := `not valid json
{"type":"message","message":{"role":"assistant","content":"SELECT * FROM users"}}`

	sql, err := parseCodexResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}
