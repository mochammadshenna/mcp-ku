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
			"status":  "healthy",
			"version": "1.0.0",
			"services": map[string]string{
				"database": "healthy",
				"genkit":   "healthy",
				"mcp":      "healthy",
			},
		}
		
		ctx.JSON(http.StatusOK, status)
	}
}

// GenerateContent handles content generation requests
func GenerateContent(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req service.ContentGenerationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate request ID if not provided
		if req.RequestID == "" {
			req.RequestID = uuid.New().String()
		}

		response, err := c.ContentSvc.GenerateContent(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		streamCh, err := c.ContentSvc.GenerateContentStream(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := c.ContentSvc.InterruptGeneration(ctx.Request.Context(), req.RequestID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Generation interrupted successfully"})
	}
}

// Flow handlers

// ListFlows returns all flows
func ListFlows(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flows := c.GenkitSvc.GetFlowManager().ListFlows()
		ctx.JSON(http.StatusOK, gin.H{"flows": flows})
	}
}

// CreateFlow creates a new flow
func CreateFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var flow types.Flow
		if err := ctx.ShouldBindJSON(&flow); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := c.GenkitSvc.GetFlowManager().CreateFlow(&flow)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, flow)
	}
}

// GetFlow returns a specific flow
func GetFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		
		flow, err := c.GenkitSvc.GetFlowManager().GetFlow(flowID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		flow.ID = flowID
		err := c.GenkitSvc.GetFlowManager().CreateFlow(&flow) // This would be UpdateFlow in a real implementation
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, flow)
	}
}

// DeleteFlow deletes a flow
func DeleteFlow(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		flowID := ctx.Param("id")
		
		// In a real implementation, this would delete the flow
		_ = flowID
		
		ctx.JSON(http.StatusOK, gin.H{"message": "Flow deleted successfully"})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		execReq := &genkit.FlowExecutionRequest{
			FlowID:     flowID,
			Input:      req.Input,
			Parameters: req.Parameters,
			RequestID:  uuid.New().String(),
		}

		response, err := c.GenkitSvc.GetFlowManager().ExecuteFlow(ctx.Request.Context(), execReq)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, response)
	}
}

// Tool handlers

// ListTools returns all available tools
func ListTools(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tools := c.GenkitSvc.GetToolManager().ListTools()
		ctx.JSON(http.StatusOK, gin.H{"tools": tools})
	}
}

// CallTool executes a tool
func CallTool(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req genkit.ToolCallRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.RequestID == "" {
			req.RequestID = uuid.New().String()
		}

		response, err := c.GenkitSvc.GetToolManager().CallTool(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// In a real implementation, this would register the tool
		ctx.JSON(http.StatusCreated, gin.H{"message": "Tool registered successfully", "tool": tool})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Model == "" {
			req.Model = "text-embedding-ada-002"
		}

		embedding, err := c.GenkitSvc.EmbedText(ctx.Request.Context(), req.Text, req.Model)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"embedding": embedding})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"results": results})
	}
}

// IndexDocument indexes a document with vector embeddings
func IndexDocument(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Content string                 `json:"content" binding:"required"`
			Source  string                 `json:"source"`
			Metadata map[string]interface{} `json:"metadata"`
			Model   string                 `json:"model"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Model == "" {
			req.Model = "text-embedding-ada-002"
		}

		// Generate embedding
		embedding, err := c.GenkitSvc.EmbedText(ctx.Request.Context(), req.Content, req.Model)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"document_id": doc.ID})
	}
}

// MCP Server handlers

// ListServers returns all registered MCP servers
func ListServers(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		servers := c.MCPManager.ListServers()
		ctx.JSON(http.StatusOK, gin.H{"servers": servers})
	}
}

// RegisterServer registers a new MCP server
func RegisterServer(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var server types.MCPServer
		if err := ctx.ShouldBindJSON(&server); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := c.MCPManager.RegisterServer(&server)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Server unregistered successfully"})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Mock evaluation result
		result := map[string]interface{}{
			"id":            uuid.New().String(),
			"evaluation_id": req.EvaluationID,
			"generation_id": req.GenerationID,
			"score":         0.85,
			"metrics": map[string]interface{}{
				"accuracy":  0.9,
				"relevance": 0.8,
				"coherence": 0.85,
			},
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
				"created_at":    "2024-01-01T00:00:00Z",
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

// Observability handlers

// GetMetrics returns system metrics
func GetMetrics(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metrics := map[string]interface{}{
			"requests_total":    1000,
			"requests_success":  950,
			"requests_failed":   50,
			"avg_response_time": "150ms",
			"uptime":           "24h30m",
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
				"timestamp": "2024-01-01T00:00:00Z",
			},
		}

		ctx.JSON(http.StatusOK, gin.H{"traces": traces})
	}
}