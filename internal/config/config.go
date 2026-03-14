package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GitHubToken   string
	LLMAPIKey     string
	Port          string
	DatabaseURL   string
	WebhookSecret string
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

	if config.GitHubToken == "" || config.LLMAPIKey == "" {
		log.Println("Warning: GITHUB_TOKEN or LLM_API_KEY is not set.")
	}

	return config
}
