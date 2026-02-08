package prompt

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"safeboxtgbot/internal/helpers"
	"strings"

	"safeboxtgbot/internal/core/logger"
)

type MessageGenerator interface {
	Generate(ctx context.Context, input LLMInput) (string, error)
}

type PromptBuilder interface {
	BuildSystem() string
	BuildUser(input LLMInput) string
}

type LLMGenerator interface {
	Generate(ctx context.Context, req LLMRequest) (string, error)
}

type LLMInput struct {
	CurrentEntity string
	TimeOfDay     string
	StyleMode     string
	RandomSeed    int
}

type defaultPromptBuilder struct {
	prompt string
	logger logger.AppLogger
}

func MustNewPromptBuilder(promptPath string, appLogger logger.AppLogger) PromptBuilder {
	data, err := os.ReadFile(promptPath)
	if err != nil {
		stdlog.Fatalf("read prompt: %v", err)
	}
	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		stdlog.Fatal("prompt is empty")
	}

	return &defaultPromptBuilder{
		prompt: prompt,
		logger: appLogger,
	}
}

func (b *defaultPromptBuilder) BuildSystem() string {
	return b.prompt
}

func (b *defaultPromptBuilder) BuildUser(input LLMInput) string {
	return buildUserPrompt(input)
}

type MessageOrchestrator struct {
	builder PromptBuilder
	llm     LLMGenerator
	logger  logger.AppLogger
}

func MustNewMessageGenerator(builder PromptBuilder, llm LLMGenerator, appLogger logger.AppLogger) *MessageOrchestrator {
	if builder == nil {
		stdlog.Fatal("prompt builder is nil")
	}
	if llm == nil {
		stdlog.Fatal("llm generator is nil")
	}

	return &MessageOrchestrator{
		builder: builder,
		llm:     llm,
		logger:  appLogger,
	}
}

func (g *MessageOrchestrator) Generate(ctx context.Context, input LLMInput) (string, error) {
	if g == nil || g.builder == nil || g.llm == nil {
		return "", errors.New("message generator is nil")
	}
	if strings.TrimSpace(input.CurrentEntity) == "" {
		return "", errors.New("current_entity is empty")
	}

	raw, err := g.llm.Generate(ctx, LLMRequest{
		SystemPrompt: g.builder.BuildSystem(),
		UserPrompt:   g.builder.BuildUser(input),
		Temperature:  0.8,
		MaxTokens:    180,
	})
	if err != nil {
		return "", err
	}

	text := helpers.CleanLLMText(raw)
	if text == "" {
		return "", errors.New("llm response is empty")
	}

	if g.logger != nil {
		g.logger.Debug(fmt.Sprintf("PromptBuilder generated text for entity=%s", input.CurrentEntity))
	}

	return text, nil
}

func buildUserPrompt(input LLMInput) string {
	return fmt.Sprintf(`{"current_entity":%q,"time_of_day":%q,"style_mode":%q,"random_seed":%d}`,
		input.CurrentEntity,
		input.TimeOfDay,
		input.StyleMode,
		input.RandomSeed,
	)
}
