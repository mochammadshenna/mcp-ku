package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-octo-enigma/internal/config"
	"mcp-octo-enigma/internal/container"
	"mcp-octo-enigma/internal/server"

	"github.com/gin-gonic/gin"
)

// @title MCP Octo Enigma API
// @version 1.0
// @description Advanced MCP Server with Genkit integration
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize dependency injection container
	c, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer c.Close()

	// Initialize server
	srv := server.NewServer(c)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("MCP Server started on port %s", cfg.Server.Port)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}