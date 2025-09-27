package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Client represents an MCP client for communicating with MCP servers
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	connected  bool
}

// MCPRequest represents a generic MCP request
type MCPRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	ID     string      `json:"id,omitempty"`
}

// MCPResponse represents a generic MCP response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
	ID     string      `json:"id,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewClient creates a new MCP client
func NewClient(baseURL string, logger *logrus.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:    logger,
		connected: false,
	}
}

// Connect establishes connection to the MCP server
func (c *Client) Connect(ctx context.Context) error {
	// Send initialize request
	initReq := &MCPRequest{
		Method: "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools":     true,
				"prompts":   true,
				"resources": true,
			},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-octo-enigma",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(ctx, initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("MCP initialization error: %s", resp.Error.Message)
	}

	c.connected = true
	c.logger.Infof("Connected to MCP server: %s", c.baseURL)

	return nil
}

// Ping sends a ping request to check server health
func (c *Client) Ping(ctx context.Context) error {
	if !c.connected {
		return fmt.Errorf("client not connected")
	}

	pingReq := &MCPRequest{
		Method: "ping",
	}

	resp, err := c.sendRequest(ctx, pingReq)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("ping error: %s", resp.Error.Message)
	}

	return nil
}

// ListTools retrieves available tools from the MCP server
func (c *Client) ListTools(ctx context.Context) ([]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "tools/list",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	// Parse tools from response
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			return tools, nil
		}
	}

	return []interface{}{}, nil
}

// CallTool calls a specific tool on the MCP server
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// ListPrompts retrieves available prompts from the MCP server
func (c *Client) ListPrompts(ctx context.Context) ([]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "prompts/list",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list prompts error: %s", resp.Error.Message)
	}

	// Parse prompts from response
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if prompts, ok := result["prompts"].([]interface{}); ok {
			return prompts, nil
		}
	}

	return []interface{}{}, nil
}

// GetPrompt retrieves a specific prompt from the MCP server
func (c *Client) GetPrompt(ctx context.Context, promptName string, arguments map[string]interface{}) (interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "prompts/get",
		Params: map[string]interface{}{
			"name":      promptName,
			"arguments": arguments,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("get prompt error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// ListResources retrieves available resources from the MCP server
func (c *Client) ListResources(ctx context.Context) ([]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "resources/list",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list resources error: %s", resp.Error.Message)
	}

	// Parse resources from response
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if resources, ok := result["resources"].([]interface{}); ok {
			return resources, nil
		}
	}

	return []interface{}{}, nil
}

// ReadResource reads a specific resource from the MCP server
func (c *Client) ReadResource(ctx context.Context, uri string) (interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "resources/read",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("read resource error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// SendRequest sends a generic request to the MCP server
func (c *Client) SendRequest(ctx context.Context, request interface{}) (interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Convert generic request to MCP request format
	var mcpReq *MCPRequest
	if req, ok := request.(*MCPRequest); ok {
		mcpReq = req
	} else {
		mcpReq = &MCPRequest{
			Method: "custom",
			Params: request,
		}
	}

	resp, err := c.sendRequest(ctx, mcpReq)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("request error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// sendRequest sends an HTTP request to the MCP server
func (c *Client) sendRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
	// Serialize request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse MCP response
	var mcpResp MCPResponse
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &mcpResp, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if !c.connected {
		return nil
	}

	// Send disconnect notification if possible
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	disconnectReq := &MCPRequest{
		Method: "notifications/cancelled",
	}

	// Best effort - don't fail if this doesn't work
	_, _ = c.sendRequest(ctx, disconnectReq)

	c.connected = false
	c.logger.Infof("Disconnected from MCP server: %s", c.baseURL)

	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// GetServerCapabilities retrieves server capabilities
func (c *Client) GetServerCapabilities(ctx context.Context) (map[string]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	req := &MCPRequest{
		Method: "capabilities",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("capabilities error: %s", resp.Error.Message)
	}

	if capabilities, ok := resp.Result.(map[string]interface{}); ok {
		return capabilities, nil
	}

	return map[string]interface{}{}, nil
}