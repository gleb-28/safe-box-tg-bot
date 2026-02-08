package message

import (
	"context"
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/handler/keyboard"

	"gopkg.in/telebot.v4"
)

func createValidateActivationKey(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID

		if ctx.Message().Text == bot.Config.ActivationKey {
			err := bot.UserService.AddUser(userID)
			if err != nil {
				bot.MustSend(userID, bot.Replies.Error)
				bot.Fsm.UserEvent(context.Background(), userID, fsmManager.InitialEvent)
				return nil
			}
			bot.MustSend(userID, bot.Replies.EnterKeySuccess)
			bot.MustSend(userID, bot.Replies.Start, keyboard.MainMenuKeyboard())
			bot.Fsm.UserEvent(context.Background(), userID, fsmManager.InitialEvent)
			return nil
		}

		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.InitialEvent)
		return ctx.Send(bot.Replies.EnterKeyWrong)
	}
}
