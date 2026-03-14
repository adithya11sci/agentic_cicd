package agents

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"

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
	logger *zap.Logger
	secret []byte
	db     *sql.DB
}

func NewMonitorAgent(logger *zap.Logger, cfg *config.Config, db *sql.DB) *MonitorAgent {
	return &MonitorAgent{
		logger: logger,
		secret: []byte(cfg.WebhookSecret),
		db:     db,
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
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	
	if deliveryID != "" && m.db != nil {
		_, err := m.db.ExecContext(c.Request.Context(), "INSERT INTO processed_webhooks (delivery_id) VALUES ($1)", deliveryID)
		if err != nil {
			// Postgres unique violation code is 23505
			if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
				m.logger.Info("Duplicate webhook delivery skipped via DB limit", zap.String("delivery_id", deliveryID))
				return nil, ErrDuplicateDelivery
			}
			m.logger.Error("Failed to persist delivery_id to DB", zap.Error(err))
		}
	} else if deliveryID != "" && m.db == nil {
		m.logger.Warn("Database not configured, skipping idempotency check for delivery", zap.String("delivery_id", deliveryID))
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

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
			return nil, nil
		}
		return &models.PipelineEvent{
			RepositoryName:  e.GetRepo().GetName(),
			RepositoryOwner: e.GetRepo().GetOwner().GetLogin(),
			CommitSHA:       e.GetWorkflowRun().GetHeadSHA(),
			Branch:          e.GetWorkflowRun().GetHeadBranch(),
			PipelineID:      e.GetWorkflowRun().GetID(),
			Status:          "failed",
			Logs:            "simulated logs: syntax error",
		}, nil
	case *github.PingEvent:
		m.logger.Info("Received ping event")
		c.JSON(http.StatusOK, gin.H{"message": "Pong!"})
		return nil, nil
	default:
		return nil, ErrUnsupportedPayload
	}
}
