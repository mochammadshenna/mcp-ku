# MCP Octo Enigma - System Diagrams

This document contains Mermaid.js diagrams that illustrate the architecture and flow of the MCP Octo Enigma system.

## 1. Flow Control: Request Path from Controller to Database

```mermaid
graph TD
    A[Client Request] --> B[Gin Router]
    B --> C[Middleware Layer]
    C --> D[Authentication]
    C --> E[Logging]
    C --> F[Rate Limiting]
    C --> G[CORS]
    
    D --> H[Handler Function]
    E --> H
    F --> H
    G --> H
    
    H --> I[Service Layer]
    I --> J[Business Logic Validation]
    J --> K[Genkit Service]
    J --> L[MCP Manager]
    J --> M[Repository Layer]
    
    K --> N[AI Provider Selection]
    N --> O[OpenAI]
    N --> P[Google AI]
    N --> Q[Anthropic]
    N --> R[Vertex AI]
    N --> S[Ollama]
    
    L --> T[MCP Server Communication]
    
    M --> U[PostgreSQL Database]
    U --> V[Vector Operations]
    U --> W[CRUD Operations]
    
    V --> X[pgvector Extension]
    W --> Y[Standard Tables]
    
    Z[Response] <-- H
    Z --> AA[JSON Serialization]
    AA --> BB[HTTP Response]
    BB --> A
```

## 2. Data Lineage: Variable Tracing Example

This diagram traces how a user prompt flows through the system to generate content and store vectors:

```mermaid
graph LR
    A[User Prompt Input] --> B[Handler Validation]
    B --> C[ContentService.GenerateContent]
    C --> D[Request Preparation]
    D --> E[Model Selection Logic]
    E --> F[AI Provider Router]
    
    F --> G[Genkit Service]
    G --> H[OpenAI Provider]
    H --> I[API Request Transformation]
    I --> J[OpenAI API Call]
    J --> K[Raw API Response]
    
    K --> L[Response Processing]
    L --> M[Content Extraction]
    M --> N[Usage Info Parsing]
    N --> O[Metadata Assembly]
    
    O --> P[Service Response Object]
    P --> Q[Database Storage]
    Q --> R[Generation Record]
    
    P --> S[Vector Processing]
    S --> T[Embedding Generation]
    T --> U[Vector Storage]
    U --> V[pgvector Table]
    
    P --> W[HTTP Response]
    W --> X[Client Response]
    
    subgraph "Transformation Points"
        Y[prompt → genkit.GenerateContentRequest]
        Z[openai.Response → genkit.GenerateContentResponse]
        AA[response → types.Generation]
        BB[content → vector embedding]
    end
```

## 3. Component Structure: Service Architecture

```mermaid
graph TB
    subgraph "Presentation Layer"
        A[Gin HTTP Server]
        B[Middleware Stack]
        C[Handler Functions]
    end
    
    subgraph "Service Layer"
        D[Content Service]
        E[Flow Service]
        F[Tool Service]
        G[Evaluation Service]
        H[Observability Service]
    end
    
    subgraph "Core Components"
        I[Genkit Service]
        J[MCP Manager]
        K[Vector Repository]
        L[Interrupt Manager]
    end
    
    subgraph "AI Providers"
        M[OpenAI Provider]
        N[Google AI Provider]
        O[Anthropic Provider]
        P[Vertex AI Provider]
        Q[Ollama Provider]
    end
    
    subgraph "MCP Ecosystem"
        R[MCP Client]
        S[Local Tools Server]
        T[Vector Search Server]
        U[External MCP Servers]
    end
    
    subgraph "Data Layer"
        V[PostgreSQL]
        W[pgvector Extension]
        X[Redis Cache]
    end
    
    subgraph "Management Components"
        Y[Flow Manager]
        Z[Prompt Manager]
        AA[Tool Manager]
        BB[Configuration Manager]
    end
    
    A --> B
    B --> C
    C --> D
    C --> E
    C --> F
    C --> G
    C --> H
    
    D --> I
    E --> I
    F --> I
    G --> I
    
    I --> M
    I --> N
    I --> O
    I --> P
    I --> Q
    
    D --> J
    F --> J
    J --> R
    R --> S
    R --> T
    R --> U
    
    D --> K
    G --> K
    K --> V
    K --> W
    
    I --> Y
    I --> Z
    I --> AA
    I --> L
    
    Y --> V
    Z --> V
    AA --> V
    
    H --> X
```

## 4. MCP Server Communication Flow

```mermaid
sequenceDiagram
    participant Client
    participant MCPManager
    participant MCPClient
    participant LocalServer
    participant VectorServer
    participant ExternalServer
    
    Client->>MCPManager: Request with server preference
    MCPManager->>MCPManager: Route request to appropriate server
    
    alt Specific Server Requested
        MCPManager->>MCPClient: Connect to specified server
        MCPClient->>LocalServer: Send MCP request
        LocalServer-->>MCPClient: MCP response
        MCPClient-->>MCPManager: Processed response
    else Broadcast to All Servers
        MCPManager->>MCPClient: Connect to all active servers
        par Parallel Requests
            MCPClient->>LocalServer: Broadcast request
            MCPClient->>VectorServer: Broadcast request
            MCPClient->>ExternalServer: Broadcast request
        end
        par Parallel Responses
            LocalServer-->>MCPClient: Response 1
            VectorServer-->>MCPClient: Response 2
            ExternalServer-->>MCPClient: Response 3
        end
        MCPClient-->>MCPManager: Aggregated responses
    end
    
    MCPManager-->>Client: Final response
```

## 5. Content Generation with RAG Flow

```mermaid
graph TD
    A[User Query] --> B[Query Analysis]
    B --> C{RAG Required?}
    
    C -->|Yes| D[Vector Search]
    C -->|No| E[Direct Generation]
    
    D --> F[Query Embedding]
    F --> G[pgvector Search]
    G --> H[Retrieve Similar Documents]
    H --> I[Context Assembly]
    I --> J[Augmented Prompt Creation]
    
    J --> K[AI Provider Selection]
    E --> K
    
    K --> L[Content Generation]
    L --> M[Response Processing]
    M --> N[Tool Calls Processing]
    N --> O[Final Response Assembly]
    
    O --> P[Store Generation Record]
    P --> Q[Update Metrics]
    Q --> R[Return to User]
    
    subgraph "Vector Operations"
        S[Document Indexing]
        T[Embedding Generation]
        U[Similarity Search]
        V[Result Ranking]
    end
    
    F --> T
    G --> U
    H --> V
```

## 6. Tool Calling Architecture

```mermaid
graph LR
    A[Tool Call Request] --> B[Tool Manager]
    B --> C[Tool Registry Lookup]
    C --> D{Tool Found?}
    
    D -->|Yes| E[Parameter Validation]
    D -->|No| F[Tool Not Found Error]
    
    E --> G{Valid Parameters?}
    G -->|Yes| H[Tool Handler Execution]
    G -->|No| I[Validation Error]
    
    H --> J[Tool Logic]
    J --> K{Tool Type}
    
    K -->|Built-in| L[Local Execution]
    K -->|MCP Tool| M[MCP Server Call]
    K -->|External API| N[API Call]
    
    L --> O[Process Result]
    M --> O
    N --> O
    
    O --> P[Tool Response]
    P --> Q[Response Formatting]
    Q --> R[Return to Caller]
    
    subgraph "Built-in Tools"
        S[Calculator]
        T[Text Analyzer]
        U[Web Search]
        V[Code Generator]
    end
    
    L --> S
    L --> T
    L --> U
    L --> V
```

## 7. Interrupt Handling System

```mermaid
stateDiagram-v2
    [*] --> RequestReceived: New generation request
    RequestReceived --> Processing: Start generation
    Processing --> CheckInterrupt: Periodic check
    
    CheckInterrupt --> Processing: No interrupt
    CheckInterrupt --> Interrupted: Interrupt detected
    CheckInterrupt --> Completed: Generation finished
    
    Interrupted --> Acknowledged: User acknowledges
    Interrupted --> Timeout: Timeout reached
    
    Acknowledged --> [*]: Cleanup
    Timeout --> [*]: Auto cleanup
    Completed --> [*]: Success
    
    Processing --> Failed: Error occurred
    Failed --> [*]: Error handling
    
    note right of CheckInterrupt
        Interrupt checks happen:
        - Every 100ms during generation
        - On streaming chunk
        - Before tool calls
    end note
```

## 8. Observability and Metrics Flow

```mermaid
graph TD
    A[Application Events] --> B[Metrics Collection]
    B --> C[Structured Logging]
    B --> D[Performance Metrics]
    B --> E[Error Tracking]
    
    C --> F[Log Aggregation]
    D --> G[Metrics Storage]
    E --> H[Error Analysis]
    
    F --> I[Logrus JSON Output]
    G --> J[Prometheus Format]
    H --> K[Error Reporting]
    
    I --> L[Log Storage]
    J --> M[Metrics Dashboard]
    K --> N[Alert System]
    
    subgraph "Observability Components"
        O[Request Tracing]
        P[Database Monitoring]
        Q[AI Provider Metrics]
        R[MCP Server Health]
    end
    
    B --> O
    B --> P
    B --> Q
    B --> R
    
    subgraph "Dashboards"
        S[System Health]
        T[Performance KPIs]
        U[Error Rates]
        V[Usage Analytics]
    end
    
    M --> S
    M --> T
    M --> U
    M --> V
```

## 9. Database Schema Relationships

```mermaid
erDiagram
    MCP_SERVERS {
        uuid id PK
        string name
        string url
        text description
        string_array capabilities
        jsonb config
        string status
        timestamp created_at
        timestamp updated_at
    }
    
    FLOWS {
        uuid id PK
        string name
        text description
        jsonb input
        jsonb output
        jsonb steps
        jsonb config
        timestamp created_at
        timestamp updated_at
    }
    
    FLOW_EXECUTIONS {
        uuid id PK
        uuid flow_id FK
        jsonb input
        jsonb output
        string status
        text error
        timestamp started_at
        timestamp ended_at
        jsonb metadata
    }
    
    PROMPTS {
        uuid id PK
        string name
        text template
        string_array variables
        jsonb config
        integer version
        timestamp created_at
        timestamp updated_at
    }
    
    VECTOR_DOCUMENTS {
        uuid id PK
        text content
        vector_768 embedding
        jsonb metadata
        string source
        timestamp created_at
    }
    
    GENERATIONS {
        uuid id PK
        string model
        text prompt
        text response
        jsonb parameters
        jsonb metadata
        string status
        string request_id
        timestamp created_at
        timestamp completed_at
    }
    
    EVALUATIONS {
        uuid id PK
        string name
        text description
        string type
        jsonb config
        timestamp created_at
        timestamp updated_at
    }
    
    EVALUATION_RESULTS {
        uuid id PK
        uuid evaluation_id FK
        uuid generation_id FK
        decimal score
        jsonb metrics
        jsonb details
        timestamp created_at
    }
    
    TOOLS {
        uuid id PK
        string name
        text description
        jsonb parameters
        string_array required_params
        string handler
        jsonb config
        timestamp created_at
        timestamp updated_at
    }
    
    TOOL_EXECUTIONS {
        uuid id PK
        uuid tool_id FK
        jsonb arguments
        jsonb result
        boolean success
        text error
        string request_id
        timestamp created_at
        timestamp completed_at
    }
    
    FLOWS ||--o{ FLOW_EXECUTIONS : "executes"
    EVALUATIONS ||--o{ EVALUATION_RESULTS : "evaluates"
    GENERATIONS ||--o{ EVALUATION_RESULTS : "evaluated_by"
    TOOLS ||--o{ TOOL_EXECUTIONS : "executed_as"
```

## 10. Deployment Architecture

```mermaid
graph TB
    subgraph "Load Balancer"
        A[Nginx/HAProxy]
    end
    
    subgraph "Application Tier"
        B[MCP Server Instance 1]
        C[MCP Server Instance 2]
        D[MCP Server Instance 3]
    end
    
    subgraph "Database Tier"
        E[PostgreSQL Primary]
        F[PostgreSQL Replica]
        G[Redis Cache]
    end
    
    subgraph "AI Services"
        H[OpenAI API]
        I[Google AI API]
        J[Anthropic API]
        K[Vertex AI]
        L[Local Ollama]
    end
    
    subgraph "MCP Ecosystem"
        M[Local Tools MCP]
        N[Vector Search MCP]
        O[External MCP Servers]
    end
    
    subgraph "Monitoring"
        P[Prometheus]
        Q[Grafana]
        R[AlertManager]
    end
    
    subgraph "Storage"
        S[Persistent Volumes]
        T[Backup Storage]
    end
    
    A --> B
    A --> C
    A --> D
    
    B --> E
    C --> E
    D --> E
    
    E --> F
    B --> G
    C --> G
    D --> G
    
    B --> H
    B --> I
    B --> J
    B --> K
    B --> L
    
    B --> M
    B --> N
    B --> O
    
    B --> P
    C --> P
    D --> P
    
    P --> Q
    P --> R
    
    E --> S
    F --> S
    G --> S
    
    S --> T
```

These diagrams provide a comprehensive view of the MCP Octo Enigma system architecture, data flow, and component interactions. They can be rendered using any Mermaid.js compatible viewer or integrated into documentation platforms that support Mermaid syntax.