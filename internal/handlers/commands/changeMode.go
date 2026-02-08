package commands

import (
	"fmt"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/handlers/auth"
	"safeboxtgbot/models"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	btnModeRofl  = telebot.Btn{Unique: "btn_mode_rofl", Text: "😜 Рофл"}
	btnModeCozy  = telebot.Btn{Unique: "btn_mode_cozy", Text: "🏡 Уют"}
	btnModeCare  = telebot.Btn{Unique: "btn_mode_care", Text: "🤍 Забота"}
	btnModeClose = telebot.Btn{Unique: "btn_mode_close", Text: "✖️ Закрыть"}
)

func initChangeModeHandler(bot *b.Bot) {
	bot.Handle("/change_mode", createChangeModeHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnModeRofl, createModeSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnModeCozy, createModeSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnModeCare, createModeSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnModeClose, createCloseModeHandler(bot), auth.CreateAuthMiddleware(bot))
}

func createChangeModeHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		user := bot.UserService.GetUser(userID)
		if user == nil || user.TelegramID == 0 {
			return nil
		}

		msg := bot.MustSend(userID, fmt.Sprintf(bot.Replies.ChangeModePrompt, humanModeName(user.Mode)), changeModeMarkup(user.Mode))
		if msg == nil {
			return ctx.Send(bot.Replies.Error)
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		return nil
	}
}

func createModeSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		raw := ctx.Data()
		if raw == "" && ctx.Callback() != nil {
			raw = ctx.Callback().Data
		}
		mode, ok := parseMode(raw)
		if !ok {
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if err := bot.UserService.UpdateMode(userID, mode); err != nil {
			bot.Logger.Error(fmt.Sprintf("Error updating mode for userID=%d: %v", userID, err))
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		msg := bot.MustSend(userID, fmt.Sprintf(bot.Replies.ChangeModeUpdated, humanModeName(mode)))
		if msg != nil {
			go func(m *telebot.Message) {
				time.Sleep(5 * time.Second)
				bot.MustDelete(m)
			}(msg)
		}

		return ctx.Respond()
	}
}

func createCloseModeHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}
		return ctx.Respond()
	}
}

func changeModeMarkup(current models.UserMode) *telebot.ReplyMarkup {
	type option struct {
		mode   models.UserMode
		text   string
		unique string
	}

	options := []option{
		{mode: constants.RoflMode, text: btnModeRofl.Text, unique: btnModeRofl.Unique},
		{mode: constants.CozyMode, text: btnModeCozy.Text, unique: btnModeCozy.Unique},
		{mode: constants.CareMode, text: btnModeCare.Text, unique: btnModeCare.Unique},
	}

	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, 0, len(options))
	for _, opt := range options {
		if current != "" && opt.mode == current {
			continue
		}
		rows = append(rows, markup.Row(markup.Data(opt.text, opt.unique, string(opt.mode))))
	}

	if len(rows) == 0 {
		for _, opt := range options {
			rows = append(rows, markup.Row(markup.Data(opt.text, opt.unique, string(opt.mode))))
		}
	}

	rows = append(rows, markup.Row(btnModeClose))

	markup.Inline(rows...)
	return markup
}

func parseMode(raw string) (models.UserMode, bool) {
	switch models.UserMode(raw) {
	case constants.RoflMode:
		return constants.RoflMode, true
	case constants.CozyMode:
		return constants.CozyMode, true
	case constants.CareMode:
		return constants.CareMode, true
	default:
		return "", false
	}
}

func humanModeName(mode models.UserMode) string {
	switch mode {
	case constants.RoflMode:
		return "Рофл"
	case constants.CareMode:
		return "Забота"
	case constants.CozyMode:
		return "Уют"
	default:
		return "Уют"
	}
}
