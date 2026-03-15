import os

files = {
    r'd:\agentic_devops\agentic-cicd\internal\config\config.go': """package config

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
""",
    r'd:\agentic_devops\agentic-cicd\internal\services\db.go': """package services

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(dbURL string) (*sql.DB, error) {
	if dbURL == "" {
		log.Println("Warning: DATABASE_URL is not set. DB operations will be skipped.")
		return nil, nil
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS governance_decisions (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		pipeline_id TEXT NOT NULL,
		risk_level  TEXT NOT NULL,
		patch_hash  TEXT NOT NULL,
		decision    TEXT NOT NULL,
		llm_reason  TEXT,
		decided_at  TIMESTAMPTZ DEFAULT now()
	);
	
	CREATE TABLE IF NOT EXISTS processed_webhooks (
		delivery_id TEXT PRIMARY KEY,
		received_at TIMESTAMPTZ DEFAULT now()
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database schemas initialized.")
	return db, nil
}
""",
    r'd:\agentic_devops\agentic-cicd\internal\agents\monitor.go': """package agents

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/config"
	"github.com/user/agentic-cicd/internal/models"
)

var (
	ErrInvalidSignature    = errors.New("invalid webhook signature")
	ErrDuplicateDelivery   = errors.New("duplicate delivery ID")
	ErrUnsupportedPayload  = errors.New("unsupported payload type")
)

type MonitorAgent struct {
	logger *zap.Logger
	secret []byte
	db     *sql.DB
}

func NewMonitorAgent(logger *zap.Logger, cfg *config.Config, db *sql.DB) *MonitorAgent {
	return &MonitorAgent{
		logger: logger,
		secret: []byte(cfg.WebhookSecret),
		db:     db,
	}
}

func (m *MonitorAgent) VerifySignature(payload []byte, signature string) error {
	if len(m.secret) == 0 {
		return nil
	}
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte("sha256="+expectedMAC), []byte(signature)) {
		return ErrInvalidSignature
	}
	return nil
}

func (m *MonitorAgent) HandleWebhook(c *gin.Context) (*models.PipelineEvent, error) {
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	
	if deliveryID != "" && m.db != nil {
		_, err := m.db.ExecContext(c.Request.Context(), "INSERT INTO processed_webhooks (delivery_id) VALUES ($1)", deliveryID)
		if err != nil {
			// Postgres unique violation code is 23505
			if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
				m.logger.Info("Duplicate webhook delivery skipped via DB limit", zap.String("delivery_id", deliveryID))
				return nil, ErrDuplicateDelivery
			}
			m.logger.Error("Failed to persist delivery_id to DB", zap.Error(err))
		}
	} else if deliveryID != "" && m.db == nil {
		m.logger.Warn("Database not configured, skipping idempotency check for delivery", zap.String("delivery_id", deliveryID))
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if err := m.VerifySignature(payload, signature); err != nil {
		m.logger.Warn("Signature verification failed", zap.Error(err))
		return nil, err
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		return nil, err
	}

	switch e := event.(type) {
	case *github.WorkflowRunEvent:
		if e.GetAction() != "completed" || e.GetWorkflowRun().GetConclusion() != "failure" {
			return nil, nil
		}
		return &models.PipelineEvent{
			RepositoryName:  e.GetRepo().GetName(),
			RepositoryOwner: e.GetRepo().GetOwner().GetLogin(),
			CommitSHA:       e.GetWorkflowRun().GetHeadSHA(),
			Branch:          e.GetWorkflowRun().GetHeadBranch(),
			PipelineID:      e.GetWorkflowRun().GetID(),
			Status:          "failed",
			Logs:            "simulated logs: syntax error",
		}, nil
	case *github.PingEvent:
		m.logger.Info("Received ping event")
		c.JSON(http.StatusOK, gin.H{"message": "Pong!"})
		return nil, nil
	default:
		return nil, ErrUnsupportedPayload
	}
}
""",
    r'd:\agentic_devops\agentic-cicd\internal\agents\governance.go': """package agents

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/models"
	"github.com/user/agentic-cicd/internal/services"
)

type GovernanceAgent struct {
	logger *zap.Logger
	llm    *services.LLMService
	db     *sql.DB
}

func NewGovernanceAgent(logger *zap.Logger, llm *services.LLMService, db *sql.DB) *GovernanceAgent {
	return &GovernanceAgent{logger: logger, llm: llm, db: db}
}

func (g *GovernanceAgent) Evaluate(ctx context.Context, pCtx *models.PipelineContext) (*models.GovernanceResult, error) {
	prompt := fmt.Sprintf("Evaluate risk for Patch: %s, Explain: %s", pCtx.Repair.Patch, pCtx.Repair.Explanation)
	resp, err := g.llm.GenerateJSON(ctx, prompt, "Evaluate risk. Return JSON with risk_level, requires_human_approval, reason.")
	if err != nil {
		return nil, err
	}

	var result models.GovernanceResult
	if err := json.Unmarshal([]byte(services.CleanJSON(resp)), &result); err != nil {
		return nil, err
	}

	if g.db != nil {
		patchHash := fmt.Sprintf("%x", sha256.Sum256([]byte(pCtx.Repair.Patch)))
		query := `INSERT INTO governance_decisions (pipeline_id, risk_level, patch_hash, decision, llm_reason) VALUES ($1, $2, $3, $4, $5)`
		_, err := g.db.ExecContext(ctx, query, fmt.Sprintf("%d", pCtx.Event.PipelineID), result.RiskLevel, patchHash, "PENDING", result.Reason)
		if err != nil {
			g.logger.Error("Failed to persist audit trail", zap.Error(err))
			return nil, fmt.Errorf("failed to persist governance decision: %w", err)
		} else {
			g.logger.Info("Decision persisted to Audit Trail")
		}
	}
	
	return &result, nil
}
""",
    r'd:\agentic_devops\agentic-cicd\internal\orchestrator\orchestrator.go': """package orchestrator

import (
	"context"

	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/agents"
	"github.com/user/agentic-cicd/internal/config"
	"github.com/user/agentic-cicd/internal/models"
)

type Orchestrator struct {
	cfg        *config.Config
	monitor    *agents.MonitorAgent
	rootCause  *agents.RootCauseAgent
	repair     *agents.RepairAgent
	governance *agents.GovernanceAgent
	prAgent    *agents.PRAgent
	logger     *zap.Logger
}

func NewOrchestrator(
	cfg *config.Config,
	m *agents.MonitorAgent,
	rc *agents.RootCauseAgent,
	re *agents.RepairAgent,
	gov *agents.GovernanceAgent,
	pr *agents.PRAgent,
	logger *zap.Logger,
) *Orchestrator {
	return &Orchestrator{cfg: cfg, monitor: m, rootCause: rc, repair: re, governance: gov, prAgent: pr, logger: logger}
}

func (o *Orchestrator) RunPipeline(ctx context.Context, event *models.PipelineEvent) {
	defer func() {
		if r := recover(); r != nil {
			o.logger.Error("Pipeline panicked, recovered", zap.Any("panic", r))
		}
	}()

	pCtx := &models.PipelineContext{Event: event}

	analysis, err := o.rootCause.Analyze(ctx, event)
	if err != nil {
		o.logger.Error("RootCause failed", zap.Error(err))
		return
	}
	pCtx.Analysis = analysis

	if analysis.Confidence < o.cfg.ConfidenceThreshold {
		o.logger.Warn("Analysis confidence too low, halting pipeline", zap.Float64("confidence", analysis.Confidence), zap.Float64("threshold", o.cfg.ConfidenceThreshold))
		return
	}

	repair, err := o.repair.GenerateFix(ctx, analysis)
	if err != nil {
		o.logger.Error("Repair failed", zap.Error(err))
		return
	}
	pCtx.Repair = repair

	govResult, err := o.governance.Evaluate(ctx, pCtx)
	if err != nil {
		o.logger.Error("Governance skipped/failed", zap.Error(err))
		return
	}
	pCtx.Governance = govResult

	if !govResult.RequiresHumanApproval {
		o.logger.Info("Auto-approving patch creation")
		err := o.prAgent.CreateFixPR(ctx, event, analysis, repair)
		if err != nil {
			o.logger.Error("PR Agent failed", zap.Error(err))
		}
	} else {
		o.logger.Info("Human approval required for patch PR, delegating to dashboard")
	}
}
""",
    r'd:\agentic_devops\agentic-cicd\cmd\server.go': """package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/agents"
	"github.com/user/agentic-cicd/internal/config"
	"github.com/user/agentic-cicd/internal/orchestrator"
	"github.com/user/agentic-cicd/internal/services"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.LoadConfig()

	db, err := services.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	llmSvc := services.NewLLMService(cfg.LLMAPIKey, logger)
	githubSvc := services.NewGitHubService(cfg.GitHubToken)

	monitorAgent := agents.NewMonitorAgent(logger, cfg, db)
	rootCauseAgent := agents.NewRootCauseAgent(logger, llmSvc)
	repairAgent := agents.NewRepairAgent(logger, llmSvc)
	govAgent := agents.NewGovernanceAgent(logger, llmSvc, db)
	prAgent := agents.NewPRAgent(githubSvc, logger)

	orch := orchestrator.NewOrchestrator(
		cfg,
		monitorAgent,
		rootCauseAgent,
		repairAgent,
		govAgent,
		prAgent,
		logger,
	)

	r := gin.Default()

	r.POST("/webhook/github", func(c *gin.Context) {
		event, err := monitorAgent.HandleWebhook(c)
		if err != nil {
			logger.Error("Failed to parse webhook", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
			return
		}

		if event != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				orch.RunPipeline(ctx, event)
			}()
			c.JSON(http.StatusOK, gin.H{"status": "accepted pipeline run", "pipeline_id": event.PipelineID})
			return
		}
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
"""
}

for path, content in files.items():
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content)
