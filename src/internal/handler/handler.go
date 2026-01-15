package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/tobilg/ai-cli-proxy/src/internal/provider"
)

// SQLRequest represents the incoming request payload.
type SQLRequest struct {
	DDL      string `json:"ddl"`
	Question string `json:"question"`
	Provider string `json:"provider,omitempty"`
}

// SQLResponse represents the response payload.
type SQLResponse struct {
	SQL   string `json:"sql,omitempty"`
	Error string `json:"error,omitempty"`
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	providers       map[string]provider.SQLGenerator
	defaultProvider string
	allowedOrigin   string
}

// New creates a new Handler with the given dependencies.
func New(providers map[string]provider.SQLGenerator, defaultProvider, allowedOrigin string) *Handler {
	return &Handler{
		providers:       providers,
		defaultProvider: defaultProvider,
		allowedOrigin:   allowedOrigin,
	}
}

// HandleGenerateSQL handles POST /generate-sql requests.
func (h *Handler) HandleGenerateSQL(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid JSON: %v", err)
		h.sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DDL == "" || req.Question == "" {
		log.Printf("[ERROR] Missing required fields: ddl=%q, question=%q", req.DDL, req.Question)
		h.sendError(w, "Both 'ddl' and 'question' fields are required", http.StatusBadRequest)
		return
	}

	// Determine which provider to use
	providerName := req.Provider
	if providerName == "" {
		providerName = h.defaultProvider
	}

	p, ok := h.providers[providerName]
	if !ok {
		log.Printf("[ERROR] Unknown provider: %s", providerName)
		h.sendError(w, fmt.Sprintf("Unknown provider: %s", providerName), http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Generating SQL using %s for question: %q", providerName, req.Question)

	sql, err := p.GenerateSQL(req.DDL, req.Question)
	if err != nil {
		log.Printf("[ERROR] %s CLI failed: %v", providerName, err)
		h.sendError(w, "Failed to generate SQL", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] Successfully generated SQL")
	h.sendJSON(w, SQLResponse{SQL: sql})
}

// setCORSHeaders sets the required CORS and Private Network Access headers.
func (h *Handler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", h.allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
}

// sendError sends an error response as JSON.
func (h *Handler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SQLResponse{Error: message})
}

// sendJSON sends a successful JSON response.
func (h *Handler) sendJSON(w http.ResponseWriter, response SQLResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
