package prompt

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"strings"
	"time"

	"safeboxtgbot/internal/core/logger"

	"github.com/revrost/go-openrouter"
)

type OpenRouterClient struct {
	client *openrouter.Client
	logger logger.AppLogger
}

func MustNewOpenRouterClient(apiKey string, appLogger logger.AppLogger) *OpenRouterClient {
	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey == "" {
		stdlog.Fatal("OpenRouter api key is empty")
	}

	config := openrouter.DefaultConfig(trimmedKey)
	if config == nil {
		stdlog.Fatal("OpenRouter config is nil")
	}
	config.HTTPClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	config.XTitle = "Nudger" // Working title

	client := openrouter.NewClientWithConfig(*config)
	if client == nil {
		stdlog.Fatal("OpenRouter client is nil")
	}

	if appLogger != nil {
		appLogger.Debug("OpenRouter client initialized")
	}

	return &OpenRouterClient{
		client: client,
		logger: appLogger,
	}
}

func (c *OpenRouterClient) Chat(ctx context.Context, reqBody openrouter.ChatCompletionRequest) (openrouter.ChatCompletionResponse, error) {
	if c == nil || c.client == nil {
		return openrouter.ChatCompletionResponse{}, errors.New("OpenRouter client is nil")
	}
	if c.logger != nil {
		c.logger.Debug(fmt.Sprintf("OpenRouter chat request model=%s messages=%d", reqBody.Model, len(reqBody.Messages)))
	}
	return c.client.CreateChatCompletion(ctx, reqBody)
}
