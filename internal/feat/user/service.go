package user

import (
	"fmt"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/models"
	"safeboxtgbot/pkg/utils"
	"time"
)

type Service struct {
	store          *session.Store
	userRepo       *repo.UserRepo
	itemRepo       *repo.ItemRepo
	messageLogRepo *repo.MessageLogRepo
	logger         logger.AppLogger
}

func NewUserService(
	userRepo *repo.UserRepo,
	itemRepo *repo.ItemRepo,
	messageLogRepo *repo.MessageLogRepo,
	store *session.Store,
	logger logger.AppLogger,
) *Service {
	return &Service{
		store:          store,
		userRepo:       userRepo,
		itemRepo:       itemRepo,
		messageLogRepo: messageLogRepo,
		logger:         logger,
	}
}

func (s *Service) GetUser(userID int64) *models.User {
	s.ensureUserSessionLoaded(userID)
	return s.store.Get(userID).User
}

func (s *Service) AddUser(userID int64) error {
	s.ensureUserSessionLoaded(userID)
	nextNotification := s.getNextRandNotification()
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.TelegramID = userID
		sess.User.NextNotification = nextNotification
	})
	return s.userRepo.Upsert(&models.User{TelegramID: userID, NextNotification: nextNotification})
}

func (s *Service) UpdateMode(userID int64, mode models.UserMode) error {
	s.ensureUserSessionLoaded(userID)
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.Mode = mode
	})
	return s.userRepo.Upsert(&models.User{TelegramID: userID, Mode: mode})
}

func (s *Service) UpdateNextNotification(userID int64, t time.Time) {
	s.ensureUserSessionLoaded(userID)
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.NextNotification = t
	})
}

func (s *Service) UpdateItems(userID int64, items []models.Item) error {
	s.ensureUserSessionLoaded(userID)
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.Items = items
	})

	for _, item := range items {
		item.UserID = userID
		if err := s.itemRepo.Upsert(&item); err != nil {
			s.logger.Error(fmt.Sprintf("Error updating item %v: %v", item.Name, err))
		}
	}
	return nil
}

func (s *Service) LogMessage(userID int64, itemID uint, text string) error {
	s.ensureUserSessionLoaded(userID)
	log := &models.MessageLog{
		UserID: userID,
		ItemID: itemID,
		SentAt: time.Now(),
		Text:   text,
	}
	return s.messageLogRepo.Create(log)
}

func (s *Service) ensureUserSessionLoaded(userID int64) {
	if s.store.IsUserLoaded(userID) {
		return
	}

	userDTO, err := s.userRepo.GetByTelegramID(userID)
	if err == nil {
		s.store.UpdateUser(userID, &userDTO)
		return
	}
}

func (s *Service) getNextRandNotification() time.Time {
	return time.Now().Add(utils.RandomDuration(1, 3))
}
