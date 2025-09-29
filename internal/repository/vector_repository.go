package repository

import (
	"database/sql"
	"fmt"
	"time"

	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// VectorRepository interface for vector operations
type VectorRepository interface {
	StoreDocument(doc *types.VectorDocument) error
	SearchSimilar(query []float64, limit int, threshold float64) ([]*types.VectorSearchResult, error)
	GetDocument(id string) (*types.VectorDocument, error)
	DeleteDocument(id string) error
	ListDocuments(source string, limit int, offset int) ([]*types.VectorDocument, error)
	GetSimilarityScore(embedding1, embedding2 []float64) (float64, error)
}

// vectorRepository implements VectorRepository
type vectorRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewVectorRepository creates a new vector repository
func NewVectorRepository(db *sql.DB, logger *logrus.Logger) VectorRepository {
	return &vectorRepository{
		db:     db,
		logger: logger,
	}
}

// StoreDocument stores a document with its vector embedding
func (r *vectorRepository) StoreDocument(doc *types.VectorDocument) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	doc.CreatedAt = time.Now()

	query := `
		INSERT INTO vector_documents (id, content, embedding, metadata, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding,
			metadata = EXCLUDED.metadata,
			source = EXCLUDED.source`

	// Convert embedding to pgvector format
	embeddingStr := formatEmbeddingForPostgres(doc.Embedding)

	_, err := r.db.Exec(query, doc.ID, doc.Content, embeddingStr, doc.Metadata, doc.Source, doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	r.logger.Debugf("Stored document: %s", doc.ID)
	return nil
}

// SearchSimilar searches for documents similar to the query vector
func (r *vectorRepository) SearchSimilar(query []float64, limit int, threshold float64) ([]*types.VectorSearchResult, error) {
	queryStr := `
		SELECT 
			id, content, embedding, metadata, source, created_at,
			1 - (embedding <=> $1) as similarity
		FROM vector_documents
		WHERE 1 - (embedding <=> $1) > $2
		ORDER BY embedding <=> $1
		LIMIT $3`

	// Convert query embedding to pgvector format
	queryEmbedding := formatEmbeddingForPostgres(query)

	rows, err := r.db.Query(queryStr, queryEmbedding, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar documents: %w", err)
	}
	defer rows.Close()

	var results []*types.VectorSearchResult
	for rows.Next() {
		var doc types.VectorDocument
		var embeddingStr string
		var similarity float64

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&embeddingStr,
			&doc.Metadata,
			&doc.Source,
			&doc.CreatedAt,
			&similarity,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan row: %v", err)
			continue
		}

		// Parse embedding from string
		doc.Embedding, err = parseEmbeddingFromPostgres(embeddingStr)
		if err != nil {
			r.logger.Errorf("Failed to parse embedding: %v", err)
			continue
		}

		result := &types.VectorSearchResult{
			Document: &doc,
			Score:    similarity,
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetDocument retrieves a document by ID
func (r *vectorRepository) GetDocument(id string) (*types.VectorDocument, error) {
	query := `
		SELECT id, content, embedding, metadata, source, created_at
		FROM vector_documents
		WHERE id = $1`

	var doc types.VectorDocument
	var embeddingStr string

	err := r.db.QueryRow(query, id).Scan(
		&doc.ID,
		&doc.Content,
		&embeddingStr,
		&doc.Metadata,
		&doc.Source,
		&doc.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Parse embedding from string
	doc.Embedding, err = parseEmbeddingFromPostgres(embeddingStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding: %w", err)
	}

	return &doc, nil
}

// DeleteDocument deletes a document by ID
func (r *vectorRepository) DeleteDocument(id string) error {
	query := `DELETE FROM vector_documents WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	r.logger.Debugf("Deleted document: %s", id)
	return nil
}

// ListDocuments lists documents with optional filtering by source
func (r *vectorRepository) ListDocuments(source string, limit int, offset int) ([]*types.VectorDocument, error) {
	var query string
	var args []interface{}

	if source != "" {
		query = `
			SELECT id, content, embedding, metadata, source, created_at
			FROM vector_documents
			WHERE source = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{source, limit, offset}
	} else {
		query = `
			SELECT id, content, embedding, metadata, source, created_at
			FROM vector_documents
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []*types.VectorDocument
	for rows.Next() {
		var doc types.VectorDocument
		var embeddingStr string

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&embeddingStr,
			&doc.Metadata,
			&doc.Source,
			&doc.CreatedAt,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan row: %v", err)
			continue
		}

		// Parse embedding from string
		doc.Embedding, err = parseEmbeddingFromPostgres(embeddingStr)
		if err != nil {
			r.logger.Errorf("Failed to parse embedding: %v", err)
			continue
		}

		documents = append(documents, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return documents, nil
}

// GetSimilarityScore calculates similarity between two embeddings
func (r *vectorRepository) GetSimilarityScore(embedding1, embedding2 []float64) (float64, error) {
	query := `SELECT 1 - ($1 <=> $2) as similarity`

	emb1Str := formatEmbeddingForPostgres(embedding1)
	emb2Str := formatEmbeddingForPostgres(embedding2)

	var similarity float64
	err := r.db.QueryRow(query, emb1Str, emb2Str).Scan(&similarity)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate similarity: %w", err)
	}

	return similarity, nil
}

// Helper functions for embedding format conversion

// formatEmbeddingForPostgres converts []float64 to pgvector format
func formatEmbeddingForPostgres(embedding []float64) interface{} {
	// pgvector expects a string representation like "[1,2,3]"
	if len(embedding) == 0 {
		return nil
	}

	return pq.Array(embedding)
}

// parseEmbeddingFromPostgres converts pgvector format back to []float64
func parseEmbeddingFromPostgres(embeddingStr string) ([]float64, error) {
	if embeddingStr == "" {
		return nil, nil
	}

	// Parse the vector string format
	var embedding pq.Float64Array
	err := embedding.Scan(embeddingStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding: %w", err)
	}

	return []float64(embedding), nil
}
