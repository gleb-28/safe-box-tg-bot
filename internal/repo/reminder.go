package repo

import (
	"errors"
	"safeboxtgbot/models"
	"time"

	"gorm.io/gorm"
)

type ReminderRepo struct {
	db *gorm.DB
}

func NewReminderRepo(db *gorm.DB) *ReminderRepo {
	return &ReminderRepo{db: db}
}

func (r *ReminderRepo) TryGet(id uint) (*models.Reminder, bool, error) {
	var n models.Reminder
	err := r.db.First(&n, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &n, true, nil
}

func (r *ReminderRepo) GetByUser(userID int64) ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := r.db.Where("user_id = ?", userID).
		Order("next_run ASC").
		Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (r *ReminderRepo) Create(n *models.Reminder) error {
	return r.db.Create(n).Error
}

func (r *ReminderRepo) Update(n *models.Reminder) error {
	return r.db.Save(n).Error
}

func (r *ReminderRepo) UpdateFields(id uint, userID int64, fields map[string]interface{}) error {
	return r.db.Model(&models.Reminder{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(fields).
		Error
}

func (r *ReminderRepo) UpdateNextRunAt(id uint, userID int64, next time.Time) error {
	return r.db.Model(&models.Reminder{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("next_run", next).
		Error
}

func (r *ReminderRepo) GetDue(now time.Time) ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := r.db.Where("enabled = ?", true).
		Where("next_run <= ?", now).
		Order("next_run ASC").
		Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (r *ReminderRepo) GetDueBySchedule(now time.Time) ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := r.db.Where("enabled = ?", true).
		Where("next_run <= ?", now).
		Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (r *ReminderRepo) SetEnabled(id uint, userID int64, enabled bool) error {
	return r.db.Model(&models.Reminder{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("enabled", enabled).
		Error
}

func (r *ReminderRepo) Delete(id uint, userID int64) (bool, error) {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Reminder{})
	return res.RowsAffected > 0, res.Error
}
