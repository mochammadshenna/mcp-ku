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
	"mcp-octo-enigma/internal/cache"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Container holds all application dependencies
type Container struct {
	Config     *config.Config
	Logger     *logrus.Logger
	DB         *sql.DB
	Cache      cache.Cache
	GenkitSvc  *genkit.Service
	MCPManager *mcp.Manager
	
	// Repositories
	VectorRepo repository.VectorRepository
	FlowRepo   repository.FlowRepository
	GenerationRepo repository.GenerationRepository
	
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

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MinConns)

	// Run migrations
	if err := database.RunMigrations(cfg.Database.URL); err != nil {
		log.Warnf("Failed to run migrations: %v", err)
	}

	// Initialize cache
	cacheClient, err := cache.NewRedisCache(cfg.Redis.URL, cfg.Redis.Password, cfg.Redis.DB, log)
	if err != nil {
		log.Warnf("Failed to initialize Redis cache: %v", err)
		cacheClient = cache.NewMemoryCache()
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
	flowRepo := repository.NewFlowRepository(db, log)
	generationRepo := repository.NewGenerationRepository(db, log)

	// Initialize services
	contentSvc := service.NewContentService(genkitSvc, mcpManager, generationRepo, log)
	flowSvc := service.NewFlowService(genkitSvc, flowRepo, log)
	toolSvc := service.NewToolService(genkitSvc, mcpManager, log)
	evalSvc := service.NewEvaluationService(genkitSvc, vectorRepo, log)
	observabilitySvc := service.NewObservabilityService(cfg, log)

	return &Container{
		Config:           cfg,
		Logger:           log,
		DB:               db,
		Cache:            cacheClient,
		GenkitSvc:        genkitSvc,
		MCPManager:       mcpManager,
		VectorRepo:       vectorRepo,
		FlowRepo:         flowRepo,
		GenerationRepo:   generationRepo,
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
	if c.Cache != nil {
		c.Cache.Close()
	}
	if c.GenkitSvc != nil {
		c.GenkitSvc.Close()
	}
	if c.MCPManager != nil {
		c.MCPManager.Close()
	}
	return nil
}