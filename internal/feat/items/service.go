package items

import (
	"errors"
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"strings"
	"unicode/utf8"

	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/models"

	"gopkg.in/telebot.v4"
)

var (
	ErrItemNameEmpty    = errors.New("item name empty")
	ErrItemNameTooLong  = errors.New("item name too long")
	ErrItemDuplicate    = errors.New("item duplicate")
	ErrItemLimitReached = errors.New("item limit reached")
	ErrItemNotFound     = errors.New("item not found")
)

type Service struct {
	store    *session.Store
	itemRepo *repo.ItemRepo
	logger   logger.AppLogger
}

func NewService(
	itemRepo *repo.ItemRepo,
	store *session.Store,
	logger logger.AppLogger,
) *Service {
	return &Service{
		store:    store,
		itemRepo: itemRepo,
		logger:   logger,
	}
}

func (s *Service) GetItemList(userID int64) ([]models.Item, error) {
	s.ensureItemsSessionLoaded(userID)
	return s.store.GetItemList(userID), nil
}

func (s *Service) CreateItem(userID int64, rawName string) error {
	s.ensureItemsSessionLoaded(userID)
	name, err := s.normalizeItemName(rawName)
	if err != nil {
		return err
	}

	items := s.store.GetItemList(userID)
	if len(items) >= constants.MaxItemsPerUser {
		return ErrItemLimitReached
	}
	for _, item := range items {
		if item.Name == name {
			return ErrItemDuplicate
		}
	}

	if err := s.itemRepo.Upsert(&models.Item{UserID: userID, Name: name}); err != nil {
		return err
	}
	err = s.refreshItems(userID)
	return err
}

func (s *Service) UpdateItemName(userID int64, itemName string, rawName string) error {
	s.ensureItemsSessionLoaded(userID)
	name, err := s.normalizeItemName(rawName)
	if err != nil {
		return err
	}
	if name == itemName {
		return nil
	}

	items := s.store.GetItemList(userID)
	found := false
	for _, item := range items {
		if item.Name == itemName {
			found = true
			continue
		}
		if item.Name == name {
			return ErrItemDuplicate
		}
	}
	if !found {
		return ErrItemNotFound
	}

	updated, err := s.itemRepo.UpdateName(userID, itemName, name)
	if err != nil {
		return err
	}
	if !updated {
		return ErrItemNotFound
	}
	return s.refreshItems(userID)
}

func (s *Service) DeleteItem(userID int64, itemName string) error {
	s.ensureItemsSessionLoaded(userID)
	items := s.store.GetItemList(userID)
	found := false
	for _, item := range items {
		if item.Name == itemName {
			found = true
			break
		}
	}
	if !found {
		return ErrItemNotFound
	}

	deleted, err := s.itemRepo.DeleteByName(userID, itemName)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrItemNotFound
	}
	err = s.refreshItems(userID)
	return err
}

func (s *Service) SetBotLastMsg(userID int64, msg *telebot.Message) {
	s.store.SetBotLastMsg(userID, msg)
}

func (s *Service) GetBotLastMsg(userID int64) *telebot.Message {
	return s.store.GetBotLastMsg(userID)
}

func (s *Service) GetEditingItemName(userID int64) string {
	return s.store.GetEditingItemName(userID)
}

func (s *Service) SetEditingItemName(userID int64, itemName string) {
	s.store.SetEditingItemName(userID, itemName)
}

func (s *Service) ClearEditingItemName(userID int64) {
	s.store.ClearEditingItemName(userID)
}

func (s *Service) ensureItemsSessionLoaded(userID int64) {
	if s.store.IsItemsLoaded(userID) {
		s.logger.Debug(fmt.Sprintf("Items session already loaded for userID=%d", userID))
		return
	}

	s.logger.Debug(fmt.Sprintf("Loading items into session for userID=%d", userID))
	items, err := s.itemRepo.GetByTelegramID(userID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error loading items from DB for userID=%d: %v", userID, err))
		return
	}

	s.store.SetItemList(userID, items)
	s.logger.Debug(fmt.Sprintf("Items loaded into session for userID=%d", userID))
}

func (s *Service) refreshItems(userID int64) error {
	items, err := s.itemRepo.GetByTelegramID(userID)
	if err != nil {
		return err
	}
	s.store.SetItemList(userID, items)
	return nil
}

func (s *Service) normalizeItemName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrItemNameEmpty
	}
	normalized := strings.ToLower(strings.Join(strings.Fields(trimmed), " "))
	if normalized == "" {
		return "", ErrItemNameEmpty
	}
	if utf8.RuneCountInString(normalized) > constants.MaxItemNameLen {
		return "", ErrItemNameTooLong
	}
	return normalized, nil
}
