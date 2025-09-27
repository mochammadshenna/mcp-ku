package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type AppConfig struct {
	ServerPort       string `mapstructure:"SERVER_PORT"`
	Env              string `mapstructure:"APP_ENV"`
	DatabaseURL      string `mapstructure:"DATABASE_URL"`
	FirebaseProjectID string `mapstructure:"FIREBASE_PROJECT_ID"`
	GenAIProvider    string `mapstructure:"GENAI_PROVIDER"`
}

func Load() *AppConfig {
	_ = godotenv.Load()

	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal env config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	}

	return &cfg
}
