package helpers

import (
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/models"
	"safeboxtgbot/pkg/utils"
	"strings"
	"time"
)

func HumanNotificationStatus(muted bool) string {
	if muted {
		return "выключены"
	}
	return "включены"
}

func FormatTimeHM(minutes int) string {
	if minutes < 0 {
		return ""
	}
	h, m := utils.MinutesToTime(minutes % 1440)
	return fmt.Sprintf("%02d:%02d", h, m)
}

func ParseNotificationPreset(raw string) (constants.NotificationPreset, bool) {
	preset, ok := constants.NotificationPresets[raw]
	return preset, ok
}

func HumanNotificationPresetName(raw string) string {
	if preset, ok := constants.NotificationPresets[raw]; ok && preset.Name != "" {
		return preset.Name
	}
	if preset, ok := constants.NotificationPresets[constants.DefaultNotificationPreset]; ok {
		return preset.Name
	}
	return "Иногда"
}

func NotificationPresetRangeText(preset constants.NotificationPreset) string {
	return FormatMinutesRange(preset.MinMinutes, preset.MaxMinutes)
}

func FormatMinutesRange(minMinutes, maxMinutes int) string {
	if minMinutes <= 0 || maxMinutes <= 0 {
		return ""
	}
	if minMinutes%60 == 0 && maxMinutes%60 == 0 {
		return fmt.Sprintf("%d–%d ч", minMinutes/60, maxMinutes/60)
	}
	return fmt.Sprintf("%d–%d мин", minMinutes, maxMinutes)
}

func UserNotificationRange(user models.User) (int, int) {
	if preset, ok := constants.NotificationPresets[user.NotificationPreset]; ok {
		return preset.MinMinutes, preset.MaxMinutes
	}

	min := int(user.NotificationIntervalMinMinutes)
	max := int(user.NotificationIntervalMaxMinutes)
	if min > 0 && max >= min {
		return min, max
	}

	if preset, ok := constants.NotificationPresets[constants.DefaultNotificationPreset]; ok {
		return preset.MinMinutes, preset.MaxMinutes
	}
	return constants.DefaultNotificationIntervalMinMinutes, constants.DefaultNotificationIntervalMaxMinutes
}

func NextNotificationTime(user models.User, nowUTC time.Time) time.Time {
	loc, _ := UserLocation(user)
	return NextNotificationTimeWithLoc(user, nowUTC, loc)
}

func NextNotificationTimeWithLoc(user models.User, nowUTC time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	minMinutes, maxMinutes := UserNotificationRange(user)
	if minMinutes <= 0 || maxMinutes <= 0 || maxMinutes < minMinutes {
		minMinutes = constants.DefaultNotificationIntervalMinMinutes
		maxMinutes = constants.DefaultNotificationIntervalMaxMinutes
	}

	localNow := nowUTC.In(loc)
	nextLocal := localNow.Add(utils.RandomDurationMinutes(minMinutes, maxMinutes))
	if IsWithinActiveWindow(user, nextLocal) {
		return nextLocal.UTC()
	}
	return NextStartTimeFromLocal(user, nextLocal, loc)
}

func UserLocation(user models.User) (*time.Location, error) {
	timezone := strings.TrimSpace(user.Timezone)
	if timezone == "" {
		timezone = constants.DefaultTimezone
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC, err
	}
	return loc, nil
}

func IsWithinActiveWindow(user models.User, local time.Time) bool {
	start, end := normalizedActiveWindow(user)
	minutes := local.Hour()*60 + local.Minute()

	if start <= end {
		return minutes >= start && minutes <= end
	}
	return minutes >= start || minutes <= end
}

func NextStartTimeFromLocal(user models.User, fromLocal time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	start, _ := normalizedActiveWindow(user)
	startHour, startMinute := utils.MinutesToTime(start)
	startTime := time.Date(fromLocal.Year(), fromLocal.Month(), fromLocal.Day(), startHour, startMinute, 0, 0, loc)
	if !fromLocal.Before(startTime) {
		startTime = startTime.AddDate(0, 0, 1)
	}

	jitterMax := ActiveWindowMinutes(user)
	if jitterMax > 60 {
		jitterMax = 60
	}
	if jitterMax > 0 {
		startTime = startTime.Add(time.Duration(utils.RandomIntRange(0, jitterMax)) * time.Minute)
	}
	return startTime.UTC()
}

func ActiveWindowMinutes(user models.User) int {
	start, end := normalizedActiveWindow(user)
	if start <= end {
		return end - start
	}
	return 24*60 - start + end
}

func normalizedActiveWindow(user models.User) (int, int) {
	start := int(user.DayStart)
	end := int(user.DayEnd)

	if start < 0 || start > 1440 || end < 0 || end > 1440 || start == end {
		return constants.DefaultDayStartMinutes, constants.DefaultDayEndMinutes
	}
	return start, end
}
