package session

import (
	"context"
	"fmt"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/models"
	"sync"
	"time"

	"gopkg.in/telebot.v4"
)

type Session struct {
	User               *models.User
	UserIsLoaded       bool
	UserLastMsg        *telebot.Message
	BotLastMsg         *telebot.Message
	ReminderBotLastMsg *telebot.Message
	Authorized         bool
	Items              ItemsState
	Daytime            DaytimeState
	Reminders          RemindersState
	ExpiresAt          time.Time
}

type ItemsState struct {
	EditingItemID uint
	ItemsLoaded   bool
	ItemList      []models.Item
}

type DaytimeState struct {
	StartMinutes int
}

type RemindersState struct {
	Pending *PendingReminder
	Loaded  bool
	List    []models.Reminder
}

type PendingReminder struct {
	EntityName       string
	ScheduleType     models.ReminderSchedule
	IntervalMinutes  *int32
	TimeOfDayMinutes *int16
	Weekday          *int8
	MonthDay         *int8
	OnceDate         *time.Time
}
type Store struct {
	sessions map[int64]*Session
	mu       sync.RWMutex
	logger   logger.AppLogger
	ttl      time.Duration
}

func NewStore(ttl time.Duration, logger logger.AppLogger) *Store {
	return &Store{
		sessions: make(map[int64]*Session),
		logger:   logger,
		ttl:      ttl,
	}
}

func (store *Store) Get(userID int64) *Session {
	store.mu.RLock()
	session, ok := store.sessions[userID]
	store.mu.RUnlock()

	if ok {
		store.mu.Lock()
		store.ensureExpiryLocked(session)
		store.mu.Unlock()
		store.logger.Debug(fmt.Sprintf("Session cache hit for userID=%d", userID))
		return session
	}

	store.logger.Debug(fmt.Sprintf("Session cache miss for userID=%d", userID))
	store.mu.Lock()
	defer store.mu.Unlock()

	session = &Session{
		User: &models.User{},
		Items: ItemsState{
			ItemList: make([]models.Item, 0),
		},
		Daytime: DaytimeState{
			StartMinutes: -1,
		},
		Reminders: RemindersState{},
	}
	store.ensureExpiryLocked(session)
	store.sessions[userID] = session
	store.logger.Info(fmt.Sprintf("New session created for userID=%d", userID))

	return session
}

func (store *Store) Update(userID int64, updater func(s *Session)) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if session, ok := store.sessions[userID]; ok {
		updater(session)
	}
}

func (store *Store) IsUserLoaded(userID int64) bool {
	return store.Get(userID).UserIsLoaded
}

func (store *Store) GetUser(userID int64) *models.User {
	session := store.Get(userID)
	store.mu.RLock()
	defer store.mu.RUnlock()
	return session.User
}

func (store *Store) UpdateUser(userID int64, user *models.User) {
	store.Update(userID, func(s *Session) {
		s.User = user
		s.UserIsLoaded = true
	})
	store.logger.Debug(fmt.Sprintf("Session user updated for userID=%d", userID))
}

func (store *Store) SetBotLastMsg(userID int64, msg *telebot.Message) {
	store.Update(userID, func(sess *Session) {
		sess.BotLastMsg = msg
	})
}

func (store *Store) SetReminderBotLastMsg(userID int64, msg *telebot.Message) {
	store.Update(userID, func(sess *Session) {
		sess.ReminderBotLastMsg = msg
	})
}

func (store *Store) GetReminderBotLastMsg(userID int64) *telebot.Message {
	return store.Get(userID).ReminderBotLastMsg
}

func (store *Store) GetBotLastMsg(userID int64) *telebot.Message {
	return store.Get(userID).BotLastMsg
}

func (store *Store) IsItemsLoaded(userID int64) bool {
	return store.Get(userID).Items.ItemsLoaded
}

func (store *Store) GetItemList(userID int64) []models.Item {
	session := store.Get(userID)
	store.mu.RLock()
	defer store.mu.RUnlock()
	return session.Items.ItemList
}

func (store *Store) SetItemList(userID int64, items []models.Item) {
	store.Update(userID, func(sess *Session) {
		sess.Items.ItemList = items
		sess.Items.ItemsLoaded = true
	})
}

func (store *Store) GetEditingItemID(userID int64) uint {
	return store.Get(userID).Items.EditingItemID
}

func (store *Store) SetEditingItemID(userID int64, itemID uint) {
	store.Update(userID, func(sess *Session) {
		sess.Items.EditingItemID = itemID
	})
}

func (store *Store) ClearEditingItemID(userID int64) {
	store.Update(userID, func(sess *Session) {
		sess.Items.EditingItemID = 0
	})
}

func (store *Store) GetDayStartSelection(userID int64) int {
	return store.Get(userID).Daytime.StartMinutes
}

func (store *Store) SetDayStartSelection(userID int64, minutes int) {
	store.Update(userID, func(sess *Session) {
		sess.Daytime.StartMinutes = minutes
	})
}

func (store *Store) ClearDayStartSelection(userID int64) {
	store.Update(userID, func(sess *Session) {
		sess.Daytime.StartMinutes = -1
	})
}

func (store *Store) GetPendingReminder(userID int64) *PendingReminder {
	return store.Get(userID).Reminders.Pending
}

func (store *Store) SetPendingReminder(userID int64, pending *PendingReminder) {
	store.Update(userID, func(sess *Session) {
		sess.Reminders.Pending = pending
	})
}

func (store *Store) ClearPendingReminder(userID int64) {
	store.Update(userID, func(sess *Session) {
		sess.Reminders.Pending = nil
	})
}

func (store *Store) IsRemindersLoaded(userID int64) bool {
	return store.Get(userID).Reminders.Loaded
}

func (store *Store) GetReminderList(userID int64) []models.Reminder {
	return store.Get(userID).Reminders.List
}

func (store *Store) SetReminderList(userID int64, reminders []models.Reminder) {
	store.Update(userID, func(sess *Session) {
		sess.Reminders.List = reminders
		sess.Reminders.Loaded = true
	})
}

func (store *Store) ClearReminders(userID int64) {
	store.Update(userID, func(sess *Session) {
		sess.Reminders.List = nil
		sess.Reminders.Loaded = false
		sess.Reminders.Pending = nil
	})
}

// Delete removes a session explicitly.
func (store *Store) Delete(userID int64) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.deleteLocked(userID)
}

// MarkAuthorized stops expiration for an authenticated user session.
func (store *Store) MarkAuthorized(userID int64) {
	session := store.Get(userID)

	store.mu.Lock()
	defer store.mu.Unlock()

	session.Authorized = true
	session.ExpiresAt = time.Time{}
}

// StartCleanupWorker launches a ticker-based cleanup loop.
func (store *Store) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				store.cleanupExpired()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (store *Store) cleanupExpired() {
	now := time.Now()

	store.mu.Lock()
	defer store.mu.Unlock()

	for userID, session := range store.sessions {
		if session.Authorized || session.ExpiresAt.IsZero() {
			continue
		}

		if now.After(session.ExpiresAt) {
			store.logger.Debug(fmt.Sprintf("Session expired for userID=%d", userID))
			store.deleteLocked(userID)
		}
	}
}

func (store *Store) deleteLocked(userID int64) {
	delete(store.sessions, userID)
}

func (store *Store) ensureExpiryLocked(session *Session) {
	if session.Authorized || store.ttl <= 0 {
		session.ExpiresAt = time.Time{}
		return
	}

	session.ExpiresAt = time.Now().Add(store.ttl)
}
