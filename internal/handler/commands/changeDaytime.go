package commands

import (
	"fmt"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/internal/middleware/auth"
	"strconv"
	"time"

	b "safeboxtgbot/internal"

	"gopkg.in/telebot.v4"
)

var (
	btnDayStartSelect = telebot.Btn{Unique: "btn_day_start_select"}
	btnDayEndSelect   = telebot.Btn{Unique: "btn_day_end_select"}
	btnDaytimeClose   = telebot.Btn{Unique: "btn_daytime_close", Text: "✖️ Закрыть"}
)

var (
	dayStartSlots = []int{360, 420, 480, 540, 600, 660, 720, 780, 840, 900, 960} // 06:00 ... 16:00 (hourly)
	dayEndSlots   = []int{1080, 1140, 1200, 1260, 1320, 1380, 0, 60, 120}        // 18:00 ... 02:00 (hourly, midnight as 0)
)

func initChangeDaytimeHandler(bot *b.Bot) {
	bot.Handle("/change_daytime", createChangeDaytimeHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnDayStartSelect, createDayStartSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnDayEndSelect, createDayEndSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnDaytimeClose, createCloseDaytimeHandler(bot), auth.CreateAuthMiddleware(bot))
}

func createChangeDaytimeHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		user := bot.UserService.GetUser(userID)
		if user == nil || user.TelegramID == 0 {
			return nil
		}

		bot.UserService.ClearDayStartSelection(userID)
		text := fmt.Sprintf(bot.Replies.ChangeDayStartPrompt, helpers.FormatTimeHM(int(user.DayStart)), helpers.FormatTimeHM(int(user.DayEnd)))
		msg := bot.MustSend(userID, text, dayStartMarkup())
		if msg == nil {
			return ctx.Send(bot.Replies.Error)
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}
		return nil
	}
}

func createDayStartSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		raw := ctx.Data()
		if raw == "" && ctx.Callback() != nil {
			raw = ctx.Callback().Data
		}
		minutes, err := strconv.Atoi(raw)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		bot.UserService.SetDayStartSelection(userID, minutes)

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		text := fmt.Sprintf(bot.Replies.ChangeDayEndPrompt, helpers.FormatTimeHM(minutes))
		msg := bot.MustSend(userID, text, dayEndMarkup(minutes))
		if msg == nil {
			return ctx.Send(bot.Replies.Error)
		}

		return ctx.Respond()
	}
}

func createDayEndSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		raw := ctx.Data()
		if raw == "" && ctx.Callback() != nil {
			raw = ctx.Callback().Data
		}

		dayStart := bot.UserService.GetDayStartSelection(userID)
		if dayStart < 0 {
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		minutes, err := strconv.Atoi(raw)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}
		if minutes == dayStart {
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if err := bot.UserService.UpdateDayWindow(userID, dayStart, minutes); err != nil {
			bot.Logger.Error(fmt.Sprintf("Error updating day window for userID=%d: %v", userID, err))
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}
		bot.UserService.ClearDayStartSelection(userID)

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		msg := bot.MustSend(userID, fmt.Sprintf(bot.Replies.ChangeDayUpdated, helpers.FormatTimeHM(dayStart), helpers.FormatTimeHM(minutes)))
		if msg != nil {
			go func(m *telebot.Message) {
				time.Sleep(5 * time.Second)
				bot.MustDelete(m)
			}(msg)
		}

		return ctx.Respond()
	}
}

func createCloseDaytimeHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		bot.UserService.ClearDayStartSelection(ctx.Chat().ID)
		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}
		return ctx.Respond()
	}
}

func dayStartMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, 0, 3)
	for i := 0; i < len(dayStartSlots); i += 3 {
		row := make([]telebot.Btn, 0, 3)
		for j := i; j < i+3 && j < len(dayStartSlots); j++ {
			minutes := dayStartSlots[j]
			row = append(row, markup.Data(helpers.FormatTimeHM(minutes), btnDayStartSelect.Unique, strconv.Itoa(minutes)))
		}
		rows = append(rows, markup.Row(row...))
	}
	rows = append(rows, markup.Row(btnDaytimeClose))
	markup.Inline(rows...)
	return markup
}

func dayEndMarkup(startMinutes int) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, 0, 3)
	for i := 0; i < len(dayEndSlots); i += 3 {
		row := make([]telebot.Btn, 0, 3)
		for j := i; j < i+3 && j < len(dayEndSlots); j++ {
			minutes := dayEndSlots[j]
			if minutes == startMinutes {
				continue
			}
			row = append(row, markup.Data(helpers.FormatTimeHM(minutes), btnDayEndSelect.Unique, strconv.Itoa(minutes)))
		}
		if len(row) > 0 {
			rows = append(rows, markup.Row(row...))
		}
	}
	rows = append(rows, markup.Row(btnDaytimeClose))
	markup.Inline(rows...)
	return markup
}
