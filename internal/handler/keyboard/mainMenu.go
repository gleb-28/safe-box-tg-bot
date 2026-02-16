package keyboard

import "gopkg.in/telebot.v4"

const OpenItemBoxLabel = "Открыть Шкатулку 👀"
const OpenReminderBoxLabel = "Открыть Напоминания"

func MainMenuKeyboard() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
	markup.Reply(
		markup.Row(markup.Text(OpenItemBoxLabel)),
		markup.Row(markup.Text(OpenReminderBoxLabel)),
	)
	return markup
}
