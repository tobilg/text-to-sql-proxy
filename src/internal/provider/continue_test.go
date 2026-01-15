package provider

import (
	"testing"
)

func TestParseContinueResponse_StructuredJSON(t *testing.T) {
	// Standard Continue CLI response format
	input := `{"response":"SELECT * FROM users","status":"success","note":"Response was not valid JSON, so it was wrapped in a JSON object"}`

	sql, err := parseContinueResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseContinueResponse_SuccessStatus(t *testing.T) {
	input := `{"response":"SELECT COUNT(*) FROM orders WHERE status = 'completed'","status":"success"}`

	sql, err := parseContinueResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT COUNT(*) FROM orders WHERE status = 'completed'"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseContinueResponse_WithMarkdownCodeBlock(t *testing.T) {
	input := `{"response":"` + "```sql\\nSELECT * FROM users\\n```" + `","status":"success"}`

	sql, err := parseContinueResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseContinueResponse_RawSQL(t *testing.T) {
	// Fallback to raw content
	input := `SELECT * FROM users`

	sql, err := parseContinueResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseContinueResponse_EmptyResponse(t *testing.T) {
	input := ``

	_, err := parseContinueResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseContinueResponse_EmptyResponseField(t *testing.T) {
	input := `{"response":"","status":"success"}`

	sql, err := parseContinueResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Falls back to trimmed raw content
	if sql != `{"response":"","status":"success"}` {
		t.Errorf("unexpected fallback: %q", sql)
	}
}

func TestParseContinueResponse_WhitespaceOnly(t *testing.T) {
	input := `   `

	_, err := parseContinueResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseContinueResponse_ComplexSQL(t *testing.T) {
	input := `{"response":"SELECT u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING COUNT(o.id) > 5","status":"success"}`

	sql, err := parseContinueResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING COUNT(o.id) > 5"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}
