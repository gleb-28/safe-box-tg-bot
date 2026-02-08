package commands

import (
	"context"
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"

	"gopkg.in/telebot.v4"
)

func createKeyHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		user := bot.UserService.GetUser(userID)
		if user.TelegramID == 0 {
			bot.Fsm.UserEvent(context.Background(), userID, fsmManager.AwaitingKeyEvent)
			return ctx.Send(bot.Replies.EnterKey)
		}

		return ctx.Send(bot.Replies.KeyAlreadyEntered)
	}
}
