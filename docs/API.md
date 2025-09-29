# API Documentation

This document provides comprehensive API documentation for the MCP Octo Enigma service.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints require authentication using Bearer tokens:

```
Authorization: Bearer <your-api-token>
```

## Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "request_id": "uuid"
}
```

## Content Generation

### Generate Content

Generate content using AI models.

**Endpoint:** `POST /content/generate`

**Request Body:**
```json
{
  "model": "gpt-4",
  "prompt": "Write a short story about AI",
  "parameters": {
    "max_tokens": 1000,
    "temperature": 0.7
  },
  "tools": [
    {
      "name": "web_search",
      "description": "Search the web for information",
      "parameters": {
        "type": "object",
        "properties": {
          "query": {
            "type": "string",
            "description": "Search query"
          }
        }
      }
    }
  ],
  "request_id": "optional-request-id"
}
```

**Response:**
```json
{
  "id": "gen_123456",
  "content": "Generated content here...",
  "model": "gpt-4",
  "metadata": {
    "finish_reason": "stop"
  },
  "tool_calls": [
    {
      "id": "call_123",
      "name": "web_search",
      "arguments": {
        "query": "AI latest developments"
      }
    }
  ],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 200,
    "total_tokens": 250
  },
  "request_id": "req_123456",
  "status": "completed",
  "created_at": "2024-01-01T12:00:00Z",
  "completed_at": "2024-01-01T12:00:05Z"
}
```

### Stream Content Generation

Generate content with real-time streaming.

**Endpoint:** `POST /content/generate/stream`

**Request Body:** Same as generate content with `"stream": true`

**Response:** Server-Sent Events (SSE) stream
```
data: {"id": "gen_123", "content": "Hello", "done": false}

data: {"id": "gen_123", "content": " world", "done": false}

data: {"id": "gen_123", "content": "!", "done": true}
```

### Interrupt Generation

Interrupt an ongoing content generation.

**Endpoint:** `POST /content/interrupt`

**Request Body:**
```json
{
  "request_id": "req_123456"
}
```

**Response:**
```json
{
  "message": "Generation interrupted successfully"
}
```

## Flow Management

### List Flows

Get all available flows.

**Endpoint:** `GET /flows/`

**Response:**
```json
{
  "flows": [
    {
      "id": "content-generation",
      "name": "Content Generation",
      "description": "Generate content using AI models",
      "input": {
        "prompt": "string",
        "model": "string"
      },
      "output": {
        "content": "string"
      },
      "steps": [
        {
          "id": "validate-input",
          "type": "validation",
          "name": "Validate Input",
          "description": "Validate the input parameters",
          "config": {
            "required_fields": ["prompt", "model"]
          }
        }
      ],
      "config": {},
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### Create Flow

Create a new flow.

**Endpoint:** `POST /flows/`

**Request Body:**
```json
{
  "name": "Custom Flow",
  "description": "A custom processing flow",
  "input": {
    "data": "string"
  },
  "output": {
    "result": "string"
  },
  "steps": [
    {
      "id": "step1",
      "type": "generate",
      "name": "Generate Content",
      "description": "Generate content based on input",
      "config": {
        "model": "gpt-4",
        "prompt": "Process: {{.data}}"
      }
    }
  ],
  "config": {
    "timeout": "30s"
  }
}
```

**Response:**
```json
{
  "id": "flow_123456",
  "name": "Custom Flow",
  "description": "A custom processing flow",
  "input": {
    "data": "string"
  },
  "output": {
    "result": "string"
  },
  "steps": [...],
  "config": {...},
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Execute Flow

Execute a flow with input data.

**Endpoint:** `POST /flows/{flow_id}/execute`

**Request Body:**
```json
{
  "input": {
    "prompt": "Tell me about quantum computing",
    "model": "gpt-4"
  },
  "parameters": {
    "timeout": "60s"
  }
}
```

**Response:**
```json
{
  "flow_id": "content-generation",
  "output": {
    "result": "Generated content about quantum computing..."
  },
  "request_id": "req_123456",
  "status": "completed",
  "metadata": {
    "execution_time": "5.2s",
    "steps_completed": 2
  }
}
```

## Tool Management

### List Tools

Get all available tools.

**Endpoint:** `GET /tools/`

**Response:**
```json
{
  "tools": [
    {
      "name": "calculator",
      "description": "Perform mathematical calculations",
      "parameters": {
        "type": "object",
        "properties": {
          "expression": {
            "type": "string",
            "description": "Mathematical expression to evaluate"
          }
        }
      },
      "required": ["expression"]
    },
    {
      "name": "web_search",
      "description": "Search the web for information",
      "parameters": {
        "type": "object",
        "properties": {
          "query": {
            "type": "string",
            "description": "Search query"
          },
          "limit": {
            "type": "number",
            "description": "Maximum number of results",
            "default": 5
          }
        }
      },
      "required": ["query"]
    }
  ]
}
```

### Call Tool

Execute a tool with arguments.

**Endpoint:** `POST /tools/call`

**Request Body:**
```json
{
  "tool_name": "calculator",
  "arguments": {
    "expression": "2 + 2 * 3"
  },
  "request_id": "req_123456"
}
```

**Response:**
```json
{
  "tool_name": "calculator",
  "result": {
    "expression": "2 + 2 * 3",
    "result": "8",
    "value": 8
  },
  "success": true,
  "request_id": "req_123456"
}
```

### Register Tool

Register a new tool (for external tools).

**Endpoint:** `POST /tools/register`

**Request Body:**
```json
{
  "name": "custom-tool",
  "description": "A custom tool for specific operations",
  "parameters": {
    "type": "object",
    "properties": {
      "input": {
        "type": "string",
        "description": "Input parameter"
      }
    }
  },
  "required": ["input"],
  "handler": "external_api",
  "config": {
    "endpoint": "https://api.example.com/tool",
    "method": "POST"
  }
}
```

## Vector Operations

### Generate Embeddings

Create vector embeddings for text.

**Endpoint:** `POST /vectors/embed`

**Request Body:**
```json
{
  "text": "This is a sample text for embedding",
  "model": "text-embedding-ada-002"
}
```

**Response:**
```json
{
  "embedding": [0.1, 0.2, 0.3, ..., 0.768]
}
```

### Vector Search

Search for similar vectors.

**Endpoint:** `POST /vectors/search`

**Request Body:**
```json
{
  "query": [0.1, 0.2, 0.3, ..., 0.768],
  "limit": 10,
  "threshold": 0.8
}
```

**Response:**
```json
{
  "results": [
    {
      "document": {
        "id": "doc_123",
        "content": "Similar document content",
        "embedding": [0.1, 0.2, 0.3, ...],
        "metadata": {
          "source": "document.pdf",
          "page": 1
        },
        "source": "document.pdf",
        "created_at": "2024-01-01T12:00:00Z"
      },
      "score": 0.92
    }
  ]
}
```

### Index Document

Index a document with vector embeddings.

**Endpoint:** `POST /vectors/index`

**Request Body:**
```json
{
  "content": "This is the document content to be indexed",
  "source": "document.pdf",
  "metadata": {
    "page": 1,
    "section": "introduction"
  },
  "model": "text-embedding-ada-002"
}
```

**Response:**
```json
{
  "document_id": "doc_123456"
}
```

## MCP Server Management

### List MCP Servers

Get all registered MCP servers.

**Endpoint:** `GET /mcp/servers`

**Response:**
```json
{
  "servers": [
    {
      "id": "local-tools",
      "config": {
        "id": "local-tools",
        "name": "Local Tools Server",
        "url": "http://localhost:8081",
        "description": "Local MCP server for tool management",
        "capabilities": ["tools", "prompts"],
        "status": "active"
      },
      "status": "active",
      "last_ping": "2024-01-01T12:00:00Z",
      "error_count": 0,
      "capabilities": ["tools", "prompts"],
      "metadata": {}
    }
  ]
}
```

### Register MCP Server

Register a new MCP server.

**Endpoint:** `POST /mcp/servers`

**Request Body:**
```json
{
  "name": "Custom MCP Server",
  "url": "http://localhost:8082",
  "description": "A custom MCP server",
  "capabilities": ["tools", "resources"],
  "config": {
    "timeout": "30s",
    "retry_count": 3
  }
}
```

**Response:**
```json
{
  "id": "server_123456",
  "name": "Custom MCP Server",
  "url": "http://localhost:8082",
  "description": "A custom MCP server",
  "capabilities": ["tools", "resources"],
  "config": {
    "timeout": "30s",
    "retry_count": 3
  },
  "status": "inactive",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Unregister MCP Server

Remove an MCP server.

**Endpoint:** `DELETE /mcp/servers/{server_id}`

**Response:**
```json
{
  "message": "Server unregistered successfully"
}
```

## Evaluation

### Run Evaluation

Execute an evaluation on generated content.

**Endpoint:** `POST /evaluation/run`

**Request Body:**
```json
{
  "evaluation_id": "quality-eval",
  "generation_id": "gen_123456",
  "config": {
    "metrics": ["accuracy", "relevance", "coherence"],
    "reference": "Expected output for comparison"
  }
}
```

**Response:**
```json
{
  "id": "eval_123456",
  "evaluation_id": "quality-eval",
  "generation_id": "gen_123456",
  "score": 0.85,
  "metrics": {
    "accuracy": 0.9,
    "relevance": 0.8,
    "coherence": 0.85
  },
  "details": {
    "strengths": ["Clear explanation", "Good structure"],
    "weaknesses": ["Minor factual error"]
  },
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Get Evaluation Results

Retrieve evaluation results.

**Endpoint:** `GET /evaluation/results`

**Query Parameters:**
- `evaluation_id` (optional): Filter by evaluation ID
- `limit` (default: 10): Number of results to return
- `offset` (default: 0): Offset for pagination

**Response:**
```json
{
  "results": [
    {
      "id": "eval_123456",
      "evaluation_id": "quality-eval",
      "score": 0.85,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 100,
  "limit": 10,
  "offset": 0
}
```

## Observability

### Get Metrics

Retrieve system metrics.

**Endpoint:** `GET /observability/metrics`

**Response:**
```json
{
  "requests_total": 1000,
  "requests_success": 950,
  "requests_failed": 50,
  "avg_response_time": "150ms",
  "uptime": "24h30m",
  "memory_usage": "512MB",
  "cpu_usage": "25%",
  "active_connections": 42,
  "ai_provider_calls": {
    "openai": 300,
    "google_ai": 200,
    "anthropic": 150
  },
  "vector_operations": {
    "searches": 500,
    "indexes": 100
  }
}
```

### Get Traces

Retrieve distributed traces.

**Endpoint:** `GET /observability/traces`

**Query Parameters:**
- `operation` (optional): Filter by operation name
- `limit` (default: 50): Number of traces to return
- `since` (optional): ISO 8601 timestamp for filtering

**Response:**
```json
{
  "traces": [
    {
      "id": "trace_123456",
      "operation": "content_generation",
      "start_time": "2024-01-01T12:00:00Z",
      "end_time": "2024-01-01T12:00:05Z",
      "duration": "5.2s",
      "status": "ok",
      "tags": {
        "model": "gpt-4",
        "user_id": "user_123"
      },
      "logs": [
        {
          "timestamp": "2024-01-01T12:00:01Z",
          "level": "info",
          "message": "Starting content generation",
          "fields": {
            "request_id": "req_123"
          }
        }
      ]
    }
  ]
}
```

## Health Check

### System Health

Check system health status.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "genkit": "healthy",
    "mcp": "healthy",
    "redis": "healthy"
  },
  "uptime": "24h30m15s"
}
```

## Rate Limiting

All endpoints are subject to rate limiting:

- **Default**: 100 requests per minute per API key
- **Rate limit headers** are included in responses:
  - `X-RateLimit-Limit`: Request limit per window
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Time when the rate limit resets

When rate limit is exceeded:

```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60
}
```

## Webhooks

### Webhook Events

The system can send webhooks for certain events:

- `generation.completed`: When content generation completes
- `generation.failed`: When content generation fails
- `evaluation.completed`: When evaluation completes
- `server.registered`: When new MCP server is registered

**Webhook Payload Example:**
```json
{
  "event": "generation.completed",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "id": "gen_123456",
    "model": "gpt-4",
    "status": "completed",
    "request_id": "req_123456"
  }
}
```

## SDKs and Libraries

### Go Client
```go
package main

import (
    "context"
    "github.com/your-org/mcp-octo-enigma-client-go"
)

func main() {
    client := mcpclient.New("http://localhost:8080", "your-api-key")
    
    resp, err := client.GenerateContent(context.Background(), &mcpclient.GenerateRequest{
        Model:  "gpt-4",
        Prompt: "Hello, world!",
    })
}
```

### Python Client
```python
from mcp_octo_enigma import Client

client = Client("http://localhost:8080", "your-api-key")

response = client.generate_content(
    model="gpt-4",
    prompt="Hello, world!"
)
```

### JavaScript/TypeScript Client
```typescript
import { MCPClient } from 'mcp-octo-enigma-client';

const client = new MCPClient('http://localhost:8080', 'your-api-key');

const response = await client.generateContent({
  model: 'gpt-4',
  prompt: 'Hello, world!'
});
```

---

For more examples and advanced usage patterns, see the [examples](./examples/) directory.