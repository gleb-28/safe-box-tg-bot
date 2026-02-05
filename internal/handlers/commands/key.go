package commands

import (
	"context"
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"

	"gopkg.in/telebot.v4"
)

func createKeyHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userId := ctx.Chat().ID
		user := bot.UserService.GetUser(userId)
		if user.TelegramID == 0 {
			bot.Fsm.UserEvent(context.Background(), userId, fsmManager.AwaitingKeyEvent)
			return ctx.Send(bot.Replies.EnterKey)
		}

		return ctx.Send(bot.Replies.KeyAlreadyEntered)
	}
}
