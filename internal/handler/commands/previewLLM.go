package commands

import (
	"context"
	"fmt"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/feat/notify"
	"safeboxtgbot/internal/feat/prompt"

	"gopkg.in/telebot.v4"
)

func createPreviewLLMHandler(bot *b.Bot, messageGenerator prompt.MessageGenerator) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if messageGenerator == nil {
			bot.Logger.Error("LLM preview: message generator is nil")
			return ctx.Send(bot.Replies.Error)
		}

		service := notify.NewLLMPreviewService(
			bot.UserService,
			bot.ItemsService,
			messageGenerator,
			bot,
			bot.Logger,
		)

		if err := service.SendPreviews(context.Background(), ctx.Chat().ID); err != nil {
			bot.Logger.Error(fmt.Sprintf("LLM preview failed for userID=%d: %v", ctx.Chat().ID, err))
			return ctx.Send(fmt.Sprintf("LLM preview error: %v", err))
		}

		return ctx.Send("LLM preview sent")
	}
}
