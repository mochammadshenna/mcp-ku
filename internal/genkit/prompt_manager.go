package genkit

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// PromptManager manages Genkit prompts using dotprompt
type PromptManager struct {
	logger  *logrus.Logger
	prompts map[string]*PromptDefinition
	mu      sync.RWMutex
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(logger *logrus.Logger) (*PromptManager, error) {
	return &PromptManager{
		logger:  logger,
		prompts: make(map[string]*PromptDefinition),
	}, nil
}

// CreatePrompt creates a new prompt
func (pm *PromptManager) CreatePrompt(ctx context.Context, prompt *PromptDefinition) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.prompts[prompt.ID]; exists {
		return fmt.Errorf("prompt with ID %s already exists", prompt.ID)
	}

	// Validate prompt definition
	if err := pm.validatePrompt(prompt); err != nil {
		return fmt.Errorf("invalid prompt definition: %w", err)
	}

	pm.prompts[prompt.ID] = prompt
	pm.logger.Infof("Created prompt: %s (%s)", prompt.Name, prompt.ID)
	
	return nil
}

// GetPrompt retrieves a prompt by ID
func (pm *PromptManager) GetPrompt(ctx context.Context, promptID string) (*PromptDefinition, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	prompt, exists := pm.prompts[promptID]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", promptID)
	}

	return prompt, nil
}

// UpdatePrompt updates an existing prompt
func (pm *PromptManager) UpdatePrompt(ctx context.Context, prompt *PromptDefinition) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.prompts[prompt.ID]; !exists {
		return fmt.Errorf("prompt not found: %s", prompt.ID)
	}

	// Validate prompt definition
	if err := pm.validatePrompt(prompt); err != nil {
		return fmt.Errorf("invalid prompt definition: %w", err)
	}

	prompt.Version++
	pm.prompts[prompt.ID] = prompt
	pm.logger.Infof("Updated prompt: %s (%s) to version %d", prompt.Name, prompt.ID, prompt.Version)
	
	return nil
}

// DeletePrompt deletes a prompt by ID
func (pm *PromptManager) DeletePrompt(ctx context.Context, promptID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.prompts[promptID]; !exists {
		return fmt.Errorf("prompt not found: %s", promptID)
	}

	delete(pm.prompts, promptID)
	pm.logger.Infof("Deleted prompt: %s", promptID)
	
	return nil
}

// ListPrompts returns all prompts
func (pm *PromptManager) ListPrompts(ctx context.Context) ([]*PromptDefinition, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	prompts := make([]*PromptDefinition, 0, len(pm.prompts))
	for _, prompt := range pm.prompts {
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// RenderPrompt renders a prompt with the given variables
func (pm *PromptManager) RenderPrompt(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	pm.mu.RLock()
	prompt, exists := pm.prompts[req.PromptName]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", req.PromptName)
	}

	pm.logger.Debugf("Rendering prompt: %s", req.PromptName)

	// Render the template
	rendered, err := pm.renderTemplate(prompt.Template, req.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	response := &PromptResponse{
		PromptName: req.PromptName,
		Rendered:   rendered,
		RequestID:  req.RequestID,
		Metadata: map[string]interface{}{
			"template_version": prompt.Version,
			"variables_used":   pm.getUsedVariables(prompt.Template),
		},
	}

	pm.logger.Debugf("Rendered prompt: %s", req.PromptName)
	return response, nil
}

// renderTemplate renders a template with variables
func (pm *PromptManager) renderTemplate(template string, variables map[string]interface{}) (string, error) {
	// Simple template rendering using Go's text/template syntax
	// In a real implementation, you would use a proper template engine
	
	rendered := template
	
	// Replace {{.variable}} patterns
	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)
	rendered = re.ReplaceAllStringFunc(rendered, func(match string) string {
		// Extract variable name
		varName := strings.Trim(match, "{{.}}")
		
		// Get variable value
		if value, exists := variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		
		// Return original match if variable not found
		return match
	})

	// Replace {{if condition}}...{{end}} blocks
	rendered = pm.renderConditionals(rendered, variables)
	
	// Replace {{range .array}}...{{end}} blocks
	rendered = pm.renderLoops(rendered, variables)

	return rendered, nil
}

// renderConditionals renders conditional blocks
func (pm *PromptManager) renderConditionals(template string, variables map[string]interface{}) string {
	// Simple conditional rendering
	// This is a basic implementation - in production, use a proper template engine
	
	re := regexp.MustCompile(`\{\{if \.(\w+)\}\}(.*?)\{\{end\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract condition and content
		matches := re.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		
		condition := matches[1]
		content := matches[2]
		
		// Check if condition is true
		if value, exists := variables[condition]; exists {
			if boolValue, ok := value.(bool); ok && boolValue {
				return content
			}
			if value != nil && value != "" {
				return content
			}
		}
		
		return ""
	})
}

// renderLoops renders loop blocks
func (pm *PromptManager) renderLoops(template string, variables map[string]interface{}) string {
	// Simple loop rendering
	// This is a basic implementation - in production, use a proper template engine
	
	re := regexp.MustCompile(`\{\{range \.(\w+)\}\}(.*?)\{\{end\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract array name and content
		matches := re.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		
		arrayName := matches[1]
		content := matches[2]
		
		// Get array from variables
		if value, exists := variables[arrayName]; exists {
			if arr, ok := value.([]interface{}); ok {
				var result strings.Builder
				for _, item := range arr {
					// Replace {{.}} with the current item
					itemContent := strings.ReplaceAll(content, "{{.}}", fmt.Sprintf("%v", item))
					result.WriteString(itemContent)
				}
				return result.String()
			}
		}
		
		return ""
	})
}

// getUsedVariables extracts variable names from a template
func (pm *PromptManager) getUsedVariables(template string) []string {
	var variables []string
	seen := make(map[string]bool)
	
	// Find all {{.variable}} patterns
	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !seen[varName] {
				variables = append(variables, varName)
				seen[varName] = true
			}
		}
	}
	
	return variables
}

// validatePrompt validates a prompt definition
func (pm *PromptManager) validatePrompt(prompt *PromptDefinition) error {
	if prompt.ID == "" {
		return fmt.Errorf("prompt ID is required")
	}
	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}
	if prompt.Template == "" {
		return fmt.Errorf("prompt template is required")
	}
	if prompt.Version < 1 {
		prompt.Version = 1
	}

	// Extract and validate variables
	variables := pm.getUsedVariables(prompt.Template)
	prompt.Variables = variables

	return nil
}

// CreateExamplePrompts creates example prompts for demonstration
func (pm *PromptManager) CreateExamplePrompts(ctx context.Context) error {
	// Content generation prompt
	contentPrompt := &PromptDefinition{
		ID:       "content-generation",
		Name:     "Content Generation Prompt",
		Template: `You are a helpful assistant. Generate content based on the following prompt: {{.prompt}}\n\nRequirements:\n- Be creative and engaging\n- Keep it concise\n- Make it relevant to the topic\n\nGenerated content:`,
		Version:  1,
	}

	if err := pm.CreatePrompt(ctx, contentPrompt); err != nil {
		return fmt.Errorf("failed to create content generation prompt: %w", err)
	}

	// RAG prompt
	ragPrompt := &PromptDefinition{
		ID:       "rag-generation",
		Name:     "RAG Generation Prompt",
		Template: `Based on the following context, answer the question: {{.question}}\n\nContext:\n{{range .context}}{{.}}\n{{end}}\n\nAnswer:`,
		Version:  1,
	}

	if err := pm.CreatePrompt(ctx, ragPrompt); err != nil {
		return fmt.Errorf("failed to create RAG prompt: %w", err)
	}

	// Code generation prompt
	codePrompt := &PromptDefinition{
		ID:       "code-generation",
		Name:     "Code Generation Prompt",
		Template: `Generate {{.language}} code for the following task: {{.task}}\n\n{{if .requirements}}Requirements:\n{{range .requirements}}- {{.}}\n{{end}}{{end}}\n\nCode:`,
		Version:  1,
	}

	if err := pm.CreatePrompt(ctx, codePrompt); err != nil {
		return fmt.Errorf("failed to create code generation prompt: %w", err)
	}

	// Summary prompt
	summaryPrompt := &PromptDefinition{
		ID:       "text-summary",
		Name:     "Text Summary Prompt",
		Template: `Summarize the following text in {{.length}} words:\n\n{{.text}}\n\nSummary:`,
		Version:  1,
	}

	if err := pm.CreatePrompt(ctx, summaryPrompt); err != nil {
		return fmt.Errorf("failed to create summary prompt: %w", err)
	}

	return nil
}

// LoadPromptsFromFile loads prompts from a .prompt file
func (pm *PromptManager) LoadPromptsFromFile(ctx context.Context, filename string) error {
	// This would load prompts from a .prompt file format
	// For now, we'll create some example prompts
	return pm.CreateExamplePrompts(ctx)
}

// ExportPromptsToFile exports prompts to a .prompt file
func (pm *PromptManager) ExportPromptsToFile(ctx context.Context, filename string) error {
	// This would export prompts to a .prompt file format
	// For now, we'll just log the action
	pm.logger.Infof("Exporting prompts to file: %s", filename)
	return nil
}