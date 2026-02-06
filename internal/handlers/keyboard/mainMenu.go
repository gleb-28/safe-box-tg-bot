package keyboard

import "gopkg.in/telebot.v4"

const OpenItemBoxLabel = "Моя Шкатулка 📦"

func MainMenuKeyboard() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
	markup.Reply(markup.Row(markup.Text(OpenItemBoxLabel)))
	return markup
}
