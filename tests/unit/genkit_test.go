package unit

import (
	"context"
	"testing"

	"mcp-octo-enigma/internal/config"
	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/types"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFlowManager_CreateFlow(t *testing.T) {
	// Setup
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a minimal service for the flow manager
	cfg := &config.Config{}
	service, err := genkit.NewService(cfg, logger)
	assert.NoError(t, err)

	flowManager := service.GetFlowManager()

	// Test data
	flow := &types.Flow{
		Name:        "test-flow",
		Description: "Test flow description",
		Steps: []types.FlowStep{
			{
				ID:          "step1",
				Type:        "validation",
				Name:        "Validate Input",
				Description: "Validate the input",
				Config: map[string]interface{}{
					"required_fields": []string{"input"},
				},
			},
		},
	}

	// Execute
	err = flowManager.CreateFlow(flow)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, flow.ID)

	// Verify flow was stored
	retrievedFlow, err := flowManager.GetFlow(flow.ID)
	assert.NoError(t, err)
	assert.Equal(t, flow.Name, retrievedFlow.Name)
	assert.Equal(t, flow.Description, retrievedFlow.Description)
}

func TestFlowManager_ExecuteFlow(t *testing.T) {
	// Setup
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	cfg := &config.Config{}
	service, err := genkit.NewService(cfg, logger)
	assert.NoError(t, err)

	flowManager := service.GetFlowManager()

	// Execute the built-in content generation flow
	req := &genkit.FlowExecutionRequest{
		FlowID: "content-generation",
		Input: map[string]interface{}{
			"prompt": "Test prompt",
			"model":  "gpt-4",
		},
	}

	ctx := context.Background()
	resp, err := flowManager.ExecuteFlow(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "content-generation", resp.FlowID)
	assert.Equal(t, "completed", resp.Status)
}

func TestPromptManager_CreatePrompt(t *testing.T) {
	// Setup
	logger := logrus.New()
	promptManager, err := genkit.NewPromptManager(logger)
	assert.NoError(t, err)

	// Test data
	prompt := &types.Prompt{
		Name:     "test-prompt",
		Template: "Hello {{.name}}, how are you?",
		Variables: []string{"name"},
		Config: map[string]interface{}{
			"temperature": 0.7,
		},
	}

	// Execute
	err = promptManager.CreatePrompt(prompt)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, prompt.ID)

	// Verify prompt was stored
	retrievedPrompt, err := promptManager.GetPrompt(prompt.ID)
	assert.NoError(t, err)
	assert.Equal(t, prompt.Name, retrievedPrompt.Name)
	assert.Equal(t, prompt.Template, retrievedPrompt.Template)
}

func TestPromptManager_RenderPrompt(t *testing.T) {
	// Setup
	logger := logrus.New()
	promptManager, err := genkit.NewPromptManager(logger)
	assert.NoError(t, err)

	// Use a built-in prompt
	req := &genkit.PromptRenderRequest{
		PromptID: "content-generation",
		Variables: map[string]interface{}{
			"content_type": "blog post",
			"topic":        "artificial intelligence",
			"style":        "professional",
		},
	}

	// Execute
	resp, err := promptManager.RenderPrompt(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Rendered)
	assert.Contains(t, resp.Rendered, "blog post")
	assert.Contains(t, resp.Rendered, "artificial intelligence")
}

func TestToolManager_RegisterTool(t *testing.T) {
	// Setup
	logger := logrus.New()
	cfg := &config.Config{}
	service, err := genkit.NewService(cfg, logger)
	assert.NoError(t, err)

	toolManager := service.GetToolManager()

	// Test data
	tool := &types.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type": "string",
				},
			},
		},
		Required: []string{"input"},
	}

	handler := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"result": "success",
			"input":  args["input"],
		}, nil
	}

	// Execute
	err = toolManager.RegisterTool(tool, handler)

	// Assertions
	assert.NoError(t, err)

	// Verify tool was registered
	retrievedTool, err := toolManager.GetTool(tool.Name)
	assert.NoError(t, err)
	assert.Equal(t, tool.Name, retrievedTool.Name)
	assert.Equal(t, tool.Description, retrievedTool.Description)
}

func TestToolManager_CallTool(t *testing.T) {
	// Setup
	logger := logrus.New()
	cfg := &config.Config{}
	service, err := genkit.NewService(cfg, logger)
	assert.NoError(t, err)

	toolManager := service.GetToolManager()

	// Use a built-in tool
	req := &genkit.ToolCallRequest{
		ToolName: "calculator",
		Arguments: map[string]interface{}{
			"expression": "2 + 2",
		},
		RequestID: "test-request",
	}

	ctx := context.Background()
	resp, err := toolManager.CallTool(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "calculator", resp.ToolName)
	assert.NotNil(t, resp.Result)
}

func TestInterruptManager_Interrupt(t *testing.T) {
	// Setup
	logger := logrus.New()
	interruptManager, err := genkit.NewInterruptManager(logger)
	assert.NoError(t, err)

	// Test data
	requestID := "test-request-123"
	reason := "User cancellation"

	// Execute
	err = interruptManager.Interrupt(requestID, reason)

	// Assertions
	assert.NoError(t, err)

	// Verify interrupt was recorded
	isInterrupted := interruptManager.IsInterrupted(requestID)
	assert.True(t, isInterrupted)

	interrupt, exists := interruptManager.GetInterrupt(requestID)
	assert.True(t, exists)
	assert.Equal(t, requestID, interrupt.RequestID)
	assert.Equal(t, reason, interrupt.Reason)
	assert.False(t, interrupt.Acknowledged)
}

func TestInterruptManager_AcknowledgeInterrupt(t *testing.T) {
	// Setup
	logger := logrus.New()
	interruptManager, err := genkit.NewInterruptManager(logger)
	assert.NoError(t, err)

	requestID := "test-request-123"

	// First interrupt
	err = interruptManager.Interrupt(requestID, "Test interrupt")
	assert.NoError(t, err)

	// Acknowledge interrupt
	err = interruptManager.AcknowledgeInterrupt(requestID)
	assert.NoError(t, err)

	// Verify interrupt is acknowledged
	interrupt, exists := interruptManager.GetInterrupt(requestID)
	assert.True(t, exists)
	assert.True(t, interrupt.Acknowledged)

	// Should no longer be considered interrupted
	isInterrupted := interruptManager.IsInterrupted(requestID)
	assert.False(t, isInterrupted)
}