# ai-cli-proxy

A local HTTP proxy that bridges web applications with AI CLI tools to generate DuckDB-compatible SQL queries.

## Overview

This proxy allows browser-based applications (like [sql-workbench.com](https://sql-workbench.com)) to leverage your local AI CLI subscriptions for SQL generation. It accepts DDL schemas and natural language questions, then returns DuckDB-compatible SQL queries.

### How It Works

```
┌─────────────────┐     HTTP POST      ┌─────────────────┐     exec      ┌─────────────────────┐
│   Web Browser   │ ─────────────────► │  ai-cli-proxy   │ ────────────► │    AI CLI           │
│ (sql-workbench) │ ◄───────────────── │  localhost:4000 │ ◄──────────── │ (claude/gemini/...) │
└─────────────────┘     SQL Response   └─────────────────┘    Response   └─────────────────────┘
```

### Supported Providers

| Provider | CLI Command | Install |
|----------|-------------|---------|
| Claude | `claude` | [Installation Guide](https://docs.anthropic.com/en/docs/claude-cli) |
| Gemini | `gemini` | [Installation Guide](https://geminicli.com/docs/installation) |
| Codex | `codex` | [Installation Guide](https://developers.openai.com/codex/cli/installation) |
| Continue | `cn` | `npm i -g @continuedev/cli` |
| OpenCode | `opencode` | [Installation Guide](https://opencode.ai/docs/cli) |

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/tobilg/ai-cli-proxy.git
cd ai-cli-proxy

# Build for your platform
make build

# Or build for all platforms
make build-all
```

### Pre-built binaries

Download the appropriate binary for your platform from the `dist/` directory after building:

| Platform | Binary |
|----------|--------|
| Windows | `ai-cli-proxy-windows-amd64.exe` |
| Linux | `ai-cli-proxy-linux-amd64` |
| macOS (Apple Silicon) | `ai-cli-proxy-darwin-arm64` |

## Usage

### Prerequisites

At least one of these CLI tools must be installed and authenticated:

```bash
# Claude CLI (Anthropic)
# Follow: https://docs.anthropic.com/en/docs/claude-cli

# Gemini CLI (Google)
# Follow: https://geminicli.com/docs/installation

# Codex CLI (OpenAI)
# Follow: https://developers.openai.com/codex/cli/installation

# Continue CLI
npm i -g @continuedev/cli

# OpenCode CLI
# Follow: https://opencode.ai/docs/cli
```

### Running the proxy

```bash
# Run with default settings (Claude provider)
./dist/ai-cli-proxy

# Run with a specific default provider
AI_CLI_PROXY_PROVIDER=gemini ./dist/ai-cli-proxy
AI_CLI_PROXY_PROVIDER=codex ./dist/ai-cli-proxy
AI_CLI_PROXY_PROVIDER=continue ./dist/ai-cli-proxy
AI_CLI_PROXY_PROVIDER=opencode ./dist/ai-cli-proxy

# Run with custom port
AI_CLI_PROXY_PORT=8080 ./dist/ai-cli-proxy

# Run with custom allowed origin
AI_CLI_PROXY_ALLOWED_ORIGIN="http://localhost:3000" ./dist/ai-cli-proxy
```

The proxy will start and display (with default settings):

```
AI CLI Proxy active at http://localhost:4000
Default provider: claude
Allowed origin: https://sql-workbench.com
Available providers: claude, gemini, codex, continue, opencode
API docs: http://localhost:4000/openapi.json
Press Ctrl+C to stop
```

### Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `AI_CLI_PROXY_PORT` | `4000` | Port the proxy listens on |
| `AI_CLI_PROXY_ALLOWED_ORIGIN` | `https://sql-workbench.com` | CORS allowed origin |
| `AI_CLI_PROXY_PROVIDER` | `claude` | Default AI provider |

Valid providers: `claude`, `gemini`, `codex`, `continue`, `opencode`

## API

### GET /health

Health check endpoint to verify the proxy is running.

**Example Request:**

```bash
curl http://localhost:4000/health
```

**Example Response (200):**

Empty response with HTTP status 200.

---

### GET /openapi.json

Returns the OpenAPI v3 specification for this API.

**Example Request:**

```bash
curl http://localhost:4000/openapi.json
```

**Example Response (200):**

```json
{
  "openapi": "3.0.3",
  "info": {
    "title": "AI CLI Proxy API",
    "version": "1.0.0"
  },
  "paths": { ... }
}
```

---

### GET /providers

Returns the list of available AI providers with their descriptions.

**Example Request:**

```bash
curl http://localhost:4000/providers
```

**Example Response (200):**

```json
{
  "providers": [
    {"name": "claude", "description": "Anthropic Claude CLI"},
    {"name": "gemini", "description": "Google Gemini CLI"},
    {"name": "codex", "description": "OpenAI Codex CLI"},
    {"name": "continue", "description": "Continue CLI"},
    {"name": "opencode", "description": "OpenCode CLI"}
  ]
}
```

---

### POST /generate-sql

Generate a DuckDB SQL query from a schema and natural language question.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ddl` | string | Yes | DDL schema (CREATE TABLE statements) |
| `question` | string | Yes | Natural language question |
| `provider` | string | No | AI provider to use (defaults to configured provider) |

**Example Request:**

```bash
curl -X POST http://localhost:4000/generate-sql \
  -H "Content-Type: application/json" \
  -d '{
    "ddl": "CREATE TABLE users (id INT, name TEXT, email TEXT);",
    "question": "Find all users whose name starts with A"
  }'
```

**Example Response (200):**

```json
{
  "sql": "SELECT * FROM users WHERE name LIKE 'A%'"
}
```

**Example Request with Provider Override:**

```bash
curl -X POST http://localhost:4000/generate-sql \
  -H "Content-Type: application/json" \
  -d '{
    "ddl": "CREATE TABLE orders (id INT, user_id INT, total DECIMAL, created_at TIMESTAMP);",
    "question": "Calculate total sales per month",
    "provider": "gemini"
  }'
```

**Example Response (200):**

```json
{
  "sql": "SELECT DATE_TRUNC('month', created_at) AS month, SUM(total) AS total_sales FROM orders GROUP BY month ORDER BY month"
}
```

**Error Responses:**

| Status | Description | Example |
|--------|-------------|---------|
| 400 | Invalid JSON or missing required fields | `{"error": "Both 'ddl' and 'question' fields are required"}` |
| 400 | Unknown provider | `{"error": "Unknown provider: invalid"}` |
| 405 | Method not allowed | `{"error": "Method not allowed"}` |
| 500 | AI CLI execution failed | `{"error": "Failed to generate SQL"}` |

## Development

### Running tests

```bash
make test
```

### Build commands

```bash
make build              # Build for current platform
make build-all          # Build for all platforms
make build-windows      # Build for Windows
make build-linux        # Build for Linux
make build-darwin-amd64 # Build for macOS Intel
make build-darwin-arm64 # Build for macOS Apple Silicon
make clean              # Remove build artifacts
```

### Project structure

```
ai-cli-proxy/
├── src/
│   ├── cmd/ai-cli-proxy/    # Application entry point
│   └── internal/
│       ├── config/          # Configuration loading
│       ├── handler/         # HTTP handlers
│       └── provider/        # AI CLI provider implementations
├── dist/                    # Built binaries
├── Makefile
└── README.md
```

## Browser Security Notes

Modern browsers enforce strict security policies for requests from HTTPS sites to local HTTP servers. This proxy includes:

- **CORS headers** for cross-origin requests
- **Private Network Access** header (`Access-Control-Allow-Private-Network: true`) for browser compatibility

If you encounter connection issues, you may need to enable `chrome://flags/#allow-insecure-localhost` in Chrome-based browsers.

## License

MIT
