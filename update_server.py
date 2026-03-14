import os

content = """package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agentic-cicd/internal/agents"
	"agentic-cicd/internal/config"
	"agentic-cicd/internal/orchestrator"
	"agentic-cicd/internal/services"
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
	githubSvc := services.NewGitHubService(cfg.GitHubToken, logger)

	monitorAgent := agents.NewMonitorAgent(logger, cfg)
	rootCauseAgent := agents.NewRootCauseAgent(logger, llmSvc)
	repairAgent := agents.NewRepairAgent(logger, llmSvc)
	govAgent := agents.NewGovernanceAgent(logger, llmSvc, db)
	prAgent := agents.NewPRAgent(logger, githubSvc)

	orch := orchestrator.NewOrchestrator(
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

with open(r'd:\agentic_devops\agentic-cicd\cmd\server.go', 'w', encoding='utf-8') as f:
    f.write(content)
