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
	NotificationCheckIntervalMinutes = 5
	NotificationIntervalMinMinutes   = 40
	NotificationIntervalMaxMinutes   = 150
	NotificationRetryMinutes         = 10
	NotificationItemCooldownMinutes  = 360
)

var FallbackEmojis = []string{"✨", "👀", "🌿", "☕", "🤍", "🍫"}
