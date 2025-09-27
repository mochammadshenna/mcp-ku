# mcp-octo-enigma

Go 1.24, Gin-based service scaffolding for MCP server and client, Genkit-ready stubs, and PostgreSQL/pgvector RAG.

## Diagrams

### Flow control
```mermaid
sequenceDiagram
    participant Client
    participant Controller as Gin Router
    participant Service as RAG Service
    participant DB as PostgreSQL/pgvector

    Client->>Controller: GET /api/v1/search?q=...
    Controller->>Service: Query(q, k)
    Service->>DB: SELECT ... ORDER BY embedding <-> $1 LIMIT k
    DB-->>Service: Rows
    Service-->>Controller: []Document
    Controller-->>Client: 200 JSON
```

### Data lineage (endpoint.path)
```mermaid
sequenceDiagram
    participant Client
    participant Router
    participant Handler
    participant Service
    participant Repo

    Client->>Router: GET /api/v1/search?q=hello
    Router->>Handler: search(q)
    Handler->>Service: Query(q, k)
    Service->>Repo: SearchKNN(embed(q), k)
    Repo-->>Service: []Document
    Service-->>Handler: []Document
    Handler-->>Client: 200 JSON
```

### Structure
```mermaid
graph TD
  A[cmd/server] --> B[internal/http/handlers]
  B --> C[internal/core/services]
  C --> D[internal/core/repository]
  D --> E[pkg/db]
  C --> F[internal/genkit/*]
  A --> G[internal/mcp/server]
  H[cmd/client] --> I[internal/mcp/client]
```

## Setup
- Env: SERVER_PORT, APP_ENV, DATABASE_URL
- Run: `go run ./cmd/server`
- Test: `go test ./...`

## Notes
- Genkit Go integrations (models, flows, dotprompt, tool-calling, interrupts) are stubbed pending SDK wiring.
- RAG uses pgvector; run migrations to create `documents`.
