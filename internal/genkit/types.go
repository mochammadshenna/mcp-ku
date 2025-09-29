package genkit

import (
	"context"
	"time"
)

// AIProvider interface for AI providers
type AIProvider interface {
	GenerateContent(ctx context.Context, req *GenerateContentRequest) (*GenerateContentResponse, error)
	GenerateContentStream(ctx context.Context, req *GenerateContentRequest) (<-chan *GenerateContentChunk, error)
	EmbedText(ctx context.Context, text string) ([]float64, error)
	ListModels(ctx context.Context) ([]string, error)
	Close() error
}

// GenerateContentRequest represents a content generation request
type GenerateContentRequest struct {
	Model      string                 `json:"model"`
	Prompt     string                 `json:"prompt"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Tools      []Tool                 `json:"tools,omitempty"`
	Stream     bool                   `json:"stream,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// GenerateContentResponse represents a content generation response
type GenerateContentResponse struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Model       string                 `json:"model"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ToolCalls   []ToolCall             `json:"tool_calls,omitempty"`
	Usage       *UsageInfo             `json:"usage,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// GenerateContentChunk represents a streaming content chunk
type GenerateContentChunk struct {
	Content   string                 `json:"content"`
	Done      bool                   `json:"done"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Tool represents a callable tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required,omitempty"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallRequest represents a request to call a tool
type ToolCallRequest struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ToolCallResponse represents a response from a tool call
type ToolCallResponse struct {
	ToolName    string                 `json:"tool_name"`
	Result      map[string]interface{} `json:"result"`
	RequestID   string                 `json:"request_id,omitempty"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// FlowDefinition represents a flow definition
type FlowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Steps       []FlowStep             `json:"steps"`
	Config      map[string]interface{} `json:"config"`
}

// FlowStep represents a step in a flow
type FlowStep struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // "generate", "tool", "prompt", etc.
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies,omitempty"`
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

// PromptDefinition represents a prompt definition
type PromptDefinition struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Template  string                 `json:"template"`
	Variables []string               `json:"variables"`
	Config    map[string]interface{} `json:"config"`
	Version   int                    `json:"version"`
}

// PromptRequest represents a prompt rendering request
type PromptRequest struct {
	PromptName string                 `json:"prompt_name"`
	Variables  map[string]interface{} `json:"variables"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// PromptResponse represents a prompt rendering response
type PromptResponse struct {
	PromptName string                 `json:"prompt_name"`
	Rendered   string                 `json:"rendered"`
	RequestID  string                 `json:"request_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// InterruptRequestType represents an interrupt request
type InterruptRequestType struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason,omitempty"`
}

// InterruptResponseType represents an interrupt response
type InterruptResponseType struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"` // "interrupted", "not_found", "already_completed"
	Message   string `json:"message,omitempty"`
}

// EvaluationRequest represents an evaluation request
type EvaluationRequest struct {
	EvaluationID string                 `json:"evaluation_id"`
	GenerationID string                 `json:"generation_id"`
	Config       map[string]interface{} `json:"config"`
	RequestID    string                 `json:"request_id,omitempty"`
}

// EvaluationResponse represents an evaluation response
type EvaluationResponse struct {
	EvaluationID string                 `json:"evaluation_id"`
	GenerationID string                 `json:"generation_id"`
	Score        float64                `json:"score"`
	Metrics      map[string]interface{} `json:"metrics"`
	Details      map[string]interface{} `json:"details"`
	RequestID    string                 `json:"request_id,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// VectorSearchRequest represents a vector search request
type VectorSearchRequest struct {
	Query     []float64 `json:"query"`
	Limit     int       `json:"limit"`
	Threshold float64   `json:"threshold"`
}

// VectorSearchResponse represents a vector search response
type VectorSearchResponse struct {
	Results []VectorSearchResult `json:"results"`
	Total   int                  `json:"total"`
}

// VectorSearchResult represents a vector search result
type VectorSearchResult struct {
	Document *VectorDocument `json:"document"`
	Score    float64         `json:"score"`
}

// VectorDocument represents a document with vector embeddings
type VectorDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Embedding []float64              `json:"embedding"`
	Metadata  map[string]interface{} `json:"metadata"`
	Source    string                 `json:"source"`
	CreatedAt time.Time              `json:"created_at"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Text  string `json:"text"`
	Model string `json:"model"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// ModelInfo represents information about an AI model
type ModelInfo struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // "text", "embedding", "image", etc.
	Provider     string                 `json:"provider"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config"`
}

// ProviderInfo represents information about an AI provider
type ProviderInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Models      []ModelInfo            `json:"models"`
	Status      string                 `json:"status"` // "active", "inactive", "error"
	Config      map[string]interface{} `json:"config"`
}
