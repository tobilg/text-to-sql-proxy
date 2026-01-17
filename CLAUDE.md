# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

text-to-sql-proxy is a local HTTP proxy server (Go) that bridges web applications with AI CLI tools to generate SQL queries. It accepts DDL schemas and natural language questions via HTTP POST, executes local AI CLI commands, and returns SQL. The target database is configurable (defaults to DuckDB).

## Build and Test Commands

```bash
make build        # Build for current platform (output: dist/text-to-sql-proxy)
make test         # Run all tests with verbose output
make build-all    # Build for all platforms (windows, linux, darwin-amd64, darwin-arm64)
make clean        # Remove build artifacts
```

Run individual tests:
```bash
go test -v ./src/internal/provider/...   # Test provider package only
go test -v ./src/internal/handler/...    # Test handler package only
go test -run TestClaudeClient ./src/internal/provider/  # Run specific test
```

## Architecture

**Entry point**: [main.go](src/cmd/text-to-sql-proxy/main.go) - Initializes all providers, sets up HTTP routes (`/generate-sql`, `/openapi.json`), and handles graceful shutdown.

**Provider pattern**: Each AI CLI (claude, gemini, codex, continue, opencode) implements the `SQLGenerator` interface in [provider.go](src/internal/provider/provider.go):
```go
type SQLGenerator interface {
    GenerateSQL(ddl, question string) (string, error)
}
```
Provider implementations call external CLI binaries via `os/exec` and parse their responses. Use `CleanSQL()` helper to strip markdown formatting from responses.

**Handler**: [handler.go](src/internal/handler/handler.go) handles HTTP requests, CORS (including Private Network Access headers for browser compatibility), and provider selection (per-request override or default from config).

**Config**: [config.go](src/internal/config/config.go) loads from environment variables:
- `TEXT_TO_SQL_PROXY_PORT` (default: 4000)
- `TEXT_TO_SQL_PROXY_ALLOWED_ORIGIN` (default: https://sql-workbench.com)
- `TEXT_TO_SQL_PROXY_PROVIDER` (default: claude)
- `TEXT_TO_SQL_PROXY_DATABASE` (default: DuckDB)

## Adding a New Provider

1. Create `src/internal/provider/<name>.go` implementing `SQLGenerator`
2. Create corresponding `<name>_test.go`
3. Register in `main.go` providers map
4. Update valid provider names in error message
