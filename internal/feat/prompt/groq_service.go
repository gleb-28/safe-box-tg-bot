package prompt

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"strings"

	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/helpers"
)

var _ LLMGenerator = (*GroqService)(nil)

type GroqService struct {
	client *GroqClient
	model  string
	logger logger.AppLogger
}

func MustNewGroqService(client *GroqClient, model string, appLogger logger.AppLogger) *GroqService {
	if client == nil {
		stdlog.Fatal("Groq client is nil")
	}
	trimmedModel := strings.TrimSpace(model)
	if trimmedModel == "" {
		stdlog.Fatal("Groq model is empty")
	}
	return &GroqService{
		client: client,
		model:  trimmedModel,
		logger: appLogger,
	}
}

func (s *GroqService) Generate(ctx context.Context, req LLMRequest) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("Groq client is nil")
	}
	if strings.TrimSpace(req.SystemPrompt) == "" {
		return "", errors.New("system prompt is empty")
	}
	if strings.TrimSpace(req.UserPrompt) == "" {
		return "", errors.New("user prompt is empty")
	}

	messages := []GroqChatMessage{
		{Role: "system", Content: req.SystemPrompt},
		{Role: "user", Content: req.UserPrompt},
	}

	request := GroqChatCompletionRequest{
		Model:    s.model,
		Messages: messages,
	}
	if req.MaxTokens > 0 {
		maxTokens := req.MaxTokens
		request.MaxTokens = &maxTokens
	}
	temp := req.Temperature
	request.Temperature = &temp

	resp, err := s.client.Chat(ctx, request)
	if err != nil {
		if s.logger != nil {
			s.logger.Debug(fmt.Sprintf("Groq generate failed model=%s: %v", s.model, err))
		}
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("Groq: empty choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	clean := helpers.CleanLLMText(content)
	if clean == "" {
		return "", errors.New("Groq: empty content")
	}

	if s.logger != nil {
		s.logger.Debug(fmt.Sprintf("Groq generate succeeded model=%s content_len=%d", s.model, len(clean)))
	}

	return clean, nil
}
