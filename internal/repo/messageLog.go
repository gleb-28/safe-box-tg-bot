package repo

import (
	"errors"
	"safeboxtgbot/models"
	"time"

	"gorm.io/gorm"
)

type MessageLogRepo struct {
	db *gorm.DB
}

func NewMessageLogRepo(db *gorm.DB) *MessageLogRepo {
	return &MessageLogRepo{db: db}
}

func (r *MessageLogRepo) TryGet(id uint) (*models.MessageLog, bool, error) {
	var log models.MessageLog
	err := r.db.First(&log, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &log, true, nil
}

func (r *MessageLogRepo) Create(log *models.MessageLog) error {
	return r.db.Create(log).Error
}

func (r *MessageLogRepo) GetByTelegramID(userID int64) ([]models.MessageLog, error) {
	var logs []models.MessageLog
	if err := r.db.Where("user_id = ?", userID).Order("sent_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *MessageLogRepo) GetByUserAndItem(userID int64, itemID uint) ([]models.MessageLog, error) {
	var logs []models.MessageLog
	if err := r.db.Where("user_id = ? AND item_id = ?", userID, itemID).
		Order("sent_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *MessageLogRepo) GetRecentItemIDs(userID int64, since time.Time) ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&models.MessageLog{}).
		Distinct("item_id").
		Where("user_id = ? AND sent_at >= ?", userID, since).
		Pluck("item_id", &ids).
		Error; err != nil {
		return nil, err
	}
	return ids, nil
}
