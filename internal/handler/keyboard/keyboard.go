package keyboard

import (
	b "safeboxtgbot/internal"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/middleware/auth"

	"gopkg.in/telebot.v4"
)

func MustInitKeyboardHandler(bot *b.Bot) {
	bot.Handle(OpenItemBoxLabel, createOpenItemBoxBtnHandler(bot), auth.CreateAuthMiddleware(bot))
	MustInitItemBoxButtons(bot)
}

func createOpenItemBoxBtnHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		if bot.Fsm.GetFSMForUser(userID).Current() == fsmManager.StateInitial {
			bot.MustDelete(ctx.Message())
			return OpenItemBox(bot, userID, nil)
		}
		bot.MustDelete(ctx.Message())
		return nil
	}
}
