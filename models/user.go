package models

import (
	"time"

	"gorm.io/gorm"
)

type UserMode string

type User struct {
	gorm.Model
	TelegramID                     int64     `gorm:"uniqueIndex"`
	Mode                           UserMode  `gorm:"not null;default:'COZY_MODE'"`
	Timezone                       string    `gorm:"not null;default:'Europe/Moscow'"`
	DayStart                       int16     `gorm:"not null;default:720;check:day_start >= 0 AND day_start <= 1440"` // minutes form start day: 12 * 60 = 720 (12:00)
	DayEnd                         int16     `gorm:"not null;default:1320;check:day_end >= 0 AND day_end <= 1440"`    // minutes form start day: 22 * 60 = 1320 (22:00)
	NotificationPreset             string    `gorm:"not null;default:'normal'"`
	NotificationIntervalMinMinutes int16     `gorm:"not null;default:60;check:notification_interval_min_minutes >= 1 AND notification_interval_min_minutes <= 1440"`
	NotificationIntervalMaxMinutes int16     `gorm:"not null;default:120;check:notification_interval_max_minutes >= notification_interval_min_minutes AND notification_interval_max_minutes <= 1440"`
	NextNotification               time.Time `gorm:"index"`
	ItemBoxClosedMsgID             int       `gorm:"not null;default:0"`
	Items                          []Item
}

type Item struct {
	gorm.Model
	UserID int64  `gorm:"not null"`
	Name   string `gorm:"not null"`

	User User `gorm:"foreignKey:UserID;references:ID"`
}

type MessageLog struct {
	gorm.Model
	UserID int64     `gorm:"not null;index:idx_user_time"`
	ItemID uint      `gorm:"not null;index"`
	SentAt time.Time `gorm:"not null;index:idx_user_time"`
	Text   string    `gorm:"not null"`
}
