package agents

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/user/agentic-cicd/internal/models"
	"go.uber.org/zap"
)

type MonitorAgent struct {
	logger *zap.Logger
}

func NewMonitorAgent(logger *zap.Logger) *MonitorAgent {
	return &MonitorAgent{logger: logger}
}

// HandleWebhook parses the incoming GitHub Workflow Run webhook
func (a *MonitorAgent) HandleWebhook(c *gin.Context) (*models.PipelineEvent, error) {
	// Simplify payload for example purposes
	// Usually we bind to GitHub's WorkflowRunEvent struct
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	// Check for GitHub Ping Event (sent when webhook is created)
	if ping, exists := payload["zen"]; exists {
		a.logger.Info("Received GitHub ping event", zap.Any("zen", ping))
		return nil, nil // Return gracefully without errors for ping events
	}

	// Basic check assuming workflow_run event
	action, ok := payload["action"].(string)
	if !ok || action != "completed" {
		a.logger.Info("Ignored webhook event", zap.String("action", action))
		return nil, nil // Not a completed workflow run
	}

	workflowRun, ok := payload["workflow_run"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload: missing workflow_run")
	}

	conclusion, _ := workflowRun["conclusion"].(string)
	if conclusion != "failure" {
		a.logger.Info("Workflow completed successfully, no action needed", zap.String("conclusion", conclusion))
		return nil, nil // We only care about failures
	}

	repoMap, _ := payload["repository"].(map[string]interface{})
	ownerMap, _ := repoMap["owner"].(map[string]interface{})

	event := &models.PipelineEvent{
		RepositoryName:  repoMap["name"].(string),
		RepositoryOwner: ownerMap["login"].(string),
		CommitSHA:       workflowRun["head_sha"].(string),
		Branch:          workflowRun["head_branch"].(string),
		PipelineID:      int64(workflowRun["id"].(float64)),
		Status:          conclusion,
	}

	a.logger.Info("Pipeline failure detected", zap.String("repo", event.RepositoryName), zap.String("sha", event.CommitSHA))

	return event, nil
}
