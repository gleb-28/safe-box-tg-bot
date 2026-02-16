package reminder

import (
	"context"
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/feat/prompt"
	"safeboxtgbot/internal/feat/user"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/models"
	"safeboxtgbot/pkg/utils"
	"time"

	"gopkg.in/telebot.v4"
)

type Worker struct {
	reminderService  *Service
	userService      *user.Service
	messageGenerator prompt.MessageGenerator
	bot              *telebot.Bot
	logger           logger.AppLogger
}

func NewWorker(reminderService *Service, userService *user.Service, messageGenerator prompt.MessageGenerator, bot *telebot.Bot, logger logger.AppLogger) *Worker {
	return &Worker{reminderService: reminderService, userService: userService, messageGenerator: messageGenerator, bot: bot, logger: logger}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(constants.ReminderWorkerIntervalSeconds) * time.Second)
	defer ticker.Stop()

	if ctx.Err() == nil {
		w.processSafe()
	}

	for {
		select {
		case <-ticker.C:
			w.processSafe()
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) processSafe() {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error(fmt.Sprintf("reminder worker panic: %v", r))
		}
	}()

	w.process()
}

func (w *Worker) process() {
	now := time.Now().UTC()
	due, err := w.reminderService.GetDue(now)
	if err != nil {
		w.logger.Error(fmt.Sprintf("get due reminders: %v", err))
		return
	}

	for _, r := range due {
		w.handle(now, r)
	}
}

func (w *Worker) handle(nowUTC time.Time, r models.Reminder) {
	userDTO := w.userService.GetUser(r.UserID)
	if userDTO == nil || userDTO.TelegramID == 0 {
		_ = w.reminderService.Disable(r.ID, r.UserID)
		return
	}

	loc := w.userLocation(*userDTO)
	localNow := nowUTC.In(loc)

	if userDTO.NotificationsMuted {
		retry := nowUTC.Add(time.Duration(constants.NotificationRetryMinutes) * time.Minute)
		_ = w.reminderService.SetNextRun(r.ID, r.UserID, retry)
		return
	}

	if !helpers.IsWithinActiveWindow(*userDTO, localNow) {
		next := helpers.NextStartTimeFromLocal(*userDTO, localNow, loc)
		_ = w.reminderService.SetNextRun(r.ID, r.UserID, next)
		return
	}

	text := helpers.FallbackText(r.Name, constants.FallbackEmojis)
	if w.messageGenerator != nil {
		if generated, err := w.generateText(localNow, *userDTO, r.Name); err == nil && generated != "" {
			text = generated
		}
	}

	if err := w.send(userDTO.TelegramID, text); err != nil {
		retry := nowUTC.Add(time.Duration(constants.NotificationRetryMinutes) * time.Minute)
		_ = w.reminderService.SetNextRun(r.ID, r.UserID, retry)
		return
	}

	if r.Schedule == models.ReminderScheduleOnce {
		if err := w.reminderService.Delete(r.ID, r.UserID); err != nil {
			w.logger.Error(fmt.Sprintf("delete once reminder %d: %v", r.ID, err))
		}
		return
	}

	if err := w.reminderService.Reschedule(&r, nowUTC, loc); err != nil {
		w.logger.Error(fmt.Sprintf("reschedule reminder %d: %v", r.ID, err))
	}
}

func (w *Worker) send(userID int64, text string) error {
	_, err := w.bot.Send(&telebot.User{ID: userID}, text)
	if err != nil {
		w.logger.Error(fmt.Sprintf("send reminder to userID=%d: %v", userID, err))
	}
	return err
}

func (w *Worker) generateText(localNow time.Time, user models.User, entityName string) (string, error) {
	input := prompt.LLMInput{
		CurrentEntity: entityName,
		TimeOfDay:     helpers.TimeOfDay(localNow),
		StyleMode:     helpers.ModeToStyle(user.Mode),
		RandomSeed:    utils.RandomIntRange(1, 1_000_000),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()
	return w.messageGenerator.Generate(ctx, input)
}

func (w *Worker) userLocation(user models.User) *time.Location {
	loc, err := helpers.UserLocation(user)
	if err != nil {
		w.logger.Error(fmt.Sprintf("invalid timezone %q for userID=%d: %v", user.Timezone, user.TelegramID, err))
		return time.UTC
	}
	return loc
}
