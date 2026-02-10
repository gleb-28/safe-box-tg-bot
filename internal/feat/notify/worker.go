package notify

import (
	"context"
	"fmt"
	"safeboxtgbot/internal/helpers"
	"strings"
	"time"

	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/feat/items"
	"safeboxtgbot/internal/feat/prompt"
	"safeboxtgbot/internal/feat/user"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/models"
	"safeboxtgbot/pkg/utils"

	"gopkg.in/telebot.v4"
)

type Worker struct {
	userService      *user.Service
	itemsService     *items.Service
	messageLogRepo   *repo.MessageLogRepo
	messageGenerator prompt.MessageGenerator
	bot              *b.Bot
	logger           logger.AppLogger
}

func NewWorker(
	userService *user.Service,
	itemsService *items.Service,
	messageLogRepo *repo.MessageLogRepo,
	messageGenerator prompt.MessageGenerator,
	bot *b.Bot,
	logger logger.AppLogger,
) *Worker {
	return &Worker{
		userService:      userService,
		itemsService:     itemsService,
		messageLogRepo:   messageLogRepo,
		messageGenerator: messageGenerator,
		bot:              bot,
		logger:           logger,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(constants.NotificationCheckIntervalMinutes) * time.Minute)
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
		if recovered := recover(); recovered != nil {
			w.logger.Error(fmt.Sprintf("Notification worker panic: %v", recovered))
		}
	}()

	w.process()
}

func (w *Worker) process() {
	nowUTC := time.Now().UTC()
	users, err := w.userService.GetUsersForNotification(nowUTC)
	if err != nil {
		w.logger.Error(fmt.Sprintf("Error fetching users for notification: %v", err))
		return
	}
	w.logger.Debug(fmt.Sprintf("Notification tick: %d users due at %s", len(users), nowUTC.Format(time.RFC3339)))

	for _, userDTO := range users {
		w.processUser(nowUTC, userDTO)
	}
}

func (w *Worker) processUser(nowUTC time.Time, user models.User) {
	if user.TelegramID == 0 {
		return
	}

	if user.NotificationsMuted {
		return
	}

	if w.isOverdue(user, nowUTC) {
		next := w.nextNotificationTime(user, nowUTC)
		w.logger.Debug(fmt.Sprintf("UserID=%d overdue; reschedule to %s", user.TelegramID, next.Format(time.RFC3339)))
		w.updateNextNotification(user, next)
		return
	}

	if !w.isInActiveHours(user, nowUTC) {
		next := w.nextStartTime(user, nowUTC)
		w.logger.Debug(fmt.Sprintf("UserID=%d outside active hours; next start %s", user.TelegramID, next.Format(time.RFC3339)))
		w.updateNextNotification(user, next)
		return
	}

	itemList, err := w.itemsService.GetItemList(user.TelegramID)
	if err != nil {
		w.logger.Error(fmt.Sprintf("Error fetching itemList for userID=%d: %v", user.TelegramID, err))
		w.updateNextNotification(user, w.retryAt(nowUTC))
		return
	}
	if len(itemList) == 0 {
		next := w.nextNotificationTime(user, nowUTC)
		w.logger.Debug(fmt.Sprintf("UserID=%d has no itemList; reschedule to %s", user.TelegramID, next.Format(time.RFC3339)))
		w.updateNextNotification(user, next)
		return
	}

	item := w.pickItem(user, itemList, nowUTC)
	if item == nil {
		next := w.nextNotificationTime(user, nowUTC)
		w.logger.Debug(fmt.Sprintf("UserID=%d no eligible itemList; reschedule to %s", user.TelegramID, next.Format(time.RFC3339)))
		w.updateNextNotification(user, next)
		return
	}

	text := helpers.FallbackText(item.Name, constants.FallbackEmojis)
	if w.messageGenerator != nil {
		generated, err := w.generateText(nowUTC, user, *item)
		if err != nil {
			w.logger.Error(fmt.Sprintf("Error generating message for userID=%d: %v", user.TelegramID, err))
		} else if strings.TrimSpace(generated) != "" {
			text = generated
		}
	}
	w.logger.Debug(fmt.Sprintf("UserID=%d selected itemID=%d name=%q", user.TelegramID, item.ID, item.Name))
	if err := w.send(user.TelegramID, text); err != nil {
		w.updateNextNotification(user, w.retryAt(nowUTC))
		return
	}
	w.logger.Info(fmt.Sprintf("Notification sent userID=%d itemID=%d name=%q text=%q", user.TelegramID, item.ID, item.Name, text))

	if err := w.messageLogRepo.Create(&models.MessageLog{
		UserID: user.TelegramID,
		ItemID: item.ID,
		SentAt: nowUTC,
		Text:   text,
	}); err != nil {
		w.logger.Error(fmt.Sprintf("Error logging message for userID=%d: %v", user.TelegramID, err))
	}

	w.updateNextNotification(user, w.nextNotificationTime(user, nowUTC))
}

func (w *Worker) pickItem(user models.User, items []models.Item, nowUTC time.Time) *models.Item {
	since := nowUTC.Add(-time.Duration(constants.NotificationItemCooldownMinutes) * time.Minute)
	recentIDs, err := w.messageLogRepo.GetRecentItemIDs(user.TelegramID, since)
	if err != nil {
		w.logger.Error(fmt.Sprintf("Error fetching recent items for userID=%d: %v", user.TelegramID, err))
		return w.randomItem(items)
	}

	recent := make(map[uint]struct{}, len(recentIDs))
	for _, id := range recentIDs {
		recent[id] = struct{}{}
	}

	candidates := make([]models.Item, 0, len(items))
	for _, item := range items {
		if _, ok := recent[item.ID]; !ok {
			candidates = append(candidates, item)
		}
	}
	if len(candidates) == 0 {
		candidates = items
	}

	if len(candidates) == 0 {
		return nil
	}

	selected := candidates[utils.RandomIndex(len(candidates))]
	return &selected
}

func (w *Worker) randomItem(items []models.Item) *models.Item {
	if len(items) == 0 {
		return nil
	}
	selected := items[utils.RandomIndex(len(items))]
	return &selected
}

func (w *Worker) send(userID int64, text string) error {
	_, err := w.bot.Send(&telebot.User{ID: userID}, text)
	if err != nil {
		w.logger.Error(fmt.Sprintf("Error sending notification to userID=%d: %v", userID, err))
	}
	return err
}

func (w *Worker) updateNextNotification(user models.User, next time.Time) {
	w.logger.Debug(fmt.Sprintf("UserID=%d NextNotification -> %s", user.TelegramID, next.Format(time.RFC3339)))
	if err := w.userService.UpdateNextNotification(user.TelegramID, next); err != nil {
		w.logger.Error(fmt.Sprintf("Error updating next notification for userID=%d: %v", user.TelegramID, err))
	}
}

func (w *Worker) generateText(nowUTC time.Time, user models.User, item models.Item) (string, error) {
	loc := w.userLocation(user)
	localNow := nowUTC.In(loc)
	input := prompt.LLMInput{
		CurrentEntity: item.Name,
		TimeOfDay:     helpers.TimeOfDay(localNow),
		StyleMode:     helpers.ModeToStyle(user.Mode),
		RandomSeed:    utils.RandomIntRange(1, 1_000_000),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()
	return w.messageGenerator.Generate(ctx, input)
}

func (w *Worker) retryAt(nowUTC time.Time) time.Time {
	return nowUTC.Add(time.Duration(constants.NotificationRetryMinutes) * time.Minute)
}

func (w *Worker) nextNotificationTime(user models.User, nowUTC time.Time) time.Time {
	loc := w.userLocation(user)
	return helpers.NextNotificationTimeWithLoc(user, nowUTC, loc)
}

func (w *Worker) nextStartTime(user models.User, nowUTC time.Time) time.Time {
	loc := w.userLocation(user)
	return helpers.NextStartTimeFromLocal(user, nowUTC.In(loc), loc)
}

func (w *Worker) nextStartTimeFromLocal(user models.User, fromLocal time.Time, loc *time.Location) time.Time {
	return helpers.NextStartTimeFromLocal(user, fromLocal, loc)
}

func (w *Worker) isInActiveHours(user models.User, nowUTC time.Time) bool {
	loc := w.userLocation(user)
	return helpers.IsWithinActiveWindow(user, nowUTC.In(loc))
}

func (w *Worker) isOverdue(user models.User, nowUTC time.Time) bool {
	if user.NextNotification.IsZero() {
		return true
	}
	overdueBy := nowUTC.Sub(user.NextNotification)
	if overdueBy <= 0 {
		return false
	}
	_, maxMinutes := w.notificationRange(user)
	maxOverdue := time.Duration(maxMinutes+constants.NotificationCheckIntervalMinutes) * time.Minute
	isOverdue := overdueBy > maxOverdue
	w.logger.Debug(fmt.Sprintf("UserID=%d overdueBy=%s maxOverdue=%s isOverdue=%t", user.TelegramID, overdueBy, maxOverdue, isOverdue))
	return isOverdue
}

func (w *Worker) userLocation(user models.User) *time.Location {
	loc, err := helpers.UserLocation(user)
	if err != nil {
		w.logger.Error(fmt.Sprintf("Invalid timezone %q for userID=%d: %v", user.Timezone, user.TelegramID, err))
		return time.UTC
	}
	return loc
}

func (w *Worker) notificationRange(user models.User) (int, int) {
	minMinutes, maxMinutes := helpers.UserNotificationRange(user)
	if minMinutes <= 0 || maxMinutes <= 0 || maxMinutes < minMinutes {
		return constants.DefaultNotificationIntervalMinMinutes, constants.DefaultNotificationIntervalMaxMinutes
	}
	return minMinutes, maxMinutes
}
