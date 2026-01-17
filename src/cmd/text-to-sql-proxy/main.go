package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tobilg/text-to-sql-proxy/src/internal/config"
	"github.com/tobilg/text-to-sql-proxy/src/internal/handler"
	"github.com/tobilg/text-to-sql-proxy/src/internal/provider"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("text-to-sql-proxy %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		os.Exit(0)
	}

	cfg := config.Load()

	// Initialize all providers
	providers := map[string]provider.SQLGenerator{
		"claude":   provider.NewClaudeClient(cfg.Database),
		"gemini":   provider.NewGeminiClient(cfg.Database),
		"codex":    provider.NewCodexClient(cfg.Database),
		"continue": provider.NewContinueClient(cfg.Database),
		"opencode": provider.NewOpenCodeClient(cfg.Database),
	}

	// Validate configured provider exists
	if _, ok := providers[cfg.Provider]; !ok {
		log.Fatalf("Unknown provider: %s (valid options: claude, gemini, codex, continue, opencode)", cfg.Provider)
	}

	h := handler.New(providers, cfg.Provider, cfg.AllowedOrigin)

	mux := http.NewServeMux()
	mux.HandleFunc("/generate-sql", h.HandleGenerateSQL)
	mux.HandleFunc("/providers", h.HandleProviders)
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/openapi.json", h.HandleOpenAPI)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for shutdown signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Text-to-SQL Proxy active at http://localhost:%d\n", cfg.Port)
		fmt.Printf("Default provider: %s\n", cfg.Provider)
		fmt.Printf("Target database: %s\n", cfg.Database)
		fmt.Printf("Allowed origin: %s\n", cfg.AllowedOrigin)
		fmt.Println("Available providers: claude, gemini, codex, continue, opencode")
		fmt.Printf("API docs: http://localhost:%d/openapi.json\n", cfg.Port)
		fmt.Println("Press Ctrl+C to stop")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	fmt.Println("\nShutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	fmt.Println("Server stopped")
}
