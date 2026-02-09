package session

import (
	"fmt"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/models"
	"sync"

	"gopkg.in/telebot.v4"
)

type Session struct {
	User         *models.User
	UserIsLoaded bool
	UserLastMsg  *telebot.Message
	BotLastMsg   *telebot.Message
	Items        ItemsState
	Daytime      DaytimeState
}

type ItemsState struct {
	EditingItemName string
	ItemsLoaded     bool
	ItemList        []models.Item
}

type DaytimeState struct {
	StartMinutes int
}
type Store struct {
	sessions map[int64]*Session
	mu       sync.RWMutex
	logger   logger.AppLogger
}

func NewStore(logger logger.AppLogger) *Store {
	return &Store{
		sessions: make(map[int64]*Session),
		logger:   logger,
	}
}

func (store *Store) Get(userID int64) *Session {
	store.mu.RLock()
	session, ok := store.sessions[userID]
	store.mu.RUnlock()

	if ok {
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
	}
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

func (store *Store) GetEditingItemName(userID int64) string {
	return store.Get(userID).Items.EditingItemName
}

func (store *Store) SetEditingItemName(userID int64, itemName string) {
	store.Update(userID, func(sess *Session) {
		sess.Items.EditingItemName = itemName
	})
}

func (store *Store) ClearEditingItemName(userID int64) {
	store.Update(userID, func(sess *Session) {
		sess.Items.EditingItemName = ""
	})
}

func (store *Store) SetDayStartSelection(userID int64, minutes int) {
	store.Update(userID, func(sess *Session) {
		sess.Daytime.StartMinutes = minutes
	})
}

func (store *Store) GetDayStartSelection(userID int64) int {
	return store.Get(userID).Daytime.StartMinutes
}

func (store *Store) ClearDayStartSelection(userID int64) {
	store.Update(userID, func(sess *Session) {
		sess.Daytime.StartMinutes = -1
	})
}
