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
}
type Store struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
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
