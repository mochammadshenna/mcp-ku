package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// GenerationRepository interface for generation operations
type GenerationRepository interface {
	CreateGeneration(generation *types.Generation) error
	GetGeneration(id string) (*types.Generation, error)
	UpdateGeneration(generation *types.Generation) error
	DeleteGeneration(id string) error
	ListGenerations(filter GenerationFilter) ([]*types.Generation, error)
	GetGenerationByRequestID(requestID string) (*types.Generation, error)
}

// GenerationFilter for filtering generation results
type GenerationFilter struct {
	Model     string
	Status    string
	RequestID string
	Limit     int
	Offset    int
}

// generationRepository implements GenerationRepository
type generationRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewGenerationRepository creates a new generation repository
func NewGenerationRepository(db *sql.DB, logger *logrus.Logger) GenerationRepository {
	return &generationRepository{
		db:     db,
		logger: logger,
	}
}

// CreateGeneration creates a new generation record
func (r *generationRepository) CreateGeneration(generation *types.Generation) error {
	if generation.ID == "" {
		generation.ID = uuid.New().String()
	}
	generation.CreatedAt = time.Now()

	query := `
		INSERT INTO generations (id, model, prompt, response, parameters, metadata, status, request_id, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	parametersJSON, err := json.Marshal(generation.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	metadataJSON, err := json.Marshal(generation.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.Exec(query, generation.ID, generation.Model, generation.Prompt, generation.Response, parametersJSON, metadataJSON, generation.Status, generation.RequestID, generation.CreatedAt, generation.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to create generation: %w", err)
	}

	r.logger.Debugf("Created generation: %s", generation.ID)
	return nil
}

// GetGeneration retrieves a generation by ID
func (r *generationRepository) GetGeneration(id string) (*types.Generation, error) {
	query := `
		SELECT id, model, prompt, response, parameters, metadata, status, request_id, created_at, completed_at
		FROM generations
		WHERE id = $1`

	var generation types.Generation
	var parametersJSON, metadataJSON string

	err := r.db.QueryRow(query, id).Scan(
		&generation.ID,
		&generation.Model,
		&generation.Prompt,
		&generation.Response,
		&parametersJSON,
		&metadataJSON,
		&generation.Status,
		&generation.RequestID,
		&generation.CreatedAt,
		&generation.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("generation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get generation: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(parametersJSON), &generation.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &generation.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &generation, nil
}

// UpdateGeneration updates an existing generation
func (r *generationRepository) UpdateGeneration(generation *types.Generation) error {
	query := `
		UPDATE generations 
		SET response = $2, parameters = $3, metadata = $4, status = $5, completed_at = $6
		WHERE id = $1`

	parametersJSON, err := json.Marshal(generation.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	metadataJSON, err := json.Marshal(generation.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.Exec(query, generation.ID, generation.Response, parametersJSON, metadataJSON, generation.Status, generation.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to update generation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("generation not found: %s", generation.ID)
	}

	r.logger.Debugf("Updated generation: %s", generation.ID)
	return nil
}

// DeleteGeneration deletes a generation by ID
func (r *generationRepository) DeleteGeneration(id string) error {
	query := `DELETE FROM generations WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete generation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("generation not found: %s", id)
	}

	r.logger.Debugf("Deleted generation: %s", id)
	return nil
}

// ListGenerations lists generations with filtering and pagination
func (r *generationRepository) ListGenerations(filter GenerationFilter) ([]*types.Generation, error) {
	query := `
		SELECT id, model, prompt, response, parameters, metadata, status, request_id, created_at, completed_at
		FROM generations
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if filter.Model != "" {
		query += fmt.Sprintf(" AND model = $%d", argIndex)
		args = append(args, filter.Model)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.RequestID != "" {
		query += fmt.Sprintf(" AND request_id = $%d", argIndex)
		args = append(args, filter.RequestID)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list generations: %w", err)
	}
	defer rows.Close()

	var generations []*types.Generation
	for rows.Next() {
		var generation types.Generation
		var parametersJSON, metadataJSON string

		err := rows.Scan(
			&generation.ID,
			&generation.Model,
			&generation.Prompt,
			&generation.Response,
			&parametersJSON,
			&metadataJSON,
			&generation.Status,
			&generation.RequestID,
			&generation.CreatedAt,
			&generation.CompletedAt,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan generation row: %v", err)
			continue
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(parametersJSON), &generation.Parameters); err != nil {
			r.logger.Errorf("Failed to unmarshal generation parameters: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(metadataJSON), &generation.Metadata); err != nil {
			r.logger.Errorf("Failed to unmarshal generation metadata: %v", err)
			continue
		}

		generations = append(generations, &generation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating generation rows: %w", err)
	}

	return generations, nil
}

// GetGenerationByRequestID retrieves a generation by request ID
func (r *generationRepository) GetGenerationByRequestID(requestID string) (*types.Generation, error) {
	query := `
		SELECT id, model, prompt, response, parameters, metadata, status, request_id, created_at, completed_at
		FROM generations
		WHERE request_id = $1`

	var generation types.Generation
	var parametersJSON, metadataJSON string

	err := r.db.QueryRow(query, requestID).Scan(
		&generation.ID,
		&generation.Model,
		&generation.Prompt,
		&generation.Response,
		&parametersJSON,
		&metadataJSON,
		&generation.Status,
		&generation.RequestID,
		&generation.CreatedAt,
		&generation.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("generation not found for request ID: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get generation by request ID: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(parametersJSON), &generation.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &generation.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &generation, nil
}