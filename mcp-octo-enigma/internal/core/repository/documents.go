package repository

import (
	"context"

	"mcp-octo-enigma/internal/core/models"
	"github.com/jackc/pgx/v5/pgxpool"
	vector "github.com/pgvector/pgvector-go"
)

type DocumentRepository interface {
	Insert(ctx context.Context, doc models.Document, embedding []float32) error
	SearchKNN(ctx context.Context, embedding []float32, topK int) ([]models.Document, error)
}

type documentRepository struct { pool *pgxpool.Pool }

func NewDocumentRepository(pool *pgxpool.Pool) DocumentRepository { return &documentRepository{pool: pool} }

func (r *documentRepository) Insert(ctx context.Context, doc models.Document, embedding []float32) error {
	_, err := r.pool.Exec(
		ctx,
		"INSERT INTO documents (id, content, embedding, metadata) VALUES ($1,$2,$3,$4) ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, embedding = EXCLUDED.embedding, metadata = EXCLUDED.metadata",
		doc.ID,
		doc.Content,
		vector.NewVector(embedding),
		doc.Metadata,
	)
	return err
}

func (r *documentRepository) SearchKNN(ctx context.Context, embedding []float32, topK int) ([]models.Document, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, content, metadata FROM documents ORDER BY embedding <-> $1 LIMIT $2",
		vector.NewVector(embedding), topK,
	)
	if err != nil { return nil, err }
	defer rows.Close()
	out := make([]models.Document, 0, topK)
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.Content, &d.Metadata); err != nil { return nil, err }
		out = append(out, d)
	}
	return out, rows.Err()
}
