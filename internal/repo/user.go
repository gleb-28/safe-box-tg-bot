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
	if err := r.db.Where("notifications_muted = ?", false).
		Where("next_notification <= ?", now).
		Find(&users).Error; err != nil {
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

func (r *UserRepo) UpdateReminderBoxClosedMsgID(telegramID int64, msgID int) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Update("reminder_box_closed_msg_id", msgID).
		Error
}

func (r *UserRepo) UpdateMode(telegramID int64, mode models.UserMode) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Update("mode", mode).
		Error
}

func (r *UserRepo) UpdateNotificationInterval(telegramID int64, preset string, min, max int16, next time.Time) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Updates(map[string]interface{}{
			"notification_preset":               preset,
			"notification_interval_min_minutes": min,
			"notification_interval_max_minutes": max,
			"next_notification":                 next,
		}).
		Error
}

func (r *UserRepo) UpdateNotificationsMuted(telegramID int64, muted bool, next time.Time) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Updates(map[string]interface{}{
			"notifications_muted": muted,
			"next_notification":   next,
		}).
		Error
}

func (r *UserRepo) UpdateDayWindow(telegramID int64, dayStart, dayEnd int16, next time.Time) error {
	return r.db.Model(&models.User{}).
		Where("telegram_id = ?", telegramID).
		Updates(map[string]interface{}{
			"day_start":         dayStart,
			"day_end":           dayEnd,
			"next_notification": next,
		}).
		Error
}
