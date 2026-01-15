package provider

import (
	"testing"
)

func TestParseOpenCodeResponse_TextEvent(t *testing.T) {
	// Standard OpenCode NDJSON response with text event
	input := `{"type":"step_start","timestamp":1234567890,"sessionID":"abc123"}
{"type":"text","timestamp":1234567891,"sessionID":"abc123","content":"SELECT * FROM users"}
{"type":"step_finish","timestamp":1234567892,"sessionID":"abc123"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_MultipleTextEvents(t *testing.T) {
	// Should return the last text event
	input := `{"type":"text","timestamp":1234567890,"sessionID":"abc123","content":"Thinking..."}
{"type":"text","timestamp":1234567891,"sessionID":"abc123","content":"SELECT COUNT(*) FROM orders"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT COUNT(*) FROM orders"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_WithToolUse(t *testing.T) {
	// Should ignore tool_use events and only extract text
	input := `{"type":"step_start","timestamp":1234567890,"sessionID":"abc123"}
{"type":"tool_use","timestamp":1234567891,"sessionID":"abc123","tool":"read_file"}
{"type":"text","timestamp":1234567892,"sessionID":"abc123","content":"SELECT * FROM products WHERE price > 100"}
{"type":"step_finish","timestamp":1234567893,"sessionID":"abc123"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM products WHERE price > 100"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_WithMarkdownCodeBlock(t *testing.T) {
	input := `{"type":"text","timestamp":1234567890,"sessionID":"abc123","content":"` + "```sql\\nSELECT * FROM users\\n```" + `"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_RawSQL(t *testing.T) {
	// Fallback to raw content
	input := `SELECT * FROM users`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_EmptyResponse(t *testing.T) {
	input := ``

	_, err := parseOpenCodeResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseOpenCodeResponse_EmptyLines(t *testing.T) {
	// Should handle empty lines in NDJSON
	input := `{"type":"step_start","timestamp":1234567890}

{"type":"text","timestamp":1234567891,"content":"SELECT 1"}
`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT 1"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_InvalidJSON(t *testing.T) {
	// Should skip invalid JSON lines and continue
	input := `not valid json
{"type":"text","timestamp":1234567890,"content":"SELECT * FROM users"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_ErrorEvent(t *testing.T) {
	// Should still extract text even with error events
	input := `{"type":"text","timestamp":1234567890,"content":"SELECT * FROM users"}
{"type":"error","timestamp":1234567891,"message":"Some warning"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseOpenCodeResponse_ComplexSQL(t *testing.T) {
	input := `{"type":"text","timestamp":1234567890,"sessionID":"abc123","content":"SELECT u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING COUNT(o.id) > 5"}`

	sql, err := parseOpenCodeResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING COUNT(o.id) > 5"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}
