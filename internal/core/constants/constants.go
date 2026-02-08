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
	NotificationCheckIntervalMinutes      = 5
	DefaultNotificationIntervalMinMinutes = 60
	DefaultNotificationIntervalMaxMinutes = 120
	NotificationRetryMinutes              = 10
	NotificationItemCooldownMinutes       = 360
)

var FallbackEmojis = []string{"✨", "👀", "🌿", "☕", "🤍", "🍫"}

const (
	DefaultDayStartMinutes = 720  // 12:00
	DefaultDayEndMinutes   = 1320 // 22:00
	DefaultTimezone        = "Europe/Moscow"
)
