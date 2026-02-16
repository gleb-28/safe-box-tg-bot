package models

import (
	"time"

	"gorm.io/gorm"
)

// Reminder is an independent scheduled reminder (separate from items).
type Reminder struct {
	gorm.Model
	UserID           int64            `gorm:"index;not null;constraint:OnDelete:CASCADE;"`
	Name             string           `gorm:"not null"`
	Schedule         ReminderSchedule `gorm:"not null"`
	NextRun          time.Time        `gorm:"index:idx_reminder_user_next;not null"`
	LastRun          time.Time        `gorm:"not null"`
	IntervalMinutes  *int32           `gorm:"check:interval_minutes IS NULL OR (interval_minutes > 0 AND interval_minutes <= 1440)"`
	TimeOfDayMinutes *int16           `gorm:"check:time_of_day_minutes IS NULL OR (time_of_day_minutes >= 0 AND time_of_day_minutes < 1440)"`
	Weekday          *int8            `gorm:"index;check:weekday IS NULL OR (weekday >= 0 AND weekday <= 6)"`
	MonthDay         *int8            `gorm:"index;check:month_day IS NULL OR (month_day >= 1 AND month_day <= 31)"`
	Enabled          bool             `gorm:"not null;default:true"`

	User User `gorm:"foreignKey:UserID;references:ID"`
}

// ReminderSchedule describes recurrence.
type ReminderSchedule string

const (
	ReminderScheduleOnce     ReminderSchedule = "once"
	ReminderScheduleInterval ReminderSchedule = "interval"
	ReminderScheduleDaily    ReminderSchedule = "daily"
	ReminderScheduleWeekly   ReminderSchedule = "weekly"
	ReminderScheduleMonthly  ReminderSchedule = "monthly"
)
