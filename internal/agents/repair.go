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

type RepairAgent struct {
	llm    *services.LLMService
	logger *zap.Logger
}

func NewRepairAgent(llm *services.LLMService, logger *zap.Logger) *RepairAgent {
	return &RepairAgent{
		llm:    llm,
		logger: logger,
	}
}

func (a *RepairAgent) GenerateFix(ctx context.Context, event *models.PipelineEvent, analysis *models.AnalysisResult) (*models.RepairResult, error) {
	a.logger.Info("Generating Fix Patch", zap.String("failure_type", analysis.FailureType))

	promptTmpl, err := services.ReadPromptTemplate("prompts/repair_prompt.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt: %v", err)
	}

	analysisJSON, _ := json.Marshal(analysis)

	prompt := strings.Replace(promptTmpl, "{analysis}", string(analysisJSON), 1)
	prompt = strings.Replace(prompt, "{repo_context}", event.Diff, 1)

	resp, err := a.llm.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM failed to generate fix: %v", err)
	}

	respClean := services.CleanJSON(resp)

	var result models.RepairResult
	if err := json.Unmarshal([]byte(respClean), &result); err != nil {
		a.logger.Error("Failed to parse LLM JSON", zap.String("response", respClean))
		return nil, fmt.Errorf("invalid json from LLM: %v", err)
	}

	return &result, nil
}
