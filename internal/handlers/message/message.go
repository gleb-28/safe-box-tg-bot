package message

import (
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"

	"gopkg.in/telebot.v4"
)

func MustInitMessagesHandler(bot *b.Bot) {
	bot.Handle(telebot.OnText, createMessageHandler(bot))
}

func createMessageHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		chatId := ctx.Chat().ID
		userFsm := bot.Fsm.GetFSMForUser(chatId)

		switch userFsm.Current() {
		case fsmManager.StateInitial:
			return nil
		case fsmManager.StateAwaitingKey:
			return createValidateActivationKey(bot)(ctx)
		default:
			return nil
		}
	}
}
