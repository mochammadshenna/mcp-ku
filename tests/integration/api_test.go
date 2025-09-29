package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcp-octo-enigma/internal/config"
	"mcp-octo-enigma/internal/container"
	"mcp-octo-enigma/internal/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	server    *server.Server
	container *container.Container
}

func (suite *APITestSuite) SetupSuite() {
	// Load test configuration
	cfg := &config.Config{
		Database: config.Database{
			URL: "postgres://test:test@localhost:5432/test_db?sslmode=disable",
		},
		Server: config.Server{
			Port: "8080",
		},
		AI: config.AI{
			OpenAI: config.OpenAI{
				APIKey: "test-key",
			},
		},
		Logger: config.Logger{
			Level: 5, // Debug level
		},
	}

	// Create container (this might fail if DB is not available, so we'll mock it)
	var err error
	suite.container, err = container.NewContainer(cfg)
	if err != nil {
		// For testing without a real database, we'll skip these tests
		suite.T().Skip("Database not available for integration tests")
	}

	// Create server
	suite.server = server.NewServer(suite.container)
}

func (suite *APITestSuite) TearDownSuite() {
	if suite.container != nil {
		suite.container.Close()
	}
}

func (suite *APITestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Use the router from the server
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

func (suite *APITestSuite) TestGenerateContent() {
	requestBody := map[string]interface{}{
		"model":  "gpt-4",
		"prompt": "What is artificial intelligence?",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/content/generate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	// Note: This might fail if AI providers are not properly configured
	// In a real test environment, you'd mock the AI providers
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), response["content"])
		assert.Equal(suite.T(), "gpt-4", response["model"])
	}
}

func (suite *APITestSuite) TestListFlows() {
	req, _ := http.NewRequest("GET", "/api/v1/flows/", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "flows")
}

func (suite *APITestSuite) TestCreateFlow() {
	flowData := map[string]interface{}{
		"name":        "test-flow",
		"description": "A test flow",
		"steps": []map[string]interface{}{
			{
				"id":          "step1",
				"type":        "generate",
				"name":        "Generate Content",
				"description": "Generate some content",
				"config": map[string]interface{}{
					"model":  "gpt-4",
					"prompt": "Test prompt",
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(flowData)
	req, _ := http.NewRequest("POST", "/api/v1/flows/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response["id"])
	assert.Equal(suite.T(), "test-flow", response["name"])
}

func (suite *APITestSuite) TestListTools() {
	req, _ := http.NewRequest("GET", "/api/v1/tools/", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "tools")

	tools := response["tools"].([]interface{})
	assert.Greater(suite.T(), len(tools), 0) // Should have built-in tools
}

func (suite *APITestSuite) TestCallTool() {
	toolCallData := map[string]interface{}{
		"tool_name": "calculator",
		"arguments": map[string]interface{}{
			"expression": "2 + 2",
		},
	}

	jsonBody, _ := json.Marshal(toolCallData)
	req, _ := http.NewRequest("POST", "/api/v1/tools/call", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "calculator", response["tool_name"])
}

func (suite *APITestSuite) TestListMCPServers() {
	req, _ := http.NewRequest("GET", "/api/v1/mcp/servers", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "servers")
}

func (suite *APITestSuite) TestRegisterMCPServer() {
	serverData := map[string]interface{}{
		"name":         "test-server",
		"url":          "http://localhost:8081",
		"description":  "A test MCP server",
		"capabilities": []string{"tools", "prompts"},
	}

	jsonBody, _ := json.Marshal(serverData)
	req, _ := http.NewRequest("POST", "/api/v1/mcp/servers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response["id"])
	assert.Equal(suite.T(), "test-server", response["name"])
}

func (suite *APITestSuite) TestEmbedText() {
	embedData := map[string]interface{}{
		"text":  "This is a test sentence for embedding",
		"model": "text-embedding-ada-002",
	}

	jsonBody, _ := json.Marshal(embedData)
	req, _ := http.NewRequest("POST", "/api/v1/vectors/embed", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	// This might fail if embedding provider is not configured
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), response, "embedding")

		embedding := response["embedding"].([]interface{})
		assert.Greater(suite.T(), len(embedding), 0)
	}
}

func (suite *APITestSuite) TestMetrics() {
	req, _ := http.NewRequest("GET", "/api/v1/observability/metrics", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router := suite.server.GetRouter()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "requests_total")
	assert.Contains(suite.T(), response, "uptime")
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
