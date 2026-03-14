package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/agentic-cicd/internal/agents"
	"github.com/user/agentic-cicd/internal/config"
	"github.com/user/agentic-cicd/internal/orchestrator"
	"github.com/user/agentic-cicd/internal/services"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Services
	llmSvc := services.NewLLMService(cfg.LLMAPIKey)
	githubSvc := services.NewGitHubService(cfg.GitHubToken)

	// Initialize Agents
	monitorAgent := agents.NewMonitorAgent(logger)
	rootCauseAgent := agents.NewRootCauseAgent(llmSvc, logger)
	repairAgent := agents.NewRepairAgent(llmSvc, logger)
	govAgent := agents.NewGovernanceAgent(llmSvc, logger)
	prAgent := agents.NewPRAgent(githubSvc, logger)

	// Initialize Orchestrator
	orch := orchestrator.NewOrchestrator(
		monitorAgent,
		rootCauseAgent,
		repairAgent,
		govAgent,
		prAgent,
		githubSvc,
		logger,
	)

	// Setup Gin router
	r := gin.Default()

	// Webhook endpoint
	r.POST("/webhook/github", func(c *gin.Context) {
		event, err := monitorAgent.HandleWebhook(c)
		if err != nil {
			logger.Error("Failed to parse webhook", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		if event != nil {
			// Run orchestration asynchronously
			go orch.ProcessEvent(event)
		}

		c.JSON(http.StatusOK, gin.H{"status": "received"})
	})

	// Health check
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

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
