package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mcp-octo-enigma/internal/config"
	"mcp-octo-enigma/internal/types"

	"github.com/sirupsen/logrus"
)

// MCPClient represents the MCP client
type MCPClient struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
	logger     *logrus.Logger
}

// GenerateRequest represents a content generation request
type GenerateRequest struct {
	Model      string                 `json:"model"`
	Prompt     string                 `json:"prompt"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Tools      []types.Tool           `json:"tools,omitempty"`
	Stream     bool                   `json:"stream,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// GenerateResponse represents a content generation response
type GenerateResponse struct {
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

// FlowRequest represents a flow execution request
type FlowRequest struct {
	FlowID     string                 `json:"flow_id"`
	Input      map[string]interface{} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// FlowResponse represents a flow execution response
type FlowResponse struct {
	FlowID    string                 `json:"flow_id"`
	Output    map[string]interface{} `json:"output"`
	RequestID string                 `json:"request_id,omitempty"`
	Status    string                 `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Done      bool                   `json:"done"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(cfg *config.Config) *MCPClient {
	logger := logrus.New()
	logger.SetLevel(cfg.Logger.Level)

	return &MCPClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Server.ClientTimeout,
		},
		baseURL: fmt.Sprintf("http://localhost:%s/api/v1", cfg.Server.Port),
		logger:  logger,
	}
}

// Connect establishes connection to the MCP server
func (c *MCPClient) Connect(ctx context.Context) error {
	// Test connection with health check
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/../health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MCP server health check failed with status: %d", resp.StatusCode)
	}

	c.logger.Info("Connected to MCP server")
	return nil
}

// GenerateContent generates content using the MCP server
func (c *MCPClient) GenerateContent(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	result, err := c.makeRequest(ctx, "POST", "/content/generate", req, &GenerateResponse{})
	if err != nil {
		return nil, err
	}
	return result.(*GenerateResponse), nil
}

// GenerateContentStream generates content with streaming
func (c *MCPClient) GenerateContentStream(ctx context.Context, req *GenerateRequest) (<-chan *StreamChunk, error) {
	req.Stream = true

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/content/generate/stream", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.Security.SecretKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	ch := make(chan *StreamChunk)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		decoder := json.NewDecoder(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var chunk StreamChunk
				if err := decoder.Decode(&chunk); err != nil {
					if err == io.EOF {
						return
					}
					c.logger.Errorf("Failed to decode stream chunk: %v", err)
					continue
				}

				ch <- &chunk

				if chunk.Done {
					return
				}
			}
		}
	}()

	return ch, nil
}

// ExecuteFlow executes a flow on the MCP server
func (c *MCPClient) ExecuteFlow(ctx context.Context, req *FlowRequest) (*FlowResponse, error) {
	result, err := c.makeRequest(ctx, "POST", "/flows/"+req.FlowID+"/execute", req, &FlowResponse{})
	if err != nil {
		return nil, err
	}
	return result.(*FlowResponse), nil
}

// CallTool calls a tool on the MCP server
func (c *MCPClient) CallTool(ctx context.Context, toolCall *types.ToolCall) (*types.ToolResult, error) {
	result, err := c.makeRequest(ctx, "POST", "/tools/call", toolCall, &types.ToolResult{})
	if err != nil {
		return nil, err
	}
	return result.(*types.ToolResult), nil
}

// ListServers lists all registered MCP servers
func (c *MCPClient) ListServers(ctx context.Context) ([]types.MCPServer, error) {
	var servers []types.MCPServer
	_, err := c.makeRequest(ctx, "GET", "/mcp/servers", nil, &servers)
	return servers, err
}

// RegisterServer registers a new MCP server
func (c *MCPClient) RegisterServer(ctx context.Context, server *types.MCPServer) error {
	_, err := c.makeRequest(ctx, "POST", "/mcp/servers", server, nil)
	return err
}

// UnregisterServer unregisters an MCP server
func (c *MCPClient) UnregisterServer(ctx context.Context, serverID string) error {
	_, err := c.makeRequest(ctx, "DELETE", "/mcp/servers/"+serverID, nil, nil)
	return err
}

// InterruptGeneration interrupts ongoing content generation
func (c *MCPClient) InterruptGeneration(ctx context.Context, requestID string) error {
	req := map[string]string{"request_id": requestID}
	_, err := c.makeRequest(ctx, "POST", "/content/interrupt", req, nil)
	return err
}

// EmbedText embeds text using vector embeddings
func (c *MCPClient) EmbedText(ctx context.Context, text string, model string) ([]float64, error) {
	req := map[string]string{"text": text, "model": model}
	var response map[string][]float64
	_, err := c.makeRequest(ctx, "POST", "/vectors/embed", req, &response)
	if err != nil {
		return nil, err
	}
	return response["embedding"], nil
}

// SearchVectors searches for similar vectors
func (c *MCPClient) SearchVectors(ctx context.Context, query []float64, limit int, threshold float64) ([]types.VectorSearchResult, error) {
	req := map[string]interface{}{
		"query":     query,
		"limit":     limit,
		"threshold": threshold,
	}
	var results []types.VectorSearchResult
	_, err := c.makeRequest(ctx, "POST", "/vectors/search", req, &results)
	return results, err
}

// IndexDocument indexes a document with vector embeddings
func (c *MCPClient) IndexDocument(ctx context.Context, content string, source string, metadata map[string]interface{}, model string) (string, error) {
	req := map[string]interface{}{
		"content":  content,
		"source":   source,
		"metadata": metadata,
		"model":    model,
	}
	var response map[string]string
	_, err := c.makeRequest(ctx, "POST", "/vectors/index", req, &response)
	if err != nil {
		return "", err
	}
	return response["document_id"], nil
}

// ListTools lists available tools
func (c *MCPClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	var response map[string][]types.Tool
	_, err := c.makeRequest(ctx, "GET", "/tools/", nil, &response)
	if err != nil {
		return nil, err
	}
	return response["tools"], nil
}

// GetHealth gets the health status of the server
func (c *MCPClient) GetHealth(ctx context.Context) (map[string]interface{}, error) {
	var health map[string]interface{}
	_, err := c.makeRequest(ctx, "GET", "/../health", nil, &health)
	return health, err
}

// GetMetrics gets system metrics
func (c *MCPClient) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	var metrics map[string]interface{}
	_, err := c.makeRequest(ctx, "GET", "/observability/metrics", nil, &metrics)
	return metrics, err
}

// Close closes the client connection
func (c *MCPClient) Close() error {
	c.logger.Info("MCP client closed")
	return nil
}

// makeRequest is a helper method for making HTTP requests
func (c *MCPClient) makeRequest(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) (interface{}, error) {
	var body io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.config.Security.SecretKey)

	c.logger.Debugf("Making %s request to %s", method, path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return respBody, nil
	}
	return nil, nil
}
