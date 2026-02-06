package repo

import (
	"safeboxtgbot/models"

	"gorm.io/gorm"
)

type ItemRepo struct {
	db *gorm.DB
}

func NewItemRepo(db *gorm.DB) *ItemRepo {
	return &ItemRepo{db: db}
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

func (r *ItemRepo) UpdateName(userID int64, oldName string, newName string) (bool, error) {
	result := r.db.Model(&models.Item{}).
		Where("user_id = ? AND name = ?", userID, oldName).
		Update("name", newName)
	return result.RowsAffected > 0, result.Error
}

func (r *ItemRepo) DeleteByName(userID int64, name string) (bool, error) {
	result := r.db.Where("user_id = ? AND name = ?", userID, name).
		Delete(&models.Item{})
	return result.RowsAffected > 0, result.Error
}
