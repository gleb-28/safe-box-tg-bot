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
	// If model returns reasoning + final text separated by newlines,
	// keep only the last non-empty line to avoid leaking deliberations.
	lines := strings.Split(trimmed, "\n")
	candidates := make([]string, 0, len(lines))
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if t != "" {
			candidates = append(candidates, t)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[len(candidates)-1]
}

func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ValidTimeOfDay returns true when minutes is within a single day.
func ValidTimeOfDay(minutes int) bool {
	return minutes >= 0 && minutes < constants.MinutesInDay
}

func NormalizeReminderName(raw string, emptyErr, tooLongErr error) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", emptyErr
	}
	if len([]rune(trimmed)) > constants.MaxReminderNameLen {
		return "", tooLongErr
	}
	return trimmed, nil
}

// ParseDateDM parses "DD.MM" using provided now for year defaults and loc.
func ParseDateDM(raw string, now time.Time, loc *time.Location) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	if loc == nil {
		loc = time.UTC
	}
	var d, m int
	n, err := fmt.Sscanf(trimmed, "%d.%d", &d, &m)
	if err != nil || n != 2 {
		return time.Time{}, fmt.Errorf("invalid format")
	}
	if d < 1 || d > 31 || m < 1 || m > 12 {
		return time.Time{}, fmt.Errorf("invalid range")
	}
	year := now.In(loc).Year()
	candidate := time.Date(year, time.Month(m), d, 0, 0, 0, 0, loc)
	return candidate, nil
}

// ComposeDateTime combines date (treated in loc) with minutes of day, returns UTC time.
func ComposeDateTime(date time.Time, minutes int, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	localDate := date.In(loc)
	return time.Date(localDate.Year(), localDate.Month(), localDate.Day(), minutes/constants.MinutesInHour, minutes%constants.MinutesInHour, 0, 0, loc).UTC()
}
