package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/user/agentic-cicd/internal/models"
	"github.com/user/agentic-cicd/internal/services"
	"go.uber.org/zap"
)

type RootCauseAgent struct {
	llm    *services.LLMService
	logger *zap.Logger
}

func NewRootCauseAgent(llm *services.LLMService, logger *zap.Logger) *RootCauseAgent {
	return &RootCauseAgent{
		llm:    llm,
		logger: logger,
	}
}

func (a *RootCauseAgent) Analyze(ctx context.Context, event *models.PipelineEvent) (*models.AnalysisResult, error) {
	a.logger.Info("Starting Root Cause Analysis", zap.String("repo", event.RepositoryName))

	promptTmpl, err := services.ReadPromptTemplate("prompts/root_cause_prompt.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt: %v", err)
	}

	prompt := strings.Replace(promptTmpl, "{logs}", event.Logs, 1)
	prompt = strings.Replace(prompt, "{diff}", event.Diff, 1)
	prompt = strings.Replace(prompt, "{tests}", event.TestReport, 1)

	resp, err := a.llm.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM failed to analyze root cause: %v", err)
	}

	respClean := services.CleanJSON(resp)

	var result models.AnalysisResult
	if err := json.Unmarshal([]byte(respClean), &result); err != nil {
		a.logger.Error("Failed to parse LLM JSON", zap.String("response", respClean))
		return nil, fmt.Errorf("invalid json from LLM: %v", err)
	}

	return &result, nil
}
