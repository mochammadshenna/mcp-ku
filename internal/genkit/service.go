package genkit

import (
	"context"
	"fmt"

	"mcp-octo-enigma/internal/config"

	"github.com/sirupsen/logrus"
)

// Service represents the Genkit service
type Service struct {
	config *config.Config
	logger *logrus.Logger
	
	// AI providers
	googleAI   *GoogleAIProvider
	openAI     *OpenAIProvider
	anthropic  *AnthropicProvider
	vertexAI   *VertexAIProvider
	ollama     *OllamaProvider
	
	// Genkit components
	flowManager    *FlowManager
	promptManager  *PromptManager
	toolManager    *ToolManager
	interruptManager *InterruptManager
}

// NewService creates a new Genkit service
func NewService(cfg *config.Config, logger *logrus.Logger) (*Service, error) {
	service := &Service{
		config: cfg,
		logger: logger,
	}

	// Initialize AI providers
	if err := service.initializeProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize AI providers: %w", err)
	}

	// Initialize Genkit components
	if err := service.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize Genkit components: %w", err)
	}

	return service, nil
}

// initializeProviders initializes all AI providers
func (s *Service) initializeProviders() error {
	var err error

	// Initialize Google AI provider
	if s.config.AI.GoogleAI.APIKey != "" {
		s.googleAI, err = NewGoogleAIProvider(s.config.AI.GoogleAI.APIKey, s.logger)
		if err != nil {
			s.logger.Warnf("Failed to initialize Google AI provider: %v", err)
		}
	}

	// Initialize OpenAI provider
	if s.config.AI.OpenAI.APIKey != "" {
		s.openAI, err = NewOpenAIProvider(s.config.AI.OpenAI.APIKey, s.logger)
		if err != nil {
			s.logger.Warnf("Failed to initialize OpenAI provider: %v", err)
		}
	}

	// Initialize Anthropic provider
	if s.config.AI.Anthropic.APIKey != "" {
		s.anthropic, err = NewAnthropicProvider(s.config.AI.Anthropic.APIKey, s.logger)
		if err != nil {
			s.logger.Warnf("Failed to initialize Anthropic provider: %v", err)
		}
	}

	// Initialize Vertex AI provider
	if s.config.AI.VertexAI.ProjectID != "" {
		s.vertexAI, err = NewVertexAIProvider(s.config.AI.VertexAI.ProjectID, s.logger)
		if err != nil {
			s.logger.Warnf("Failed to initialize Vertex AI provider: %v", err)
		}
	}

	// Initialize Ollama provider
	if s.config.AI.Ollama.Host != "" {
		s.ollama, err = NewOllamaProvider(s.config.AI.Ollama.Host, s.logger)
		if err != nil {
			s.logger.Warnf("Failed to initialize Ollama provider: %v", err)
		}
	}

	return nil
}

// initializeComponents initializes Genkit components
func (s *Service) initializeComponents() error {
	var err error

	// Initialize flow manager
	s.flowManager, err = NewFlowManager(s, s.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize flow manager: %w", err)
	}

	// Initialize prompt manager
	s.promptManager, err = NewPromptManager(s.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize prompt manager: %w", err)
	}

	// Initialize tool manager
	s.toolManager, err = NewToolManager(s, s.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize tool manager: %w", err)
	}

	// Initialize interrupt manager
	s.interruptManager, err = NewInterruptManager(s.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize interrupt manager: %w", err)
	}

	return nil
}

// GenerateContent generates content using the specified model
func (s *Service) GenerateContent(ctx context.Context, req *GenerateContentRequest) (*GenerateContentResponse, error) {
	provider, err := s.getProvider(req.Model)
	if err != nil {
		return nil, err
	}

	// Check for interrupts
	if s.interruptManager.IsInterrupted(req.RequestID) {
		return nil, fmt.Errorf("generation interrupted")
	}

	return provider.GenerateContent(ctx, req)
}

// GenerateContentStream generates content with streaming
func (s *Service) GenerateContentStream(ctx context.Context, req *GenerateContentRequest) (<-chan *GenerateContentChunk, error) {
	provider, err := s.getProvider(req.Model)
	if err != nil {
		return nil, err
	}

	return provider.GenerateContentStream(ctx, req)
}

// EmbedText creates vector embeddings for text
func (s *Service) EmbedText(ctx context.Context, text string, model string) ([]float64, error) {
	provider, err := s.getProvider(model)
	if err != nil {
		return nil, err
	}

	return provider.EmbedText(ctx, text)
}

// getProvider returns the appropriate AI provider for the given model
func (s *Service) getProvider(model string) (AIProvider, error) {
	switch {
	case s.isGoogleAIModel(model) && s.googleAI != nil:
		return s.googleAI, nil
	case s.isOpenAIModel(model) && s.openAI != nil:
		return s.openAI, nil
	case s.isAnthropicModel(model) && s.anthropic != nil:
		return s.anthropic, nil
	case s.isVertexAIModel(model) && s.vertexAI != nil:
		return s.vertexAI, nil
	case s.isOllamaModel(model) && s.ollama != nil:
		return s.ollama, nil
	default:
		return nil, fmt.Errorf("no provider available for model: %s", model)
	}
}

// Model detection methods
func (s *Service) isGoogleAIModel(model string) bool {
	googleModels := []string{"gemini-pro", "gemini-pro-vision", "text-embedding-004"}
	for _, m := range googleModels {
		if m == model {
			return true
		}
	}
	return false
}

func (s *Service) isOpenAIModel(model string) bool {
	openAIModels := []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo", "text-embedding-ada-002"}
	for _, m := range openAIModels {
		if m == model {
			return true
		}
	}
	return false
}

func (s *Service) isAnthropicModel(model string) bool {
	anthropicModels := []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}
	for _, m := range anthropicModels {
		if m == model {
			return true
		}
	}
	return false
}

func (s *Service) isVertexAIModel(model string) bool {
	vertexModels := []string{"text-bison", "chat-bison", "textembedding-gecko"}
	for _, m := range vertexModels {
		if m == model {
			return true
		}
	}
	return false
}

func (s *Service) isOllamaModel(model string) bool {
	// Ollama can run any model, so we'll assume any model not matching others is Ollama
	return !s.isGoogleAIModel(model) && !s.isOpenAIModel(model) && 
		   !s.isAnthropicModel(model) && !s.isVertexAIModel(model)
}

// GetFlowManager returns the flow manager
func (s *Service) GetFlowManager() *FlowManager {
	return s.flowManager
}

// GetPromptManager returns the prompt manager
func (s *Service) GetPromptManager() *PromptManager {
	return s.promptManager
}

// GetToolManager returns the tool manager
func (s *Service) GetToolManager() *ToolManager {
	return s.toolManager
}

// GetInterruptManager returns the interrupt manager
func (s *Service) GetInterruptManager() *InterruptManager {
	return s.interruptManager
}

// Close closes the Genkit service
func (s *Service) Close() error {
	// Close all providers and components
	if s.googleAI != nil {
		s.googleAI.Close()
	}
	if s.openAI != nil {
		s.openAI.Close()
	}
	if s.anthropic != nil {
		s.anthropic.Close()
	}
	if s.vertexAI != nil {
		s.vertexAI.Close()
	}
	if s.ollama != nil {
		s.ollama.Close()
	}

	s.logger.Info("Genkit service closed")
	return nil
}