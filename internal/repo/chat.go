package repo

import (
	"safeboxtgbot/models"

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
