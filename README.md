# MCP Octo Enigma

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](docker-compose.yml)

**Advanced MCP (Model Context Protocol) Server with Genkit Integration**

MCP Octo Enigma is a sophisticated MCP server implementation built with Go 1.24, featuring advanced Genkit integration, multiple AI providers, vector search with pgvector, and comprehensive observability.

## 🚀 Features

### Core MCP Features
- **Full MCP Protocol Implementation** - Complete server and client implementation
- **Multi-Server Management** - Register, manage, and monitor multiple MCP servers
- **Streaming Support** - Real-time content generation with Server-Sent Events
- **Interrupt Handling** - Pause and resume generation processes

### AI Integration
- **Multiple AI Providers** - OpenAI, Google AI, Anthropic, Vertex AI, Ollama
- **Advanced Genkit Integration** - Content generation, flows, prompts, and tools
- **Vector Search & RAG** - PostgreSQL with pgvector for retrieval-augmented generation
- **Tool Calling** - Extensible tool system with built-in tools

### Advanced Features
- **Flow Management** - Complex workflow orchestration with dependencies
- **Prompt Management** - dotprompt integration with template rendering
- **Evaluation Framework** - Quality, safety, and performance evaluation
- **Observability** - Metrics, tracing, and logging with Prometheus, Grafana, and Jaeger

### Development & Production
- **Hot Reloading** - Air for development with instant code reloading
- **Docker Support** - Complete containerization with Docker Compose
- **Comprehensive Testing** - Unit, integration, and E2E tests
- **Security** - Rate limiting, authentication, and input validation

## 📋 Prerequisites

- **Go 1.24+** - [Install Go](https://golang.org/doc/install)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **PostgreSQL 16+** - For local development (or use Docker)
- **Redis** - For caching (or use Docker)
- **Make** - For build automation

### Optional
- **Ollama** - For local AI models
- **API Keys** - OpenAI, Google AI, Anthropic for cloud providers

## 🛠️ Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd mcp-octo-enigma
make dev-setup
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your API keys and configuration
```

### 3. Start with Docker (Recommended)

```bash
make docker-up
```

### 4. Or Start Locally

```bash
# Install dependencies
make install-tools

# Setup database
make db-setup
make migrate-up

# Build and run
make build
make run
```

### 5. Verify Installation

```bash
# Check health
make health

# View logs
make logs

# Check status
make status
```

## 🏗️ Development

### Project Structure

```
mcp-octo-enigma/
├── cmd/                    # Application binaries
│   ├── server/            # MCP Server
│   └── client/            # MCP Client
├── internal/              # Core application code
│   ├── config/           # Configuration management
│   ├── genkit/          # Genkit service integration
│   ├── mcp/             # MCP protocol implementation
│   ├── server/          # HTTP server with Gin
│   ├── handlers/        # API handlers
│   ├── service/         # Business logic
│   ├── repository/      # Data access layer
│   ├── middleware/      # HTTP middleware
│   ├── client/          # MCP client implementation
│   ├── cache/           # Caching layer
│   └── types/           # Shared types
├── migrations/           # Database migrations
├── tests/               # Test suites
├── docs/               # Documentation
├── monitoring/         # Monitoring configuration
├── docker-compose.yml  # Container orchestration
├── Makefile           # Build automation
└── README.md          # This file
```

### Development Commands

```bash
# Development with hot reloading
make dev

# Run tests
make test

# Code quality checks
make check

# Format code
make format

# Run linter
make lint

# Generate documentation
make docs

# Database operations
make migrate-up
make migrate-down
make db-backup
```

### Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Coverage report
make coverage

# Benchmarks
make bench
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://mcp_user:mcp_password@localhost:5432/mcp_octo_enigma?sslmode=disable` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `MCP_SERVER_PORT` | Server port | `8080` |
| `LOG_LEVEL` | Logging level | `info` |
| `GOOGLE_AI_API_KEY` | Google AI API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |
| `ANTHROPIC_API_KEY` | Anthropic API key | - |
| `API_SECRET_KEY` | JWT secret key | - |

### AI Providers

Configure your AI providers by setting the appropriate API keys in `.env`:

```bash
# Google AI
GOOGLE_AI_API_KEY=your-google-ai-api-key

# OpenAI
OPENAI_API_KEY=your-openai-api-key

# Anthropic
ANTHROPIC_API_KEY=your-anthropic-api-key

# Vertex AI
VERTEX_AI_PROJECT_ID=your-vertex-project-id

# Ollama (local)
OLLAMA_HOST=http://localhost:11434
```

## 🚀 API Usage

### Content Generation

```bash
# Generate content
curl -X POST http://localhost:8080/api/v1/content/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "gpt-4",
    "prompt": "Tell me a story about a brave knight",
    "parameters": {
      "temperature": 0.7,
      "max_tokens": 1000
    }
  }'

# Streaming generation
curl -X POST http://localhost:8080/api/v1/content/generate/stream \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "gpt-4",
    "prompt": "Tell me a story about a brave knight",
    "stream": true
  }'
```

### Flow Execution

```bash
# Execute a flow
curl -X POST http://localhost:8080/api/v1/flows/content-generation/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "input": {
      "prompt": "Write a poem about the ocean",
      "model": "gpt-4"
    }
  }'
```

### Vector Search

```bash
# Search similar documents
curl -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "query": [0.1, 0.2, 0.3],
    "limit": 10,
    "threshold": 0.7
  }'
```

## 📊 Monitoring

### Prometheus Metrics

Access Prometheus at `http://localhost:9090`

### Grafana Dashboards

Access Grafana at `http://localhost:3000` (admin/admin)

### Jaeger Tracing

Access Jaeger at `http://localhost:16686`

### Health Checks

```bash
# Application health
curl http://localhost:8080/health

# Database health
make status
```

## 🐳 Docker Deployment

### Production Deployment

```bash
# Build production images
make docker-build

# Deploy with Docker Compose
make docker-up

# View logs
make docker-logs

# Scale services
docker-compose up -d --scale server=3
```

### Environment-specific Configurations

```bash
# Development
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
```

## 🧪 Testing

### Test Structure

```
tests/
├── unit/               # Unit tests
├── integration/        # Integration tests
├── e2e/               # End-to-end tests
└── fixtures/          # Test data
```

### Running Tests

```bash
# All tests
make test

# Specific test types
make test-unit
make test-integration
make test-e2e

# With coverage
make coverage

# Performance tests
make bench
```

## 📚 Documentation

### API Documentation

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI Spec**: `docs/swagger/swagger.json`

### Generated Documentation

```bash
# Generate Swagger docs
make swagger

# Serve documentation
make docs-serve
```

## 🔒 Security

### Authentication

The API uses JWT-based authentication. Include the token in the Authorization header:

```bash
Authorization: Bearer your-jwt-token
```

### Rate Limiting

- Default: 100 requests per minute per IP
- Configurable via `RATE_LIMIT_REQUESTS_PER_MINUTE`

### Input Validation

- All inputs are validated using Go struct tags
- SQL injection protection via parameterized queries
- XSS protection via content sanitization

## 🚀 Performance

### Optimization Features

- **Connection Pooling** - Database and Redis connection pooling
- **Caching** - Redis-based caching for frequently accessed data
- **Vector Indexing** - HNSW algorithm for fast vector similarity search
- **Async Processing** - Non-blocking operations where possible

### Benchmarks

Run benchmarks to check performance:

```bash
make bench
```

## 🛠️ Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database status
   make status
   
   # Restart database
   docker-compose restart postgres
   ```

2. **API Key Issues**
   ```bash
   # Verify API keys in .env
   cat .env | grep API_KEY
   ```

3. **Port Conflicts**
   ```bash
   # Check port usage
   lsof -i :8080
   
   # Change port in .env
   echo "MCP_SERVER_PORT=8081" >> .env
   ```

### Logs

```bash
# Application logs
make logs

# Docker logs
make docker-logs

# Specific service logs
docker-compose logs server
docker-compose logs postgres
```

### Debug Mode

```bash
# Enable debug logging
echo "LOG_LEVEL=debug" >> .env

# Restart services
make docker-down && make docker-up
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Write tests for new features
- Update documentation
- Ensure all tests pass: `make check`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Genkit](https://github.com/firebase/genkit) - Google's AI framework
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [pgvector](https://github.com/pgvector/pgvector) - Vector similarity search
- [Prometheus](https://prometheus.io/) - Metrics and monitoring
- [Grafana](https://grafana.com/) - Visualization and dashboards

## 📞 Support

- **Documentation**: [Wiki](https://github.com/your-org/mcp-octo-enigma/wiki)
- **Issues**: [GitHub Issues](https://github.com/your-org/mcp-octo-enigma/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/mcp-octo-enigma/discussions)

---

**Made with ❤️ by the MCP Octo Enigma Team**