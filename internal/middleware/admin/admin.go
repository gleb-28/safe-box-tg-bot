package admin

import (
	b "safeboxtgbot/internal"

	"gopkg.in/telebot.v4"
)

func CreateAdminMiddleware(bot *b.Bot) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			if bot == nil || bot.Config == nil || ctx == nil || ctx.Chat() == nil {
				return nil
			}
			if ctx.Chat().ID != bot.Config.AdminID {
				return nil
			}
			return next(ctx)
		}
	}
}
