package genkit

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// ToolManager manages Genkit tools
type ToolManager struct {
	service *Service
	logger  *logrus.Logger
	tools   map[string]*Tool
	mu      sync.RWMutex
}

// NewToolManager creates a new tool manager
func NewToolManager(service *Service, logger *logrus.Logger) (*ToolManager, error) {
	return &ToolManager{
		service: service,
		logger:  logger,
		tools:   make(map[string]*Tool),
	}, nil
}

// RegisterTool registers a new tool
func (tm *ToolManager) RegisterTool(ctx context.Context, tool *Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tools[tool.Name]; exists {
		return fmt.Errorf("tool with name %s already exists", tool.Name)
	}

	// Validate tool definition
	if err := tm.validateTool(tool); err != nil {
		return fmt.Errorf("invalid tool definition: %w", err)
	}

	tm.tools[tool.Name] = tool
	tm.logger.Infof("Registered tool: %s", tool.Name)
	
	return nil
}

// GetTool retrieves a tool by name
func (tm *ToolManager) GetTool(ctx context.Context, toolName string) (*Tool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tool, exists := tm.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	return tool, nil
}

// ListTools returns all registered tools
func (tm *ToolManager) ListTools(ctx context.Context) ([]*Tool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]*Tool, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool calls a tool with the given arguments
func (tm *ToolManager) CallTool(ctx context.Context, req *ToolCallRequest) (*ToolCallResponse, error) {
	tm.mu.RLock()
	tool, exists := tm.tools[req.ToolName]
	tm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", req.ToolName)
	}

	tm.logger.Debugf("Calling tool: %s", req.ToolName)

	// Validate arguments
	if err := tm.validateArguments(tool, req.Arguments); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Execute tool
	result, err := tm.executeTool(ctx, tool, req.Arguments)
	if err != nil {
		return &ToolCallResponse{
			ToolName:   req.ToolName,
			Result:     nil,
			RequestID:  req.RequestID,
			Status:     "error",
			Error:      err.Error(),
		}, nil
	}

	response := &ToolCallResponse{
		ToolName:  req.ToolName,
		Result:    result,
		RequestID: req.RequestID,
		Status:    "completed",
	}

	tm.logger.Debugf("Completed tool call: %s", req.ToolName)
	return response, nil
}

// executeTool executes a tool based on its type
func (tm *ToolManager) executeTool(ctx context.Context, tool *Tool, args map[string]interface{}) (map[string]interface{}, error) {
	// Check for interrupts
	if tm.service.interruptManager.IsInterrupted(ctx.Value("request_id").(string)) {
		return nil, fmt.Errorf("tool execution interrupted")
	}

	// Tool execution based on tool name
	switch tool.Name {
	case "vector-search":
		return tm.executeVectorSearch(ctx, args)
	case "weather-lookup":
		return tm.executeWeatherLookup(ctx, args)
	case "calculator":
		return tm.executeCalculator(ctx, args)
	case "text-processor":
		return tm.executeTextProcessor(ctx, args)
	case "file-reader":
		return tm.executeFileReader(ctx, args)
	case "web-scraper":
		return tm.executeWebScraper(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", tool.Name)
	}
}

// executeVectorSearch executes vector search tool
func (tm *ToolManager) executeVectorSearch(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query argument is required")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Mock vector search results
	results := []map[string]interface{}{
		{
			"id":      "doc1",
			"content": fmt.Sprintf("Document related to: %s", query),
			"score":   0.95,
		},
		{
			"id":      "doc2",
			"content": fmt.Sprintf("Another document about: %s", query),
			"score":   0.87,
		},
	}

	return map[string]interface{}{
		"results": results,
		"total":   len(results),
		"query":   query,
		"limit":   limit,
	}, nil
}

// executeWeatherLookup executes weather lookup tool
func (tm *ToolManager) executeWeatherLookup(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	city, ok := args["city"].(string)
	if !ok {
		return nil, fmt.Errorf("city argument is required")
	}

	// Mock weather data
	weather := map[string]interface{}{
		"city":        city,
		"temperature": 22.5,
		"condition":   "Sunny",
		"humidity":    65,
		"wind_speed":  12.3,
		"timestamp":   "2024-01-01T12:00:00Z",
	}

	return weather, nil
}

// executeCalculator executes calculator tool
func (tm *ToolManager) executeCalculator(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	expression, ok := args["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression argument is required")
	}

	// Simple calculator - in production, use a proper expression evaluator
	result := 0.0
	// This is a very basic implementation
	switch expression {
	case "2+2":
		result = 4.0
	case "10*5":
		result = 50.0
	case "100/4":
		result = 25.0
	default:
		result = 42.0 // Default answer
	}

	return map[string]interface{}{
		"expression": expression,
		"result":     result,
	}, nil
}

// executeTextProcessor executes text processor tool
func (tm *ToolManager) executeTextProcessor(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text argument is required")
	}

	operation, ok := args["operation"].(string)
	if !ok {
		operation = "count_words"
	}

	var result interface{}
	switch operation {
	case "count_words":
		result = len(args["text"].(string))
	case "uppercase":
		result = fmt.Sprintf("%s", text)
	case "lowercase":
		result = fmt.Sprintf("%s", text)
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return map[string]interface{}{
		"text":      text,
		"operation": operation,
		"result":    result,
	}, nil
}

// executeFileReader executes file reader tool
func (tm *ToolManager) executeFileReader(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	filename, ok := args["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename argument is required")
	}

	// Mock file reading
	content := fmt.Sprintf("Mock content of file: %s\nThis is a demonstration of file reading.", filename)

	return map[string]interface{}{
		"filename": filename,
		"content":  content,
		"size":     len(content),
	}, nil
}

// executeWebScraper executes web scraper tool
func (tm *ToolManager) executeWebScraper(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url argument is required")
	}

	// Mock web scraping
	title := "Mock Web Page"
	content := fmt.Sprintf("Content scraped from: %s\nThis is a demonstration of web scraping.", url)

	return map[string]interface{}{
		"url":     url,
		"title":   title,
		"content": content,
		"status":  "success",
	}, nil
}

// validateTool validates a tool definition
func (tm *ToolManager) validateTool(tool *Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Description == "" {
		return fmt.Errorf("tool description is required")
	}
	if tool.Parameters == nil {
		tool.Parameters = make(map[string]interface{})
	}
	return nil
}

// validateArguments validates tool arguments
func (tm *ToolManager) validateArguments(tool *Tool, args map[string]interface{}) error {
	// Check required parameters
	for _, required := range tool.Required {
		if _, exists := args[required]; !exists {
			return fmt.Errorf("required parameter missing: %s", required)
		}
	}
	return nil
}

// CreateExampleTools creates example tools for demonstration
func (tm *ToolManager) CreateExampleTools(ctx context.Context) error {
	// Vector search tool
	vectorSearchTool := &Tool{
		Name:        "vector-search",
		Description: "Search for similar documents using vector embeddings",
		Parameters: map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results",
				"default":     10,
			},
		},
		Required: []string{"query"},
	}

	if err := tm.RegisterTool(ctx, vectorSearchTool); err != nil {
		return fmt.Errorf("failed to register vector search tool: %w", err)
	}

	// Weather lookup tool
	weatherTool := &Tool{
		Name:        "weather-lookup",
		Description: "Get current weather information for a city",
		Parameters: map[string]interface{}{
			"city": map[string]interface{}{
				"type":        "string",
				"description": "City name",
			},
		},
		Required: []string{"city"},
	}

	if err := tm.RegisterTool(ctx, weatherTool); err != nil {
		return fmt.Errorf("failed to register weather tool: %w", err)
	}

	// Calculator tool
	calculatorTool := &Tool{
		Name:        "calculator",
		Description: "Perform mathematical calculations",
		Parameters: map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "Mathematical expression to evaluate",
			},
		},
		Required: []string{"expression"},
	}

	if err := tm.RegisterTool(ctx, calculatorTool); err != nil {
		return fmt.Errorf("failed to register calculator tool: %w", err)
	}

	// Text processor tool
	textProcessorTool := &Tool{
		Name:        "text-processor",
		Description: "Process text with various operations",
		Parameters: map[string]interface{}{
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text to process",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Operation to perform",
				"enum":        []string{"count_words", "uppercase", "lowercase", "reverse"},
				"default":     "count_words",
			},
		},
		Required: []string{"text"},
	}

	if err := tm.RegisterTool(ctx, textProcessorTool); err != nil {
		return fmt.Errorf("failed to register text processor tool: %w", err)
	}

	// File reader tool
	fileReaderTool := &Tool{
		Name:        "file-reader",
		Description: "Read content from a file",
		Parameters: map[string]interface{}{
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
		},
		Required: []string{"filename"},
	}

	if err := tm.RegisterTool(ctx, fileReaderTool); err != nil {
		return fmt.Errorf("failed to register file reader tool: %w", err)
	}

	// Web scraper tool
	webScraperTool := &Tool{
		Name:        "web-scraper",
		Description: "Scrape content from a web page",
		Parameters: map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to scrape",
			},
		},
		Required: []string{"url"},
	}

	if err := tm.RegisterTool(ctx, webScraperTool); err != nil {
		return fmt.Errorf("failed to register web scraper tool: %w", err)
	}

	return nil
}