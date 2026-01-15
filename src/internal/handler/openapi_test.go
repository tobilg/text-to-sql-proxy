package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleOpenAPI_ReturnsValidJSON(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	handler.HandleOpenAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify it's valid JSON
	var spec map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

func TestHandleOpenAPI_HasCorrectContentType(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	handler.HandleOpenAPI(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

func TestHandleOpenAPI_HasOpenAPIVersion(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	handler.HandleOpenAPI(w, req)

	var spec map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &spec)

	openapi, ok := spec["openapi"].(string)
	if !ok {
		t.Fatal("missing 'openapi' field")
	}

	if openapi != "3.0.3" {
		t.Errorf("expected openapi version '3.0.3', got %q", openapi)
	}
}

func TestHandleOpenAPI_HasPaths(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	handler.HandleOpenAPI(w, req)

	var spec map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &spec)

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("missing 'paths' field")
	}

	if _, ok := paths["/generate-sql"]; !ok {
		t.Error("missing '/generate-sql' path")
	}

	if _, ok := paths["/openapi.json"]; !ok {
		t.Error("missing '/openapi.json' path")
	}
}

func TestHandleOpenAPI_HasProviderEnum(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	handler.HandleOpenAPI(w, req)

	var spec map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &spec)

	// Navigate to components.schemas.SQLRequest.properties.provider.enum
	components := spec["components"].(map[string]interface{})
	schemas := components["schemas"].(map[string]interface{})
	sqlRequest := schemas["SQLRequest"].(map[string]interface{})
	properties := sqlRequest["properties"].(map[string]interface{})
	provider := properties["provider"].(map[string]interface{})
	enum := provider["enum"].([]interface{})

	expectedProviders := []string{"claude", "gemini", "codex", "continue", "opencode"}
	if len(enum) != len(expectedProviders) {
		t.Errorf("expected %d providers, got %d", len(expectedProviders), len(enum))
	}

	for i, expected := range expectedProviders {
		if enum[i].(string) != expected {
			t.Errorf("expected provider %q at index %d, got %q", expected, i, enum[i])
		}
	}
}

func TestHandleOpenAPI_AllowsCORS(t *testing.T) {
	handler := newTestHandler(&mockSQLGenerator{})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	handler.HandleOpenAPI(w, req)

	cors := w.Header().Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("expected CORS header '*', got %q", cors)
	}
}
