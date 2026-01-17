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
		{
			name:     "with backticks",
			input:    "SELECT COUNT(*) FROM `aws_iam`.actions",
			expected: "SELECT COUNT(*) FROM aws_iam.actions",
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
	template := "You are a %s expert.\nDDL: %s\nQuestion: %s"
	database := "DuckDB"
	ddl := "CREATE TABLE users (id INT)"
	question := "Select all users"

	result := FormatPrompt(template, database, ddl, question)
	expected := "You are a DuckDB expert.\nDDL: CREATE TABLE users (id INT)\nQuestion: Select all users"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatPrompt_CustomDatabase(t *testing.T) {
	template := "You are a %s expert.\nDDL: %s\nQuestion: %s"
	database := "PostgreSQL"
	ddl := "CREATE TABLE users (id INT)"
	question := "Select all users"

	result := FormatPrompt(template, database, ddl, question)
	expected := "You are a PostgreSQL expert.\nDDL: CREATE TABLE users (id INT)\nQuestion: Select all users"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
