package service

import (
	"context"
	"fmt"

	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/types"

	"github.com/sirupsen/logrus"
)

// ToolService provides methods for tool management
type ToolService struct {
	genkitService *genkit.Service
	mcpManager    interface{} // MCP Manager interface
	logger        *logrus.Logger
}

// NewToolService creates a new ToolService
func NewToolService(gs *genkit.Service, mcpManager interface{}, logger *logrus.Logger) *ToolService {
	return &ToolService{
		genkitService: gs,
		mcpManager:    mcpManager,
		logger:        logger,
	}
}

// RegisterTool registers a new tool
func (s *ToolService) RegisterTool(ctx context.Context, tool *types.Tool) error {
	toolManager := s.genkitService.GetToolManager()
	if toolManager == nil {
		return fmt.Errorf("tool manager not available")
	}

	// Convert to Genkit tool
	genkitTool := &genkit.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  tool.Parameters,
		Required:    tool.Required,
	}

	// Register in Genkit
	if err := toolManager.RegisterTool(ctx, genkitTool); err != nil {
		return fmt.Errorf("failed to register tool in Genkit: %w", err)
	}

	return nil
}

// ListTools returns all available tools
func (s *ToolService) ListTools(ctx context.Context) []types.Tool {
	toolManager := s.genkitService.GetToolManager()
	if toolManager == nil {
		return []types.Tool{}
	}

	genkitTools, err := toolManager.ListTools(ctx)
	if err != nil {
		s.logger.Errorf("Failed to list tools: %v", err)
		return []types.Tool{}
	}

	var tools []types.Tool
	for _, genkitTool := range genkitTools {
		tools = append(tools, types.Tool{
			Name:        genkitTool.Name,
			Description: genkitTool.Description,
			Parameters:  genkitTool.Parameters,
			Required:    genkitTool.Required,
		})
	}

	return tools
}

// CallTool calls a tool with the given arguments
func (s *ToolService) CallTool(ctx context.Context, req *genkit.ToolCallRequest) (*genkit.ToolCallResponse, error) {
	toolManager := s.genkitService.GetToolManager()
	if toolManager == nil {
		return nil, fmt.Errorf("tool manager not available")
	}

	return toolManager.CallTool(ctx, req)
}
