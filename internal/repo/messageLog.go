package repo

import (
	"safeboxtgbot/models"

	"gorm.io/gorm"
)

type MessageLogRepo struct {
	db *gorm.DB
}

func NewMessageLogRepo(db *gorm.DB) *MessageLogRepo {
	return &MessageLogRepo{db: db}
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
