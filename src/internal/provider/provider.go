package provider

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrCLIExecution = errors.New("CLI execution failed")
	ErrParsing      = errors.New("failed to parse response")
)

// SQLGenerator defines the interface for SQL generation providers.
type SQLGenerator interface {
	GenerateSQL(ddl, question string) (string, error)
}

// CleanSQL removes any markdown code blocks or extra formatting from SQL.
func CleanSQL(sql string) string {
	// Remove markdown code blocks like ```sql ... ``` or ``` ... ```
	codeBlockRegex := regexp.MustCompile("(?s)```(?:sql)?\\s*(.+?)\\s*```")
	if matches := codeBlockRegex.FindStringSubmatch(sql); len(matches) > 1 {
		sql = matches[1]
	}

	// Remove backticks (MySQL-style quoting) - DuckDB doesn't use them
	sql = strings.ReplaceAll(sql, "`", "")

	return strings.TrimSpace(sql)
}

// FormatPrompt is a helper to format prompts with DDL and question.
func FormatPrompt(template, ddl, question string) string {
	return strings.Replace(
		strings.Replace(template, "%s", ddl, 1),
		"%s", question, 1,
	)
}
