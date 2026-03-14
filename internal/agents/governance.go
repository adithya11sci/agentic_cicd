package agents

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"crypto/sha256"

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

	// Gap 2: Governance Audit Trail Persistence
	if g.db != nil {
		patchHash := fmt.Sprintf("%x", sha256.Sum256([]byte(pCtx.Repair.Patch)))
		query := `INSERT INTO governance_decisions (pipeline_id, risk_level, patch_hash, decision, llm_reason) VALUES ($1, $2, $3, $4, $5)`
		_, err := g.db.ExecContext(ctx, query, fmt.Sprintf("%d", pCtx.Event.PipelineID), result.RiskLevel, patchHash, "PENDING", result.Reason)
		if err != nil {
			g.logger.Error("Failed to persist audit trail", zap.Error(err))
		} else {
			g.logger.Info("Decision persisted to Audit Trail")
		}
	}
	
	return &result, nil
}
