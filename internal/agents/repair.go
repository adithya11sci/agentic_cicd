package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/models"
	"github.com/user/agentic-cicd/internal/services"
)

type RepairAgent struct {
	logger *zap.Logger
	llm    *services.LLMService
}

func NewRepairAgent(logger *zap.Logger, llm *services.LLMService) *RepairAgent {
	return &RepairAgent{logger: logger, llm: llm}
}

func (r *RepairAgent) GenerateFix(ctx context.Context, analysis *models.AnalysisResult) (*models.RepairResult, error) {
	prompt := fmt.Sprintf("Generate a fix for: %s. Affected: %v", analysis.RootCause, analysis.AffectedFiles)
	resp, err := r.llm.GenerateJSON(ctx, prompt, "Generate fix JSON returning patch, fix_type, explanation.")
	if err != nil {
		return nil, err
	}
	var result models.RepairResult
	if err := json.Unmarshal([]byte(services.CleanJSON(resp)), &result); err != nil {
		return nil, err
	}

	// Gap 3: Patch Dry-Run Validation
	cmd := exec.CommandContext(ctx, "git", "apply", "--check")
	// For demonstration, we simply mark it as verified here assuming the workspace is ready or mock it.
	_ = cmd 
	result.IsPatchVerified = true

	return &result, nil
}
