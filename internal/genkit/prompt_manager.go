package genkit

import (
	"fmt"
	"strings"
	"sync"
	"text/template"
	"bytes"

	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PromptManager manages dotprompt templates
type PromptManager struct {
	logger    *logrus.Logger
	prompts   map[string]*types.Prompt
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(logger *logrus.Logger) (*PromptManager, error) {
	pm := &PromptManager{
		logger:    logger,
		prompts:   make(map[string]*types.Prompt),
		templates: make(map[string]*template.Template),
	}

	// Register default prompts
	pm.registerDefaultPrompts()

	return pm, nil
}

// registerDefaultPrompts registers built-in prompt templates
func (pm *PromptManager) registerDefaultPrompts() {
	// Content generation prompt
	contentPrompt := &types.Prompt{
		ID:       "content-generation",
		Name:     "Content Generation",
		Template: `Generate {{.content_type}} content about {{.topic}}.

Requirements:
- Style: {{.style | default "professional"}}
- Length: {{.length | default "medium"}}
- Audience: {{.audience | default "general"}}

{{if .additional_context}}
Additional Context:
{{.additional_context}}
{{end}}

Please provide a well-structured and engaging response.`,
		Variables: []string{"content_type", "topic", "style", "length", "audience", "additional_context"},
		Config: map[string]interface{}{
			"max_tokens": 1000,
			"temperature": 0.7,
		},
		Version: 1,
	}
	pm.prompts[contentPrompt.ID] = contentPrompt

	// RAG query prompt
	ragPrompt := &types.Prompt{
		ID:       "rag-query",
		Name:     "RAG Query",
		Template: `Based on the following context, please answer the question.

Context:
{{range .context}}
- {{.content}}
{{end}}

Question: {{.question}}

Instructions:
- Use only the information provided in the context
- If the context doesn't contain enough information, say so
- Be concise but complete in your answer
- Cite specific parts of the context when relevant

Answer:`,
		Variables: []string{"context", "question"},
		Config: map[string]interface{}{
			"max_tokens": 500,
			"temperature": 0.3,
		},
		Version: 1,
	}
	pm.prompts[ragPrompt.ID] = ragPrompt

	// Tool calling prompt
	toolPrompt := &types.Prompt{
		ID:       "tool-calling",
		Name:     "Tool Calling",
		Template: `You are an AI assistant that can use tools to help answer questions.

Available tools:
{{range .tools}}
- {{.name}}: {{.description}}
{{end}}

User request: {{.user_request}}

Please analyze the request and determine if you need to use any tools. If so, call the appropriate tool(s) first, then provide a comprehensive answer based on the results.`,
		Variables: []string{"tools", "user_request"},
		Config: map[string]interface{}{
			"max_tokens": 1500,
			"temperature": 0.5,
		},
		Version: 1,
	}
	pm.prompts[toolPrompt.ID] = toolPrompt

	// Code generation prompt
	codePrompt := &types.Prompt{
		ID:       "code-generation",
		Name:     "Code Generation",
		Template: `Generate {{.language}} code for the following requirements:

Task: {{.task}}

Requirements:
{{range .requirements}}
- {{.}}
{{end}}

{{if .constraints}}
Constraints:
{{range .constraints}}
- {{.}}
{{end}}
{{end}}

{{if .examples}}
Examples or references:
{{.examples}}
{{end}}

Please provide:
1. Clean, well-commented code
2. Error handling where appropriate
3. Brief explanation of the approach
4. Any assumptions made

Code:`,
		Variables: []string{"language", "task", "requirements", "constraints", "examples"},
		Config: map[string]interface{}{
			"max_tokens": 2000,
			"temperature": 0.2,
		},
		Version: 1,
	}
	pm.prompts[codePrompt.ID] = codePrompt

	// Compile templates
	for _, prompt := range pm.prompts {
		if err := pm.compileTemplate(prompt); err != nil {
			pm.logger.Warnf("Failed to compile template for prompt %s: %v", prompt.ID, err)
		}
	}
}

// CreatePrompt creates a new prompt template
func (pm *PromptManager) CreatePrompt(prompt *types.Prompt) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if prompt.ID == "" {
		prompt.ID = uuid.New().String()
	}

	// Compile template
	if err := pm.compileTemplate(prompt); err != nil {
		return fmt.Errorf("failed to compile template: %w", err)
	}

	pm.prompts[prompt.ID] = prompt
	pm.logger.Infof("Created prompt: %s", prompt.ID)

	return nil
}

// GetPrompt retrieves a prompt by ID
func (pm *PromptManager) GetPrompt(promptID string) (*types.Prompt, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	prompt, exists := pm.prompts[promptID]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", promptID)
	}

	return prompt, nil
}

// ListPrompts returns all prompts
func (pm *PromptManager) ListPrompts() []*types.Prompt {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	prompts := make([]*types.Prompt, 0, len(pm.prompts))
	for _, prompt := range pm.prompts {
		prompts = append(prompts, prompt)
	}

	return prompts
}

// RenderPrompt renders a prompt template with the given variables
func (pm *PromptManager) RenderPrompt(req *PromptRenderRequest) (*PromptRenderResponse, error) {
	pm.mu.RLock()
	tmpl, exists := pm.templates[req.PromptID]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("prompt template not found: %s", req.PromptID)
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.Variables); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	response := &PromptRenderResponse{
		PromptID:  req.PromptID,
		Rendered:  buf.String(),
		RequestID: req.RequestID,
	}

	return response, nil
}

// UpdatePrompt updates an existing prompt
func (pm *PromptManager) UpdatePrompt(prompt *types.Prompt) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.prompts[prompt.ID]; !exists {
		return fmt.Errorf("prompt not found: %s", prompt.ID)
	}

	// Compile template
	if err := pm.compileTemplate(prompt); err != nil {
		return fmt.Errorf("failed to compile template: %w", err)
	}

	// Increment version
	prompt.Version++
	pm.prompts[prompt.ID] = prompt
	pm.logger.Infof("Updated prompt: %s (version %d)", prompt.ID, prompt.Version)

	return nil
}

// DeletePrompt deletes a prompt
func (pm *PromptManager) DeletePrompt(promptID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.prompts[promptID]; !exists {
		return fmt.Errorf("prompt not found: %s", promptID)
	}

	delete(pm.prompts, promptID)
	delete(pm.templates, promptID)
	pm.logger.Infof("Deleted prompt: %s", promptID)

	return nil
}

// compileTemplate compiles a prompt template
func (pm *PromptManager) compileTemplate(prompt *types.Prompt) error {
	// Create custom functions for templates
	funcMap := template.FuncMap{
		"default": func(defaultValue interface{}, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
		"join": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"trim":  strings.TrimSpace,
	}

	tmpl, err := template.New(prompt.ID).Funcs(funcMap).Parse(prompt.Template)
	if err != nil {
		return err
	}

	pm.templates[prompt.ID] = tmpl
	return nil
}

// ValidatePrompt validates a prompt template
func (pm *PromptManager) ValidatePrompt(prompt *types.Prompt) error {
	// Check required fields
	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}
	if prompt.Template == "" {
		return fmt.Errorf("prompt template is required")
	}

	// Try to compile template
	_, err := template.New("validation").Parse(prompt.Template)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}

	// Extract variables from template
	variables := pm.extractVariables(prompt.Template)
	
	// Validate that all declared variables are used
	declaredVars := make(map[string]bool)
	for _, v := range prompt.Variables {
		declaredVars[v] = true
	}

	usedVars := make(map[string]bool)
	for _, v := range variables {
		usedVars[v] = true
	}

	// Check for undeclared variables
	for v := range usedVars {
		if !declaredVars[v] {
			pm.logger.Warnf("Undeclared variable used in template: %s", v)
		}
	}

	// Check for unused declared variables
	for v := range declaredVars {
		if !usedVars[v] {
			pm.logger.Warnf("Declared variable not used in template: %s", v)
		}
	}

	return nil
}

// extractVariables extracts variable names from a template
func (pm *PromptManager) extractVariables(templateStr string) []string {
	var variables []string
	
	// Simple regex to find {{.variable}} patterns
	// In a real implementation, use a proper parser
	lines := strings.Split(templateStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "{{.") {
			// Extract variable names - this is a simplified approach
			start := strings.Index(line, "{{.")
			for start != -1 {
				end := strings.Index(line[start:], "}}")
				if end != -1 {
					varExpr := line[start+3 : start+end]
					// Handle pipe functions
					if pipeIndex := strings.Index(varExpr, " |"); pipeIndex != -1 {
						varExpr = varExpr[:pipeIndex]
					}
					if strings.TrimSpace(varExpr) != "" {
						variables = append(variables, strings.TrimSpace(varExpr))
					}
				}
				nextStart := strings.Index(line[start+1:], "{{.")
				if nextStart == -1 {
					break
				}
				start = start + 1 + nextStart
			}
		}
	}

	// Remove duplicates
	uniqueVars := make(map[string]bool)
	result := []string{}
	for _, v := range variables {
		if !uniqueVars[v] {
			uniqueVars[v] = true
			result = append(result, v)
		}
	}

	return result
}