package constants

type NotificationPreset struct {
	Key        string
	Name       string
	MinMinutes int
	MaxMinutes int
}

const (
	NotificationPresetRare   = "rare"
	NotificationPresetNormal = "normal"
	NotificationPresetOften  = "often"
	NotificationPresetChaos  = "chaos"

	DefaultNotificationPreset = NotificationPresetNormal
)

var NotificationPresets = map[string]NotificationPreset{
	NotificationPresetRare: {
		Key:        NotificationPresetRare,
		Name:       "Редко",
		MinMinutes: 120,
		MaxMinutes: 240,
	},
	NotificationPresetNormal: {
		Key:        NotificationPresetNormal,
		Name:       "Иногда",
		MinMinutes: 60,
		MaxMinutes: 120,
	},
	NotificationPresetOften: {
		Key:        NotificationPresetOften,
		Name:       "Часто",
		MinMinutes: 40,
		MaxMinutes: 90,
	},
	NotificationPresetChaos: {
		Key:        NotificationPresetChaos,
		Name:       "Хаос",
		MinMinutes: 30,
		MaxMinutes: 180,
	},
}
