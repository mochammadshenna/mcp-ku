# Development Guide

This document provides comprehensive guidance for developers working on the MCP Octo Enigma project.

## 🏗️ Development Environment Setup

### Prerequisites
- Go 1.24+
- PostgreSQL with pgvector
- Redis (optional)
- Docker & Docker Compose
- Make
- Git

### Initial Setup

1. **Clone and setup**
```bash
git clone https://github.com/your-org/mcp-octo-enigma.git
cd mcp-octo-enigma
make install-tools
```

2. **Environment configuration**
```bash
cp .env.example .env
# Edit .env with your local configuration
```

3. **Start development services**
```bash
make test-db  # Starts PostgreSQL and Redis
```

4. **Run migrations**
```bash
make migrate-up
```

5. **Start development server**
```bash
make dev  # Uses Air for hot reloading
```

## 📁 Project Structure

```
mcp-octo-enigma/
├── cmd/                    # Application entry points
│   ├── server/            # MCP server main
│   └── client/            # MCP client main
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── container/        # Dependency injection container
│   ├── database/         # Database setup and migrations
│   ├── genkit/          # Genkit service integration
│   │   ├── service.go   # Main Genkit service
│   │   ├── providers.go # AI provider implementations
│   │   ├── flow_manager.go     # Flow execution engine
│   │   ├── prompt_manager.go   # Prompt template system
│   │   ├── tool_manager.go     # Tool registry and execution
│   │   └── interrupt_manager.go # Interrupt handling
│   ├── handlers/        # HTTP request handlers
│   ├── logger/          # Structured logging utilities
│   ├── mcp/            # MCP protocol implementation
│   │   ├── manager.go   # Multi-server management
│   │   └── client.go    # MCP client implementation
│   ├── middleware/     # HTTP middleware (auth, logging, etc.)
│   ├── repository/     # Data access layer
│   ├── server/         # HTTP server configuration
│   ├── service/        # Business logic services
│   └── types/          # Shared type definitions
├── tests/                # Test files
│   ├── unit/           # Unit tests
│   └── integration/    # Integration tests
├── migrations/           # Database schema migrations
├── docs/                # Documentation
└── config/              # Configuration files
```

## 🧩 Core Components

### 1. Configuration System (`internal/config`)
- Environment-based configuration
- Validation and defaults
- Type-safe configuration structs

### 2. Dependency Injection (`internal/container`)
- Centralized dependency management
- Service lifecycle management
- Clean testing interfaces

### 3. Genkit Integration (`internal/genkit`)
- **Service**: Main coordinator for AI operations
- **Providers**: AI provider implementations (OpenAI, Google AI, etc.)
- **Flow Manager**: Workflow execution engine
- **Prompt Manager**: Template system with dotprompt support
- **Tool Manager**: Tool registry and execution
- **Interrupt Manager**: Generation control and cancellation

### 4. MCP Implementation (`internal/mcp`)
- **Manager**: Multi-server coordination
- **Client**: Protocol implementation for MCP communication

### 5. HTTP Layer (`internal/server`, `internal/handlers`, `internal/middleware`)
- RESTful API implementation
- Request/response handling
- Authentication and rate limiting

## 🔧 Development Workflow

### Making Changes

1. **Create feature branch**
```bash
git checkout -b feature/your-feature-name
```

2. **Implement changes**
   - Follow Go conventions and project patterns
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
```bash
make test          # Run all tests
make test-unit     # Unit tests only
make test-integration  # Integration tests
make lint          # Code linting
make fmt           # Code formatting
```

4. **Commit and push**
```bash
git add .
git commit -m "feat: descriptive commit message"
git push origin feature/your-feature-name
```

### Code Style Guidelines

#### Go Code Standards
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Write godoc comments for public APIs
- Keep functions small and focused
- Use dependency injection for testability

#### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Use custom error types when appropriate
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}
```

#### Logging
```go
// Use structured logging with logrus
logger.WithFields(logrus.Fields{
    "request_id": requestID,
    "user_id":    userID,
    "operation":  "content_generation",
}).Info("Starting content generation")
```

#### Testing
```go
func TestContentService_GenerateContent(t *testing.T) {
    // Arrange
    mockProvider := new(MockAIProvider)
    service := NewContentService(mockProvider, logger)
    
    // Act
    result, err := service.GenerateContent(ctx, request)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Content)
    mockProvider.AssertExpectations(t)
}
```

## 🧪 Testing Strategy

### Test Categories

1. **Unit Tests** (`tests/unit/`)
   - Test individual components in isolation
   - Use mocks for dependencies
   - Fast execution, no external dependencies

2. **Integration Tests** (`tests/integration/`)
   - Test component interactions
   - Use real database (test instance)
   - Test API endpoints end-to-end

3. **Benchmarks**
   - Performance testing for critical paths
   - Vector search performance
   - AI provider response times

### Running Tests

```bash
# All tests
make test

# Specific test categories
make test-unit
make test-integration

# With coverage
make test-coverage

# Start test database for integration tests
make test-db

# Clean up test environment
make test-db-down
```

### Writing Tests

#### Unit Test Example
```go
func TestPromptManager_RenderPrompt(t *testing.T) {
    // Setup
    manager, err := genkit.NewPromptManager(logger)
    require.NoError(t, err)
    
    // Test data
    req := &genkit.PromptRenderRequest{
        PromptID: "test-prompt",
        Variables: map[string]interface{}{
            "name": "Alice",
            "age":  30,
        },
    }
    
    // Execute
    resp, err := manager.RenderPrompt(req)
    
    // Verify
    assert.NoError(t, err)
    assert.Contains(t, resp.Rendered, "Alice")
    assert.Contains(t, resp.Rendered, "30")
}
```

#### Integration Test Example
```go
func (suite *APITestSuite) TestGenerateContent() {
    // Setup request
    reqBody := map[string]interface{}{
        "model":  "gpt-4",
        "prompt": "Hello, world!",
    }
    
    // Make request
    resp := suite.makeRequest("POST", "/api/v1/content/generate", reqBody)
    
    // Verify response
    assert.Equal(suite.T(), http.StatusOK, resp.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(resp.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    assert.NotEmpty(suite.T(), response["content"])
}
```

## 🔌 Adding New Features

### Adding an AI Provider

1. **Implement the Provider Interface**
```go
// internal/genkit/providers.go
type NewAIProvider struct {
    apiKey string
    logger *logrus.Logger
}

func (p *NewAIProvider) GenerateContent(ctx context.Context, req *GenerateContentRequest) (*GenerateContentResponse, error) {
    // Implementation
}

func (p *NewAIProvider) GenerateContentStream(ctx context.Context, req *GenerateContentRequest) (<-chan *GenerateContentChunk, error) {
    // Implementation
}

func (p *NewAIProvider) EmbedText(ctx context.Context, text string) ([]float64, error) {
    // Implementation
}
```

2. **Update Configuration**
```go
// internal/config/config.go
type AI struct {
    // ... existing providers
    NewProvider NewProviderConfig
}

type NewProviderConfig struct {
    APIKey string
    BaseURL string
}
```

3. **Register in Service**
```go
// internal/genkit/service.go
func (s *Service) initializeProviders() error {
    // ... existing providers
    
    if s.config.AI.NewProvider.APIKey != "" {
        s.newProvider, err = NewNewAIProvider(s.config.AI.NewProvider.APIKey, s.logger)
        if err != nil {
            s.logger.Warnf("Failed to initialize New AI provider: %v", err)
        }
    }
}
```

### Adding a Custom Tool

1. **Define Tool Schema**
```go
// internal/genkit/tool_manager.go
func (tm *ToolManager) registerCustomTool() {
    tool := &types.Tool{
        Name:        "custom-tool",
        Description: "A custom tool for specific operations",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "input": map[string]interface{}{
                    "type":        "string",
                    "description": "Input parameter",
                },
            },
        },
        Required: []string{"input"},
    }
    
    tm.RegisterTool(tool, tm.customToolHandler)
}
```

2. **Implement Tool Handler**
```go
func (tm *ToolManager) customToolHandler(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
    input, ok := args["input"].(string)
    if !ok {
        return nil, fmt.Errorf("input parameter is required")
    }
    
    // Tool logic here
    result := processInput(input)
    
    return map[string]interface{}{
        "result": result,
    }, nil
}
```

### Creating Custom Flows

1. **Define Flow Structure**
```go
// internal/genkit/flow_manager.go
func (fm *FlowManager) registerCustomFlow() {
    flow := &types.Flow{
        ID:          "custom-flow",
        Name:        "Custom Processing Flow",
        Description: "A custom workflow for specific processing",
        Steps: []types.FlowStep{
            {
                ID:          "validate",
                Type:        "validation",
                Name:        "Validate Input",
                Description: "Validate the input parameters",
                Config: map[string]interface{}{
                    "required_fields": []string{"data"},
                },
            },
            {
                ID:          "process",
                Type:        "custom",
                Name:        "Process Data",
                Description: "Custom processing step",
                Dependencies: []string{"validate"},
                Config: map[string]interface{}{
                    "operation": "custom_operation",
                },
            },
        },
    }
    
    fm.flows[flow.ID] = flow
}
```

2. **Implement Step Handler**
```go
func (fm *FlowManager) executeCustomStep(ctx context.Context, step types.FlowStep, context map[string]interface{}) (interface{}, error) {
    operation := step.Config["operation"].(string)
    
    switch operation {
    case "custom_operation":
        // Custom logic here
        return map[string]interface{}{
            "processed": true,
            "result":    "custom result",
        }, nil
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
}
```

## 🐛 Debugging

### Local Debugging

1. **Enable debug logging**
```bash
export LOG_LEVEL=debug
```

2. **Use VS Code debugger**
   - Set breakpoints in your code
   - Use the debug configuration in `.vscode/launch.json`

3. **Debug with delve**
```bash
dlv debug ./cmd/server
```

### Common Issues

#### Database Connection Issues
```bash
# Check PostgreSQL status
docker-compose ps postgres

# View database logs
docker-compose logs postgres

# Connect to database manually
psql -h localhost -p 5432 -U mcp_user -d mcp_octo_enigma
```

#### AI Provider Issues
```bash
# Test API connectivity
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models

# Check API key configuration
echo $OPENAI_API_KEY
```

#### Vector Search Issues
```bash
# Check pgvector extension
psql -h localhost -p 5432 -U mcp_user -d mcp_octo_enigma -c "SELECT * FROM pg_extension WHERE extname = 'vector';"

# Test vector operations
psql -h localhost -p 5432 -U mcp_user -d mcp_octo_enigma -c "SELECT '[1,2,3]'::vector <-> '[1,2,4]'::vector;"
```

## 📊 Performance Optimization

### Database Performance
- Use appropriate indexes for query patterns
- Monitor query performance with `EXPLAIN ANALYZE`
- Use connection pooling
- Consider read replicas for heavy read workloads

### Vector Search Optimization
- Tune HNSW index parameters
- Use appropriate similarity thresholds
- Consider embedding dimensionality vs. accuracy trade-offs
- Implement result caching for common queries

### AI Provider Optimization
- Implement request batching where possible
- Use appropriate model selection for tasks
- Implement response caching
- Monitor and optimize token usage

## 🚀 Deployment

### Environment Setup

1. **Staging Environment**
```bash
# Use staging configuration
cp .env.staging .env

# Deploy to staging
make docker-build
docker-compose -f docker-compose.staging.yml up -d
```

2. **Production Environment**
```bash
# Use production configuration
cp .env.production .env

# Deploy to production
make release
# Deploy using your CI/CD pipeline
```

### Monitoring Setup

1. **Set up Prometheus metrics collection**
2. **Configure Grafana dashboards**
3. **Set up alerting rules**
4. **Configure log aggregation**

## 🤝 Contributing Guidelines

### Pull Request Process

1. Ensure all tests pass
2. Update documentation for new features
3. Add appropriate logging and metrics
4. Follow semantic commit conventions
5. Request review from maintainers

### Code Review Checklist

- [ ] Code follows project conventions
- [ ] Tests are comprehensive and pass
- [ ] Documentation is updated
- [ ] Performance implications considered
- [ ] Security implications reviewed
- [ ] Error handling is appropriate
- [ ] Logging is adequate

## 📚 Additional Resources

- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Genkit Documentation](https://genkit.dev/docs/)
- [Model Context Protocol Specification](https://github.com/modelcontextprotocol/specification)
- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [Gin Framework Guide](https://gin-gonic.com/docs/)

---

For questions or issues, please open a GitHub issue or reach out to the development team.