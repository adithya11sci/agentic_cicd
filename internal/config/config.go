package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GitHubToken string
	LLMAPIKey   string
	Port        string
	DatabaseURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Ignore error, as it might run in Docker where env vars are already set

	config := &Config{
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		LLMAPIKey:   os.Getenv("LLM_API_KEY"),
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	if config.GitHubToken == "" || config.LLMAPIKey == "" {
		log.Println("Warning: GITHUB_TOKEN or LLM_API_KEY is not set.")
	}

	return config
}
