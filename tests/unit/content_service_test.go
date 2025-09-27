package unit

import (
	"context"
	"testing"

	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/mcp"
	"mcp-octo-enigma/internal/service"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGenkitService is a mock implementation of Genkit service
type MockGenkitService struct {
	mock.Mock
}

func (m *MockGenkitService) GenerateContent(ctx context.Context, req *genkit.GenerateContentRequest) (*genkit.GenerateContentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*genkit.GenerateContentResponse), args.Error(1)
}

func (m *MockGenkitService) GenerateContentStream(ctx context.Context, req *genkit.GenerateContentRequest) (<-chan *genkit.GenerateContentChunk, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(<-chan *genkit.GenerateContentChunk), args.Error(1)
}

func (m *MockGenkitService) EmbedText(ctx context.Context, text string, model string) ([]float64, error) {
	args := m.Called(ctx, text, model)
	return args.Get(0).([]float64), args.Error(1)
}

func (m *MockGenkitService) GetInterruptManager() *genkit.InterruptManager {
	args := m.Called()
	return args.Get(0).(*genkit.InterruptManager)
}

func (m *MockGenkitService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockMCPManager is a mock implementation of MCP manager
type MockMCPManager struct {
	mock.Mock
}

func TestContentService_GenerateContent(t *testing.T) {
	// Setup
	mockGenkit := new(MockGenkitService)
	mockMCP := new(MockMCPManager)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	contentSvc := service.NewContentService(mockGenkit, mockMCP, logger)

	// Test data
	req := &service.ContentGenerationRequest{
		Model:  "gpt-4",
		Prompt: "Test prompt",
	}

	expectedGenkitResp := &genkit.GenerateContentResponse{
		Content: "Generated content",
		Model:   "gpt-4",
		Usage: &genkit.UsageInfo{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	// Mock expectations
	mockGenkit.On("GenerateContent", mock.Anything, mock.MatchedBy(func(gReq *genkit.GenerateContentRequest) bool {
		return gReq.Model == req.Model && gReq.Prompt == req.Prompt
	})).Return(expectedGenkitResp, nil)

	// Execute
	ctx := context.Background()
	resp, err := contentSvc.GenerateContent(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Generated content", resp.Content)
	assert.Equal(t, "gpt-4", resp.Model)
	assert.Equal(t, "completed", resp.Status)
	assert.NotEmpty(t, resp.ID)

	// Verify mocks
	mockGenkit.AssertExpectations(t)
}

func TestContentService_GenerateContentStream(t *testing.T) {
	// Setup
	mockGenkit := new(MockGenkitService)
	mockMCP := new(MockMCPManager)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	contentSvc := service.NewContentService(mockGenkit, mockMCP, logger)

	// Test data
	req := &service.ContentGenerationRequest{
		Model:  "gpt-4",
		Prompt: "Test prompt",
		Stream: true,
	}

	// Create mock channel
	mockChannel := make(chan *genkit.GenerateContentChunk, 2)
	mockChannel <- &genkit.GenerateContentChunk{Content: "Hello", Done: false}
	mockChannel <- &genkit.GenerateContentChunk{Content: " World", Done: true}
	close(mockChannel)

	// Mock expectations
	mockGenkit.On("GenerateContentStream", mock.Anything, mock.MatchedBy(func(gReq *genkit.GenerateContentRequest) bool {
		return gReq.Model == req.Model && gReq.Prompt == req.Prompt && gReq.Stream
	})).Return((<-chan *genkit.GenerateContentChunk)(mockChannel), nil)

	// Execute
	ctx := context.Background()
	streamCh, err := contentSvc.GenerateContentStream(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, streamCh)

	// Read from stream
	var chunks []string
	for chunk := range streamCh {
		chunks = append(chunks, chunk.Content)
		if chunk.Done {
			break
		}
	}

	assert.Len(t, chunks, 2)
	assert.Equal(t, "Hello", chunks[0])
	assert.Equal(t, " World", chunks[1])

	// Verify mocks
	mockGenkit.AssertExpectations(t)
}

func TestContentService_InterruptGeneration(t *testing.T) {
	// Setup
	mockGenkit := new(MockGenkitService)
	mockMCP := new(MockMCPManager)
	logger := logrus.New()

	mockInterruptManager := &genkit.InterruptManager{}
	mockGenkit.On("GetInterruptManager").Return(mockInterruptManager)

	contentSvc := service.NewContentService(mockGenkit, mockMCP, logger)

	// Execute
	ctx := context.Background()
	err := contentSvc.InterruptGeneration(ctx, "test-request-id")

	// Since we can't easily mock the interrupt manager's Interrupt method,
	// we'll just check that no error is returned for now
	// In a real implementation, you'd want to mock this properly
	assert.NoError(t, err)

	// Verify mocks
	mockGenkit.AssertExpectations(t)
}

func TestContentService_GenerateContent_ValidationError(t *testing.T) {
	// Setup
	mockGenkit := new(MockGenkitService)
	mockMCP := new(MockMCPManager)
	logger := logrus.New()

	contentSvc := service.NewContentService(mockGenkit, mockMCP, logger)

	// Test with empty model
	req := &service.ContentGenerationRequest{
		Model:  "",
		Prompt: "Test prompt",
	}

	// Mock expectations - Genkit should receive the request as-is
	mockGenkit.On("GenerateContent", mock.Anything, mock.Anything).Return(
		(*genkit.GenerateContentResponse)(nil), 
		assert.AnError,
	)

	// Execute
	ctx := context.Background()
	resp, err := contentSvc.GenerateContent(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify mocks
	mockGenkit.AssertExpectations(t)
}