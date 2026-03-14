package orchestrator

import (
	"context"
	"time"

	"github.com/user/agentic-cicd/internal/agents"
	"github.com/user/agentic-cicd/internal/models"
	"github.com/user/agentic-cicd/internal/services"
	"go.uber.org/zap"
)

type Orchestrator struct {
	monitor    *agents.MonitorAgent
	rootCause  *agents.RootCauseAgent
	repair     *agents.RepairAgent
	governance *agents.GovernanceAgent
	prAgent    *agents.PRAgent
	githubSvc  *services.GitHubService
	logger     *zap.Logger
}

func NewOrchestrator(
	monitor *agents.MonitorAgent,
	rootCause *agents.RootCauseAgent,
	repair *agents.RepairAgent,
	governance *agents.GovernanceAgent,
	prAgent *agents.PRAgent,
	githubSvc *services.GitHubService,
	logger *zap.Logger,
) *Orchestrator {
	return &Orchestrator{
		monitor:    monitor,
		rootCause:  rootCause,
		repair:     repair,
		governance: governance,
		prAgent:    prAgent,
		githubSvc:  githubSvc,
		logger:     logger,
	}
}

// ProcessEvent runs the multi-agent workflow
func (o *Orchestrator) ProcessEvent(event *models.PipelineEvent) {
	defer func() {
		if r := recover(); r != nil {
			o.logger.Error("Panic recovered in ProcessEvent", zap.Any("recover", r))
		}
	}()

	// Create a context with timeout to prevent hanging forever
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	o.logger.Info("Coordinator starting workflow", zap.String("repo", event.RepositoryName))

	// Step 1: Fetch Extra Context (Logs, Diff)
	logs, _ := o.githubSvc.FetchPipelineLogs(ctx, event.RepositoryOwner, event.RepositoryName, event.PipelineID)
	diff, _ := o.githubSvc.FetchCommitDiff(ctx, event.RepositoryOwner, event.RepositoryName, event.CommitSHA)
	event.Logs = logs
	event.Diff = diff
	event.TestReport = "Mocked Test Report: 1 test failing."

	// Step 2: Root Cause Analysis
	analysis, err := o.rootCause.Analyze(ctx, event)
	if err != nil {
		o.logger.Error("Root cause analysis failed", zap.Error(err))
		return
	}

	// Step 3: Auto Repair Gen
	fix, err := o.repair.GenerateFix(ctx, event, analysis)
	if err != nil {
		o.logger.Error("Repair generation failed", zap.Error(err))
		return
	}

	// Step 4: Governance
	govResult, err := o.governance.EvaluateRisk(ctx, fix)
	if err != nil {
		o.logger.Error("Governance evaluation failed", zap.Error(err))
		return
	}

	// Step 5: Check Deployment Risk and take action
	o.logger.Info("Governance decision", zap.String("risk", govResult.RiskLevel), zap.Bool("needs_approval", govResult.RequiresHumanApproval))

	if govResult.RiskLevel == "LOW" || !govResult.RequiresHumanApproval {
		// Create PR
		err := o.prAgent.CreateFixPR(ctx, event, analysis, fix)
		if err != nil {
			o.logger.Error("Failed to create PR", zap.Error(err))
		}
	} else {
		o.logger.Info("Fix requires human approval. Sending notification... (Mocked Slack ping)", zap.String("reason", govResult.Reason))
	}
}
