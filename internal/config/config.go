package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GitHubToken         string
	LLMAPIKey           string
	Port                string
	DatabaseURL         string
	WebhookSecret       string
	ConfidenceThreshold float64
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	config := &Config{
		GitHubToken:   os.Getenv("GITHUB_TOKEN"),
		LLMAPIKey:     os.Getenv("LLM_API_KEY"),
		Port:          os.Getenv("PORT"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	confStr := os.Getenv("RCA_CONFIDENCE_THRESHOLD")
	if confStr != "" {
		if val, err := strconv.ParseFloat(confStr, 64); err == nil {
			config.ConfidenceThreshold = val
		} else {
			config.ConfidenceThreshold = 0.75
		}
	} else {
		config.ConfidenceThreshold = 0.75
	}

	if config.GitHubToken == "" || config.LLMAPIKey == "" {
		log.Println("Warning: GITHUB_TOKEN or LLM_API_KEY is not set.")
	}

	return config
}
