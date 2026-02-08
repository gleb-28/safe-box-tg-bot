package commands

import (
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/handler/keyboard"

	"gopkg.in/telebot.v4"
)

func createStartHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		return ctx.Send(bot.Replies.Start, keyboard.MainMenuKeyboard())
	}
}
