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

func (r *ItemRepo) GetByTelegramID(userID uint) ([]models.Item, error) {
	var items []models.Item
	if err := r.db.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ItemRepo) Upsert(item *models.Item) error {
	return r.db.Where("id = ?", item.ID).
		Assign(item).
		FirstOrCreate(item).
		Error
}

func (r *ItemRepo) Delete(itemID uint) error {
	return r.db.Delete(&models.Item{}, itemID).Error
}
