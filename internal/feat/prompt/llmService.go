package prompt

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"strings"

	"safeboxtgbot/internal/core/logger"

	"github.com/goforj/godump"
	"github.com/revrost/go-openrouter"
)

type LLMRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

type LLMService struct {
	client   *OpenRouterClient
	model    string
	fallback LLMGenerator
	logger   logger.AppLogger
}

var _ LLMGenerator = (*LLMService)(nil)

func MustNewLLMService(client *OpenRouterClient, model string, fallback LLMGenerator, appLogger logger.AppLogger) *LLMService {
	if client == nil {
		stdlog.Fatal("llm client is nil")
	}
	trimmedModel := strings.TrimSpace(model)
	if trimmedModel == "" {
		stdlog.Fatal("llm model is empty")
	}
	return &LLMService{
		client:   client,
		model:    trimmedModel,
		fallback: fallback,
		logger:   appLogger,
	}
}

func (s *LLMService) Generate(ctx context.Context, req LLMRequest) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("llm client is nil")
	}
	if strings.TrimSpace(req.SystemPrompt) == "" {
		return "", errors.New("system prompt is empty")
	}
	if strings.TrimSpace(req.UserPrompt) == "" {
		return "", errors.New("user prompt is empty")
	}

	resp, err := s.client.Chat(ctx, openrouter.ChatCompletionRequest{
		Model: s.model,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.SystemMessage(req.SystemPrompt),
			openrouter.UserMessage(req.UserPrompt),
		},
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeText,
		},
	})
	if err != nil {
		if s.logger != nil {
			s.logger.Debug(fmt.Sprintf("LLM generate failed model=%s: %v", s.model, err))
		}
		return s.tryFallback(ctx, req, fmt.Errorf("OpenRouter chat failed: %w", err))
	}
	if len(resp.Choices) == 0 {
		return s.tryFallback(ctx, req, errors.New("OpenRouter: empty choices"))
	}

	message := resp.Choices[0].Message
	content := extractMessageContent(message)
	if content == "" {
		if s.logger != nil {
			s.logger.Error(godump.DumpJSONStr(resp))
		}
		godump.DumpJSON(resp)
		return s.tryFallback(ctx, req, errors.New("OpenRouter: empty content"))
	}

	if s.logger != nil {
		s.logger.Debug(fmt.Sprintf("LLM generate succeeded model=%s content_len=%d", s.model, len(content)))
	}

	return content, nil
}

func extractMessageContent(message openrouter.ChatCompletionMessage) string {
	if trimmed := strings.TrimSpace(message.Content.Text); trimmed != "" {
		return trimmed
	}
	if len(message.Content.Multi) > 0 {
		for _, part := range message.Content.Multi {
			if part.Type == openrouter.ChatMessagePartTypeText {
				if trimmed := strings.TrimSpace(part.Text); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return messageReasoning(message)
}

func messageReasoning(message openrouter.ChatCompletionMessage) string {
	if message.Reasoning != nil {
		if trimmed := strings.TrimSpace(*message.Reasoning); trimmed != "" {
			return trimmed
		}
	}
	for _, detail := range message.ReasoningDetails {
		if trimmed := strings.TrimSpace(detail.Text); trimmed != "" {
			return trimmed
		}
		if trimmed := strings.TrimSpace(detail.Summary); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (s *LLMService) tryFallback(ctx context.Context, req LLMRequest, cause error) (string, error) {
	if s.fallback == nil {
		return "", cause
	}
	if s.logger != nil {
		s.logger.Debug(fmt.Sprintf("OpenRouter failed, falling back: %v", cause))
	}

	result, err := s.fallback.Generate(ctx, req)
	if err != nil {
		if s.logger != nil {
			s.logger.Debug(fmt.Sprintf("Fallback generate failed: %v", err))
		}
		return "", fmt.Errorf("%v; fallback: %w", cause, err)
	}

	if s.logger != nil {
		s.logger.Debug("Fallback generate succeeded")
	}
	return result, nil
}
