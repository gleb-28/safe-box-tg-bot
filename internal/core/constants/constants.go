package constants

import (
	"safeboxtgbot/models"
	"time"
)

const (
	RoflMode models.UserMode = "ROFL_MODE"
	CozyMode models.UserMode = "COZY_MODE"
	CareMode models.UserMode = "CARE_MODE"
)

const (
	MaxItemsPerUser    = 200
	MaxItemNameLen     = 40
	MaxReminderNameLen = 120
	ReminderPrefix     = "⏰ "
)

const (
	NotificationCheckIntervalMinutes      = 5
	DefaultNotificationIntervalMinMinutes = 60
	DefaultNotificationIntervalMaxMinutes = 120
	NotificationRetryMinutes              = 10
	NotificationItemCooldownMinutes       = 360
	ReminderWorkerIntervalSeconds         = 30
	NonAuthSessionTTL                     = 10 * time.Minute
)

const (
	DefaultDayStartMinutes = 720  // 12:00
	DefaultDayEndMinutes   = 1320 // 22:00
	DefaultTimezone        = "Europe/Moscow"
)

const (
	MinutesInHour = 60
	MinutesInDay  = 24 * MinutesInHour
	DaysInWeek    = 7
)

var (
	FallbackEmojis = []string{"✨", "👀", "🌿", "☕", "🤍", "🍫"}
	WeekdayShortRu = []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}
)
