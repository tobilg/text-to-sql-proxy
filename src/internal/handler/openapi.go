package handler

import (
	"net/http"
)

// OpenAPISpec is the OpenAPI v3 specification for the AI CLI Proxy API.
const OpenAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "AI CLI Proxy API",
    "description": "A local HTTP proxy that bridges web applications with AI CLI tools to generate DuckDB-compatible SQL queries.",
    "version": "1.0.0",
    "license": {
      "name": "MIT",
      "url": "https://opensource.org/licenses/MIT"
    }
  },
  "servers": [
    {
      "url": "http://localhost:4000",
      "description": "Local development server"
    }
  ],
  "paths": {
    "/generate-sql": {
      "post": {
        "summary": "Generate SQL Query",
        "description": "Generate a DuckDB-compatible SQL query from a DDL schema and natural language question using an AI CLI tool.",
        "operationId": "generateSQL",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/SQLRequest"
              },
              "examples": {
                "basic": {
                  "summary": "Basic query",
                  "value": {
                    "ddl": "CREATE TABLE users (id INT, name TEXT, email TEXT);",
                    "question": "Find all users whose name starts with 'A'"
                  }
                },
                "with_provider": {
                  "summary": "Query with specific provider",
                  "value": {
                    "ddl": "CREATE TABLE orders (id INT, user_id INT, total DECIMAL, created_at TIMESTAMP);",
                    "question": "Calculate total sales per month",
                    "provider": "gemini"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successfully generated SQL query",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/SQLResponse"
                },
                "example": {
                  "sql": "SELECT * FROM users WHERE name LIKE 'A%'"
                }
              }
            }
          },
          "400": {
            "description": "Bad request - invalid JSON, missing fields, or unknown provider",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                },
                "examples": {
                  "invalid_json": {
                    "summary": "Invalid JSON",
                    "value": {
                      "error": "Invalid JSON"
                    }
                  },
                  "missing_fields": {
                    "summary": "Missing required fields",
                    "value": {
                      "error": "Both 'ddl' and 'question' fields are required"
                    }
                  },
                  "unknown_provider": {
                    "summary": "Unknown provider",
                    "value": {
                      "error": "Unknown provider: invalid"
                    }
                  }
                }
              }
            }
          },
          "405": {
            "description": "Method not allowed - only POST and OPTIONS are supported",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                },
                "example": {
                  "error": "Method not allowed"
                }
              }
            }
          },
          "500": {
            "description": "Internal server error - AI CLI execution failed",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                },
                "example": {
                  "error": "Failed to generate SQL"
                }
              }
            }
          }
        }
      },
      "options": {
        "summary": "CORS Preflight",
        "description": "Handle CORS preflight requests for cross-origin access.",
        "operationId": "generateSQLOptions",
        "responses": {
          "200": {
            "description": "CORS preflight response with appropriate headers"
          }
        }
      }
    },
    "/openapi.json": {
      "get": {
        "summary": "OpenAPI Specification",
        "description": "Returns the OpenAPI v3 specification for this API.",
        "operationId": "getOpenAPISpec",
        "responses": {
          "200": {
            "description": "OpenAPI v3 specification",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object"
                }
              }
            }
          }
        }
      }
    },
    "/providers": {
      "get": {
        "summary": "List Providers",
        "description": "Returns the list of available AI providers.",
        "operationId": "listProviders",
        "responses": {
          "200": {
            "description": "List of available providers",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ProvidersResponse"
                },
                "example": {
                  "providers": ["claude", "gemini", "codex", "continue", "opencode"]
                }
              }
            }
          }
        }
      }
    },
    "/health": {
      "get": {
        "summary": "Health Check",
        "description": "Returns HTTP 200 if the proxy is running. Used by tools to check if the proxy is available.",
        "operationId": "healthCheck",
        "responses": {
          "200": {
            "description": "Proxy is running"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "SQLRequest": {
        "type": "object",
        "required": ["ddl", "question"],
        "properties": {
          "ddl": {
            "type": "string",
            "description": "DDL schema definition (CREATE TABLE statements)",
            "example": "CREATE TABLE users (id INT, name TEXT, email TEXT);"
          },
          "question": {
            "type": "string",
            "description": "Natural language question describing the desired SQL query",
            "example": "Find all users whose name starts with 'A'"
          },
          "provider": {
            "type": "string",
            "description": "AI provider to use for SQL generation. If omitted, uses the default configured provider.",
            "enum": ["claude", "gemini", "codex", "continue", "opencode"],
            "example": "claude"
          }
        }
      },
      "SQLResponse": {
        "type": "object",
        "properties": {
          "sql": {
            "type": "string",
            "description": "Generated DuckDB-compatible SQL query",
            "example": "SELECT * FROM users WHERE name LIKE 'A%'"
          },
          "error": {
            "type": "string",
            "description": "Error message if the request failed"
          }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "error": {
            "type": "string",
            "description": "Error message describing what went wrong",
            "example": "Failed to generate SQL"
          }
        }
      },
      "ProvidersResponse": {
        "type": "object",
        "properties": {
          "providers": {
            "type": "array",
            "items": {
              "type": "string"
            },
            "description": "List of available AI provider names",
            "example": ["claude", "gemini", "codex", "continue", "opencode"]
          }
        }
      }
    }
  }
}`

// HandleOpenAPI serves the OpenAPI v3 specification.
func (h *Handler) HandleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(OpenAPISpec))
}
