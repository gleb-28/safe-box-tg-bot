package repo

import (
	"safeboxtgbot/models"
	"time"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByTelegramID(userID int64) (models.User, error) {
	var user models.User
	resp := r.db.Preload("Items").
		Find(&models.User{}, &models.User{TelegramID: userID}).
		Scan(&user)

	return user, resp.Error
}

func (r *UserRepo) Upsert(user *models.User) error {
	return r.db.Where("telegram_id = ?", user.TelegramID).
		Assign(user).
		FirstOrCreate(user).
		Error
}

func (r *UserRepo) GetUsersForNotification(now time.Time) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("next_notification <= ?", now).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) UpdateNextNotification(telegramID int64, next time.Time) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Update("next_notification", next).
		Error
}

func (r *UserRepo) UpdateItemBoxClosedMsgID(telegramID int64, msgID int) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Update("item_box_closed_msg_id", msgID).
		Error
}
