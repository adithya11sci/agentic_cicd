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

type GovernanceAgent struct {
	llm    *services.LLMService
	logger *zap.Logger
}

func NewGovernanceAgent(llm *services.LLMService, logger *zap.Logger) *GovernanceAgent {
	return &GovernanceAgent{
		llm:    llm,
		logger: logger,
	}
}

func (a *GovernanceAgent) EvaluateRisk(ctx context.Context, fix *models.RepairResult) (*models.GovernanceResult, error) {
	a.logger.Info("Evaluating Governance Risk", zap.String("fix_type", fix.FixType))

	promptTmpl, err := services.ReadPromptTemplate("prompts/governance_prompt.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt: %v", err)
	}

	fixJSON, _ := json.Marshal(fix)
	prompt := strings.Replace(promptTmpl, "{fix}", string(fixJSON), 1)

	resp, err := a.llm.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM failed to evaluate risk: %v", err)
	}

	respClean := services.CleanJSON(resp)

	var result models.GovernanceResult
	if err := json.Unmarshal([]byte(respClean), &result); err != nil {
		a.logger.Error("Failed to parse LLM JSON", zap.String("response", respClean))
		return nil, fmt.Errorf("invalid json from LLM: %v", err)
	}

	return &result, nil
}
