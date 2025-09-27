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
)

// MCPClient represents the MCP client
type MCPClient struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
}

// GenerateRequest represents a content generation request
type GenerateRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Tools       []types.Tool           `json:"tools,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
}

// GenerateResponse represents a content generation response
type GenerateResponse struct {
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ToolCalls []types.ToolCall       `json:"tool_calls,omitempty"`
}

// FlowRequest represents a flow execution request
type FlowRequest struct {
	FlowID     string                 `json:"flow_id"`
	Input      map[string]interface{} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// FlowResponse represents a flow execution response
type FlowResponse struct {
	Output   map[string]interface{} `json:"output"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(cfg *config.Config) *MCPClient {
	return &MCPClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Server.ClientTimeout,
		},
		baseURL: fmt.Sprintf("http://localhost:%s/api/v1", cfg.Server.Port),
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

	return nil
}

// GenerateContent generates content using the MCP server
func (c *MCPClient) GenerateContent(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	return c.makeRequest(ctx, "POST", "/content/generate", req, &GenerateResponse{})
}

// GenerateContentStream generates content with streaming
func (c *MCPClient) GenerateContentStream(ctx context.Context, req *GenerateRequest) (<-chan string, error) {
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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	ch := make(chan string)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		decoder := json.NewDecoder(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var chunk map[string]interface{}
				if err := decoder.Decode(&chunk); err != nil {
					if err == io.EOF {
						return
					}
					// Log error but continue
					continue
				}

				if content, ok := chunk["content"].(string); ok {
					ch <- content
				}
			}
		}
	}()

	return ch, nil
}

// ExecuteFlow executes a flow on the MCP server
func (c *MCPClient) ExecuteFlow(ctx context.Context, req *FlowRequest) (*FlowResponse, error) {
	return c.makeRequest(ctx, "POST", "/flows/"+req.FlowID+"/execute", req, &FlowResponse{})
}

// CallTool calls a tool on the MCP server
func (c *MCPClient) CallTool(ctx context.Context, toolCall *types.ToolCall) (*types.ToolResult, error) {
	return c.makeRequest(ctx, "POST", "/tools/call", toolCall, &types.ToolResult{})
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

// InterruptGeneration interrupts ongoing content generation
func (c *MCPClient) InterruptGeneration(ctx context.Context, requestID string) error {
	req := map[string]string{"request_id": requestID}
	_, err := c.makeRequest(ctx, "POST", "/content/interrupt", req, nil)
	return err
}

// EmbedText embeds text using vector embeddings
func (c *MCPClient) EmbedText(ctx context.Context, text string) ([]float64, error) {
	req := map[string]string{"text": text}
	var response map[string][]float64
	_, err := c.makeRequest(ctx, "POST", "/vectors/embed", req, &response)
	if err != nil {
		return nil, err
	}
	return response["embedding"], nil
}

// SearchVectors searches for similar vectors
func (c *MCPClient) SearchVectors(ctx context.Context, query []float64, limit int) ([]types.VectorSearchResult, error) {
	req := map[string]interface{}{
		"query": query,
		"limit": limit,
	}
	var results []types.VectorSearchResult
	_, err := c.makeRequest(ctx, "POST", "/vectors/search", req, &results)
	return results, err
}

// Close closes the client connection
func (c *MCPClient) Close() error {
	// Close any open connections
	return nil
}

// makeRequest is a helper method for making HTTP requests
func (c *MCPClient) makeRequest(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) (*http.Response, error) {
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return resp, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return resp, nil
}