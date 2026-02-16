package keyboard

import (
	"context"
	"errors"
	"fmt"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/feat/reminder"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/internal/middleware/auth"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/models"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	btnAddReminder            = telebot.Btn{Unique: "btn_add_reminder", Text: "➕ Добавить"}
	btnDeleteReminder         = telebot.Btn{Unique: "btn_delete_reminder", Text: "🗑 Удалить"}
	btnCloseReminderBox       = telebot.Btn{Unique: "btn_close_reminder_box", Text: "✖️ Закрыть"}
	btnBackToReminderBox      = telebot.Btn{Unique: "btn_back_to_reminder_box", Text: "⬅️ Назад"}
	btnSelectReminderToDelete = telebot.Btn{Unique: "btn_select_reminder_to_delete"}

	btnReminderDaily    = telebot.Btn{Unique: "btn_reminder_daily", Text: "Ежедневно"}
	btnReminderWeekly   = telebot.Btn{Unique: "btn_reminder_weekly", Text: "Еженедельно"}
	btnReminderMonthly  = telebot.Btn{Unique: "btn_reminder_monthly", Text: "Ежемесячно"}
	btnReminderInterval = telebot.Btn{Unique: "btn_reminder_interval", Text: "Через интервал"}
	btnReminderOnce     = telebot.Btn{Unique: "btn_reminder_once", Text: "Один раз"}
	btnSelectWeekday    = telebot.Btn{Unique: "btn_select_weekday"}
)

func MustInitReminderBoxButtons(bot *b.Bot) {
	bot.Handle(&btnAddReminder, createAddReminderHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnDeleteReminder, createDeleteReminderHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnCloseReminderBox, createCloseReminderBoxHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnBackToReminderBox, createBackToReminderBoxHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnSelectReminderToDelete, createDeleteReminderSelectHandler(bot), auth.CreateAuthMiddleware(bot))

	bot.Handle(&btnReminderDaily, createSelectScheduleHandler(bot, models.ReminderScheduleDaily), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnReminderWeekly, createSelectScheduleHandler(bot, models.ReminderScheduleWeekly), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnReminderMonthly, createSelectScheduleHandler(bot, models.ReminderScheduleMonthly), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnReminderInterval, createSelectScheduleHandler(bot, models.ReminderScheduleInterval), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnReminderOnce, createSelectScheduleHandler(bot, models.ReminderScheduleOnce), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnSelectWeekday, createSelectWeekdayHandler(bot), auth.CreateAuthMiddleware(bot))
}

func OpenReminderBox(bot *b.Bot, userID int64, sourceMsg *telebot.Message) error {
	clearClosedReminderBoxMessage(bot, userID)
	bot.Fsm.UserEvent(context.Background(), userID, fsmManager.RemindersMenuOpenedEvent)
	return renderReminderBox(bot, userID, sourceMsg, "")
}

func createAddReminderHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		bot.ReminderService.ClearPending(userID)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.AwaitingReminderAddEvent)
		return renderSchedulePrompt(bot, userID, ctx.Message(), "")
	}
}

func CreateValidateAddReminderHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		raw := strings.TrimSpace(ctx.Message().Text)
		bot.MustDelete(ctx.Message())
		loc := safeUserLoc(bot, userID)

		pending := bot.ReminderService.GetPending(userID)
		if pending == nil || pending.ScheduleType == "" {
			return renderSchedulePrompt(bot, userID, nil, bot.Replies.ReminderSelectTypeFirst)
		}

		if handled, err := handleNameStep(bot, userID, pending, raw); err != nil {
			return err
		} else if handled {
			return nil
		}

		if handled, err := handleScheduleSpecificStep(bot, userID, pending, raw, loc); err != nil {
			return err
		} else if handled {
			return nil
		}

		if handled, err := handleTimeStep(bot, userID, pending, raw, loc); err != nil {
			return err
		} else if handled {
			return nil
		}

		return finalizeReminder(bot, userID, pending, "", loc)
	}
}

func finalizeReminder(bot *b.Bot, userID int64, pending *session.PendingReminder, note string, loc *time.Location) error {
	if loc == nil {
		loc = safeUserLoc(bot, userID)
	}
	nowUTC := time.Now().UTC()
	nowLocal := time.Now().In(loc)

	switch pending.ScheduleType {
	case models.ReminderScheduleInterval:
		_, err := bot.ReminderService.CreateInterval(userID, pending.EntityName, *pending.IntervalMinutes, nowUTC, loc)
		if err != nil {
			return handleReminderInputError(bot, userID, err)
		}
	case models.ReminderScheduleDaily:
		_, err := bot.ReminderService.CreateDaily(userID, pending.EntityName, *pending.TimeOfDayMinutes, nowUTC, loc)
		if err != nil {
			return handleReminderInputError(bot, userID, err)
		}
	case models.ReminderScheduleWeekly:
		_, err := bot.ReminderService.CreateWeekly(userID, pending.EntityName, *pending.Weekday, *pending.TimeOfDayMinutes, nowUTC, loc)
		if err != nil {
			return handleReminderInputError(bot, userID, err)
		}
	case models.ReminderScheduleMonthly:
		_, err := bot.ReminderService.CreateMonthly(userID, pending.EntityName, *pending.MonthDay, *pending.TimeOfDayMinutes, nowUTC, loc)
		if err != nil {
			return handleReminderInputError(bot, userID, err)
		}
	case models.ReminderScheduleOnce:
		if pending.OnceDate == nil || pending.TimeOfDayMinutes == nil {
			return handleReminderInputError(bot, userID, reminder.ErrInvalidSchedule)
		}
		runAt := helpers.ComposeDateTime(*pending.OnceDate, int(*pending.TimeOfDayMinutes), loc)
		if runAt.In(loc).Before(nowLocal) {
			pending.TimeOfDayMinutes = nil
			return renderTimePrompt(bot, userID, nil, bot.Replies.ReminderOnceTimePast)
		}
		if _, err := bot.ReminderService.CreateOnce(userID, pending.EntityName, runAt, nowUTC, loc); err != nil {
			return handleReminderInputError(bot, userID, err)
		}
	}

	bot.Fsm.UserEvent(context.Background(), userID, fsmManager.RemindersMenuOpenedEvent)
	bot.ReminderService.ClearPending(userID)

	if note != "" {
		notifyAndDelete(bot, userID, note, 5*time.Second)
	}
	return renderReminderBox(bot, userID, nil, "")
}

func safeUserLoc(bot *b.Bot, userID int64) *time.Location {
	user := bot.UserService.GetUser(userID)
	loc, err := helpers.UserLocation(*user)
	if err != nil {
		return time.UTC
	}
	return loc
}

func createDeleteReminderHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ReminderDeleteSelectEvent)
		return renderDeleteReminderSelect(bot, userID, ctx.Message())
	}
}

func createCloseReminderBoxHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.InitialEvent)
		bot.ReminderService.ClearPending(userID)
		bot.ReminderService.SetBotLastMsg(userID, nil)
		clearClosedReminderBoxMessage(bot, userID)
		msg := bot.MustSend(userID, bot.Replies.ReminderBoxClosed, MainMenuKeyboard())
		saveClosedReminderBoxMessage(bot, userID, msg)
		bot.MustDelete(ctx.Message())
		return nil
	}
}

func createBackToReminderBoxHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		bot.ReminderService.ClearPending(userID)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.RemindersMenuOpenedEvent)
		return renderReminderBox(bot, userID, ctx.Message(), "")
	}
}

func createSelectScheduleHandler(bot *b.Bot, schedule models.ReminderSchedule) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		pending := &session.PendingReminder{ScheduleType: schedule}
		bot.ReminderService.SetPending(userID, pending)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.AwaitingReminderAddEvent)
		return renderNamePrompt(bot, userID, ctx.Message(), "")
	}
}

func createSelectWeekdayHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		pending := bot.ReminderService.GetPending(userID)
		if pending == nil || pending.ScheduleType != models.ReminderScheduleWeekly {
			return renderSchedulePrompt(bot, userID, ctx.Message(), bot.Replies.ReminderSelectTypeFirst)
		}
		raw := strings.TrimSpace(ctx.Data())
		val, err := strconv.Atoi(raw)
		if err != nil || val < 0 || val > 6 {
			return renderWeekdayPrompt(bot, userID, ctx.Message(), bot.Replies.ReminderWeekdayInvalid)
		}
		day := int8(val)
		pending.Weekday = &day
		bot.ReminderService.SetPending(userID, pending)
		return renderTimePrompt(bot, userID, ctx.Message(), "")
	}
}

func createDeleteReminderSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		bot.RespondSilently(ctx)
		reminderID, err := parseUintData(ctx)
		if err != nil {
			return renderReminderBox(bot, userID, ctx.Message(), "")
		}

		if err := bot.ReminderService.Delete(reminderID, userID); err != nil {
			if errors.Is(err, reminder.ErrReminderNotFound) {
				bot.Fsm.UserEvent(context.Background(), userID, fsmManager.RemindersMenuOpenedEvent)
				return renderReminderBox(bot, userID, ctx.Message(), "")
			}
			return upsertReminderLastMessage(bot, userID, ctx.Message(), bot.Replies.Error, reminderBoxMarkup())
		}

		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.RemindersMenuOpenedEvent)
		return renderReminderBox(bot, userID, ctx.Message(), "")
	}
}

func renderReminderBox(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	user := bot.UserService.GetUser(userID)
	reminders, err := bot.ReminderService.GetList(userID)
	if err != nil {
		return upsertReminderLastMessage(bot, userID, sourceMsg, bot.Replies.Error, reminderBoxMarkup())
	}

	status := buildReminderBoxStatus(bot, user, len(reminders))
	var text string
	loc, _ := helpers.UserLocation(*user)

	if len(reminders) == 0 {
		text = fmt.Sprintf(bot.Replies.RemindersMenuEmpty, status)
	} else {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf(bot.Replies.RemindersMenuHeader, status))
		for _, r := range reminders {
			builder.WriteString(fmt.Sprintf(bot.Replies.RemindersMenuItemRow, r.Name, helpers.HumanReminderSchedule(r, loc, bot.Replies)))
		}
		builder.WriteString(bot.Replies.RemindersMenuFooter)
		text = builder.String()
	}

	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, reminderBoxMarkup())
}

func renderDeleteReminderSelect(bot *b.Bot, userID int64, sourceMsg *telebot.Message) error {
	list, err := bot.ReminderService.GetList(userID)
	if err != nil {
		return upsertReminderLastMessage(bot, userID, sourceMsg, bot.Replies.Error, reminderBoxMarkup())
	}
	text := bot.Replies.WhatDoWeDelete
	if len(list) == 0 {
		text = bot.Replies.ListIsEmpty
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, selectReminderMarkup(list))
}

func buildReminderBoxStatus(bot *b.Bot, user *models.User, count int) string {
	if user == nil {
		user = &models.User{}
	}
	mode := helpers.HumanModeName(user.Mode)
	dayStart := helpers.FormatTimeHM(int(user.DayStart))
	dayEnd := helpers.FormatTimeHM(int(user.DayEnd))
	return fmt.Sprintf(bot.Replies.ItemsMenuStatus, mode, count, dayStart, dayEnd)
}

func reminderBoxMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnAddReminder, btnDeleteReminder),
		markup.Row(btnCloseReminderBox),
	)
	return markup
}

func backToReminderBoxMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.Inline(markup.Row(btnBackToReminderBox))
	return markup
}

func scheduleMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnReminderDaily, btnReminderWeekly),
		markup.Row(btnReminderMonthly, btnReminderInterval),
		markup.Row(btnReminderOnce),
		markup.Row(btnBackToReminderBox),
	)
	return markup
}

func weekdayMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	order := []int{1, 2, 3, 4, 5, 6, 0} // Monday..Sunday with Sunday last
	btns := make([]telebot.Btn, 0, len(order))
	for _, idx := range order {
		title := constants.WeekdayShortRu[idx%len(constants.WeekdayShortRu)]
		btns = append(btns, markup.Data(title, btnSelectWeekday.Unique, fmt.Sprintf("%d", idx)))
	}
	markup.Inline(
		markup.Row(btns[0], btns[1], btns[2]),
		markup.Row(btns[3], btns[4], btns[5]),
		markup.Row(btns[6]),
		markup.Row(btnBackToReminderBox),
	)
	return markup
}

func selectReminderMarkup(reminders []models.Reminder) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	if len(reminders) > 0 {
		rows := make([]telebot.Row, 0, len(reminders)+1)
		for _, r := range reminders {
			btn := markup.Data(r.Name, btnSelectReminderToDelete.Unique, fmt.Sprintf("%d", r.ID))
			rows = append(rows, markup.Row(btn))
		}
		rows = append(rows, markup.Row(btnBackToReminderBox))
		markup.Inline(rows...)
		return markup
	}
	markup.Inline(markup.Row(btnBackToReminderBox))
	return markup
}

func upsertReminderLastMessage(bot *b.Bot, userID int64, sourceMsg *telebot.Message, text string, markup *telebot.ReplyMarkup) error {
	msg := sourceMsg
	if msg == nil {
		msg = bot.ReminderService.GetBotLastMsg(userID)
	}

	if msg != nil {
		edited := bot.MustEdit(msg, text, markup)
		if edited != nil {
			bot.ReminderService.SetBotLastMsg(userID, edited)
			return nil
		}
		// If edit failed (likely "message is not modified"), keep existing message and avoid sending duplicates.
		bot.ReminderService.SetBotLastMsg(userID, msg)
		return nil
	}

	sent := bot.MustSend(userID, text, markup)
	bot.ReminderService.SetBotLastMsg(userID, sent)
	return nil
}

func handleReminderInputError(bot *b.Bot, userID int64, err error) error {
	switch {
	case errors.Is(err, reminder.ErrInvalidTimeOfDay):
		return renderTimePrompt(bot, userID, nil, bot.Replies.ReminderTimeFormatError)
	case errors.Is(err, reminder.ErrInvalidSchedule):
		return renderSchedulePrompt(bot, userID, nil, bot.Replies.ReminderScheduleInvalid)
	case errors.Is(err, reminder.ErrInvalidInterval):
		return renderIntervalPrompt(bot, userID, nil, bot.Replies.ReminderIntervalInvalid)
	case errors.Is(err, reminder.ErrEmptyEntityName):
		return renderNamePrompt(bot, userID, nil, bot.Replies.ReminderNameEmpty)
	case errors.Is(err, reminder.ErrEntityNameTooLong):
		return renderNamePrompt(bot, userID, nil, bot.Replies.ReminderNameTooLong)
	case errors.Is(err, reminder.ErrReminderDuplicate):
		return renderNamePrompt(bot, userID, nil, bot.Replies.ReminderDuplicate)
	default:
		return upsertReminderLastMessage(bot, userID, nil, bot.Replies.Error, reminderBoxMarkup())
	}
}

func parseUintData(ctx telebot.Context) (uint, error) {
	raw := ctx.Data()
	if raw == "" && ctx.Callback() != nil {
		raw = ctx.Callback().Data
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty id")
	}
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return uint(id64), nil
}

func clearClosedReminderBoxMessage(bot *b.Bot, userID int64) {
	userDTO := bot.UserService.GetUser(userID)
	if userDTO == nil || userDTO.ReminderBoxClosedMsgID == 0 {
		return
	}
	bot.MustDelete(&telebot.Message{ID: userDTO.ReminderBoxClosedMsgID, Chat: &telebot.Chat{ID: userID}})
	if err := bot.UserService.UpdateReminderBoxClosedMsgID(userID, 0); err != nil {
		bot.Logger.Error(fmt.Sprintf("Error clearing closed reminder box message for userID=%d: %v", userID, err))
	}
}

func saveClosedReminderBoxMessage(bot *b.Bot, userID int64, msg *telebot.Message) {
	if msg == nil {
		return
	}
	if err := bot.UserService.UpdateReminderBoxClosedMsgID(userID, msg.ID); err != nil {
		bot.Logger.Error(fmt.Sprintf("Error saving closed reminder box message for userID=%d: %v", userID, err))
	}
}

// clampToActiveWindow moves the time to the start of the user's active window if it's outside.
func notifyAndDelete(bot *b.Bot, userID int64, text string, delay time.Duration) {
	msg := bot.MustSend(userID, text)
	if msg == nil {
		return
	}
	go func(m *telebot.Message) {
		time.Sleep(delay)
		bot.MustDelete(m)
	}(msg)
}

func handleNameStep(bot *b.Bot, userID int64, pending *session.PendingReminder, raw string) (bool, error) {
	if pending.EntityName != "" {
		return false, nil
	}

	name, err := helpers.NormalizeReminderName(raw, reminder.ErrEmptyEntityName, reminder.ErrEntityNameTooLong)
	if err != nil {
		return true, renderNamePrompt(bot, userID, nil, err.Error())
	}
	if dup, err := bot.ReminderService.IsDuplicateName(userID, name); err == nil && dup {
		return true, upsertReminderLastMessage(bot, userID, bot.ReminderService.GetBotLastMsg(userID), bot.Replies.ReminderDuplicate+"\n\n"+bot.Replies.ReminderNamePrompt, backToReminderBoxMarkup())
	} else if err != nil {
		return true, upsertReminderLastMessage(bot, userID, nil, bot.Replies.Error, reminderBoxMarkup())
	}

	pending.EntityName = name
	bot.ReminderService.SetPending(userID, pending)
	switch pending.ScheduleType {
	case models.ReminderScheduleInterval:
		return true, renderIntervalPrompt(bot, userID, nil, "")
	case models.ReminderScheduleWeekly:
		return true, renderWeekdayPrompt(bot, userID, nil, "")
	case models.ReminderScheduleMonthly:
		return true, renderMonthDayPrompt(bot, userID, nil, "")
	case models.ReminderScheduleOnce:
		return true, renderOncePrompt(bot, userID, nil, "")
	default:
		return true, renderTimePrompt(bot, userID, nil, "")
	}
}

func handleScheduleSpecificStep(bot *b.Bot, userID int64, pending *session.PendingReminder, raw string, loc *time.Location) (bool, error) {
	switch pending.ScheduleType {
	case models.ReminderScheduleInterval:
		if pending.IntervalMinutes != nil {
			return false, nil
		}
		minutes, err := strconv.Atoi(raw)
		if err != nil || minutes <= 0 {
			return true, renderIntervalPrompt(bot, userID, nil, bot.Replies.ReminderIntervalInvalid)
		}
		val := int32(minutes)
		pending.IntervalMinutes = &val
		bot.ReminderService.SetPending(userID, pending)
	case models.ReminderScheduleMonthly:
		if pending.MonthDay != nil {
			return false, nil
		}
		dayVal, err := strconv.Atoi(raw)
		if err != nil || dayVal < 1 || dayVal > 31 {
			return true, renderMonthDayPrompt(bot, userID, nil, bot.Replies.ReminderMonthDayInvalid)
		}
		day := int8(dayVal)
		pending.MonthDay = &day
		bot.ReminderService.SetPending(userID, pending)
		return true, renderTimePrompt(bot, userID, nil, "")
	case models.ReminderScheduleWeekly:
		if pending.Weekday != nil {
			return false, nil
		}
		dayVal, err := strconv.Atoi(raw)
		if err != nil || dayVal < 0 || dayVal > 6 {
			return true, renderWeekdayPrompt(bot, userID, nil, bot.Replies.ReminderWeekdayInvalid)
		}
		day := int8(dayVal)
		pending.Weekday = &day
		bot.ReminderService.SetPending(userID, pending)
		return true, renderTimePrompt(bot, userID, nil, "")
	case models.ReminderScheduleOnce:
		if pending.OnceDate != nil {
			return false, nil
		}
		now := time.Now().In(loc)
		date, err := helpers.ParseDateDM(raw, now, loc)
		if err != nil {
			return true, renderOncePrompt(bot, userID, nil, bot.Replies.ReminderOnceDateInvalid)
		}
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		dateMidnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
		if dateMidnight.Before(today) {
			return true, renderOncePrompt(bot, userID, nil, bot.Replies.ReminderOnceDatePast)
		}
		pending.OnceDate = &date
		bot.ReminderService.SetPending(userID, pending)
		return true, renderTimePrompt(bot, userID, nil, "")
	}

	return false, nil
}

func handleTimeStep(bot *b.Bot, userID int64, pending *session.PendingReminder, raw string, loc *time.Location) (bool, error) {
	if pending.ScheduleType == models.ReminderScheduleInterval || pending.TimeOfDayMinutes != nil {
		return false, nil
	}

	minutes, err := helpers.ParseTimeHM(raw)
	if err != nil {
		text := bot.Replies.ReminderTimeFormatError + "\n\n" + bot.Replies.ReminderTimePrompt
		return true, upsertReminderLastMessage(bot, userID, bot.ReminderService.GetBotLastMsg(userID), text, backToReminderBoxMarkup())
	}

	user := bot.UserService.GetUser(userID)
	adjusted, clamped := reminder.ClampMinutesToWindow(minutes, int(user.DayStart), int(user.DayEnd))
	min := int16(adjusted)
	pending.TimeOfDayMinutes = &min
	bot.ReminderService.SetPending(userID, pending)

	if clamped {
		note := fmt.Sprintf("Вне окна %s–%s, поставил на %s",
			helpers.FormatTimeHM(int(user.DayStart)),
			helpers.FormatTimeHM(int(user.DayEnd)),
			helpers.FormatTimeHM(adjusted))
		return true, finalizeReminder(bot, userID, pending, note, loc)
	}

	return false, nil
}
func renderSchedulePrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderSchedulePrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, scheduleMarkup())
}

func renderNamePrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderNamePrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, backToReminderBoxMarkup())
}

func renderWeekdayPrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderWeekdayPrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, weekdayMarkup())
}

func renderMonthDayPrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderMonthDayPrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, backToReminderBoxMarkup())
}

func renderIntervalPrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderIntervalPrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, backToReminderBoxMarkup())
}

func renderOncePrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderOnceDatePrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, backToReminderBoxMarkup())
}

func renderTimePrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.ReminderTimePrompt
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertReminderLastMessage(bot, userID, sourceMsg, text, backToReminderBoxMarkup())
}
