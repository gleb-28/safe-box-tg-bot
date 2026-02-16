package reminder

import (
	"safeboxtgbot/models"
	"testing"
	"time"
)

func TestComputeNextInterval(t *testing.T) {
	s := NewScheduler()
	interval := int32(30)
	now := time.Date(2025, time.January, 1, 10, 0, 0, 0, time.UTC)

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:        models.ReminderScheduleInterval,
		IntervalMinutes: &interval,
	}, now, time.UTC)

	if !ok {
		t.Fatalf("expected ok for interval schedule")
	}
	want := now.Add(30 * time.Minute)
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextDailyPastToday(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(9 * 60) // 09:00
	loc := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2025, time.January, 1, 10, 0, 0, 0, loc).UTC()

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleDaily,
		TimeOfDayMinutes: &timeOfDay,
	}, now, loc)

	if !ok {
		t.Fatalf("expected ok for daily schedule")
	}
	want := time.Date(2025, time.January, 2, 9, 0, 0, 0, loc).UTC()
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextDailyLaterToday(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(15 * 60) // 15:00
	loc := time.UTC
	now := time.Date(2025, time.January, 1, 14, 0, 0, 0, loc)

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleDaily,
		TimeOfDayMinutes: &timeOfDay,
	}, now, loc)

	if !ok {
		t.Fatalf("expected ok for daily schedule")
	}
	want := time.Date(2025, time.January, 1, 15, 0, 0, 0, loc)
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextDailyInvalidTime(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(24 * 60) // invalid
	now := time.Now()

	_, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleDaily,
		TimeOfDayMinutes: &timeOfDay,
	}, now, time.UTC)

	if ok {
		t.Fatalf("expected ok=false for invalid time of day")
	}
}

func TestComputeNextWeeklySameDayInPast(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(8 * 60) // 08:00
	weekday := int8(time.Monday)
	now := time.Date(2025, time.January, 6, 9, 0, 0, 0, time.UTC) // Monday 09:00

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleWeekly,
		TimeOfDayMinutes: &timeOfDay,
		Weekday:          &weekday,
	}, now, time.UTC)

	if !ok {
		t.Fatalf("expected ok for weekly schedule")
	}
	want := time.Date(2025, time.January, 13, 8, 0, 0, 0, time.UTC) // next Monday
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextWeeklySameDayLaterToday(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(12 * 60) // 12:00
	weekday := int8(time.Tuesday)
	now := time.Date(2025, time.January, 7, 10, 0, 0, 0, time.UTC) // Tuesday 10:00

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleWeekly,
		TimeOfDayMinutes: &timeOfDay,
		Weekday:          &weekday,
	}, now, time.UTC)

	if !ok {
		t.Fatalf("expected ok for weekly schedule")
	}
	want := time.Date(2025, time.January, 7, 12, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextWeeklyInvalidWeekday(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(8 * 60)
	weekday := int8(7) // invalid
	now := time.Now()

	_, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleWeekly,
		TimeOfDayMinutes: &timeOfDay,
		Weekday:          &weekday,
	}, now, time.UTC)

	if ok {
		t.Fatalf("expected ok=false for invalid weekday")
	}
}

func TestComputeNextMonthlyClampsDay(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(8 * 60) // 08:00
	day := int8(31)
	now := time.Date(2025, time.February, 1, 12, 0, 0, 0, time.UTC)

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleMonthly,
		TimeOfDayMinutes: &timeOfDay,
		MonthDay:         &day,
	}, now, time.UTC)

	if !ok {
		t.Fatalf("expected ok for monthly schedule")
	}
	// February 2025 has 28 days, so the 31st clamps to 28th.
	want := time.Date(2025, time.February, 28, 8, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextMonthlyFutureThisMonth(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(9*60 + 30) // 09:30
	day := int8(15)
	now := time.Date(2025, time.January, 10, 12, 0, 0, 0, time.UTC)

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleMonthly,
		TimeOfDayMinutes: &timeOfDay,
		MonthDay:         &day,
	}, now, time.UTC)

	if !ok {
		t.Fatalf("expected ok for monthly schedule")
	}
	want := time.Date(2025, time.January, 15, 9, 30, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}

func TestComputeNextMonthlyInvalidDay(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(8 * 60)
	day := int8(0) // invalid
	now := time.Now()

	_, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleMonthly,
		TimeOfDayMinutes: &timeOfDay,
		MonthDay:         &day,
	}, now, time.UTC)

	if ok {
		t.Fatalf("expected ok=false for invalid month day")
	}
}

func TestComputeNextInvalidInterval(t *testing.T) {
	s := NewScheduler()
	now := time.Now()

	next, ok := s.ComputeNext(models.Reminder{
		Schedule: models.ReminderScheduleInterval,
	}, now, nil)

	if ok {
		t.Fatalf("expected ok=false, got true with next=%v", next)
	}
}

func TestComputeNextOnce(t *testing.T) {
	s := NewScheduler()
	now := time.Now().UTC()
	runAt := now.Add(2 * time.Hour)

	next, ok := s.ComputeNext(models.Reminder{
		Schedule: models.ReminderScheduleOnce,
		NextRun:  runAt,
	}, now, nil)

	if !ok || !next.Equal(runAt) {
		t.Fatalf("expected once schedule to return next=%v ok=true, got next=%v ok=%v", runAt, next, ok)
	}
}

func TestComputeNextUnknownSchedule(t *testing.T) {
	s := NewScheduler()
	now := time.Now()

	next, ok := s.ComputeNext(models.Reminder{
		Schedule: "unknown",
	}, now, nil)

	if ok || !next.IsZero() {
		t.Fatalf("expected unknown schedule to return zero time and ok=false, got next=%v ok=%v", next, ok)
	}
}

func TestComputeNextDefaultsToUTC(t *testing.T) {
	s := NewScheduler()
	timeOfDay := int16(60) // 01:00
	loc := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2025, time.January, 1, 0, 30, 0, 0, loc)

	next, ok := s.ComputeNext(models.Reminder{
		Schedule:         models.ReminderScheduleDaily,
		TimeOfDayMinutes: &timeOfDay,
	}, now, nil)

	if !ok {
		t.Fatalf("expected ok for daily schedule")
	}
	want := time.Date(2025, time.January, 1, 1, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("next = %v, want %v", next, want)
	}
}
