package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"mcp-octo-enigma/pkg/config"
	"mcp-octo-enigma/pkg/db"
	"mcp-octo-enigma/pkg/logger"
	"mcp-octo-enigma/internal/core/repository"
	"mcp-octo-enigma/internal/core/services"
	"mcp-octo-enigma/internal/http/handlers"
)

func main() {
	cfg := config.Load()
	_, _ = logger.Init(cfg.Env)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	if cfg.DatabaseURL != "" {
		pg, err := db.NewPostgres(context.Background(), cfg.DatabaseURL)
		if err != nil { log.Fatalf("db connect error: %v", err) }
		defer pg.Close()
		repo := repository.NewDocumentRepository(pg.Pool)
		rag := services.NewRAGService(repo, services.DummyEmbedder{})
		h := handlers.NewSearchHandler(rag)
		h.Register(r)
	}

	if err := r.Run(":" + cfg.ServerPort); err != nil { log.Fatal(err) }
}
