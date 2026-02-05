package message

import (
	"context"
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"

	"gopkg.in/telebot.v4"
)

func createValidateActivationKey(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userId := ctx.Chat().ID

		if ctx.Message().Text == bot.Config.ActivationKey {
			err := bot.UserService.AddUser(userId)
			if err != nil {
				bot.MustSend(userId, bot.Replies.Error)
				bot.Fsm.UserEvent(context.Background(), userId, fsmManager.InitialEvent)
				return nil
			}
			bot.MustSend(userId, bot.Replies.EnterKeySuccess)
			bot.MustSend(userId, bot.Replies.Start)
			bot.Fsm.UserEvent(context.Background(), userId, fsmManager.InitialEvent)
			return nil
		}

		bot.Fsm.UserEvent(context.Background(), userId, fsmManager.InitialEvent)
		return ctx.Send(bot.Replies.EnterKeyWrong)
	}
}
