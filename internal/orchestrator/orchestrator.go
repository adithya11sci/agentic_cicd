package orchestrator

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
