package commands

import (
	"fmt"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/internal/middleware/auth"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	btnNotifyOn    = telebot.Btn{Unique: "btn_notify_on", Text: "🔔 Включить"}
	btnNotifyOff   = telebot.Btn{Unique: "btn_notify_off", Text: "🔕 Выключить"}
	btnNotifyClose = telebot.Btn{Unique: "btn_notify_close", Text: "✖️ Закрыть"}
)

func initToggleNotificationsHandler(bot *b.Bot) {
	bot.Handle("/toggle_notifications", createToggleNotificationsHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnNotifyOn, createToggleNotificationsSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnNotifyOff, createToggleNotificationsSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnNotifyClose, createCloseToggleNotificationsHandler(bot), auth.CreateAuthMiddleware(bot))
}

func createToggleNotificationsHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		user := bot.UserService.GetUser(userID)
		if user == nil || user.TelegramID == 0 {
			return nil
		}

		msg := bot.MustSend(userID, fmt.Sprintf(bot.Replies.ToggleNotificationsPrompt, helpers.HumanNotificationStatus(user.NotificationsMuted)), toggleNotificationsMarkup(user.NotificationsMuted))
		if msg == nil {
			return ctx.Send(bot.Replies.Error)
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		return nil
	}
}

func createToggleNotificationsSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		raw := ctx.Data()
		if raw == "" && ctx.Callback() != nil {
			raw = ctx.Callback().Data
		}

		var muted bool
		switch raw {
		case "on":
			muted = false
		case "off":
			muted = true
		default:
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if err := bot.UserService.SetNotificationsMuted(userID, muted); err != nil {
			bot.Logger.Error(fmt.Sprintf("Error toggling notifications for userID=%d: %v", userID, err))
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		msg := bot.MustSend(
			userID,
			fmt.Sprintf(bot.Replies.ToggleNotificationsUpdated, helpers.HumanNotificationStatus(muted)),
		)
		if msg != nil {
			go func(m *telebot.Message) {
				time.Sleep(5 * time.Second)
				bot.MustDelete(m)
			}(msg)
		}

		return ctx.Respond()
	}
}

func createCloseToggleNotificationsHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}
		return ctx.Respond()
	}
}

func toggleNotificationsMarkup(muted bool) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, 0, 3)

	if muted {
		rows = append(rows, markup.Row(markup.Data(btnNotifyOn.Text, btnNotifyOn.Unique, "on")))
	} else {
		rows = append(rows, markup.Row(markup.Data(btnNotifyOff.Text, btnNotifyOff.Unique, "off")))
	}

	rows = append(rows, markup.Row(btnNotifyClose))
	markup.Inline(rows...)
	return markup
}
