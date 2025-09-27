package genkit

import (
	"context"
	"fmt"
	"sync"

	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ToolManager manages tool registration and execution
type ToolManager struct {
	service *Service
	logger  *logrus.Logger
	tools   map[string]*ToolHandler
	mu      sync.RWMutex
}

// ToolHandler represents a tool handler function
type ToolHandler struct {
	Tool     *types.Tool
	Handler  ToolHandlerFunc
	Config   map[string]interface{}
}

// ToolHandlerFunc is the function signature for tool handlers
type ToolHandlerFunc func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)

// NewToolManager creates a new tool manager
func NewToolManager(service *Service, logger *logrus.Logger) (*ToolManager, error) {
	tm := &ToolManager{
		service: service,
		logger:  logger,
		tools:   make(map[string]*ToolHandler),
	}

	// Register built-in tools
	tm.registerBuiltinTools()

	return tm, nil
}

// registerBuiltinTools registers default tools
func (tm *ToolManager) registerBuiltinTools() {
	// Web search tool
	webSearchTool := &types.Tool{
		Name:        "web_search",
		Description: "Search the web for information",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results",
					"default":     5,
				},
			},
		},
		Required: []string{"query"},
	}

	tm.RegisterTool(webSearchTool, tm.webSearchHandler)

	// Calculator tool
	calcTool := &types.Tool{
		Name:        "calculator",
		Description: "Perform mathematical calculations",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "Mathematical expression to evaluate",
				},
			},
		},
		Required: []string{"expression"},
	}

	tm.RegisterTool(calcTool, tm.calculatorHandler)

	// Text analyzer tool
	textAnalyzerTool := &types.Tool{
		Name:        "text_analyzer",
		Description: "Analyze text for various properties",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to analyze",
				},
				"analysis_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of analysis: sentiment, readability, keywords",
					"enum":        []string{"sentiment", "readability", "keywords"},
					"default":     "sentiment",
				},
			},
		},
		Required: []string{"text"},
	}

	tm.RegisterTool(textAnalyzerTool, tm.textAnalyzerHandler)

	// Vector search tool
	vectorSearchTool := &types.Tool{
		Name:        "vector_search",
		Description: "Search documents using vector similarity",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results",
					"default":     5,
				},
				"threshold": map[string]interface{}{
					"type":        "number",
					"description": "Similarity threshold (0-1)",
					"default":     0.7,
				},
			},
		},
		Required: []string{"query"},
	}

	tm.RegisterTool(vectorSearchTool, tm.vectorSearchHandler)

	// Code generator tool
	codeGenTool := &types.Tool{
		Name:        "code_generator",
		Description: "Generate code in various programming languages",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Programming language",
					"enum":        []string{"python", "javascript", "go", "java", "rust"},
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of what the code should do",
				},
				"style": map[string]interface{}{
					"type":        "string",
					"description": "Coding style preference",
					"default":     "clean",
				},
			},
		},
		Required: []string{"language", "description"},
	}

	tm.RegisterTool(codeGenTool, tm.codeGeneratorHandler)
}

// RegisterTool registers a new tool
func (tm *ToolManager) RegisterTool(tool *types.Tool, handler ToolHandlerFunc) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tools[tool.Name]; exists {
		return fmt.Errorf("tool already registered: %s", tool.Name)
	}

	tm.tools[tool.Name] = &ToolHandler{
		Tool:    tool,
		Handler: handler,
	}

	tm.logger.Infof("Registered tool: %s", tool.Name)
	return nil
}

// GetTool retrieves a tool by name
func (tm *ToolManager) GetTool(name string) (*types.Tool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	handler, exists := tm.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return handler.Tool, nil
}

// ListTools returns all registered tools
func (tm *ToolManager) ListTools() []*types.Tool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]*types.Tool, 0, len(tm.tools))
	for _, handler := range tm.tools {
		tools = append(tools, handler.Tool)
	}

	return tools
}

// CallTool executes a tool with the given arguments
func (tm *ToolManager) CallTool(ctx context.Context, req *ToolCallRequest) (*ToolCallResponse, error) {
	tm.mu.RLock()
	handler, exists := tm.tools[req.ToolName]
	tm.mu.RUnlock()

	if !exists {
		return &ToolCallResponse{
			ToolName:  req.ToolName,
			Success:   false,
			Error:     fmt.Sprintf("tool not found: %s", req.ToolName),
			RequestID: req.RequestID,
		}, nil
	}

	tm.logger.Infof("Executing tool: %s", req.ToolName)

	// Execute the tool
	result, err := handler.Handler(ctx, req.Arguments)
	if err != nil {
		tm.logger.Errorf("Tool execution failed: %v", err)
		return &ToolCallResponse{
			ToolName:  req.ToolName,
			Success:   false,
			Error:     err.Error(),
			RequestID: req.RequestID,
		}, nil
	}

	response := &ToolCallResponse{
		ToolName:  req.ToolName,
		Result:    result,
		Success:   true,
		RequestID: req.RequestID,
	}

	tm.logger.Infof("Tool executed successfully: %s", req.ToolName)
	return response, nil
}

// Built-in tool handlers

// webSearchHandler handles web search requests
func (tm *ToolManager) webSearchHandler(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required and must be a string")
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Mock web search results
	results := []map[string]interface{}{
		{
			"title": "Mock Result 1",
			"url":   "https://example.com/1",
			"snippet": fmt.Sprintf("This is a mock search result for query: %s", query),
		},
		{
			"title": "Mock Result 2",
			"url":   "https://example.com/2",
			"snippet": fmt.Sprintf("Another mock search result about: %s", query),
		},
	}

	if limit < len(results) {
		results = results[:limit]
	}

	return map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	}, nil
}

// calculatorHandler handles mathematical calculations
func (tm *ToolManager) calculatorHandler(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	expression, ok := args["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression parameter is required and must be a string")
	}

	// Simple calculator - in production, use a proper expression evaluator
	// For now, return a mock result
	result := fmt.Sprintf("Result of '%s' = 42", expression)

	return map[string]interface{}{
		"expression": expression,
		"result":     result,
		"value":      42,
	}, nil
}

// textAnalyzerHandler handles text analysis
func (tm *ToolManager) textAnalyzerHandler(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter is required and must be a string")
	}

	analysisType := "sentiment"
	if at, ok := args["analysis_type"].(string); ok {
		analysisType = at
	}

	var analysis map[string]interface{}

	switch analysisType {
	case "sentiment":
		analysis = map[string]interface{}{
			"sentiment": "positive",
			"score":     0.8,
			"confidence": 0.9,
		}
	case "readability":
		analysis = map[string]interface{}{
			"reading_level": "college",
			"avg_sentence_length": 15.5,
			"complexity_score": 12.3,
		}
	case "keywords":
		analysis = map[string]interface{}{
			"keywords": []string{"example", "text", "analysis"},
			"topics":   []string{"technology", "AI"},
		}
	default:
		return nil, fmt.Errorf("unsupported analysis type: %s", analysisType)
	}

	return map[string]interface{}{
		"text":          text,
		"analysis_type": analysisType,
		"analysis":      analysis,
		"word_count":    len(text) / 5, // rough approximation
	}, nil
}

// vectorSearchHandler handles vector-based document search
func (tm *ToolManager) vectorSearchHandler(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required and must be a string")
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	threshold := 0.7
	if t, ok := args["threshold"].(float64); ok {
		threshold = t
	}

	// In a real implementation, this would:
	// 1. Generate embedding for the query
	// 2. Search vector database
	// 3. Return similar documents

	// Mock results
	results := []map[string]interface{}{
		{
			"id":      uuid.New().String(),
			"content": fmt.Sprintf("Document about %s", query),
			"score":   0.9,
			"source":  "document1.txt",
		},
		{
			"id":      uuid.New().String(),
			"content": fmt.Sprintf("Related content to %s", query),
			"score":   0.8,
			"source":  "document2.txt",
		},
	}

	return map[string]interface{}{
		"query":     query,
		"results":   results,
		"count":     len(results),
		"threshold": threshold,
	}, nil
}

// codeGeneratorHandler handles code generation
func (tm *ToolManager) codeGeneratorHandler(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	language, ok := args["language"].(string)
	if !ok {
		return nil, fmt.Errorf("language parameter is required and must be a string")
	}

	description, ok := args["description"].(string)
	if !ok {
		return nil, fmt.Errorf("description parameter is required and must be a string")
	}

	style := "clean"
	if s, ok := args["style"].(string); ok {
		style = s
	}

	// Generate mock code based on language and description
	var code string
	switch language {
	case "python":
		code = fmt.Sprintf(`# %s
def main():
    """
    %s
    """
    print("Hello, World!")
    return True

if __name__ == "__main__":
    main()`, description, description)
	case "javascript":
		code = fmt.Sprintf(`// %s
function main() {
    /**
     * %s
     */
    console.log("Hello, World!");
    return true;
}

main();`, description, description)
	case "go":
		code = fmt.Sprintf(`// %s
package main

import "fmt"

// main %s
func main() {
    fmt.Println("Hello, World!")
}`, description, description)
	default:
		code = fmt.Sprintf("// Generated code for: %s\n// Language: %s", description, language)
	}

	return map[string]interface{}{
		"language":    language,
		"description": description,
		"style":       style,
		"code":        code,
		"lines":       len(code) / 50, // rough approximation
	}, nil
}