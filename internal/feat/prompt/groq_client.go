package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"strings"
	"time"

	"safeboxtgbot/internal/core/logger"
)

const groqChatCompletionsURL = "https://api.groq.com/openai/v1/chat/completions"

type GroqClient struct {
	httpClient *http.Client
	apiKey     string
	logger     logger.AppLogger
}

type GroqChatCompletionRequest struct {
	Model               string                 `json:"model"`
	Messages            []GroqChatMessage      `json:"messages"`
	Temperature         *float64               `json:"temperature,omitempty"`
	MaxTokens           *int                   `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                   `json:"max_completion_tokens,omitempty"`
	TopP                *float64               `json:"top_p,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

type GroqChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqChatCompletionResponse struct {
	ID      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Model   string                     `json:"model"`
	Choices []GroqChatCompletionChoice `json:"choices"`
	Usage   *GroqUsage                 `json:"usage,omitempty"`
}

type GroqChatCompletionChoice struct {
	Index        int                       `json:"index"`
	Message      GroqChatCompletionMessage `json:"message"`
	FinishReason string                    `json:"finish_reason"`
}

type GroqChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func MustNewGroqClient(apiKey string, appLogger logger.AppLogger) *GroqClient {
	trimmed := strings.TrimSpace(apiKey)
	if trimmed == "" {
		stdlog.Fatal("Groq api key is empty")
	}
	return &GroqClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiKey: trimmed,
		logger: appLogger,
	}
}

func (c *GroqClient) Chat(ctx context.Context, req GroqChatCompletionRequest) (GroqChatCompletionResponse, error) {
	if c == nil {
		return GroqChatCompletionResponse{}, errors.New("Groq client is nil")
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return GroqChatCompletionResponse{}, fmt.Errorf("marshal Groq request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, groqChatCompletionsURL, bytes.NewReader(payload))
	if err != nil {
		return GroqChatCompletionResponse{}, fmt.Errorf("new Groq request: %w", err)
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	if c.logger != nil {
		c.logger.Debug(fmt.Sprintf("Groq chat request model=%s messages=%d", req.Model, len(req.Messages)))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GroqChatCompletionResponse{}, fmt.Errorf("Groq request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GroqChatCompletionResponse{}, fmt.Errorf("read Groq response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return GroqChatCompletionResponse{}, fmt.Errorf("Groq status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out GroqChatCompletionResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&out); err != nil {
		return GroqChatCompletionResponse{}, fmt.Errorf("decode Groq response: %w", err)
	}

	return out, nil
}
