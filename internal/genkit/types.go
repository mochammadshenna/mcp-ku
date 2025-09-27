package genkit

import (
	"context"
	"mcp-octo-enigma/internal/types"
)

// GenerateContentRequest represents a content generation request
type GenerateContentRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Tools       []types.Tool           `json:"tools,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
}

// GenerateContentResponse represents a content generation response
type GenerateContentResponse struct {
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ToolCalls  []types.ToolCall       `json:"tool_calls,omitempty"`
	Usage      *UsageInfo             `json:"usage,omitempty"`
	Model      string                 `json:"model"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// GenerateContentChunk represents a streaming content chunk
type GenerateContentChunk struct {
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Done      bool                   `json:"done"`
	RequestID string                 `json:"request_id,omitempty"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIProvider interface for AI model providers
type AIProvider interface {
	GenerateContent(ctx context.Context, req *GenerateContentRequest) (*GenerateContentResponse, error)
	GenerateContentStream(ctx context.Context, req *GenerateContentRequest) (<-chan *GenerateContentChunk, error)
	EmbedText(ctx context.Context, text string) ([]float64, error)
	ListModels(ctx context.Context) ([]string, error)
	Close() error
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
	FlowID    string                 `json:"flow_id"`
	Output    map[string]interface{} `json:"output"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Status    string                 `json:"status"`
}

// PromptRenderRequest represents a prompt rendering request
type PromptRenderRequest struct {
	PromptID   string                 `json:"prompt_id"`
	Variables  map[string]interface{} `json:"variables"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// PromptRenderResponse represents a prompt rendering response
type PromptRenderResponse struct {
	PromptID   string `json:"prompt_id"`
	Rendered   string `json:"rendered"`
	RequestID  string `json:"request_id,omitempty"`
}

// ToolCallRequest represents a tool call request
type ToolCallRequest struct {
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// ToolCallResponse represents a tool call response
type ToolCallResponse struct {
	ToolName  string                 `json:"tool_name"`
	Result    map[string]interface{} `json:"result"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}