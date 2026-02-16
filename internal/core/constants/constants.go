package constants

import "safeboxtgbot/models"

const (
	RoflMode models.UserMode = "ROFL_MODE"
	CozyMode models.UserMode = "COZY_MODE"
	CareMode models.UserMode = "CARE_MODE"
)

const (
	MaxItemsPerUser = 200
	MaxItemNameLen  = 40
)

const (
	MaxReminderNameLen = 120
)

const (
	NotificationCheckIntervalMinutes      = 5
	DefaultNotificationIntervalMinMinutes = 60
	DefaultNotificationIntervalMaxMinutes = 120
	NotificationRetryMinutes              = 10
	NotificationItemCooldownMinutes       = 360
	ReminderWorkerIntervalSeconds         = 30
)

var FallbackEmojis = []string{"✨", "👀", "🌿", "☕", "🤍", "🍫"}
var WeekdayShortRu = []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}

const ReminderPrefix = "⏰ "

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
