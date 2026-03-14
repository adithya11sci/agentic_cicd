import os

config_content = """package config

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
"""

models_content = """package models

type PipelineEvent struct {
	RepositoryName  string `json:"repository_name"`
	RepositoryOwner string `json:"repository_owner"`
	CommitSHA       string `json:"commit_sha"`
	Branch          string `json:"branch"`
	PipelineID      int64  `json:"pipeline_id"`
	Status          string `json:"status"` // expected: "failed"
	Logs            string `json:"-"`
	Diff            string `json:"-"`
	TestReport      string `json:"-"`
}

type AnalysisResult struct {
	FailureType   string   `json:"failure_type"`
	RootCause     string   `json:"root_cause"`
	AffectedFiles []string `json:"affected_files"`
	Confidence    float64  `json:"confidence"`
}

type RepairResult struct {
	FixType         string `json:"fix_type"`
	Patch           string `json:"patch"`
	Explanation     string `json:"explanation"`
	IsPatchVerified bool   `json:"is_patch_verified"`
}

type GovernanceResult struct {
	RiskLevel             string `json:"risk_level"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
	Reason                string `json:"reason"`
}

type PipelineContext struct {
	Event         *PipelineEvent
	Analysis      *AnalysisResult
	Repair        *RepairResult
	Governance    *GovernanceResult
	ExecutionLogs []string
}
"""

with open(r'd:\agentic_devops\agentic-cicd\internal\config\config.go', 'w', encoding='utf-8') as f:
    f.write(config_content)

with open(r'd:\agentic_devops\agentic-cicd\internal\models\event.go', 'w', encoding='utf-8') as f:
    f.write(models_content)
