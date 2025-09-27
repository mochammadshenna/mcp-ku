package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Database Database
	Server   Server
	Firebase Firebase
	AI       AI
	Logger   Logger
}

type Database struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type Server struct {
	Port           string
	ClientTimeout  time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

type Firebase struct {
	ProjectID       string
	CredentialsPath string
}

type AI struct {
	GoogleAI     GoogleAI
	OpenAI       OpenAI
	Anthropic    Anthropic
	VertexAI     VertexAI
	Ollama       Ollama
}

type GoogleAI struct {
	APIKey string
}

type OpenAI struct {
	APIKey string
}

type Anthropic struct {
	APIKey string
}

type VertexAI struct {
	ProjectID string
}

type Ollama struct {
	Host string
}

type Logger struct {
	Level logrus.Level
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	return &Config{
		Database: Database{
			URL:      getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/mcp_octo_enigma?sslmode=disable"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "mcp_octo_enigma"),
		},
		Server: Server{
			Port:           getEnv("MCP_SERVER_PORT", "8080"),
			ClientTimeout:  getEnvAsDuration("MCP_CLIENT_TIMEOUT", 30*time.Second),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		Firebase: Firebase{
			ProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
			CredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./config/firebase-credentials.json"),
		},
		AI: AI{
			GoogleAI: GoogleAI{
				APIKey: getEnv("GOOGLE_AI_API_KEY", ""),
			},
			OpenAI: OpenAI{
				APIKey: getEnv("OPENAI_API_KEY", ""),
			},
			Anthropic: Anthropic{
				APIKey: getEnv("ANTHROPIC_API_KEY", ""),
			},
			VertexAI: VertexAI{
				ProjectID: getEnv("VERTEX_AI_PROJECT_ID", ""),
			},
			Ollama: Ollama{
				Host: getEnv("OLLAMA_HOST", "http://localhost:11434"),
			},
		},
		Logger: Logger{
			Level: getLogLevel(getEnv("LOG_LEVEL", "info")),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}