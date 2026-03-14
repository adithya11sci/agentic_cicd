package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type LLMService struct {
	client *openai.Client
}

func NewLLMService(apiKey string) *LLMService {
	return &LLMService{
		client: openai.NewClient(apiKey),
	}
}

func (s *LLMService) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       openai.GPT4,
			Temperature: 0.2, // low temperature for more deterministic/JSON outputs
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful DevOps assistant. Always output valid JSON only, without markdown wrappers like ```json.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("LLM completion error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// CleanJSON string from LLM responses by removing markdown wrappers
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

func ReadPromptTemplate(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
