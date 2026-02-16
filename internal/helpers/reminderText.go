package helpers

import (
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/text"
	"safeboxtgbot/models"
	"strings"
	"time"
)

func HumanReminderSchedule(r models.Reminder, loc *time.Location, replies *text.Replies) string {
	switch r.Schedule {
	case models.ReminderScheduleOnce:
		next := r.NextRun
		if !next.IsZero() {
			if loc != nil {
				next = next.In(loc)
			}
			return fmt.Sprintf(replies.ReminderHumanOnce, next.Format("02.01 15:04"))
		}
	case models.ReminderScheduleInterval:
		if r.IntervalMinutes != nil && *r.IntervalMinutes > 0 {
			return fmt.Sprintf(replies.ReminderHumanInterval, *r.IntervalMinutes)
		}
	case models.ReminderScheduleDaily:
		if r.TimeOfDayMinutes != nil {
			return fmt.Sprintf(replies.ReminderHumanDaily, FormatTimeHM(int(*r.TimeOfDayMinutes)))
		}
	case models.ReminderScheduleWeekly:
		if r.TimeOfDayMinutes != nil && r.Weekday != nil {
			wd := int(*r.Weekday)
			name := constants.WeekdayShortRu[wd%len(constants.WeekdayShortRu)]
			return fmt.Sprintf(replies.ReminderHumanWeekly, name, FormatTimeHM(int(*r.TimeOfDayMinutes)))
		}
	case models.ReminderScheduleMonthly:
		if r.TimeOfDayMinutes != nil && r.MonthDay != nil {
			day := int(*r.MonthDay)
			return fmt.Sprintf(replies.ReminderHumanMonthly, day, FormatTimeHM(int(*r.TimeOfDayMinutes)))
		}
	}
	return replies.ReminderHumanFallback
}

func ParseTimeHM(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("empty time")
	}
	var h, m int
	n, err := fmt.Sscanf(trimmed, "%d:%d", &h, &m)
	if err != nil || n != 2 {
		return 0, fmt.Errorf("invalid time")
	}
	if h < 0 || h >= constants.MinutesInDay/constants.MinutesInHour || m < 0 || m >= constants.MinutesInHour {
		return 0, fmt.Errorf("invalid time range")
	}
	return h*constants.MinutesInHour + m, nil
}
