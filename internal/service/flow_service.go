package service

import (
	"context"
	"fmt"
	"time"

	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/repository"
	"mcp-octo-enigma/internal/types"

	"github.com/sirupsen/logrus"
)

// FlowService provides methods for flow management
type FlowService struct {
	genkitService *genkit.Service
	flowRepo      repository.FlowRepository
	logger        *logrus.Logger
}

// NewFlowService creates a new FlowService
func NewFlowService(gs *genkit.Service, flowRepo repository.FlowRepository, logger *logrus.Logger) *FlowService {
	return &FlowService{
		genkitService: gs,
		flowRepo:      flowRepo,
		logger:        logger,
	}
}

// FlowExecutionRequest represents a flow execution request
type FlowExecutionRequest struct {
	FlowID     string                 `json:"flow_id"`
	Input      map[string]interface{} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// FlowExecutionResponse represents a flow execution response
type FlowExecutionResponse struct {
	FlowID      string                 `json:"flow_id"`
	Output      map[string]interface{} `json:"output"`
	RequestID   string                 `json:"request_id,omitempty"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// CreateFlow creates a new flow
func (s *FlowService) CreateFlow(ctx context.Context, flow *types.Flow) error {
	flowManager := s.genkitService.GetFlowManager()
	if flowManager == nil {
		return fmt.Errorf("flow manager not available")
	}

	// Convert to Genkit flow definition
	genkitFlow := &genkit.FlowDefinition{
		ID:          flow.ID,
		Name:        flow.Name,
		Description: flow.Description,
		Input:       flow.Input,
		Output:      flow.Output,
		Steps:       convertFlowSteps(flow.Steps),
		Config:      flow.Config,
	}

	// Create in Genkit
	if err := flowManager.CreateFlow(ctx, genkitFlow); err != nil {
		return fmt.Errorf("failed to create flow in Genkit: %w", err)
	}

	// Save to database
	if err := s.flowRepo.CreateFlow(flow); err != nil {
		return fmt.Errorf("failed to save flow to database: %w", err)
	}

	return nil
}

// GetFlow retrieves a flow by ID
func (s *FlowService) GetFlow(ctx context.Context, flowID string) (*types.Flow, error) {
	return s.flowRepo.GetFlow(flowID)
}

// UpdateFlow updates an existing flow
func (s *FlowService) UpdateFlow(ctx context.Context, flow *types.Flow) error {
	flowManager := s.genkitService.GetFlowManager()
	if flowManager == nil {
		return fmt.Errorf("flow manager not available")
	}

	// Convert to Genkit flow definition
	genkitFlow := &genkit.FlowDefinition{
		ID:          flow.ID,
		Name:        flow.Name,
		Description: flow.Description,
		Input:       flow.Input,
		Output:      flow.Output,
		Steps:       convertFlowSteps(flow.Steps),
		Config:      flow.Config,
	}

	// Update in Genkit
	if err := flowManager.UpdateFlow(ctx, genkitFlow); err != nil {
		return fmt.Errorf("failed to update flow in Genkit: %w", err)
	}

	// Update in database
	if err := s.flowRepo.UpdateFlow(flow); err != nil {
		return fmt.Errorf("failed to update flow in database: %w", err)
	}

	return nil
}

// DeleteFlow deletes a flow
func (s *FlowService) DeleteFlow(ctx context.Context, flowID string) error {
	flowManager := s.genkitService.GetFlowManager()
	if flowManager == nil {
		return fmt.Errorf("flow manager not available")
	}

	// Delete from Genkit
	if err := flowManager.DeleteFlow(ctx, flowID); err != nil {
		return fmt.Errorf("failed to delete flow from Genkit: %w", err)
	}

	// Delete from database
	if err := s.flowRepo.DeleteFlow(flowID); err != nil {
		return fmt.Errorf("failed to delete flow from database: %w", err)
	}

	return nil
}

// ListFlows lists flows with pagination
func (s *FlowService) ListFlows(ctx context.Context, limit int, offset int) ([]*types.Flow, error) {
	return s.flowRepo.ListFlows(limit, offset)
}

// ExecuteFlow executes a flow
func (s *FlowService) ExecuteFlow(ctx context.Context, req *FlowExecutionRequest) (*FlowExecutionResponse, error) {
	flowManager := s.genkitService.GetFlowManager()
	if flowManager == nil {
		return nil, fmt.Errorf("flow manager not available")
	}

	// Create Genkit flow execution request
	genkitReq := &genkit.FlowExecutionRequest{
		FlowID:     req.FlowID,
		Input:      req.Input,
		Parameters: req.Parameters,
		RequestID:  req.RequestID,
	}

	// Execute flow in Genkit
	genkitResp, err := flowManager.ExecuteFlow(ctx, genkitReq)
	if err != nil {
		return nil, fmt.Errorf("flow execution failed: %w", err)
	}

	// Create flow execution record
	execution := &types.FlowExecution{
		FlowID:   req.FlowID,
		Input:    req.Input,
		Output:   genkitResp.Output,
		Status:   genkitResp.Status,
		Metadata: genkitResp.Metadata,
	}

	// Save execution to database
	if err := s.flowRepo.CreateFlowExecution(execution); err != nil {
		s.logger.Errorf("Failed to save flow execution: %v", err)
	}

	response := &FlowExecutionResponse{
		FlowID:      genkitResp.FlowID,
		Output:      genkitResp.Output,
		RequestID:   genkitResp.RequestID,
		Status:      genkitResp.Status,
		Metadata:    genkitResp.Metadata,
		CreatedAt:   genkitResp.CreatedAt,
		CompletedAt: genkitResp.CompletedAt,
	}

	return response, nil
}

// ListFlowExecutions lists executions for a flow
func (s *FlowService) ListFlowExecutions(ctx context.Context, flowID string, limit int, offset int) ([]*types.FlowExecution, error) {
	return s.flowRepo.ListFlowExecutions(flowID, limit, offset)
}

// convertFlowSteps converts types.FlowStep to genkit.FlowStep
func convertFlowSteps(steps []types.FlowStep) []genkit.FlowStep {
	var genkitSteps []genkit.FlowStep
	for _, step := range steps {
		genkitSteps = append(genkitSteps, genkit.FlowStep{
			ID:           step.ID,
			Type:         step.Type,
			Name:         step.Name,
			Description:  step.Description,
			Config:       step.Config,
			Dependencies: step.Dependencies,
		})
	}
	return genkitSteps
}
