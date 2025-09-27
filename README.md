# MCP Octo Enigma

A sophisticated Model Context Protocol (MCP) server and client implementation built with Go 1.24, featuring advanced Genkit integration, multiple AI provider support, and comprehensive RAG capabilities with PostgreSQL and pgvector.

## 🚀 Features

### Core MCP Implementation
- **Full MCP Protocol Support**: Complete implementation of the Model Context Protocol specification
- **Multi-Server Management**: Manage and route requests across multiple MCP servers
- **Real-time Communication**: WebSocket and HTTP-based server communication
- **Health Monitoring**: Automatic health checking and failover for MCP servers

### Advanced Genkit Integration
- **Content Generation**: Support for multiple AI providers (OpenAI, Google AI, Anthropic, Vertex AI, Ollama)
- **Flow Management**: Create and execute complex AI workflows with dependency management
- **Prompt Management**: dotprompt template system with variable substitution and versioning
- **Tool Calling**: Extensible tool system with built-in and custom tools
- **Interrupt Handling**: Pause and resume generation processes with graceful interruption

### Vector Database & RAG
- **pgvector Integration**: High-performance vector similarity search
- **Document Indexing**: Automatic embedding generation and storage
- **RAG Workflows**: Retrieval-augmented generation with context injection
- **Semantic Search**: Advanced similarity search with configurable thresholds

### Observability & Evaluation
- **Comprehensive Metrics**: Request tracing, performance monitoring, and usage analytics
- **Structured Logging**: JSON-based logging with correlation IDs
- **Evaluation Framework**: Automated quality assessment for generated content
- **Health Checks**: System health monitoring and alerting

## 🏗️ Architecture

The system follows a layered architecture pattern:

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │ Gin Server  │ │ Middleware  │ │     HTTP Handlers       │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │   Content   │ │    Flow     │ │    Tool & Evaluation    │ │
│  │   Service   │ │   Service   │ │       Services          │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                      Core Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │   Genkit    │ │     MCP     │ │    Vector Repository    │ │
│  │   Service   │ │   Manager   │ │      & Database         │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

For detailed architecture diagrams, see [docs/diagrams.md](docs/diagrams.md).

## 🛠️ Prerequisites

- **Go 1.24+**
- **PostgreSQL 15+** with pgvector extension
- **Redis** (optional, for caching)
- **Docker & Docker Compose** (for containerized deployment)

### AI Provider API Keys (at least one required)
- OpenAI API Key
- Google AI API Key
- Anthropic API Key
- Vertex AI Project ID
- Local Ollama installation

## 📦 Installation

### Option 1: Local Development

1. **Clone the repository**
```bash
git clone https://github.com/your-org/mcp-octo-enigma.git
cd mcp-octo-enigma
```

2. **Install dependencies**
```bash
make deps
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Start PostgreSQL with pgvector**
```bash
make test-db
```

5. **Run migrations**
```bash
make migrate-up
```

6. **Build and run**
```bash
make build
./build/mcp-server
```

### Option 2: Docker Deployment

1. **Clone and configure**
```bash
git clone https://github.com/your-org/mcp-octo-enigma.git
cd mcp-octo-enigma
cp .env.example .env
# Edit .env with your configuration
```

2. **Deploy with Docker Compose**
```bash
make docker-up
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:password@localhost:5432/mcp_octo_enigma?sslmode=disable` |
| `MCP_SERVER_PORT` | Server port | `8080` |
| `GOOGLE_AI_API_KEY` | Google AI API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |
| `ANTHROPIC_API_KEY` | Anthropic API key | - |
| `VERTEX_AI_PROJECT_ID` | Vertex AI project ID | - |
| `OLLAMA_HOST` | Ollama server URL | `http://localhost:11434` |
| `LOG_LEVEL` | Logging level | `info` |

### Configuration File

Advanced configuration can be managed through the config system. See [internal/config/config.go](internal/config/config.go) for all available options.

## 🔧 Usage

### Starting the Server

```bash
# Development mode
make dev

# Production mode
./build/mcp-server
```

### Basic API Examples

#### Generate Content
```bash
curl -X POST http://localhost:8080/api/v1/content/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "gpt-4",
    "prompt": "Explain quantum computing in simple terms"
  }'
```

#### Execute a Flow
```bash
curl -X POST http://localhost:8080/api/v1/flows/content-generation/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "input": {
      "prompt": "Write a blog post about AI",
      "model": "gpt-4"
    }
  }'
```

#### Call a Tool
```bash
curl -X POST http://localhost:8080/api/v1/tools/call \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "tool_name": "calculator",
    "arguments": {
      "expression": "2 + 2 * 3"
    }
  }'
```

#### Vector Search
```bash
curl -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "query": [0.1, 0.2, 0.3, ...],
    "limit": 10,
    "threshold": 0.8
  }'
```

## 🔄 API Documentation

### Core Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/content/generate` | Generate content |
| `POST` | `/api/v1/content/generate/stream` | Stream content generation |
| `POST` | `/api/v1/content/interrupt` | Interrupt generation |

### Flow Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/flows/` | List all flows |
| `POST` | `/api/v1/flows/` | Create new flow |
| `GET` | `/api/v1/flows/{id}` | Get flow details |
| `PUT` | `/api/v1/flows/{id}` | Update flow |
| `DELETE` | `/api/v1/flows/{id}` | Delete flow |
| `POST` | `/api/v1/flows/{id}/execute` | Execute flow |

### Tool Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/tools/` | List available tools |
| `POST` | `/api/v1/tools/call` | Execute tool |
| `POST` | `/api/v1/tools/register` | Register new tool |

### Vector Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/vectors/embed` | Generate embeddings |
| `POST` | `/api/v1/vectors/search` | Vector similarity search |
| `POST` | `/api/v1/vectors/index` | Index document |

### MCP Server Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/mcp/servers` | List MCP servers |
| `POST` | `/api/v1/mcp/servers` | Register MCP server |
| `DELETE` | `/api/v1/mcp/servers/{id}` | Unregister MCP server |

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Unit Tests Only
```bash
make test-unit
```

### Integration Tests
```bash
make test-integration
```

### Test Coverage
```bash
make test-coverage
```

### Test with Database
```bash
make test-db        # Start test database
make test          # Run tests
make test-db-down  # Clean up
```

## 🚀 Development

### Development Setup
```bash
# Install development tools
make install-tools

# Start development environment
make dev

# Format code
make fmt

# Lint code
make lint

# Run security scan
make security
```

### Project Structure
```
mcp-octo-enigma/
├── cmd/                    # Application entry points
│   ├── server/            # MCP server binary
│   └── client/            # MCP client binary
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── container/        # Dependency injection
│   ├── database/         # Database operations
│   ├── genkit/          # Genkit integration
│   ├── handlers/        # HTTP handlers
│   ├── logger/          # Logging utilities
│   ├── mcp/            # MCP protocol implementation
│   ├── middleware/     # HTTP middleware
│   ├── repository/     # Data access layer
│   ├── server/         # HTTP server setup
│   ├── service/        # Business logic
│   └── types/          # Common types
├── tests/                # Test files
│   ├── unit/           # Unit tests
│   └── integration/    # Integration tests
├── migrations/           # Database migrations
├── docs/                # Documentation
└── config/              # Configuration files
```

### Adding New Features

1. **Add New AI Provider**
   - Implement the `AIProvider` interface in `internal/genkit/providers.go`
   - Add provider configuration to `internal/config/config.go`
   - Update model detection logic in `internal/genkit/service.go`

2. **Create Custom Tools**
   - Implement `ToolHandlerFunc` in `internal/genkit/tool_manager.go`
   - Register tool in the tool manager
   - Add appropriate tests

3. **Extend Flow Types**
   - Add new step types in `internal/genkit/flow_manager.go`
   - Implement step execution logic
   - Update flow templates

## 📊 Monitoring & Observability

### Metrics Endpoints
- `GET /api/v1/observability/metrics` - System metrics
- `GET /api/v1/observability/traces` - Distributed traces

### Available Metrics
- Request count and latency
- AI provider usage and costs
- Vector search performance
- Error rates and types
- MCP server health status

### Logging
All logs are structured JSON with the following fields:
- `timestamp` - ISO 8601 timestamp
- `level` - Log level (debug, info, warn, error)
- `message` - Log message
- `request_id` - Unique request identifier
- `service` - Service name
- Additional contextual fields

## 🔒 Security

### Authentication
- API key authentication for all endpoints
- Request validation and sanitization
- Rate limiting per client

### Data Protection
- Encrypted database connections
- Secure API key storage
- Input validation and output sanitization

### Best Practices
- Regular dependency updates
- Security scanning with `gosec`
- Vulnerability checking with `govulncheck`

## 🚢 Deployment

### Production Deployment

1. **Build release binaries**
```bash
make release
```

2. **Deploy with Docker**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

3. **Configure monitoring**
   - Set up Prometheus metrics collection
   - Configure Grafana dashboards
   - Set up alerting rules

### Environment-Specific Configurations

- **Development**: Full logging, debug endpoints enabled
- **Staging**: Production-like setup with test data
- **Production**: Minimal logging, security hardened, monitoring enabled

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run the test suite (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Guidelines
- Follow Go best practices and idioms
- Write comprehensive tests
- Add documentation for new features
- Use meaningful commit messages
- Ensure all CI checks pass

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Model Context Protocol](https://github.com/modelcontextprotocol) specification
- [Google Genkit](https://github.com/firebase/genkit) framework
- [pgvector](https://github.com/pgvector/pgvector) for vector similarity search
- [Gin](https://github.com/gin-gonic/gin) web framework
- All the amazing AI providers and open-source contributors

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Email**: support@your-domain.com

---

**Made with ❤️ by the MCP Octo Enigma Team**