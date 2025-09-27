package service

import (
	"context"
	"fmt"
	"time"

	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/mcp"
	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ContentService interface for content generation operations
type ContentService interface {
	GenerateContent(ctx context.Context, req *ContentGenerationRequest) (*ContentGenerationResponse, error)
	GenerateContentStream(ctx context.Context, req *ContentGenerationRequest) (<-chan *ContentStreamChunk, error)
	InterruptGeneration(ctx context.Context, requestID string) error
	GetGeneration(ctx context.Context, id string) (*types.Generation, error)
	ListGenerations(ctx context.Context, filter GenerationFilter) ([]*types.Generation, error)
}

// ContentGenerationRequest represents a content generation request
type ContentGenerationRequest struct {
	Model       string                 `json:"model" validate:"required"`
	Prompt      string                 `json:"prompt" validate:"required"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Tools       []types.Tool           `json:"tools,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// ContentGenerationResponse represents a content generation response
type ContentGenerationResponse struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Model       string                 `json:"model"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ToolCalls   []types.ToolCall       `json:"tool_calls,omitempty"`
	Usage       *genkit.UsageInfo      `json:"usage,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// ContentStreamChunk represents a streaming content chunk
type ContentStreamChunk struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Done      bool                   `json:"done"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// GenerationFilter for filtering generation results
type GenerationFilter struct {
	Model     string `json:"model,omitempty"`
	Status    string `json:"status,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// contentService implements ContentService
type contentService struct {
	genkitSvc  *genkit.Service
	mcpManager *mcp.Manager
	logger     *logrus.Logger
}

// NewContentService creates a new content service
func NewContentService(genkitSvc *genkit.Service, mcpManager *mcp.Manager, logger *logrus.Logger) ContentService {
	return &contentService{
		genkitSvc:  genkitSvc,
		mcpManager: mcpManager,
		logger:     logger,
	}
}

// GenerateContent generates content using AI models
func (s *contentService) GenerateContent(ctx context.Context, req *ContentGenerationRequest) (*ContentGenerationResponse, error) {
	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Create generation record
	generation := &types.Generation{
		ID:        uuid.New().String(),
		Model:     req.Model,
		Prompt:    req.Prompt,
		Status:    "generating",
		RequestID: req.RequestID,
		CreatedAt: time.Now(),
	}

	s.logger.Infof("Starting content generation with model %s (ID: %s)", req.Model, generation.ID)

	// Prepare Genkit request
	genkitReq := &genkit.GenerateContentRequest{
		Model:     req.Model,
		Prompt:    req.Prompt,
		RequestID: req.RequestID,
	}

	// Add parameters if provided
	if req.Parameters != nil {
		genkitReq.Parameters = req.Parameters
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		genkitReq.Tools = req.Tools
	}

	// Generate content using Genkit
	genkitResp, err := s.genkitSvc.GenerateContent(ctx, genkitReq)
	if err != nil {
		generation.Status = "failed"
		generation.Error = err.Error()
		s.logger.Errorf("Content generation failed: %v", err)
		return nil, fmt.Errorf("content generation failed: %w", err)
	}

	// Update generation record
	now := time.Now()
	generation.Status = "completed"
	generation.Response = genkitResp.Content
	generation.Metadata = genkitResp.Metadata
	generation.CompletedAt = &now

	response := &ContentGenerationResponse{
		ID:          generation.ID,
		Content:     genkitResp.Content,
		Model:       req.Model,
		Metadata:    genkitResp.Metadata,
		ToolCalls:   genkitResp.ToolCalls,
		Usage:       genkitResp.Usage,
		RequestID:   req.RequestID,
		Status:      generation.Status,
		CreatedAt:   generation.CreatedAt,
		CompletedAt: generation.CompletedAt,
	}

	s.logger.Infof("Content generation completed (ID: %s)", generation.ID)
	return response, nil
}

// GenerateContentStream generates content with streaming
func (s *contentService) GenerateContentStream(ctx context.Context, req *ContentGenerationRequest) (<-chan *ContentStreamChunk, error) {
	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	generationID := uuid.New().String()
	s.logger.Infof("Starting streaming content generation with model %s (ID: %s)", req.Model, generationID)

	// Prepare Genkit request
	genkitReq := &genkit.GenerateContentRequest{
		Model:     req.Model,
		Prompt:    req.Prompt,
		RequestID: req.RequestID,
		Stream:    true,
	}

	// Add parameters if provided
	if req.Parameters != nil {
		genkitReq.Parameters = req.Parameters
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		genkitReq.Tools = req.Tools
	}

	// Get streaming channel from Genkit
	genkitCh, err := s.genkitSvc.GenerateContentStream(ctx, genkitReq)
	if err != nil {
		s.logger.Errorf("Streaming content generation failed: %v", err)
		return nil, fmt.Errorf("streaming content generation failed: %w", err)
	}

	// Create output channel
	outputCh := make(chan *ContentStreamChunk)

	// Transform Genkit chunks to service chunks
	go func() {
		defer close(outputCh)

		for genkitChunk := range genkitCh {
			chunk := &ContentStreamChunk{
				ID:        generationID,
				Content:   genkitChunk.Content,
				Done:      genkitChunk.Done,
				Metadata:  genkitChunk.Metadata,
				RequestID: req.RequestID,
			}

			select {
			case outputCh <- chunk:
			case <-ctx.Done():
				return
			}

			if genkitChunk.Done {
				s.logger.Infof("Streaming content generation completed (ID: %s)", generationID)
				return
			}
		}
	}()

	return outputCh, nil
}

// InterruptGeneration interrupts ongoing content generation
func (s *contentService) InterruptGeneration(ctx context.Context, requestID string) error {
	s.logger.Infof("Interrupting content generation (Request ID: %s)", requestID)

	// Use Genkit's interrupt manager
	interruptManager := s.genkitSvc.GetInterruptManager()
	if interruptManager == nil {
		return fmt.Errorf("interrupt manager not available")
	}

	err := interruptManager.Interrupt(requestID, "User requested interruption")
	if err != nil {
		s.logger.Errorf("Failed to interrupt generation: %v", err)
		return fmt.Errorf("failed to interrupt generation: %w", err)
	}

	s.logger.Infof("Content generation interrupted (Request ID: %s)", requestID)
	return nil
}

// GetGeneration retrieves a generation by ID
func (s *contentService) GetGeneration(ctx context.Context, id string) (*types.Generation, error) {
	// In a real implementation, this would query the database
	// For now, return a mock response
	generation := &types.Generation{
		ID:        id,
		Model:     "mock-model",
		Prompt:    "Mock prompt",
		Response:  "Mock response",
		Status:    "completed",
		CreatedAt: time.Now(),
	}

	return generation, nil
}

// ListGenerations lists generations with optional filtering
func (s *contentService) ListGenerations(ctx context.Context, filter GenerationFilter) ([]*types.Generation, error) {
	// In a real implementation, this would query the database with filters
	// For now, return mock data
	generations := []*types.Generation{
		{
			ID:        uuid.New().String(),
			Model:     "gpt-4",
			Prompt:    "What is AI?",
			Response:  "AI is artificial intelligence...",
			Status:    "completed",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			Model:     "claude-3-sonnet",
			Prompt:    "Explain machine learning",
			Response:  "Machine learning is a subset of AI...",
			Status:    "completed",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	// Apply filters
	var filtered []*types.Generation
	for _, gen := range generations {
		if filter.Model != "" && gen.Model != filter.Model {
			continue
		}
		if filter.Status != "" && gen.Status != filter.Status {
			continue
		}
		if filter.RequestID != "" && gen.RequestID != filter.RequestID {
			continue
		}
		filtered = append(filtered, gen)
	}

	// Apply pagination
	if filter.Limit > 0 {
		start := filter.Offset
		end := start + filter.Limit
		if start >= len(filtered) {
			return []*types.Generation{}, nil
		}
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	return filtered, nil
}