package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mcp-octo-enigma/internal/config"
	"mcp-octo-enigma/internal/types"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Manager manages multiple MCP servers
type Manager struct {
	config  *config.Config
	logger  *logrus.Logger
	servers map[string]*MCPServerClient
	mu      sync.RWMutex
}

// MCPServerClient represents a client connection to an MCP server
type MCPServerClient struct {
	ID           string                 `json:"id"`
	Config       *types.MCPServer       `json:"config"`
	Client       *Client                `json:"-"`
	Status       string                 `json:"status"` // "active", "inactive", "error"
	LastPing     time.Time              `json:"last_ping"`
	ErrorCount   int                    `json:"error_count"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewManager creates a new MCP manager
func NewManager(cfg *config.Config, logger *logrus.Logger) *Manager {
	manager := &Manager{
		config:  cfg,
		logger:  logger,
		servers: make(map[string]*MCPServerClient),
	}

	// Start health check routine
	go manager.healthCheckRoutine()

	// Register default MCP servers if configured
	manager.registerDefaultServers()

	return manager
}

// registerDefaultServers registers default MCP servers
func (m *Manager) registerDefaultServers() {
	// Example default servers - in production, these would be loaded from config
	defaultServers := []*types.MCPServer{
		{
			ID:          "local-tools",
			Name:        "Local Tools Server",
			URL:         "http://localhost:8081",
			Description: "Local MCP server for tool management",
			Capabilities: []string{"tools", "prompts"},
			Status:      "inactive",
		},
		{
			ID:          "vector-search",
			Name:        "Vector Search Server",
			URL:         "http://localhost:8082",
			Description: "MCP server for vector search operations",
			Capabilities: []string{"vectors", "embeddings"},
			Status:      "inactive",
		},
	}

	for _, server := range defaultServers {
		if err := m.RegisterServer(server); err != nil {
			m.logger.Warnf("Failed to register default server %s: %v", server.Name, err)
		}
	}
}

// RegisterServer registers a new MCP server
func (m *Manager) RegisterServer(serverConfig *types.MCPServer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if serverConfig.ID == "" {
		serverConfig.ID = uuid.New().String()
	}

	// Create client for the server
	client := NewClient(serverConfig.URL, m.logger)

	serverClient := &MCPServerClient{
		ID:           serverConfig.ID,
		Config:       serverConfig,
		Client:       client,
		Status:       "inactive",
		LastPing:     time.Time{},
		ErrorCount:   0,
		Capabilities: serverConfig.Capabilities,
		Metadata:     make(map[string]interface{}),
	}

	m.servers[serverConfig.ID] = serverClient
	m.logger.Infof("Registered MCP server: %s (%s)", serverConfig.Name, serverConfig.ID)

	// Try to connect immediately
	go m.connectServer(serverClient)

	return nil
}

// UnregisterServer removes an MCP server
func (m *Manager) UnregisterServer(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	serverClient, exists := m.servers[serverID]
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}

	// Disconnect from server
	if serverClient.Client != nil {
		serverClient.Client.Close()
	}

	delete(m.servers, serverID)
	m.logger.Infof("Unregistered MCP server: %s", serverID)

	return nil
}

// GetServer retrieves an MCP server by ID
func (m *Manager) GetServer(serverID string) (*MCPServerClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serverClient, exists := m.servers[serverID]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	return serverClient, nil
}

// ListServers returns all registered MCP servers
func (m *Manager) ListServers() []*MCPServerClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]*MCPServerClient, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}

	return servers
}

// GetActiveServers returns only active MCP servers
func (m *Manager) GetActiveServers() []*MCPServerClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var activeServers []*MCPServerClient
	for _, server := range m.servers {
		if server.Status == "active" {
			activeServers = append(activeServers, server)
		}
	}

	return activeServers
}

// BroadcastRequest sends a request to all active servers
func (m *Manager) BroadcastRequest(ctx context.Context, request interface{}) map[string]interface{} {
	activeServers := m.GetActiveServers()
	results := make(map[string]interface{})

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, server := range activeServers {
		wg.Add(1)
		go func(srv *MCPServerClient) {
			defer wg.Done()

			// Send request to server (implementation depends on request type)
			result, err := srv.Client.SendRequest(ctx, request)
			
			mu.Lock()
			if err != nil {
				results[srv.ID] = map[string]interface{}{
					"error": err.Error(),
					"server": srv.Config.Name,
				}
			} else {
				results[srv.ID] = map[string]interface{}{
					"result": result,
					"server": srv.Config.Name,
				}
			}
			mu.Unlock()
		}(server)
	}

	wg.Wait()
	return results
}

// RouteRequest routes a request to a specific server or the best available server
func (m *Manager) RouteRequest(ctx context.Context, serverID string, request interface{}) (interface{}, error) {
	var targetServer *MCPServerClient

	if serverID != "" {
		// Route to specific server
		server, err := m.GetServer(serverID)
		if err != nil {
			return nil, err
		}
		targetServer = server
	} else {
		// Route to best available server (load balancing logic)
		targetServer = m.selectBestServer()
		if targetServer == nil {
			return nil, fmt.Errorf("no active servers available")
		}
	}

	if targetServer.Status != "active" {
		return nil, fmt.Errorf("server not active: %s", targetServer.Config.Name)
	}

	return targetServer.Client.SendRequest(ctx, request)
}

// selectBestServer selects the best server for a request (simple round-robin for now)
func (m *Manager) selectBestServer() *MCPServerClient {
	activeServers := m.GetActiveServers()
	if len(activeServers) == 0 {
		return nil
	}

	// Simple selection - in production, use more sophisticated load balancing
	return activeServers[0]
}

// connectServer attempts to connect to an MCP server
func (m *Manager) connectServer(serverClient *MCPServerClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := serverClient.Client.Connect(ctx)
	if err != nil {
		m.mu.Lock()
		serverClient.Status = "error"
		serverClient.ErrorCount++
		m.mu.Unlock()
		
		m.logger.Errorf("Failed to connect to MCP server %s: %v", serverClient.Config.Name, err)
		return
	}

	m.mu.Lock()
	serverClient.Status = "active"
	serverClient.LastPing = time.Now()
	serverClient.ErrorCount = 0
	m.mu.Unlock()

	m.logger.Infof("Connected to MCP server: %s", serverClient.Config.Name)
}

// healthCheckRoutine periodically checks server health
func (m *Manager) healthCheckRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.performHealthChecks()
	}
}

// performHealthChecks checks health of all registered servers
func (m *Manager) performHealthChecks() {
	m.mu.RLock()
	servers := make([]*MCPServerClient, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}
	m.mu.RUnlock()

	for _, server := range servers {
		go m.checkServerHealth(server)
	}
}

// checkServerHealth checks health of a single server
func (m *Manager) checkServerHealth(serverClient *MCPServerClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := serverClient.Client.Ping(ctx)
	
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		serverClient.ErrorCount++
		if serverClient.ErrorCount >= 3 {
			serverClient.Status = "error"
		}
		m.logger.Warnf("Health check failed for server %s: %v", serverClient.Config.Name, err)
	} else {
		serverClient.Status = "active"
		serverClient.LastPing = time.Now()
		serverClient.ErrorCount = 0
	}
}

// GetServerStats returns statistics about all servers
func (m *Manager) GetServerStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_servers":  len(m.servers),
		"active_servers": 0,
		"error_servers":  0,
		"servers": make([]map[string]interface{}, 0, len(m.servers)),
	}

	for _, server := range m.servers {
		if server.Status == "active" {
			stats["active_servers"] = stats["active_servers"].(int) + 1
		} else if server.Status == "error" {
			stats["error_servers"] = stats["error_servers"].(int) + 1
		}

		serverStats := map[string]interface{}{
			"id":           server.ID,
			"name":         server.Config.Name,
			"status":       server.Status,
			"last_ping":    server.LastPing,
			"error_count":  server.ErrorCount,
			"capabilities": server.Capabilities,
		}

		stats["servers"] = append(stats["servers"].([]map[string]interface{}), serverStats)
	}

	return stats
}

// Close closes all server connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, server := range m.servers {
		if server.Client != nil {
			server.Client.Close()
		}
	}

	m.logger.Info("MCP manager closed")
	return nil
}