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
	NotificationIntervalMinMinutes   = 60
	NotificationIntervalMaxMinutes   = 180
	NotificationRetryMinutes         = 10
	NotificationItemCooldownMinutes  = 360
)
