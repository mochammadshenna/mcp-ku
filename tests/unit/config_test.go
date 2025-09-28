package unit

import (
	"testing"

	"mcp-octo-enigma/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	cfg := config.Load()

	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Server.Port)
	assert.NotEmpty(t, cfg.Database.URL)
}

func TestConfigDefaults(t *testing.T) {
	cfg := config.Load()

	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "info", cfg.Logger.Level.String())
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
}
