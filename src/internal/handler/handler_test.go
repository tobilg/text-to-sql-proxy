package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tobilg/ai-cli-proxy/src/internal/provider"
)

// mockSQLGenerator implements provider.SQLGenerator for testing.
type mockSQLGenerator struct {
	sql string
	err error
}

func (m *mockSQLGenerator) GenerateSQL(ddl, question string) (string, error) {
	return m.sql, m.err
}

func newTestHandler(mock *mockSQLGenerator) *Handler {
	providers := map[string]provider.SQLGenerator{
		"claude": mock,
	}
	return New(providers, "claude", "https://sql-workbench.com")
}

func newTestHandlerWithProviders(providers map[string]provider.SQLGenerator, defaultProvider string) *Handler {
	return New(providers, defaultProvider, "https://sql-workbench.com")
}

func TestHandleGenerateSQL_CORSHeaders(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{sql: "SELECT 1"})

	req := httptest.NewRequest(http.MethodOptions, "/generate-sql", nil)
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":          "https://sql-workbench.com",
		"Access-Control-Allow-Methods":         "POST, GET, OPTIONS",
		"Access-Control-Allow-Headers":         "Content-Type",
		"Access-Control-Allow-Private-Network": "true",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("header %s: expected %q, got %q", header, expected, got)
		}
	}
}

func TestHandleGenerateSQL_OptionsRequest(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodOptions, "/generate-sql", nil)
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestHandleGenerateSQL_InvalidJSON(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp SQLResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "Invalid JSON" {
		t.Errorf("expected error 'Invalid JSON', got %q", resp.Error)
	}
}

func TestHandleGenerateSQL_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body SQLRequest
	}{
		{"missing ddl", SQLRequest{Question: "select users"}},
		{"missing question", SQLRequest{DDL: "CREATE TABLE users (id INT)"}},
		{"both missing", SQLRequest{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := newTestHandler(&mockSQLGenerator{})

			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleGenerateSQL(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}

			var resp SQLResponse
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.Error != "Both 'ddl' and 'question' fields are required" {
				t.Errorf("unexpected error: %q", resp.Error)
			}
		})
	}
}

func TestHandleGenerateSQL_Success(t *testing.T) {
	expectedSQL := "SELECT * FROM users"
	handler := newTestHandler(&mockSQLGenerator{sql: expectedSQL})

	body, _ := json.Marshal(SQLRequest{
		DDL:      "CREATE TABLE users (id INT, name TEXT)",
		Question: "Select all users",
	})
	req := httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SQLResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.SQL != expectedSQL {
		t.Errorf("expected SQL %q, got %q", expectedSQL, resp.SQL)
	}
	if resp.Error != "" {
		t.Errorf("unexpected error: %q", resp.Error)
	}
}

func TestHandleGenerateSQL_ProviderError(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{err: errors.New("CLI failed")})

	body, _ := json.Marshal(SQLRequest{
		DDL:      "CREATE TABLE users (id INT)",
		Question: "Select all users",
	})
	req := httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var resp SQLResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "Failed to generate SQL" {
		t.Errorf("expected error 'Failed to generate SQL', got %q", resp.Error)
	}
}

func TestHandleGenerateSQL_MethodNotAllowed(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/generate-sql", nil)
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleGenerateSQL_ProviderSelection(t *testing.T) {
	claudeMock := &mockSQLGenerator{sql: "SELECT claude"}
	geminiMock := &mockSQLGenerator{sql: "SELECT gemini"}

	providers := map[string]provider.SQLGenerator{
		"claude": claudeMock,
		"gemini": geminiMock,
	}
	handler := newTestHandlerWithProviders(providers, "claude")

	// Test default provider (claude)
	body, _ := json.Marshal(SQLRequest{
		DDL:      "CREATE TABLE users (id INT)",
		Question: "Select all",
	})
	req := httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	var resp SQLResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.SQL != "SELECT claude" {
		t.Errorf("expected default provider (claude), got SQL %q", resp.SQL)
	}

	// Test explicit provider selection (gemini)
	body, _ = json.Marshal(SQLRequest{
		DDL:      "CREATE TABLE users (id INT)",
		Question: "Select all",
		Provider: "gemini",
	})
	req = httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBuffer(body))
	w = httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	json.NewDecoder(w.Body).Decode(&resp)
	if resp.SQL != "SELECT gemini" {
		t.Errorf("expected gemini provider, got SQL %q", resp.SQL)
	}
}

func TestHandleGenerateSQL_UnknownProvider(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{sql: "SELECT 1"})

	body, _ := json.Marshal(SQLRequest{
		DDL:      "CREATE TABLE users (id INT)",
		Question: "Select all",
		Provider: "unknown",
	})
	req := httptest.NewRequest(http.MethodPost, "/generate-sql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.HandleGenerateSQL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp SQLResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "Unknown provider: unknown" {
		t.Errorf("expected 'Unknown provider: unknown', got %q", resp.Error)
	}
}

func TestHandleProviders_Success(t *testing.T) {
	providers := map[string]provider.SQLGenerator{
		"claude": &mockSQLGenerator{},
		"gemini": &mockSQLGenerator{},
	}
	handler := newTestHandlerWithProviders(providers, "claude")

	req := httptest.NewRequest(http.MethodGet, "/providers", nil)
	w := httptest.NewRecorder()

	handler.HandleProviders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var resp struct {
		Providers []string `json:"providers"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(resp.Providers))
	}
}

func TestHandleProviders_MethodNotAllowed(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodPost, "/providers", nil)
	w := httptest.NewRecorder()

	handler.HandleProviders(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
