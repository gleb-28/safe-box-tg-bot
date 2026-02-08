package commands

import (
	"fmt"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/internal/middleware/auth"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	btnIntervalRare   = telebot.Btn{Unique: "btn_interval_rare", Text: "⏳ Редко"}
	btnIntervalNormal = telebot.Btn{Unique: "btn_interval_normal", Text: "⏱ Иногда"}
	btnIntervalOften  = telebot.Btn{Unique: "btn_interval_often", Text: "🔔 Часто"}
	btnIntervalChaos  = telebot.Btn{Unique: "btn_interval_chaos", Text: "🎲 Хаос"}
	btnIntervalClose  = telebot.Btn{Unique: "btn_interval_close", Text: "✖️ Закрыть"}
)

func initChangeIntervalHandler(bot *b.Bot) {
	bot.Handle("/change_interval", createChangeIntervalHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnIntervalRare, createIntervalSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnIntervalNormal, createIntervalSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnIntervalOften, createIntervalSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnIntervalChaos, createIntervalSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnIntervalClose, createCloseIntervalHandler(bot), auth.CreateAuthMiddleware(bot))
}

func createChangeIntervalHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		user := bot.UserService.GetUser(userID)
		if user == nil || user.TelegramID == 0 {
			return nil
		}

		msg := bot.MustSend(userID, intervalPromptText(bot, user.NotificationPreset), changeIntervalMarkup(user.NotificationPreset))
		if msg == nil {
			return ctx.Send(bot.Replies.Error)
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		return nil
	}
}

func createIntervalSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		raw := ctx.Data()
		if raw == "" && ctx.Callback() != nil {
			raw = ctx.Callback().Data
		}
		preset, ok := helpers.ParseNotificationPreset(raw)
		if !ok {
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if err := bot.UserService.UpdateNotificationPreset(userID, preset); err != nil {
			bot.Logger.Error(fmt.Sprintf("Error updating notification preset for userID=%d: %v", userID, err))
			return ctx.Respond(&telebot.CallbackResponse{Text: bot.Replies.Error, ShowAlert: true})
		}

		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}

		msg := bot.MustSend(
			userID,
			fmt.Sprintf(bot.Replies.ChangeIntervalUpdated, helpers.HumanNotificationPresetName(preset.Key), helpers.NotificationPresetRangeText(preset)),
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

func createCloseIntervalHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if ctx.Message() != nil {
			bot.MustDelete(ctx.Message())
		}
		return ctx.Respond()
	}
}

func changeIntervalMarkup(current string) *telebot.ReplyMarkup {
	type option struct {
		preset constants.NotificationPreset
		btn    telebot.Btn
	}

	options := []option{
		{preset: constants.NotificationPresets[constants.NotificationPresetRare], btn: btnIntervalRare},
		{preset: constants.NotificationPresets[constants.NotificationPresetNormal], btn: btnIntervalNormal},
		{preset: constants.NotificationPresets[constants.NotificationPresetOften], btn: btnIntervalOften},
		{preset: constants.NotificationPresets[constants.NotificationPresetChaos], btn: btnIntervalChaos},
	}

	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, 0, len(options))
	for _, opt := range options {
		if opt.preset.Key == "" {
			continue
		}
		if current != "" && opt.preset.Key == current {
			continue
		}
		text := opt.preset.Name
		rangeText := helpers.NotificationPresetRangeText(opt.preset)
		if rangeText != "" {
			text = fmt.Sprintf("%s • %s", text, rangeText)
		}
		rows = append(rows, markup.Row(markup.Data(text, opt.btn.Unique, opt.preset.Key)))
	}

	if len(rows) == 0 {
		for _, opt := range options {
			if opt.preset.Key == "" {
				continue
			}
			text := opt.preset.Name
			rangeText := helpers.NotificationPresetRangeText(opt.preset)
			if rangeText != "" {
				text = fmt.Sprintf("%s • %s", text, rangeText)
			}
			rows = append(rows, markup.Row(markup.Data(text, opt.btn.Unique, opt.preset.Key)))
		}
	}

	rows = append(rows, markup.Row(btnIntervalClose))

	markup.Inline(rows...)
	return markup
}

func intervalPromptText(bot *b.Bot, currentPreset string) string {
	currentName := helpers.HumanNotificationPresetName(currentPreset)
	lines := []string{
		formatPresetLine(constants.NotificationPresets[constants.NotificationPresetRare]),
		formatPresetLine(constants.NotificationPresets[constants.NotificationPresetNormal]),
		formatPresetLine(constants.NotificationPresets[constants.NotificationPresetOften]),
		formatPresetLine(constants.NotificationPresets[constants.NotificationPresetChaos]),
	}
	return fmt.Sprintf("%s\n\n%s", fmt.Sprintf(bot.Replies.ChangeIntervalPrompt, currentName), strings.Join(lines, "\n"))
}

func formatPresetLine(preset constants.NotificationPreset) string {
	rangeText := helpers.FormatMinutesRange(preset.MinMinutes, preset.MaxMinutes)
	if preset.Key == constants.NotificationPresetChaos {
		if rangeText == "" {
			return fmt.Sprintf("%s — как повезёт 😄", preset.Name)
		}
		return fmt.Sprintf("%s — как повезёт 😄 (%s)", preset.Name, rangeText)
	}
	if rangeText == "" {
		return preset.Name
	}
	return fmt.Sprintf("%s — раз в %s", preset.Name, rangeText)
}
