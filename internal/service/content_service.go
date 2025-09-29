package service

import (
	"context"
	"fmt"
	"time"

	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/repository"
	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ContentService provides methods for content generation and RAG
type ContentService struct {
	genkitService  *genkit.Service
	mcpManager     interface{} // MCP Manager interface
	generationRepo repository.GenerationRepository
	logger         *logrus.Logger
}

// NewContentService creates a new ContentService
func NewContentService(gs *genkit.Service, mcpManager interface{}, generationRepo repository.GenerationRepository, logger *logrus.Logger) *ContentService {
	return &ContentService{
		genkitService:  gs,
		mcpManager:     mcpManager,
		generationRepo: generationRepo,
		logger:         logger,
	}
}

// ContentGenerationRequest represents a content generation request
type ContentGenerationRequest struct {
	Model      string                 `json:"model"`
	Prompt     string                 `json:"prompt"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Tools      []types.Tool           `json:"tools,omitempty"`
	Stream     bool                   `json:"stream,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// ContentGenerationResponse represents a content generation response
type ContentGenerationResponse struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Model       string                 `json:"model"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ToolCalls   []types.ToolCall       `json:"tool_calls,omitempty"`
	Usage       *UsageInfo             `json:"usage,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ContentGenerationChunk represents a streaming content chunk
type ContentGenerationChunk struct {
	Content   string                 `json:"content"`
	Done      bool                   `json:"done"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// GenerationFilter for filtering generation results
type GenerationFilter struct {
	Model     string
	Status    string
	RequestID string
	Limit     int
	Offset    int
}

// GenerateContent generates content using the specified model
func (s *ContentService) GenerateContent(ctx context.Context, req *ContentGenerationRequest) (*ContentGenerationResponse, error) {
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Create generation record
	generation := &types.Generation{
		ID:         uuid.New().String(),
		Model:      req.Model,
		Prompt:     req.Prompt,
		Status:     "generating",
		RequestID:  req.RequestID,
		Parameters: req.Parameters,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	// Save to database
	if err := s.generationRepo.CreateGeneration(generation); err != nil {
		s.logger.Errorf("Failed to create generation record: %v", err)
	}

	// Convert tools
	var genkitTools []genkit.Tool
	for _, tool := range req.Tools {
		genkitTools = append(genkitTools, genkit.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
			Required:    tool.Required,
		})
	}

	// Create Genkit request
	genkitReq := &genkit.GenerateContentRequest{
		Model:      req.Model,
		Prompt:     req.Prompt,
		Parameters: req.Parameters,
		Tools:      genkitTools,
		RequestID:  req.RequestID,
	}

	// Generate content
	genkitResp, err := s.genkitService.GenerateContent(ctx, genkitReq)
	if err != nil {
		generation.Status = "failed"
		generation.CompletedAt = &[]time.Time{time.Now()}[0]
		s.generationRepo.UpdateGeneration(generation)
		return nil, fmt.Errorf("content generation failed: %w", err)
	}

	// Update generation record
	generation.Response = genkitResp.Content
	generation.Status = "completed"
	generation.CompletedAt = &[]time.Time{time.Now()}[0]
	s.generationRepo.UpdateGeneration(generation)

	// Convert tool calls
	var toolCalls []types.ToolCall
	for _, toolCall := range genkitResp.ToolCalls {
		toolCalls = append(toolCalls, types.ToolCall{
			ID:        toolCall.ID,
			Name:      toolCall.Name,
			Arguments: toolCall.Arguments,
		})
	}

	response := &ContentGenerationResponse{
		ID:          genkitResp.ID,
		Content:     genkitResp.Content,
		Model:       genkitResp.Model,
		Metadata:    genkitResp.Metadata,
		ToolCalls:   toolCalls,
		Usage:       (*UsageInfo)(genkitResp.Usage),
		RequestID:   genkitResp.RequestID,
		Status:      genkitResp.Status,
		CreatedAt:   genkitResp.CreatedAt,
		CompletedAt: genkitResp.CompletedAt,
	}

	return response, nil
}

// GenerateContentStream generates content with streaming
func (s *ContentService) GenerateContentStream(ctx context.Context, req *ContentGenerationRequest) (<-chan *ContentGenerationChunk, error) {
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Create generation record
	generation := &types.Generation{
		ID:         uuid.New().String(),
		Model:      req.Model,
		Prompt:     req.Prompt,
		Status:     "generating",
		RequestID:  req.RequestID,
		Parameters: req.Parameters,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	// Save to database
	if err := s.generationRepo.CreateGeneration(generation); err != nil {
		s.logger.Errorf("Failed to create generation record: %v", err)
	}

	// Convert tools
	var genkitTools []genkit.Tool
	for _, tool := range req.Tools {
		genkitTools = append(genkitTools, genkit.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
			Required:    tool.Required,
		})
	}

	// Create Genkit request
	genkitReq := &genkit.GenerateContentRequest{
		Model:      req.Model,
		Prompt:     req.Prompt,
		Parameters: req.Parameters,
		Tools:      genkitTools,
		Stream:     true,
		RequestID:  req.RequestID,
	}

	// Generate streaming content
	streamCh, err := s.genkitService.GenerateContentStream(ctx, genkitReq)
	if err != nil {
		generation.Status = "failed"
		generation.CompletedAt = &[]time.Time{time.Now()}[0]
		s.generationRepo.UpdateGeneration(generation)
		return nil, fmt.Errorf("streaming content generation failed: %w", err)
	}

	// Create response channel
	responseCh := make(chan *ContentGenerationChunk)

	go func() {
		defer close(responseCh)
		defer func() {
			generation.Status = "completed"
			generation.CompletedAt = &[]time.Time{time.Now()}[0]
			s.generationRepo.UpdateGeneration(generation)
		}()

		for chunk := range streamCh {
			responseChunk := &ContentGenerationChunk{
				Content:   chunk.Content,
				Done:      chunk.Done,
				Metadata:  chunk.Metadata,
				RequestID: chunk.RequestID,
			}
			responseCh <- responseChunk

			if chunk.Done {
				break
			}
		}
	}()

	return responseCh, nil
}

// InterruptGeneration interrupts ongoing content generation
func (s *ContentService) InterruptGeneration(ctx context.Context, requestID string) error {
	interruptManager := s.genkitService.GetInterruptManager()
	if interruptManager == nil {
		return fmt.Errorf("interrupt manager not available")
	}

	return interruptManager.InterruptGeneration(ctx, requestID)
}

// GetGeneration retrieves a generation by ID
func (s *ContentService) GetGeneration(ctx context.Context, generationID string) (*types.Generation, error) {
	return s.generationRepo.GetGeneration(generationID)
}

// GetGenerationByRequestID retrieves a generation by request ID
func (s *ContentService) GetGenerationByRequestID(ctx context.Context, requestID string) (*types.Generation, error) {
	return s.generationRepo.GetGenerationByRequestID(requestID)
}

// DeleteGeneration deletes a generation
func (s *ContentService) DeleteGeneration(ctx context.Context, generationID string) error {
	return s.generationRepo.DeleteGeneration(generationID)
}

// ListGenerations lists generations with filtering
func (s *ContentService) ListGenerations(ctx context.Context, filter GenerationFilter) ([]*types.Generation, error) {
	repoFilter := repository.GenerationFilter{
		Model:     filter.Model,
		Status:    filter.Status,
		RequestID: filter.RequestID,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}
	return s.generationRepo.ListGenerations(repoFilter)
}
