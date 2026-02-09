package helpers

import (
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/models"
	"safeboxtgbot/pkg/utils"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

func HumanModeName(mode models.UserMode) string {
	switch mode {
	case constants.RoflMode:
		return "Рофл"
	case constants.CareMode:
		return "Забота"
	case constants.CozyMode:
		return "Уют"
	default:
		return "Уют"
	}
}

func ParseMode(raw string) (models.UserMode, bool) {
	switch models.UserMode(raw) {
	case constants.RoflMode:
		return constants.RoflMode, true
	case constants.CozyMode:
		return constants.CozyMode, true
	case constants.CareMode:
		return constants.CareMode, true
	default:
		return "", false
	}
}

func ModeToStyle(mode models.UserMode) string {
	switch mode {
	case constants.RoflMode:
		return "rofl"
	case constants.CareMode:
		return "care"
	case constants.CozyMode:
		return "cozy"
	default:
		return "cozy"
	}
}

func TimeOfDay(local time.Time) string {
	hour := local.Hour()
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "day"
	case hour >= 18 && hour < 23:
		return "evening"
	default:
		return ""
	}
}

func FallbackText(name string, fallbackEmojis []string) string {
	trimmed := strings.TrimSpace(name)
	emoji := ""
	if len(fallbackEmojis) > 0 {
		emoji = fallbackEmojis[utils.RandomIndex(len(fallbackEmojis))]
	}
	if trimmed == "" {
		return strings.TrimSpace(emoji)
	}
	if emoji == "" {
		return trimmed
	}
	return trimmed + " " + emoji
}

func ParseItemID(ctx telebot.Context) (uint, error) {
	raw := ctx.Data()
	if raw == "" && ctx.Callback() != nil {
		raw = ctx.Callback().Data
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty item id")
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("invalid item id")
	}
	return uint(value), nil
}

func CleanLLMText(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
		if idx := strings.LastIndex(trimmed, "```"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
		if idx := strings.Index(trimmed, "\n"); idx >= 0 {
			first := strings.TrimSpace(trimmed[:idx])
			if strings.EqualFold(first, "json") || strings.EqualFold(first, "text") {
				trimmed = strings.TrimSpace(trimmed[idx+1:])
			}
		}
	}
	return strings.TrimSpace(trimmed)
}
