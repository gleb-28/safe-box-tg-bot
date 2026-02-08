package message

import (
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/handler/keyboard"

	"gopkg.in/telebot.v4"
)

func MustInitMessagesHandler(bot *b.Bot) {
	bot.Handle(telebot.OnText, createMessageHandler(bot))
}

func createMessageHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		userFsm := bot.Fsm.GetFSMForUser(userID)

		switch userFsm.Current() {
		case fsmManager.StateInitial:
			return nil
		case fsmManager.StateAwaitingKey:
			return createValidateActivationKey(bot)(ctx)
		case fsmManager.StateAwaitingItemAdd:
			return keyboard.CreateValidateAddItemHandler(bot)(ctx)
		case fsmManager.StateAwaitingItemEdit:
			return keyboard.CreateValidateEditItemHandler(bot)(ctx)
		default:
			return nil
		}
	}
}
