package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/models"
	"github.com/user/agentic-cicd/internal/services"
)

type RootCauseAgent struct {
	logger *zap.Logger
	llm    *services.LLMService
}

func NewRootCauseAgent(logger *zap.Logger, llm *services.LLMService) *RootCauseAgent {
	return &RootCauseAgent{logger: logger, llm: llm}
}

func (r *RootCauseAgent) Analyze(ctx context.Context, event *models.PipelineEvent) (*models.AnalysisResult, error) {
	prompt := fmt.Sprintf("Analyze this failure.\nLogs: %s\nDiff: %s", event.Logs, event.Diff)
	resp, err := r.llm.GenerateJSON(ctx, prompt, "Analyze failure. Return JSON with failure_type, root_cause, affected_files, confidence (float 0.0-1.0)")
	if err != nil {
		return nil, err
	}
	var result models.AnalysisResult
	if err := json.Unmarshal([]byte(services.CleanJSON(resp)), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
