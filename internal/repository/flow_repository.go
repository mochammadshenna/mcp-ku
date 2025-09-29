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

// FlowRepository interface for flow operations
type FlowRepository interface {
	CreateFlow(flow *types.Flow) error
	GetFlow(id string) (*types.Flow, error)
	UpdateFlow(flow *types.Flow) error
	DeleteFlow(id string) error
	ListFlows(limit int, offset int) ([]*types.Flow, error)
	CreateFlowExecution(execution *types.FlowExecution) error
	UpdateFlowExecution(execution *types.FlowExecution) error
	GetFlowExecution(id string) (*types.FlowExecution, error)
	ListFlowExecutions(flowID string, limit int, offset int) ([]*types.FlowExecution, error)
}

// flowRepository implements FlowRepository
type flowRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewFlowRepository creates a new flow repository
func NewFlowRepository(db *sql.DB, logger *logrus.Logger) FlowRepository {
	return &flowRepository{
		db:     db,
		logger: logger,
	}
}

// CreateFlow creates a new flow
func (r *flowRepository) CreateFlow(flow *types.Flow) error {
	if flow.ID == "" {
		flow.ID = uuid.New().String()
	}
	flow.CreatedAt = time.Now()
	flow.UpdatedAt = time.Now()

	query := `
		INSERT INTO flows (id, name, description, input, output, steps, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	inputJSON, err := json.Marshal(flow.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	outputJSON, err := json.Marshal(flow.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	stepsJSON, err := json.Marshal(flow.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	configJSON, err := json.Marshal(flow.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = r.db.Exec(query, flow.ID, flow.Name, flow.Description, inputJSON, outputJSON, stepsJSON, configJSON, flow.CreatedAt, flow.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	r.logger.Debugf("Created flow: %s", flow.ID)
	return nil
}

// GetFlow retrieves a flow by ID
func (r *flowRepository) GetFlow(id string) (*types.Flow, error) {
	query := `
		SELECT id, name, description, input, output, steps, config, created_at, updated_at
		FROM flows
		WHERE id = $1`

	var flow types.Flow
	var inputJSON, outputJSON, stepsJSON, configJSON string

	err := r.db.QueryRow(query, id).Scan(
		&flow.ID,
		&flow.Name,
		&flow.Description,
		&inputJSON,
		&outputJSON,
		&stepsJSON,
		&configJSON,
		&flow.CreatedAt,
		&flow.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(inputJSON), &flow.Input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}
	if err := json.Unmarshal([]byte(outputJSON), &flow.Output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w", err)
	}
	if err := json.Unmarshal([]byte(stepsJSON), &flow.Steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}
	if err := json.Unmarshal([]byte(configJSON), &flow.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &flow, nil
}

// UpdateFlow updates an existing flow
func (r *flowRepository) UpdateFlow(flow *types.Flow) error {
	flow.UpdatedAt = time.Now()

	query := `
		UPDATE flows 
		SET name = $2, description = $3, input = $4, output = $5, steps = $6, config = $7, updated_at = $8
		WHERE id = $1`

	inputJSON, err := json.Marshal(flow.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	outputJSON, err := json.Marshal(flow.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	stepsJSON, err := json.Marshal(flow.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	configJSON, err := json.Marshal(flow.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	result, err := r.db.Exec(query, flow.ID, flow.Name, flow.Description, inputJSON, outputJSON, stepsJSON, configJSON, flow.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flow not found: %s", flow.ID)
	}

	r.logger.Debugf("Updated flow: %s", flow.ID)
	return nil
}

// DeleteFlow deletes a flow by ID
func (r *flowRepository) DeleteFlow(id string) error {
	query := `DELETE FROM flows WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flow not found: %s", id)
	}

	r.logger.Debugf("Deleted flow: %s", id)
	return nil
}

// ListFlows lists flows with pagination
func (r *flowRepository) ListFlows(limit int, offset int) ([]*types.Flow, error) {
	query := `
		SELECT id, name, description, input, output, steps, config, created_at, updated_at
		FROM flows
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list flows: %w", err)
	}
	defer rows.Close()

	var flows []*types.Flow
	for rows.Next() {
		var flow types.Flow
		var inputJSON, outputJSON, stepsJSON, configJSON string

		err := rows.Scan(
			&flow.ID,
			&flow.Name,
			&flow.Description,
			&inputJSON,
			&outputJSON,
			&stepsJSON,
			&configJSON,
			&flow.CreatedAt,
			&flow.UpdatedAt,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan flow row: %v", err)
			continue
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(inputJSON), &flow.Input); err != nil {
			r.logger.Errorf("Failed to unmarshal flow input: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(outputJSON), &flow.Output); err != nil {
			r.logger.Errorf("Failed to unmarshal flow output: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(stepsJSON), &flow.Steps); err != nil {
			r.logger.Errorf("Failed to unmarshal flow steps: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(configJSON), &flow.Config); err != nil {
			r.logger.Errorf("Failed to unmarshal flow config: %v", err)
			continue
		}

		flows = append(flows, &flow)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flow rows: %w", err)
	}

	return flows, nil
}

// CreateFlowExecution creates a new flow execution
func (r *flowRepository) CreateFlowExecution(execution *types.FlowExecution) error {
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}
	execution.StartedAt = time.Now()

	query := `
		INSERT INTO flow_executions (id, flow_id, input, output, status, error, started_at, ended_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	inputJSON, err := json.Marshal(execution.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal execution input: %w", err)
	}

	outputJSON, err := json.Marshal(execution.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal execution output: %w", err)
	}

	metadataJSON, err := json.Marshal(execution.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal execution metadata: %w", err)
	}

	_, err = r.db.Exec(query, execution.ID, execution.FlowID, inputJSON, outputJSON, execution.Status, execution.Error, execution.StartedAt, execution.EndedAt, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to create flow execution: %w", err)
	}

	r.logger.Debugf("Created flow execution: %s", execution.ID)
	return nil
}

// UpdateFlowExecution updates an existing flow execution
func (r *flowRepository) UpdateFlowExecution(execution *types.FlowExecution) error {
	query := `
		UPDATE flow_executions 
		SET output = $2, status = $3, error = $4, ended_at = $5, metadata = $6
		WHERE id = $1`

	outputJSON, err := json.Marshal(execution.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal execution output: %w", err)
	}

	metadataJSON, err := json.Marshal(execution.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal execution metadata: %w", err)
	}

	result, err := r.db.Exec(query, execution.ID, outputJSON, execution.Status, execution.Error, execution.EndedAt, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to update flow execution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flow execution not found: %s", execution.ID)
	}

	r.logger.Debugf("Updated flow execution: %s", execution.ID)
	return nil
}

// GetFlowExecution retrieves a flow execution by ID
func (r *flowRepository) GetFlowExecution(id string) (*types.FlowExecution, error) {
	query := `
		SELECT id, flow_id, input, output, status, error, started_at, ended_at, metadata
		FROM flow_executions
		WHERE id = $1`

	var execution types.FlowExecution
	var inputJSON, outputJSON, metadataJSON string

	err := r.db.QueryRow(query, id).Scan(
		&execution.ID,
		&execution.FlowID,
		&inputJSON,
		&outputJSON,
		&execution.Status,
		&execution.Error,
		&execution.StartedAt,
		&execution.EndedAt,
		&metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flow execution not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get flow execution: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(inputJSON), &execution.Input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution input: %w", err)
	}
	if err := json.Unmarshal([]byte(outputJSON), &execution.Output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution output: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &execution.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution metadata: %w", err)
	}

	return &execution, nil
}

// ListFlowExecutions lists flow executions with pagination
func (r *flowRepository) ListFlowExecutions(flowID string, limit int, offset int) ([]*types.FlowExecution, error) {
	query := `
		SELECT id, flow_id, input, output, status, error, started_at, ended_at, metadata
		FROM flow_executions
		WHERE flow_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, flowID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list flow executions: %w", err)
	}
	defer rows.Close()

	var executions []*types.FlowExecution
	for rows.Next() {
		var execution types.FlowExecution
		var inputJSON, outputJSON, metadataJSON string

		err := rows.Scan(
			&execution.ID,
			&execution.FlowID,
			&inputJSON,
			&outputJSON,
			&execution.Status,
			&execution.Error,
			&execution.StartedAt,
			&execution.EndedAt,
			&metadataJSON,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan flow execution row: %v", err)
			continue
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(inputJSON), &execution.Input); err != nil {
			r.logger.Errorf("Failed to unmarshal execution input: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(outputJSON), &execution.Output); err != nil {
			r.logger.Errorf("Failed to unmarshal execution output: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(metadataJSON), &execution.Metadata); err != nil {
			r.logger.Errorf("Failed to unmarshal execution metadata: %v", err)
			continue
		}

		executions = append(executions, &execution)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flow execution rows: %w", err)
	}

	return executions, nil
}
