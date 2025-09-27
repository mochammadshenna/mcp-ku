package container

import (
	"database/sql"
	"fmt"

	"mcp-octo-enigma/internal/config"
	"mcp-octo-enigma/internal/database"
	"mcp-octo-enigma/internal/genkit"
	"mcp-octo-enigma/internal/logger"
	"mcp-octo-enigma/internal/mcp"
	"mcp-octo-enigma/internal/repository"
	"mcp-octo-enigma/internal/service"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Container holds all application dependencies
type Container struct {
	Config     *config.Config
	Logger     *logrus.Logger
	DB         *sql.DB
	GenkitSvc  *genkit.Service
	MCPManager *mcp.Manager
	
	// Repositories
	VectorRepo repository.VectorRepository
	
	// Services
	ContentSvc    service.ContentService
	FlowSvc       service.FlowService
	ToolSvc       service.ToolService
	EvalSvc       service.EvaluationService
	ObservabilitySvc service.ObservabilityService
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	// Initialize logger
	log := logger.New(cfg.Logger.Level)

	// Initialize database
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := database.RunMigrations(cfg.Database.URL); err != nil {
		log.Warnf("Failed to run migrations: %v", err)
	}

	// Initialize Genkit service
	genkitSvc, err := genkit.NewService(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Genkit service: %w", err)
	}

	// Initialize MCP manager
	mcpManager := mcp.NewManager(cfg, log)

	// Initialize repositories
	vectorRepo := repository.NewVectorRepository(db, log)

	// Initialize services
	contentSvc := service.NewContentService(genkitSvc, mcpManager, log)
	flowSvc := service.NewFlowService(genkitSvc, log)
	toolSvc := service.NewToolService(genkitSvc, mcpManager, log)
	evalSvc := service.NewEvaluationService(genkitSvc, vectorRepo, log)
	observabilitySvc := service.NewObservabilityService(cfg, log)

	return &Container{
		Config:           cfg,
		Logger:           log,
		DB:               db,
		GenkitSvc:        genkitSvc,
		MCPManager:       mcpManager,
		VectorRepo:       vectorRepo,
		ContentSvc:       contentSvc,
		FlowSvc:          flowSvc,
		ToolSvc:          toolSvc,
		EvalSvc:          evalSvc,
		ObservabilitySvc: observabilitySvc,
	}, nil
}

// Close closes all resources
func (c *Container) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.GenkitSvc != nil {
		c.GenkitSvc.Close()
	}
	if c.MCPManager != nil {
		c.MCPManager.Close()
	}
	return nil
}