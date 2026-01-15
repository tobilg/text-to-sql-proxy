package provider

import (
	"testing"
)

func TestCleanSQL_RemovesMarkdownCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "sql code block",
			input:    "```sql\nSELECT * FROM users\n```",
			expected: "SELECT * FROM users",
		},
		{
			name:     "plain code block",
			input:    "```\nSELECT * FROM users\n```",
			expected: "SELECT * FROM users",
		},
		{
			name:     "no code block",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "with extra whitespace",
			input:    "  SELECT * FROM users  ",
			expected: "SELECT * FROM users",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanSQL(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFormatPrompt(t *testing.T) {
	template := "DDL: %s\nQuestion: %s"
	ddl := "CREATE TABLE users (id INT)"
	question := "Select all users"

	result := FormatPrompt(template, ddl, question)
	expected := "DDL: CREATE TABLE users (id INT)\nQuestion: Select all users"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
