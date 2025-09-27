package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"mcp-octo-enigma/internal/container"
	"mcp-octo-enigma/internal/handlers"
	"mcp-octo-enigma/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	container  *container.Container
	httpServer *http.Server
	router     *gin.Engine
}

// NewServer creates a new server instance
func NewServer(c *container.Container) *Server {
	// Set Gin mode
	if c.Logger.Level == c.Logger.Level.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := &Server{
		container: c,
		router:    router,
		httpServer: &http.Server{
			Addr:           ":" + c.Config.Server.Port,
			Handler:        router,
			ReadTimeout:    c.Config.Server.ReadTimeout,
			WriteTimeout:   c.Config.Server.WriteTimeout,
			MaxHeaderBytes: c.Config.Server.MaxHeaderBytes,
		},
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// CORS middleware
	s.router.Use(middleware.CORS())

	// Request ID middleware
	s.router.Use(middleware.RequestID())

	// Logging middleware
	s.router.Use(middleware.Logger(s.container.Logger))

	// Rate limiting middleware
	s.router.Use(middleware.RateLimit())

	// Authentication middleware for protected routes
	s.router.Use(middleware.Auth())
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", handlers.HealthCheck(s.container))

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// MCP routes
		mcp := v1.Group("/mcp")
		{
			mcp.POST("/generate", handlers.GenerateContent(s.container))
			mcp.POST("/flow", handlers.ExecuteFlow(s.container))
			mcp.POST("/tools/call", handlers.CallTool(s.container))
			mcp.GET("/servers", handlers.ListServers(s.container))
			mcp.POST("/servers", handlers.RegisterServer(s.container))
			mcp.DELETE("/servers/:id", handlers.UnregisterServer(s.container))
		}

		// Content generation routes
		content := v1.Group("/content")
		{
			content.POST("/generate", handlers.GenerateContent(s.container))
			content.POST("/generate/stream", handlers.GenerateContentStream(s.container))
			content.POST("/interrupt", handlers.InterruptGeneration(s.container))
		}

		// Flow routes
		flows := v1.Group("/flows")
		{
			flows.GET("/", handlers.ListFlows(s.container))
			flows.POST("/", handlers.CreateFlow(s.container))
			flows.GET("/:id", handlers.GetFlow(s.container))
			flows.PUT("/:id", handlers.UpdateFlow(s.container))
			flows.DELETE("/:id", handlers.DeleteFlow(s.container))
			flows.POST("/:id/execute", handlers.ExecuteFlow(s.container))
		}

		// Tool routes
		tools := v1.Group("/tools")
		{
			tools.GET("/", handlers.ListTools(s.container))
			tools.POST("/call", handlers.CallTool(s.container))
			tools.POST("/register", handlers.RegisterTool(s.container))
		}

		// Vector/RAG routes
		vectors := v1.Group("/vectors")
		{
			vectors.POST("/embed", handlers.EmbedText(s.container))
			vectors.POST("/search", handlers.SearchVectors(s.container))
			vectors.POST("/index", handlers.IndexDocument(s.container))
		}

		// Evaluation routes
		eval := v1.Group("/evaluation")
		{
			eval.POST("/run", handlers.RunEvaluation(s.container))
			eval.GET("/results", handlers.GetEvaluationResults(s.container))
		}

		// Observability routes
		observability := v1.Group("/observability")
		{
			observability.GET("/metrics", handlers.GetMetrics(s.container))
			observability.GET("/traces", handlers.GetTraces(s.container))
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.container.Logger.Infof("Starting server on port %s", s.container.Config.Server.Port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.container.Logger.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// GetRouter returns the Gin router for testing purposes
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}