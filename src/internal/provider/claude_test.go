package provider

import (
	"testing"
)

func TestParseClaudeResponse_StructuredJSON(t *testing.T) {
	input := `{"structured_output":{"sql":"SELECT * FROM users"},"type":"result"}`

	sql, err := parseClaudeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseClaudeResponse_RawSQL(t *testing.T) {
	input := `SELECT * FROM users WHERE id = 1`

	sql, err := parseClaudeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sql != input {
		t.Errorf("expected %q, got %q", input, sql)
	}
}

func TestParseClaudeResponse_EmptyResponse(t *testing.T) {
	input := ``

	_, err := parseClaudeResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseClaudeResponse_WhitespaceOnly(t *testing.T) {
	input := `   `

	_, err := parseClaudeResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseClaudeResponse_NestedJSON(t *testing.T) {
	input := `{"structured_output":{"sql":"SELECT json_extract(data, '$.name') FROM users"}}`

	sql, err := parseClaudeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT json_extract(data, '$.name') FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseClaudeResponse_EmptySQL(t *testing.T) {
	input := `{"structured_output":{"sql":""}}`

	sql, err := parseClaudeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Falls back to trimmed raw content
	if sql != `{"structured_output":{"sql":""}}` {
		t.Errorf("unexpected fallback: %q", sql)
	}
}

func TestParseClaudeResponse_FullResponse(t *testing.T) {
	// Test with a more complete response like the actual CLI returns
	input := `{"type":"result","subtype":"success","structured_output":{"sql":"SELECT * FROM users WHERE name LIKE 'A%';"},"session_id":"abc123"}`

	sql, err := parseClaudeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users WHERE name LIKE 'A%';"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}
