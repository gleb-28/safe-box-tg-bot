package reminder

import (
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/models"
	"time"
)

type Scheduler interface {
	ComputeNext(r models.Reminder, now time.Time, loc *time.Location) (next time.Time, ok bool)
}

type DefaultScheduler struct{}

func NewScheduler() *DefaultScheduler { return &DefaultScheduler{} }

func (s DefaultScheduler) ComputeNext(r models.Reminder, now time.Time, loc *time.Location) (next time.Time, ok bool) {
	if loc == nil {
		loc = time.UTC
	}
	switch r.Schedule {
	case models.ReminderScheduleInterval:
		return s.computeInterval(r, now)
	case models.ReminderScheduleDaily:
		return s.computeDaily(r, now, loc)
	case models.ReminderScheduleWeekly:
		return s.computeWeekly(r, now, loc)
	case models.ReminderScheduleMonthly:
		return s.computeMonthly(r, now, loc)
	case models.ReminderScheduleOnce:
		if r.NextRun.IsZero() {
			return time.Time{}, false
		}
		return r.NextRun, true
	default:
		return time.Time{}, false
	}
}

func (DefaultScheduler) computeInterval(n models.Reminder, now time.Time) (time.Time, bool) {
	if n.IntervalMinutes == nil || *n.IntervalMinutes <= 0 {
		return time.Time{}, false
	}
	return now.Add(time.Duration(*n.IntervalMinutes) * time.Minute), true
}

func (DefaultScheduler) computeDaily(n models.Reminder, now time.Time, loc *time.Location) (time.Time, bool) {
	if n.TimeOfDayMinutes == nil {
		return time.Time{}, false
	}
	minutes := int(*n.TimeOfDayMinutes)
	if !helpers.ValidTimeOfDay(minutes) {
		return time.Time{}, false
	}
	localNow := now.In(loc)
	target := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), minutes/constants.MinutesInHour, minutes%constants.MinutesInHour, 0, 0, loc)
	if !target.After(localNow) {
		target = target.AddDate(0, 0, 1)
	}
	return target.UTC(), true
}

func (DefaultScheduler) computeWeekly(n models.Reminder, now time.Time, loc *time.Location) (time.Time, bool) {
	if n.TimeOfDayMinutes == nil || n.Weekday == nil {
		return time.Time{}, false
	}
	minutes := int(*n.TimeOfDayMinutes)
	if !helpers.ValidTimeOfDay(minutes) {
		return time.Time{}, false
	}
	localNow := now.In(loc)
	target := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), minutes/constants.MinutesInHour, minutes%constants.MinutesInHour, 0, 0, loc)
	wd := int(localNow.Weekday())
	desired := int(*n.Weekday)
	if desired < 0 || desired >= constants.DaysInWeek {
		return time.Time{}, false
	}
	daysUntil := (desired - wd + constants.DaysInWeek) % constants.DaysInWeek
	if daysUntil == 0 && !target.After(localNow) {
		daysUntil = constants.DaysInWeek
	}
	target = target.AddDate(0, 0, daysUntil)
	return target.UTC(), true
}

func (DefaultScheduler) computeMonthly(n models.Reminder, now time.Time, loc *time.Location) (time.Time, bool) {
	if n.TimeOfDayMinutes == nil || n.MonthDay == nil {
		return time.Time{}, false
	}
	minutes := int(*n.TimeOfDayMinutes)
	if !helpers.ValidTimeOfDay(minutes) {
		return time.Time{}, false
	}
	localNow := now.In(loc)
	day := int(*n.MonthDay)
	if day < 1 || day > 31 {
		return time.Time{}, false
	}
	dim := helpers.DaysInMonth(localNow.Year(), localNow.Month())
	if day > dim {
		day = dim
	}
	target := time.Date(localNow.Year(), localNow.Month(), day, minutes/constants.MinutesInHour, minutes%constants.MinutesInHour, 0, 0, loc)
	if !target.After(localNow) {
		nextMonth := localNow.AddDate(0, 1, 0)
		dimNext := helpers.DaysInMonth(nextMonth.Year(), nextMonth.Month())
		if day > dimNext {
			day = dimNext
		}
		target = time.Date(nextMonth.Year(), nextMonth.Month(), day, minutes/constants.MinutesInHour, minutes%constants.MinutesInHour, 0, 0, loc)
	}
	return target.UTC(), true
}
