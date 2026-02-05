package auth

import (
	b "safeboxtgbot/internal"

	"gopkg.in/telebot.v4"
)

func CreateAuthMiddleware(bot *b.Bot) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			userId := ctx.Chat().ID
			user := bot.UserService.GetUser(userId)

			if user.TelegramID == 0 {
				return nil
			}

			return next(ctx)
		}
	}
}
