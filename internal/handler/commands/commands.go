package commands

import (
	"log"
	bot "safeboxtgbot/internal"
	"safeboxtgbot/internal/middleware/auth"

	"gopkg.in/telebot.v4"
)

var commands = []telebot.Command{
	{Text: "start", Description: "Старт"},
	{Text: "key", Description: "Ввести секретный ключ"},
	{Text: "change_mode", Description: "Сменить стиль сообщений"},
	{Text: "change_interval", Description: "Сменить частоту напоминаний"},
	{Text: "change_daytime", Description: "Настроить время для уведомлений"},
	{Text: "toggle_notifications", Description: "Включить/выключить уведомления"},
}

func MustInitCommandsHandler(bot *bot.Bot) {
	err := bot.SetCommands(commands)
	if err != nil {
		log.Fatal("Failed to set commands: " + err.Error())
	}

	bot.Handle("/start", createStartHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle("/key", createKeyHandler(bot))
	initChangeModeHandler(bot)
	initChangeIntervalHandler(bot)
	initToggleNotificationsHandler(bot)
	initChangeDaytimeHandler(bot)
}
