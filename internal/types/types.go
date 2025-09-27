package types

import (
	"time"
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	URL         string            `json:"url" db:"url"`
	Description string            `json:"description" db:"description"`
	Capabilities []string         `json:"capabilities" db:"capabilities"`
	Config      map[string]interface{} `json:"config" db:"config"`
	Status      string            `json:"status" db:"status"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
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

// ToolResult represents the result of a tool call
type ToolResult struct {
	ID      string                 `json:"id"`
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// Flow represents a Genkit flow
type Flow struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Input       map[string]interface{} `json:"input" db:"input"`
	Output      map[string]interface{} `json:"output" db:"output"`
	Steps       []FlowStep             `json:"steps" db:"steps"`
	Config      map[string]interface{} `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// FlowStep represents a step in a flow
type FlowStep struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "generate", "tool", "prompt", etc.
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Dependencies []string              `json:"dependencies,omitempty"`
}

// FlowExecution represents a flow execution instance
type FlowExecution struct {
	ID        string                 `json:"id" db:"id"`
	FlowID    string                 `json:"flow_id" db:"flow_id"`
	Input     map[string]interface{} `json:"input" db:"input"`
	Output    map[string]interface{} `json:"output" db:"output"`
	Status    string                 `json:"status" db:"status"` // "running", "completed", "failed", "interrupted"
	Error     string                 `json:"error,omitempty" db:"error"`
	StartedAt time.Time              `json:"started_at" db:"started_at"`
	EndedAt   *time.Time             `json:"ended_at,omitempty" db:"ended_at"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
}

// Prompt represents a dotprompt template
type Prompt struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Template    string                 `json:"template" db:"template"`
	Variables   []string               `json:"variables" db:"variables"`
	Config      map[string]interface{} `json:"config" db:"config"`
	Version     int                    `json:"version" db:"version"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// VectorDocument represents a document with vector embeddings
type VectorDocument struct {
	ID          string    `json:"id" db:"id"`
	Content     string    `json:"content" db:"content"`
	Embedding   []float64 `json:"embedding" db:"embedding"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Source      string    `json:"source" db:"source"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// VectorSearchResult represents a vector search result
type VectorSearchResult struct {
	Document *VectorDocument `json:"document"`
	Score    float64         `json:"score"`
}

// Generation represents a content generation request/response
type Generation struct {
	ID          string                 `json:"id" db:"id"`
	Model       string                 `json:"model" db:"model"`
	Prompt      string                 `json:"prompt" db:"prompt"`
	Response    string                 `json:"response" db:"response"`
	Parameters  map[string]interface{} `json:"parameters" db:"parameters"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Status      string                 `json:"status" db:"status"` // "pending", "generating", "completed", "failed", "interrupted"
	RequestID   string                 `json:"request_id" db:"request_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// Evaluation represents an evaluation configuration
type Evaluation struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Type        string                 `json:"type" db:"type"` // "quality", "safety", "performance", etc.
	Config      map[string]interface{} `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// EvaluationResult represents the result of an evaluation
type EvaluationResult struct {
	ID           string                 `json:"id" db:"id"`
	EvaluationID string                 `json:"evaluation_id" db:"evaluation_id"`
	GenerationID string                 `json:"generation_id" db:"generation_id"`
	Score        float64                `json:"score" db:"score"`
	Metrics      map[string]interface{} `json:"metrics" db:"metrics"`
	Details      map[string]interface{} `json:"details" db:"details"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// Metric represents an observability metric
type Metric struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"` // "counter", "gauge", "histogram", etc.
}

// Trace represents a distributed trace
type Trace struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Duration  *time.Duration         `json:"duration,omitempty"`
	Status    string                 `json:"status"` // "ok", "error", "timeout"
	Tags      map[string]string      `json:"tags"`
	Logs      []TraceLog             `json:"logs"`
	Children  []Trace                `json:"children,omitempty"`
}

// TraceLog represents a log entry within a trace
type TraceLog struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
}