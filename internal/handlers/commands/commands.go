package commands

import (
	"log"
	bot "safeboxtgbot/internal"
	"safeboxtgbot/internal/handlers/auth"

	"gopkg.in/telebot.v4"
)

var commands = []telebot.Command{
	{Text: "start", Description: "Старт"},
	{Text: "key", Description: "Ввести секретный ключ"},
}

func MustInitCommandsHandler(bot *bot.Bot) {
	err := bot.SetCommands(commands)
	if err != nil {
		log.Fatal("Failed to set commands: " + err.Error())
	}

	bot.Handle("/start", createStartHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle("/key", createKeyHandler(bot))
}
