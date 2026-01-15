package provider

import (
	"testing"
)

func TestParseGeminiResponse_StructuredJSON(t *testing.T) {
	input := `{"response":"SELECT * FROM users"}`

	sql, err := parseGeminiResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseGeminiResponse_WithMarkdownCodeBlock(t *testing.T) {
	input := `{"response":"` + "```sql\\nSELECT * FROM users\\n```" + `"}`

	sql, err := parseGeminiResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseGeminiResponse_RawSQL(t *testing.T) {
	input := `SELECT * FROM users WHERE id = 1`

	sql, err := parseGeminiResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SELECT * FROM users WHERE id = 1"
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
}

func TestParseGeminiResponse_EmptyResponse(t *testing.T) {
	input := ``

	_, err := parseGeminiResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}

func TestParseGeminiResponse_WhitespaceOnly(t *testing.T) {
	input := `   `

	_, err := parseGeminiResponse([]byte(input))
	if err != ErrParsing {
		t.Errorf("expected ErrParsing, got %v", err)
	}
}
