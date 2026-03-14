package agents

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"go.uber.org/zap"

	"github.com/user/agentic-cicd/internal/config"
	"github.com/user/agentic-cicd/internal/models"
)

var (
	ErrInvalidSignature    = errors.New("invalid webhook signature")
	ErrDuplicateDelivery   = errors.New("duplicate delivery ID")
	ErrUnsupportedPayload  = errors.New("unsupported payload type")
)

type MonitorAgent struct {
	logger        *zap.Logger
	secret        []byte
	deliveryCache sync.Map
}

func NewMonitorAgent(logger *zap.Logger, cfg *config.Config) *MonitorAgent {
	return &MonitorAgent{
		logger: logger,
		secret: []byte(cfg.WebhookSecret),
	}
}

func (m *MonitorAgent) VerifySignature(payload []byte, signature string) error {
	if len(m.secret) == 0 {
		return nil
	}
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte("sha256="+expectedMAC), []byte(signature)) {
		return ErrInvalidSignature
	}
	return nil
}

func (m *MonitorAgent) HandleWebhook(c *gin.Context) (*models.PipelineEvent, error) {
    // Gap 4: Idempotency keys
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	if deliveryID != "" {
		if _, exists := m.deliveryCache.LoadOrStore(deliveryID, true); exists {
			m.logger.Info("Duplicate webhook delivery skipped", zap.String("delivery_id", deliveryID))
			return nil, ErrDuplicateDelivery
		}
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

    // Gap 7: Webhook Signature Verification
	signature := c.GetHeader("X-Hub-Signature-256")
	if err := m.VerifySignature(payload, signature); err != nil {
		m.logger.Warn("Signature verification failed", zap.Error(err))
		return nil, err
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		return nil, err
	}

	switch e := event.(type) {
	case *github.WorkflowRunEvent:
		if e.GetAction() != "completed" || e.GetWorkflowRun().GetConclusion() != "failure" {
			return nil, nil // Not an error, just ignore
		}
		return &models.PipelineEvent{
			RepositoryName:  e.GetRepo().GetName(),
			RepositoryOwner: e.GetRepo().GetOwner().GetLogin(),
			CommitSHA:       e.GetWorkflowRun().GetHeadSHA(),
			Branch:          e.GetWorkflowRun().GetHeadBranch(),
			PipelineID:      e.GetWorkflowRun().GetID(),
			Status:          "failed",
			Logs:            "simulated logs: syntax error", // Normally fetched via GitHub API using action run ID
		}, nil
	case *github.PingEvent:
		m.logger.Info("Received ping event")
		c.JSON(http.StatusOK, gin.H{"message": "Pong!"})
		return nil, nil
	default:
		return nil, ErrUnsupportedPayload
	}
}
