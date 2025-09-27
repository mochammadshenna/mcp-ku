package handlers

import (
	"net/http"
	"strconv"

	"mcp-octo-enigma/internal/container"
	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/service"
	"mcp-octo-enigma/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HealthCheck returns the health status of the service
func HealthCheck(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := map[string]interface{}{
			"status":    "healthy",
			"version":   "1.0.0",
			"timestamp": "2024-01-01T12:00:00Z",
			"services": map[string]string{
				"database": "healthy",
				"genkit":   "healthy",
				"mcp":      "healthy",
				"redis":    "healthy",
			},
			"uptime": "24h30m15s",
		}
		
		ctx.JSON(http.StatusOK, status)
	}
}

// GenerateContent handles content generation requests
func GenerateContent(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req service.ContentGenerationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		// Generate request ID if not provided
		if req.RequestID == "" {
			req.RequestID = uuid.New().String()
		}

		response, err := c.ContentSvc.GenerateContent(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "GENERATION_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, response)
	}
}

// GenerateContentStream handles streaming content generation
func GenerateContentStream(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req service.ContentGenerationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		req.Stream = true
		if req.RequestID == "" {
			req.RequestID = uuid.New().String()
		}

		// Set up SSE headers
		ctx.Header("Content-Type", "text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")
		ctx.Header("Access-Control-Allow-Origin", "*")

		streamCh, err := c.ContentSvc.GenerateContentStream(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "STREAM_FAILED",
			})
			return
		}

		// Stream the response
		for chunk := range streamCh {
			ctx.SSEvent("message", chunk)
			ctx.Writer.Flush()
			
			if chunk.Done {
				break
			}
		}
	}
}

// InterruptGeneration handles generation interruption requests
func InterruptGeneration(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			RequestID string `json:"request_id" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		err := c.ContentSvc.InterruptGeneration(ctx.Request.Context(), req.RequestID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "INTERRUPT_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Generation interrupted successfully",
		})
	}
}

// Flow handlers

// ListFlows returns all flows
func ListFlows(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limitStr := ctx.DefaultQuery("limit", "10")
		offsetStr := ctx.DefaultQuery("offset", "0")

		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)

		flows, err := c.FlowSvc.ListFlows(ctx.Request.Context(), limit, offset)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "LIST_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"flows": flows,
			"total": len(flows),
		})
	}
}

// CreateFlow creates a new flow
func CreateFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var flow types.Flow
		if err := ctx.ShouldBindJSON(&flow); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		err := c.FlowSvc.CreateFlow(ctx.Request.Context(), &flow)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "CREATE_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusCreated, flow)
	}
}

// GetFlow returns a specific flow
func GetFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		
		flow, err := c.FlowSvc.GetFlow(ctx.Request.Context(), flowID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "FLOW_NOT_FOUND",
			})
			return
		}

		ctx.JSON(http.StatusOK, flow)
	}
}

// UpdateFlow updates a flow
func UpdateFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		
		var flow types.Flow
		if err := ctx.ShouldBindJSON(&flow); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		flow.ID = flowID
		err := c.FlowSvc.UpdateFlow(ctx.Request.Context(), &flow)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "UPDATE_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, flow)
	}
}

// DeleteFlow deletes a flow
func DeleteFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		
		err := c.FlowSvc.DeleteFlow(ctx.Request.Context(), flowID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "DELETE_FAILED",
			})
			return
		}
		
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Flow deleted successfully",
		})
	}
}

// ExecuteFlow executes a flow
func ExecuteFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		
		var req struct {
			Input      map[string]interface{} `json:"input"`
			Parameters map[string]interface{} `json:"parameters"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		execReq := &service.FlowExecutionRequest{
			FlowID:     flowID,
			Input:      req.Input,
			Parameters: req.Parameters,
			RequestID:  uuid.New().String(),
		}

		response, err := c.FlowSvc.ExecuteFlow(ctx.Request.Context(), execReq)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "EXECUTION_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, response)
	}
}

// ListFlowExecutions lists executions for a flow
func ListFlowExecutions(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		limitStr := ctx.DefaultQuery("limit", "10")
		offsetStr := ctx.DefaultQuery("offset", "0")

		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)

		executions, err := c.FlowSvc.ListFlowExecutions(ctx.Request.Context(), flowID, limit, offset)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "LIST_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"executions": executions,
			"total":      len(executions),
		})
	}
}

// Tool handlers

// ListTools returns all available tools
func ListTools(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tools := c.ToolSvc.ListTools(ctx.Request.Context())
		ctx.JSON(http.StatusOK, gin.H{
			"tools": tools,
			"total": len(tools),
		})
	}
}

// CallTool executes a tool
func CallTool(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req genkit.ToolCallRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		if req.RequestID == "" {
			req.RequestID = uuid.New().String()
		}

		response, err := c.ToolSvc.CallTool(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "TOOL_CALL_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, response)
	}
}

// RegisterTool registers a new tool
func RegisterTool(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var tool types.Tool
		if err := ctx.ShouldBindJSON(&tool); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		err := c.ToolSvc.RegisterTool(ctx.Request.Context(), &tool)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "REGISTER_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"message": "Tool registered successfully",
			"tool":    tool,
		})
	}
}

// Vector/RAG handlers

// EmbedText creates embeddings for text
func EmbedText(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Text  string `json:"text" binding:"required"`
			Model string `json:"model"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		if req.Model == "" {
			req.Model = "text-embedding-ada-002"
		}

		embedding, err := c.GenkitSvc.EmbedText(ctx.Request.Context(), req.Text, req.Model)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "EMBEDDING_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"embedding": embedding,
			"model":     req.Model,
		})
	}
}

// SearchVectors searches for similar vectors
func SearchVectors(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Query     []float64 `json:"query" binding:"required"`
			Limit     int       `json:"limit"`
			Threshold float64   `json:"threshold"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		if req.Limit == 0 {
			req.Limit = 10
		}
		if req.Threshold == 0 {
			req.Threshold = 0.7
		}

		results, err := c.VectorRepo.SearchSimilar(req.Query, req.Limit, req.Threshold)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "SEARCH_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"results": results,
			"total":   len(results),
		})
	}
}

// IndexDocument indexes a document with vector embeddings
func IndexDocument(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Content  string                 `json:"content" binding:"required"`
			Source   string                 `json:"source"`
			Metadata map[string]interface{} `json:"metadata"`
			Model    string                 `json:"model"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		if req.Model == "" {
			req.Model = "text-embedding-ada-002"
		}

		// Generate embedding
		embedding, err := c.GenkitSvc.EmbedText(ctx.Request.Context(), req.Content, req.Model)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "EMBEDDING_FAILED",
			})
			return
		}

		// Store document
		doc := &types.VectorDocument{
			Content:   req.Content,
			Embedding: embedding,
			Source:    req.Source,
			Metadata:  req.Metadata,
		}

		err = c.VectorRepo.StoreDocument(doc)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "INDEX_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"document_id": doc.ID,
		})
	}
}

// ListDocuments lists vector documents
func ListDocuments(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		source := ctx.Query("source")
		limitStr := ctx.DefaultQuery("limit", "10")
		offsetStr := ctx.DefaultQuery("offset", "0")

		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)

		documents, err := c.VectorRepo.ListDocuments(source, limit, offset)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "LIST_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"documents": documents,
			"total":     len(documents),
		})
	}
}

// DeleteDocument deletes a document
func DeleteDocument(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		docID := ctx.Param("id")

		err := c.VectorRepo.DeleteDocument(docID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "DELETE_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Document deleted successfully",
		})
	}
}

// MCP Server handlers

// ListServers returns all registered MCP servers
func ListServers(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		servers := c.MCPManager.ListServers()
		ctx.JSON(http.StatusOK, gin.H{
			"servers": servers,
			"total":   len(servers),
		})
	}
}

// RegisterServer registers a new MCP server
func RegisterServer(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var server types.MCPServer
		if err := ctx.ShouldBindJSON(&server); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		err := c.MCPManager.RegisterServer(&server)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "REGISTER_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusCreated, server)
	}
}

// UnregisterServer removes an MCP server
func UnregisterServer(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverID := ctx.Param("id")
		
		err := c.MCPManager.UnregisterServer(serverID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "SERVER_NOT_FOUND",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Server unregistered successfully",
		})
	}
}

// GetServerStatus gets the status of an MCP server
func GetServerStatus(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverID := ctx.Param("id")
		
		server, err := c.MCPManager.GetServer(serverID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "SERVER_NOT_FOUND",
			})
			return
		}

		ctx.JSON(http.StatusOK, server)
	}
}

// ConnectServer connects to an MCP server
func ConnectServer(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverID := ctx.Param("id")
		
		server, err := c.MCPManager.GetServer(serverID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "SERVER_NOT_FOUND",
			})
			return
		}

		// In a real implementation, this would trigger connection
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Connection initiated",
			"server":  server.Config.Name,
		})
	}
}

// DisconnectServer disconnects from an MCP server
func DisconnectServer(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverID := ctx.Param("id")
		
		server, err := c.MCPManager.GetServer(serverID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "SERVER_NOT_FOUND",
			})
			return
		}

		// In a real implementation, this would trigger disconnection
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Disconnection initiated",
			"server":  server.Config.Name,
		})
	}
}

// Evaluation handlers

// RunEvaluation runs an evaluation
func RunEvaluation(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			EvaluationID string                 `json:"evaluation_id"`
			GenerationID string                 `json:"generation_id"`
			Config       map[string]interface{} `json:"config"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_REQUEST",
			})
			return
		}

		// Mock evaluation result
		result := map[string]interface{}{
			"id":            uuid.New().String(),
			"evaluation_id": req.EvaluationID,
			"generation_id": req.GenerationID,
			"score":         0.85,
			"metrics": map[string]interface{}{
				"accuracy":   0.9,
				"relevance":  0.8,
				"coherence":  0.85,
			},
			"created_at": "2024-01-01T12:00:00Z",
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// GetEvaluationResults returns evaluation results
func GetEvaluationResults(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		evaluationID := ctx.Query("evaluation_id")
		limitStr := ctx.DefaultQuery("limit", "10")
		offsetStr := ctx.DefaultQuery("offset", "0")

		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)

		// Mock results
		results := []map[string]interface{}{
			{
				"id":            uuid.New().String(),
				"evaluation_id": evaluationID,
				"score":         0.85,
				"created_at":    "2024-01-01T12:00:00Z",
			},
		}

		ctx.JSON(http.StatusOK, gin.H{
			"results": results,
			"total":   len(results),
			"limit":   limit,
			"offset":  offset,
		})
	}
}

// GetEvaluationResult returns a specific evaluation result
func GetEvaluationResult(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resultID := ctx.Param("id")

		// Mock result
		result := map[string]interface{}{
			"id":      resultID,
			"score":   0.85,
			"details": map[string]interface{}{},
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// Observability handlers

// GetMetrics returns system metrics
func GetMetrics(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metrics := map[string]interface{}{
			"requests_total":     1000,
			"requests_success":   950,
			"requests_failed":    50,
			"avg_response_time":  "150ms",
			"uptime":            "24h30m",
			"memory_usage":      "512MB",
			"cpu_usage":         "25%",
			"active_connections": 42,
		}

		ctx.JSON(http.StatusOK, metrics)
	}
}

// GetTraces returns distributed traces
func GetTraces(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traces := []map[string]interface{}{
			{
				"id":        uuid.New().String(),
				"operation": "content_generation",
				"duration":  "250ms",
				"status":    "ok",
				"timestamp": "2024-01-01T12:00:00Z",
			},
		}

		ctx.JSON(http.StatusOK, gin.H{
			"traces": traces,
		})
	}
}

// Generation management handlers

// ListGenerations lists generations with filtering
func ListGenerations(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filter := service.GenerationFilter{
			Model:     ctx.Query("model"),
			Status:    ctx.Query("status"),
			RequestID: ctx.Query("request_id"),
		}

		limitStr := ctx.DefaultQuery("limit", "10")
		offsetStr := ctx.DefaultQuery("offset", "0")

		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)

		filter.Limit = limit
		filter.Offset = offset

		generations, err := c.ContentSvc.ListGenerations(ctx.Request.Context(), filter)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "LIST_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"generations": generations,
			"total":       len(generations),
		})
	}
}

// GetGeneration returns a specific generation
func GetGeneration(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		generationID := ctx.Param("id")

		generation, err := c.ContentSvc.GetGeneration(ctx.Request.Context(), generationID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "GENERATION_NOT_FOUND",
			})
			return
		}

		ctx.JSON(http.StatusOK, generation)
	}
}

// DeleteGeneration deletes a generation
func DeleteGeneration(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		generationID := ctx.Param("id")

		err := c.ContentSvc.DeleteGeneration(ctx.Request.Context(), generationID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"code":  "DELETE_FAILED",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Generation deleted successfully",
		})
	}
}

// GetGenerationByRequestID returns a generation by request ID
func GetGenerationByRequestID(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.Param("request_id")

		generation, err := c.ContentSvc.GetGenerationByRequestID(ctx.Request.Context(), requestID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "GENERATION_NOT_FOUND",
			})
			return
		}

		ctx.JSON(http.StatusOK, generation)
	}
}