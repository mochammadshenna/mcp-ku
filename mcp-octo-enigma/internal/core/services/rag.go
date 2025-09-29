package services

import (
	"context"

	"mcp-octo-enigma/internal/core/models"
	"mcp-octo-enigma/internal/core/repository"
)

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type RAGService struct {
	repo     repository.DocumentRepository
	embedder Embedder
}

func NewRAGService(repo repository.DocumentRepository, embedder Embedder) *RAGService {
	return &RAGService{repo: repo, embedder: embedder}
}

func (s *RAGService) Upsert(ctx context.Context, doc models.Document) error {
	emb, err := s.embedder.Embed(ctx, doc.Content)
	if err != nil { return err }
	return s.repo.Insert(ctx, doc, emb)
}

func (s *RAGService) Query(ctx context.Context, query string, topK int) ([]models.Document, error) {
	emb, err := s.embedder.Embed(ctx, query)
	if err != nil { return nil, err }
	return s.repo.SearchKNN(ctx, emb, topK)
}
