package genkit

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// FlowManager manages Genkit flows
type FlowManager struct {
	service *Service
	logger  *logrus.Logger
	flows   map[string]*FlowDefinition
	mu      sync.RWMutex
}

// NewFlowManager creates a new flow manager
func NewFlowManager(service *Service, logger *logrus.Logger) (*FlowManager, error) {
	return &FlowManager{
		service: service,
		logger:  logger,
		flows:   make(map[string]*FlowDefinition),
	}, nil
}

// CreateFlow creates a new flow
func (fm *FlowManager) CreateFlow(ctx context.Context, flow *FlowDefinition) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.flows[flow.ID]; exists {
		return fmt.Errorf("flow with ID %s already exists", flow.ID)
	}

	// Validate flow definition
	if err := fm.validateFlow(flow); err != nil {
		return fmt.Errorf("invalid flow definition: %w", err)
	}

	fm.flows[flow.ID] = flow
	fm.logger.Infof("Created flow: %s (%s)", flow.Name, flow.ID)

	return nil
}

// GetFlow retrieves a flow by ID
func (fm *FlowManager) GetFlow(ctx context.Context, flowID string) (*FlowDefinition, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flow, exists := fm.flows[flowID]
	if !exists {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}

	return flow, nil
}

// UpdateFlow updates an existing flow
func (fm *FlowManager) UpdateFlow(ctx context.Context, flow *FlowDefinition) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.flows[flow.ID]; !exists {
		return fmt.Errorf("flow not found: %s", flow.ID)
	}

	// Validate flow definition
	if err := fm.validateFlow(flow); err != nil {
		return fmt.Errorf("invalid flow definition: %w", err)
	}

	fm.flows[flow.ID] = flow
	fm.logger.Infof("Updated flow: %s (%s)", flow.Name, flow.ID)

	return nil
}

// DeleteFlow deletes a flow by ID
func (fm *FlowManager) DeleteFlow(ctx context.Context, flowID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.flows[flowID]; !exists {
		return fmt.Errorf("flow not found: %s", flowID)
	}

	delete(fm.flows, flowID)
	fm.logger.Infof("Deleted flow: %s", flowID)

	return nil
}

// ListFlows returns all flows
func (fm *FlowManager) ListFlows(ctx context.Context) ([]*FlowDefinition, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flows := make([]*FlowDefinition, 0, len(fm.flows))
	for _, flow := range fm.flows {
		flows = append(flows, flow)
	}

	return flows, nil
}

// ExecuteFlow executes a flow with the given input
func (fm *FlowManager) ExecuteFlow(ctx context.Context, req *FlowExecutionRequest) (*FlowExecutionResponse, error) {
	// Get flow definition
	flow, err := fm.GetFlow(ctx, req.FlowID)
	if err != nil {
		return nil, err
	}

	// Check for interrupts
	if fm.service.interruptManager.IsInterrupted(req.RequestID) {
		return nil, fmt.Errorf("execution interrupted")
	}

	fm.logger.Infof("Executing flow: %s (%s)", flow.Name, req.FlowID)

	// Execute flow steps
	output := make(map[string]interface{})
	metadata := make(map[string]interface{})

	for _, step := range flow.Steps {
		stepResult, err := fm.executeStep(ctx, step, req.Input, output, req.RequestID)
		if err != nil {
			return nil, fmt.Errorf("failed to execute step %s: %w", step.ID, err)
		}

		output[step.ID] = stepResult
		metadata[step.ID+"_executed_at"] = "2024-01-01T12:00:00Z"
	}

	response := &FlowExecutionResponse{
		FlowID:    req.FlowID,
		Output:    output,
		RequestID: req.RequestID,
		Status:    "completed",
		Metadata:  metadata,
	}

	fm.logger.Infof("Completed flow execution: %s", req.FlowID)
	return response, nil
}

// executeStep executes a single flow step
func (fm *FlowManager) executeStep(ctx context.Context, step FlowStep, input, output map[string]interface{}, requestID string) (interface{}, error) {
	fm.logger.Debugf("Executing step: %s (%s)", step.Name, step.ID)

	switch step.Type {
	case "generate":
		return fm.executeGenerateStep(ctx, step, input, output, requestID)
	case "tool":
		return fm.executeToolStep(ctx, step, input, output, requestID)
	case "prompt":
		return fm.executePromptStep(ctx, step, input, output, requestID)
	case "transform":
		return fm.executeTransformStep(ctx, step, input, output, requestID)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeGenerateStep executes a generation step
func (fm *FlowManager) executeGenerateStep(ctx context.Context, step FlowStep, input, output map[string]interface{}, requestID string) (interface{}, error) {
	// Get prompt from step config
	prompt, ok := step.Config["prompt"].(string)
	if !ok {
		return nil, fmt.Errorf("prompt not found in step config")
	}

	// Get model from step config
	model, ok := step.Config["model"].(string)
	if !ok {
		model = "gpt-3.5-turbo" // default model
	}

	// Create generation request
	genReq := &GenerateContentRequest{
		Model:     model,
		Prompt:    prompt,
		RequestID: requestID,
	}

	// Generate content
	response, err := fm.service.GenerateContent(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return response.Content, nil
}

// executeToolStep executes a tool step
func (fm *FlowManager) executeToolStep(ctx context.Context, step FlowStep, input, output map[string]interface{}, requestID string) (interface{}, error) {
	// Get tool name from step config
	toolName, ok := step.Config["tool_name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool_name not found in step config")
	}

	// Get tool arguments from step config
	args, ok := step.Config["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Create tool call request
	toolReq := &ToolCallRequest{
		ToolName:  toolName,
		Arguments: args,
		RequestID: requestID,
	}

	// Call tool
	response, err := fm.service.toolManager.CallTool(ctx, toolReq)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	return response.Result, nil
}

// executePromptStep executes a prompt step
func (fm *FlowManager) executePromptStep(ctx context.Context, step FlowStep, input, output map[string]interface{}, requestID string) (interface{}, error) {
	// Get prompt name from step config
	promptName, ok := step.Config["prompt_name"].(string)
	if !ok {
		return nil, fmt.Errorf("prompt_name not found in step config")
	}

	// Get variables from step config
	variables, ok := step.Config["variables"].(map[string]interface{})
	if !ok {
		variables = make(map[string]interface{})
	}

	// Create prompt request
	promptReq := &PromptRequest{
		PromptName: promptName,
		Variables:  variables,
		RequestID:  requestID,
	}

	// Render prompt
	response, err := fm.service.promptManager.RenderPrompt(ctx, promptReq)
	if err != nil {
		return nil, fmt.Errorf("prompt rendering failed: %w", err)
	}

	return response.Rendered, nil
}

// executeTransformStep executes a transform step
func (fm *FlowManager) executeTransformStep(ctx context.Context, step FlowStep, input, output map[string]interface{}, requestID string) (interface{}, error) {
	// Get transformation function from step config
	transform, ok := step.Config["transform"].(string)
	if !ok {
		return nil, fmt.Errorf("transform not found in step config")
	}

	// Simple transformations for demo
	switch transform {
	case "uppercase":
		if content, ok := output["previous_step"].(string); ok {
			return fmt.Sprintf("%s", content), nil
		}
		return nil, fmt.Errorf("previous step output not found")
	case "lowercase":
		if content, ok := output["previous_step"].(string); ok {
			return fmt.Sprintf("%s", content), nil
		}
		return nil, fmt.Errorf("previous step output not found")
	default:
		return nil, fmt.Errorf("unknown transform: %s", transform)
	}
}

// validateFlow validates a flow definition
func (fm *FlowManager) validateFlow(flow *FlowDefinition) error {
	if flow.ID == "" {
		return fmt.Errorf("flow ID is required")
	}
	if flow.Name == "" {
		return fmt.Errorf("flow name is required")
	}
	if len(flow.Steps) == 0 {
		return fmt.Errorf("flow must have at least one step")
	}

	// Validate steps
	stepIDs := make(map[string]bool)
	for _, step := range flow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step ID is required")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true

		// Validate dependencies
		for _, dep := range step.Dependencies {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
			}
		}
	}

	return nil
}

// CreateExampleFlows creates example flows for demonstration
func (fm *FlowManager) CreateExampleFlows(ctx context.Context) error {
	// Simple content generation flow
	contentFlow := &FlowDefinition{
		ID:          "content-generation",
		Name:        "Content Generation Flow",
		Description: "Generates content based on a prompt",
		Input: map[string]interface{}{
			"prompt": "string",
			"model":  "string",
		},
		Output: map[string]interface{}{
			"content": "string",
		},
		Steps: []FlowStep{
			{
				ID:          "generate",
				Type:        "generate",
				Name:        "Generate Content",
				Description: "Generate content using AI model",
				Config: map[string]interface{}{
					"prompt": "{{.prompt}}",
					"model":  "{{.model}}",
				},
			},
		},
	}

	if err := fm.CreateFlow(ctx, contentFlow); err != nil {
		return fmt.Errorf("failed to create content generation flow: %w", err)
	}

	// RAG flow
	ragFlow := &FlowDefinition{
		ID:          "rag-generation",
		Name:        "RAG Generation Flow",
		Description: "Retrieval-Augmented Generation flow",
		Input: map[string]interface{}{
			"query": "string",
		},
		Output: map[string]interface{}{
			"content": "string",
			"sources": "[]string",
		},
		Steps: []FlowStep{
			{
				ID:          "search",
				Type:        "tool",
				Name:        "Search Documents",
				Description: "Search for relevant documents",
				Config: map[string]interface{}{
					"tool_name": "vector-search",
					"arguments": map[string]interface{}{
						"query": "{{.query}}",
						"limit": 5,
					},
				},
			},
			{
				ID:           "generate",
				Type:         "generate",
				Name:         "Generate Answer",
				Description:  "Generate answer based on retrieved documents",
				Dependencies: []string{"search"},
				Config: map[string]interface{}{
					"prompt": "Based on the following context, answer the question: {{.query}}\n\nContext: {{.search}}",
					"model":  "gpt-4",
				},
			},
		},
	}

	if err := fm.CreateFlow(ctx, ragFlow); err != nil {
		return fmt.Errorf("failed to create RAG flow: %w", err)
	}

	return nil
}
