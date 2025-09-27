package genkit

import (
	"context"
	"fmt"
	"sync"

	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FlowManager manages Genkit flows
type FlowManager struct {
	service *Service
	logger  *logrus.Logger
	flows   map[string]*types.Flow
	executions map[string]*types.FlowExecution
	mu      sync.RWMutex
}

// NewFlowManager creates a new flow manager
func NewFlowManager(service *Service, logger *logrus.Logger) (*FlowManager, error) {
	fm := &FlowManager{
		service: service,
		logger:  logger,
		flows:   make(map[string]*types.Flow),
		executions: make(map[string]*types.FlowExecution),
	}

	// Register default flows
	fm.registerDefaultFlows()

	return fm, nil
}

// registerDefaultFlows registers built-in flows
func (fm *FlowManager) registerDefaultFlows() {
	// Content generation flow
	contentFlow := &types.Flow{
		ID:          "content-generation",
		Name:        "Content Generation",
		Description: "Generate content using AI models",
		Steps: []types.FlowStep{
			{
				ID:          "validate-input",
				Type:        "validation",
				Name:        "Validate Input",
				Description: "Validate the input parameters",
				Config: map[string]interface{}{
					"required_fields": []string{"prompt", "model"},
				},
			},
			{
				ID:          "generate-content",
				Type:        "generate",
				Name:        "Generate Content",
				Description: "Generate content using AI model",
				Dependencies: []string{"validate-input"},
				Config: map[string]interface{}{
					"model": "{{.model}}",
					"prompt": "{{.prompt}}",
				},
			},
		},
	}
	fm.flows[contentFlow.ID] = contentFlow

	// RAG flow
	ragFlow := &types.Flow{
		ID:          "rag-query",
		Name:        "RAG Query",
		Description: "Retrieval-augmented generation flow",
		Steps: []types.FlowStep{
			{
				ID:          "embed-query",
				Type:        "embed",
				Name:        "Embed Query",
				Description: "Create embeddings for the query",
				Config: map[string]interface{}{
					"text": "{{.query}}",
					"model": "text-embedding-ada-002",
				},
			},
			{
				ID:          "search-vectors",
				Type:        "vector-search",
				Name:        "Search Vectors",
				Description: "Search for similar vectors",
				Dependencies: []string{"embed-query"},
				Config: map[string]interface{}{
					"query": "{{.embed-query.output.embedding}}",
					"limit": 5,
				},
			},
			{
				ID:          "generate-answer",
				Type:        "generate",
				Name:        "Generate Answer",
				Description: "Generate answer based on retrieved context",
				Dependencies: []string{"search-vectors"},
				Config: map[string]interface{}{
					"model": "{{.model}}",
					"prompt": "Context: {{.search-vectors.output.results}}\n\nQuestion: {{.query}}\n\nAnswer:",
				},
			},
		},
	}
	fm.flows[ragFlow.ID] = ragFlow
}

// CreateFlow creates a new flow
func (fm *FlowManager) CreateFlow(flow *types.Flow) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if flow.ID == "" {
		flow.ID = uuid.New().String()
	}

	fm.flows[flow.ID] = flow
	fm.logger.Infof("Created flow: %s", flow.ID)

	return nil
}

// GetFlow retrieves a flow by ID
func (fm *FlowManager) GetFlow(flowID string) (*types.Flow, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flow, exists := fm.flows[flowID]
	if !exists {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}

	return flow, nil
}

// ListFlows returns all flows
func (fm *FlowManager) ListFlows() []*types.Flow {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flows := make([]*types.Flow, 0, len(fm.flows))
	for _, flow := range fm.flows {
		flows = append(flows, flow)
	}

	return flows
}

// ExecuteFlow executes a flow with the given input
func (fm *FlowManager) ExecuteFlow(ctx context.Context, req *FlowExecutionRequest) (*FlowExecutionResponse, error) {
	flow, err := fm.GetFlow(req.FlowID)
	if err != nil {
		return nil, err
	}

	execution := &types.FlowExecution{
		ID:     uuid.New().String(),
		FlowID: req.FlowID,
		Input:  req.Input,
		Status: "running",
		Metadata: make(map[string]interface{}),
	}

	fm.mu.Lock()
	fm.executions[execution.ID] = execution
	fm.mu.Unlock()

	// Execute flow steps
	output, err := fm.executeSteps(ctx, flow, execution, req.Input)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		return nil, err
	}

	execution.Status = "completed"
	execution.Output = output

	response := &FlowExecutionResponse{
		FlowID:    req.FlowID,
		Output:    output,
		RequestID: req.RequestID,
		Status:    execution.Status,
		Metadata:  execution.Metadata,
	}

	return response, nil
}

// executeSteps executes the flow steps in dependency order
func (fm *FlowManager) executeSteps(ctx context.Context, flow *types.Flow, execution *types.FlowExecution, input map[string]interface{}) (map[string]interface{}, error) {
	stepResults := make(map[string]interface{})
	stepResults["input"] = input

	// Build dependency graph and execute in order
	executed := make(map[string]bool)
	
	for len(executed) < len(flow.Steps) {
		progress := false
		
		for _, step := range flow.Steps {
			if executed[step.ID] {
				continue
			}

			// Check if all dependencies are satisfied
			canExecute := true
			for _, dep := range step.Dependencies {
				if !executed[dep] {
					canExecute = false
					break
				}
			}

			if !canExecute {
				continue
			}

			// Execute step
			result, err := fm.executeStep(ctx, step, stepResults)
			if err != nil {
				return nil, fmt.Errorf("step %s failed: %w", step.ID, err)
			}

			stepResults[step.ID] = map[string]interface{}{
				"output": result,
			}
			executed[step.ID] = true
			progress = true

			fm.logger.Debugf("Executed step: %s", step.ID)
		}

		if !progress {
			return nil, fmt.Errorf("circular dependency detected in flow")
		}
	}

	// Return the output from the final step or all step results
	if len(flow.Steps) > 0 {
		finalStep := flow.Steps[len(flow.Steps)-1]
		if result, ok := stepResults[finalStep.ID]; ok {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if output, ok := resultMap["output"]; ok {
					return map[string]interface{}{"result": output}, nil
				}
			}
		}
	}

	return stepResults, nil
}

// executeStep executes a single flow step
func (fm *FlowManager) executeStep(ctx context.Context, step types.FlowStep, context map[string]interface{}) (interface{}, error) {
	switch step.Type {
	case "validation":
		return fm.executeValidationStep(step, context)
	case "generate":
		return fm.executeGenerateStep(ctx, step, context)
	case "embed":
		return fm.executeEmbedStep(ctx, step, context)
	case "vector-search":
		return fm.executeVectorSearchStep(ctx, step, context)
	case "tool":
		return fm.executeToolStep(ctx, step, context)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeValidationStep executes a validation step
func (fm *FlowManager) executeValidationStep(step types.FlowStep, context map[string]interface{}) (interface{}, error) {
	requiredFields, ok := step.Config["required_fields"].([]string)
	if !ok {
		return nil, fmt.Errorf("required_fields not specified in validation step")
	}

	input, ok := context["input"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid input format")
	}

	for _, field := range requiredFields {
		if _, exists := input[field]; !exists {
			return nil, fmt.Errorf("required field missing: %s", field)
		}
	}

	return map[string]interface{}{"valid": true}, nil
}

// executeGenerateStep executes a content generation step
func (fm *FlowManager) executeGenerateStep(ctx context.Context, step types.FlowStep, context map[string]interface{}) (interface{}, error) {
	// Render template values
	model, err := fm.renderTemplate(step.Config["model"], context)
	if err != nil {
		return nil, fmt.Errorf("failed to render model template: %w", err)
	}

	prompt, err := fm.renderTemplate(step.Config["prompt"], context)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt template: %w", err)
	}

	req := &GenerateContentRequest{
		Model:  model.(string),
		Prompt: prompt.(string),
	}

	response, err := fm.service.GenerateContent(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": response.Content,
		"usage":   response.Usage,
	}, nil
}

// executeEmbedStep executes an embedding step
func (fm *FlowManager) executeEmbedStep(ctx context.Context, step types.FlowStep, context map[string]interface{}) (interface{}, error) {
	text, err := fm.renderTemplate(step.Config["text"], context)
	if err != nil {
		return nil, fmt.Errorf("failed to render text template: %w", err)
	}

	model, err := fm.renderTemplate(step.Config["model"], context)
	if err != nil {
		return nil, fmt.Errorf("failed to render model template: %w", err)
	}

	embedding, err := fm.service.EmbedText(ctx, text.(string), model.(string))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"embedding": embedding,
	}, nil
}

// executeVectorSearchStep executes a vector search step
func (fm *FlowManager) executeVectorSearchStep(ctx context.Context, step types.FlowStep, context map[string]interface{}) (interface{}, error) {
	// This would integrate with the vector repository
	// For now, return mock results
	return map[string]interface{}{
		"results": []map[string]interface{}{
			{"content": "Mock search result 1", "score": 0.9},
			{"content": "Mock search result 2", "score": 0.8},
		},
	}, nil
}

// executeToolStep executes a tool calling step
func (fm *FlowManager) executeToolStep(ctx context.Context, step types.FlowStep, context map[string]interface{}) (interface{}, error) {
	// This would integrate with the tool manager
	// For now, return mock results
	return map[string]interface{}{
		"result": "Mock tool result",
	}, nil
}

// renderTemplate renders a template with the given context
func (fm *FlowManager) renderTemplate(template interface{}, context map[string]interface{}) (interface{}, error) {
	templateStr, ok := template.(string)
	if !ok {
		return template, nil
	}

	// Simple template rendering - replace {{.key}} with values from context
	// In a real implementation, use a proper template engine
	result := templateStr
	
	if input, ok := context["input"].(map[string]interface{}); ok {
		for key, value := range input {
			placeholder := fmt.Sprintf("{{.%s}}", key)
			if valueStr, ok := value.(string); ok {
				result = fmt.Sprintf("%s", valueStr)
				if result == placeholder {
					result = valueStr
				}
			}
		}
	}

	return result, nil
}