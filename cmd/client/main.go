package main

import (
	"context"
	"log"

	"mcp-octo-enigma/internal/client"
	"mcp-octo-enigma/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize MCP client
	mcpClient := client.NewMCPClient(cfg)

	// Example usage
	ctx := context.Background()

	// Connect to MCP server
	if err := mcpClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer mcpClient.Close()

	// Example: Generate content
	response, err := mcpClient.GenerateContent(ctx, &client.GenerateRequest{
		Model:  "gemini-pro",
		Prompt: "What is the meaning of life?",
	})
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	log.Printf("Generated content: %s", response.Content)
}
