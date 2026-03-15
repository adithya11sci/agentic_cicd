package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type LLMService struct {
	client *openai.Client
	logger *zap.Logger
}

func NewLLMService(apiKey string, logger *zap.Logger) *LLMService {
	clientConfig := openai.DefaultConfig(apiKey)
	if strings.HasPrefix(apiKey, "gsk_") {
		clientConfig.BaseURL = "https://api.groq.com/openai/v1"
	}
	return &LLMService{
		client: openai.NewClientWithConfig(clientConfig),
		logger: logger,
	}
}

func CleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

// Gap 1: No Retry on LLM Fallback (implemented exponential backoff)
func (s *LLMService) GenerateJSON(ctx context.Context, userPrompt, systemPrompt string) (string, error) {
	maxRetries := 3
	var lastErr error

	model := openai.GPT4o
	apiKey := os.Getenv("LLM_API_KEY")
	if strings.HasPrefix(apiKey, "gsk_") {
		// Automatically use Groq's high speed Llama 3 model
		model = "llama-3.3-70b-versatile"
	}

	for i := 0; i < maxRetries; i++ {
		req := openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: userPrompt},
			},
			Temperature: 0.2,
		}

		resp, err := s.client.CreateChatCompletion(ctx, req)
		if err == nil && len(resp.Choices) > 0 {
			return resp.Choices[0].Message.Content, nil
		}
		lastErr = err

		s.logger.Warn("LLM call failed, retrying", zap.Int("attempt", i+1), zap.Error(err))
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(1<<i) * time.Second): // 1s, 2s, 4s
		}
	}
	return "", fmt.Errorf("LLM generation failed after %d retries: %v", maxRetries, lastErr)
}
