package user

import (
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/models"
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

func (s *Service) GetUsersForNotification(now time.Time) ([]models.User, error) {
	return s.userRepo.GetUsersForNotification(now)
}

func (s *Service) GetUser(userID int64) *models.User {
	s.ensureUserSessionLoaded(userID)
	return s.store.Get(userID).User
}

func (s *Service) AddUser(userID int64) error {
	s.ensureUserSessionLoaded(userID)
	preset := s.defaultPreset()
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.TelegramID = userID
		sess.User.NotificationPreset = preset.Key
		sess.User.NotificationIntervalMinMinutes = int16(preset.MinMinutes)
		sess.User.NotificationIntervalMaxMinutes = int16(preset.MaxMinutes)
	})
	nextNotification := helpers.NextNotificationTime(*s.store.GetUser(userID), time.Now().UTC())
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.NextNotification = nextNotification
	})
	return s.userRepo.Upsert(&models.User{
		TelegramID:                     userID,
		NotificationPreset:             preset.Key,
		NotificationIntervalMinMinutes: int16(preset.MinMinutes),
		NotificationIntervalMaxMinutes: int16(preset.MaxMinutes),
		NextNotification:               nextNotification,
	})
}

func (s *Service) UpdateMode(userID int64, mode models.UserMode) error {
	s.ensureUserSessionLoaded(userID)
	current := s.store.GetUser(userID).Mode
	if current == mode {
		return nil
	}
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.Mode = mode
	})
	return s.userRepo.UpdateMode(userID, mode)
}

func (s *Service) UpdateNextNotification(userID int64, t time.Time) error {
	s.ensureUserSessionLoaded(userID)
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.NextNotification = t
	})
	return s.userRepo.UpdateNextNotification(userID, t)
}

func (s *Service) SetNotificationsMuted(userID int64, muted bool) error {
	s.ensureUserSessionLoaded(userID)
	var next time.Time
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.NotificationsMuted = muted
		if muted {
			next = time.Time{}
			sess.User.NextNotification = time.Time{}
			return
		}
		next = helpers.NextNotificationTime(*sess.User, time.Now().UTC())
		sess.User.NextNotification = next
	})

	return s.userRepo.UpdateNotificationsMuted(userID, muted, next)
}

func (s *Service) UpdateNotificationPreset(userID int64, preset constants.NotificationPreset) error {
	s.ensureUserSessionLoaded(userID)
	sanitized := s.normalizePreset(preset)
	var nextNotification time.Time
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.NotificationPreset = sanitized.Key
		sess.User.NotificationIntervalMinMinutes = int16(sanitized.MinMinutes)
		sess.User.NotificationIntervalMaxMinutes = int16(sanitized.MaxMinutes)
		nextNotification = helpers.NextNotificationTime(*sess.User, time.Now().UTC())
		sess.User.NextNotification = nextNotification
	})

	return s.userRepo.UpdateNotificationInterval(
		userID,
		sanitized.Key,
		int16(sanitized.MinMinutes),
		int16(sanitized.MaxMinutes),
		nextNotification,
	)
}

func (s *Service) UpdateItemBoxClosedMsgID(userID int64, msgID int) error {
	s.ensureUserSessionLoaded(userID)
	s.store.Update(userID, func(sess *session.Session) {
		sess.User.ItemBoxClosedMsgID = msgID
	})
	return s.userRepo.UpdateItemBoxClosedMsgID(userID, msgID)
}

func (s *Service) UpdateItems(userID int64, items []models.Item) error {
	s.ensureUserSessionLoaded(userID)
	s.store.SetItemList(userID, items)

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
		s.logger.Debug(fmt.Sprintf("Session already loaded for userID=%d", userID))
		return
	}

	s.logger.Debug(fmt.Sprintf("Loading user into session for userID=%d", userID))
	userDTO, err := s.userRepo.GetByTelegramID(userID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error loading user from DB for userID=%d: %v", userID, err))
		return
	}

	if userDTO.TelegramID != 0 {
		s.store.UpdateUser(userID, &userDTO)
		s.logger.Debug(fmt.Sprintf("User loaded into session for userID=%d", userID))
		return
	}

	s.store.Update(userID, func(sess *session.Session) {
		sess.UserIsLoaded = true
	})
	s.logger.Debug(fmt.Sprintf("User not found in DB for userID=%d", userID))
}

func (s *Service) defaultPreset() constants.NotificationPreset {
	if preset, ok := constants.NotificationPresets[constants.DefaultNotificationPreset]; ok {
		return preset
	}
	return constants.NotificationPreset{
		Key:        constants.DefaultNotificationPreset,
		Name:       "Иногда",
		MinMinutes: constants.DefaultNotificationIntervalMinMinutes,
		MaxMinutes: constants.DefaultNotificationIntervalMaxMinutes,
	}
}

func (s *Service) normalizePreset(preset constants.NotificationPreset) constants.NotificationPreset {
	if preset.MinMinutes <= 0 || preset.MaxMinutes <= 0 || preset.MaxMinutes < preset.MinMinutes {
		return s.defaultPreset()
	}
	if preset.Key == "" {
		preset.Key = constants.DefaultNotificationPreset
	}
	return preset
}
