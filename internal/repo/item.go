package repo

import (
	"errors"
	"safeboxtgbot/models"

	"gorm.io/gorm"
)

type ItemRepo struct {
	db *gorm.DB
}

func NewItemRepo(db *gorm.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

func (r *ItemRepo) TryGet(userID int64, itemID uint) (*models.Item, bool, error) {
	var item models.Item
	err := r.db.Where("user_id = ? AND id = ?", userID, itemID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &item, true, nil
}

func (r *ItemRepo) GetByTelegramID(userID int64) ([]models.Item, error) {
	var items []models.Item
	if err := r.db.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ItemRepo) Upsert(item *models.Item) error {
	return r.db.Where("user_id = ? AND name = ?", item.UserID, item.Name).
		Assign(item).
		FirstOrCreate(item).
		Error
}

func (r *ItemRepo) UpdateName(userID int64, itemID uint, newName string) (bool, error) {
	result := r.db.Model(&models.Item{}).
		Where("user_id = ? AND id = ?", userID, itemID).
		Update("name", newName)
	return result.RowsAffected > 0, result.Error
}

func (r *ItemRepo) Delete(userID int64, itemID uint) (bool, error) {
	result := r.db.Where("user_id = ? AND id = ?", userID, itemID).
		Delete(&models.Item{})
	return result.RowsAffected > 0, result.Error
}
