package reminder

import (
	"errors"
	"fmt"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/models"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	ErrInvalidSchedule   = errors.New("invalid schedule")
	ErrInvalidTimeOfDay  = errors.New("invalid time of day")
	ErrInvalidInterval   = errors.New("invalid interval")
	ErrInvalidWeekday    = errors.New("invalid weekday")
	ErrEmptyEntityName   = errors.New("entity name empty")
	ErrEntityNameTooLong = errors.New("entity name too long")
	ErrReminderNotFound  = errors.New("reminder not found")
	ErrReminderDuplicate = errors.New("reminder duplicate")
)

type Service struct {
	reminderRepo *repo.ReminderRepo
	scheduler    Scheduler
	store        *session.Store
	logger       logger.AppLogger
}

func NewService(reminderRepo *repo.ReminderRepo, scheduler Scheduler, store *session.Store, logger logger.AppLogger) *Service {
	return &Service{reminderRepo: reminderRepo, scheduler: scheduler, store: store, logger: logger}
}

func (s *Service) SetBotLastMsg(userID int64, msg *telebot.Message) {
	s.store.SetReminderBotLastMsg(userID, msg)
}

func (s *Service) GetBotLastMsg(userID int64) *telebot.Message {
	return s.store.GetReminderBotLastMsg(userID)
}

func (s *Service) ClearPending(userID int64) {
	s.store.ClearPendingReminder(userID)
}

func (s *Service) SetPending(userID int64, pending *session.PendingReminder) {
	s.store.SetPendingReminder(userID, pending)
}

func (s *Service) GetPending(userID int64) *session.PendingReminder {
	return s.store.GetPendingReminder(userID)
}

func (s *Service) GetList(userID int64) ([]models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	return s.store.GetReminderList(userID), nil
}

func (s *Service) CreateInterval(userID int64, entityName string, intervalMinutes int32, now time.Time, loc *time.Location) (*models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	name, err := helpers.NormalizeReminderName(entityName, ErrEmptyEntityName, ErrEntityNameTooLong)
	if err != nil {
		return nil, err
	}
	if s.isDuplicateName(userID, name) {
		return nil, ErrReminderDuplicate
	}
	if intervalMinutes <= 0 {
		return nil, ErrInvalidInterval
	}

	r := models.Reminder{
		UserID:          userID,
		Name:            name,
		Schedule:        models.ReminderScheduleInterval,
		IntervalMinutes: &intervalMinutes,
		Enabled:         true,
	}

	next, ok := s.scheduler.ComputeNext(r, now, loc)
	if !ok {
		return nil, ErrInvalidSchedule
	}
	r.NextRun = next

	if err := s.reminderRepo.Create(&r); err != nil {
		return nil, err
	}
	s.upsertReminderInStore(userID, r)
	return &r, nil
}

func (s *Service) CreateDaily(userID int64, entityName string, timeOfDayMinutes int16, now time.Time, loc *time.Location) (*models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	name, err := helpers.NormalizeReminderName(entityName, ErrEmptyEntityName, ErrEntityNameTooLong)
	if err != nil {
		return nil, err
	}
	if s.isDuplicateName(userID, name) {
		return nil, ErrReminderDuplicate
	}
	if timeOfDayMinutes < 0 || timeOfDayMinutes >= 24*60 {
		return nil, ErrInvalidTimeOfDay
	}

	r := models.Reminder{
		UserID:           userID,
		Name:             name,
		Schedule:         models.ReminderScheduleDaily,
		TimeOfDayMinutes: &timeOfDayMinutes,
		Enabled:          true,
	}

	next, ok := s.scheduler.ComputeNext(r, now, loc)
	if !ok {
		return nil, ErrInvalidSchedule
	}
	r.NextRun = next

	if err := s.reminderRepo.Create(&r); err != nil {
		return nil, err
	}
	s.upsertReminderInStore(userID, r)
	return &r, nil
}

func (s *Service) CreateWeekly(userID int64, entityName string, weekday int8, timeOfDayMinutes int16, now time.Time, loc *time.Location) (*models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	name, err := helpers.NormalizeReminderName(entityName, ErrEmptyEntityName, ErrEntityNameTooLong)
	if err != nil {
		return nil, err
	}
	if s.isDuplicateName(userID, name) {
		return nil, ErrReminderDuplicate
	}
	if weekday < 0 || weekday > 6 {
		return nil, ErrInvalidWeekday
	}
	if timeOfDayMinutes < 0 || timeOfDayMinutes >= 24*60 {
		return nil, ErrInvalidTimeOfDay
	}

	r := models.Reminder{
		UserID:           userID,
		Name:             name,
		Schedule:         models.ReminderScheduleWeekly,
		TimeOfDayMinutes: &timeOfDayMinutes,
		Weekday:          &weekday,
		Enabled:          true,
	}

	next, ok := s.scheduler.ComputeNext(r, now, loc)
	if !ok {
		return nil, ErrInvalidSchedule
	}
	r.NextRun = next

	if err := s.reminderRepo.Create(&r); err != nil {
		return nil, err
	}
	s.upsertReminderInStore(userID, r)
	return &r, nil
}

func (s *Service) CreateMonthly(userID int64, entityName string, day int8, timeOfDayMinutes int16, now time.Time, loc *time.Location) (*models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	name, err := helpers.NormalizeReminderName(entityName, ErrEmptyEntityName, ErrEntityNameTooLong)
	if err != nil {
		return nil, err
	}
	if s.isDuplicateName(userID, name) {
		return nil, ErrReminderDuplicate
	}
	if day < 1 || day > 31 {
		return nil, ErrInvalidWeekday
	}
	if timeOfDayMinutes < 0 || timeOfDayMinutes >= 24*60 {
		return nil, ErrInvalidTimeOfDay
	}

	dayCopy := day
	r := models.Reminder{
		UserID:           userID,
		Name:             name,
		Schedule:         models.ReminderScheduleMonthly,
		TimeOfDayMinutes: &timeOfDayMinutes,
		MonthDay:         &dayCopy,
		Enabled:          true,
	}

	next, ok := s.scheduler.ComputeNext(r, now, loc)
	if !ok {
		return nil, ErrInvalidSchedule
	}
	r.NextRun = next

	if err := s.reminderRepo.Create(&r); err != nil {
		return nil, err
	}
	s.upsertReminderInStore(userID, r)
	return &r, nil
}

func (s *Service) CreateOnce(userID int64, entityName string, runAt time.Time, now time.Time, loc *time.Location) (*models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	name, err := helpers.NormalizeReminderName(entityName, ErrEmptyEntityName, ErrEntityNameTooLong)
	if err != nil {
		return nil, err
	}
	if s.isDuplicateName(userID, name) {
		return nil, ErrReminderDuplicate
	}
	if runAt.IsZero() {
		return nil, ErrInvalidSchedule
	}

	r := models.Reminder{
		UserID:   userID,
		Name:     name,
		Schedule: models.ReminderScheduleOnce,
		NextRun:  runAt.UTC(),
		Enabled:  true,
	}

	if err := s.reminderRepo.Create(&r); err != nil {
		return nil, err
	}
	s.upsertReminderInStore(userID, r)
	return &r, nil
}

func (s *Service) Enable(id uint, userID int64, now time.Time, loc *time.Location) error {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return err
	}
	r, found, err := s.reminderRepo.TryGet(id)
	if err != nil {
		return err
	}
	if !found || r.UserID != userID {
		return ErrReminderNotFound
	}
	if r.Enabled {
		return nil
	}
	next := r.NextRun
	if next.IsZero() {
		if computed, ok := s.scheduler.ComputeNext(*r, now, loc); ok {
			next = computed
		} else {
			next = now.Add(time.Duration(constants.DefaultNotificationIntervalMinMinutes) * time.Minute)
		}
	}
	r.Enabled = true
	r.NextRun = next
	if err := s.reminderRepo.Update(r); err != nil {
		return err
	}
	s.upsertReminderInStore(userID, *r)
	return nil
}

func (s *Service) Disable(id uint, userID int64) error {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return err
	}
	r, found, err := s.reminderRepo.TryGet(id)
	if err != nil {
		return err
	}
	if !found || r.UserID != userID {
		return ErrReminderNotFound
	}
	if !r.Enabled {
		return nil
	}
	r.Enabled = false
	if err := s.reminderRepo.Update(r); err != nil {
		return err
	}
	s.upsertReminderInStore(userID, *r)
	return nil
}

func (s *Service) Delete(id uint, userID int64) error {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return err
	}
	deleted, err := s.reminderRepo.Delete(id, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrReminderNotFound
	}
	s.removeReminderFromStore(userID, id)
	return nil
}

func (s *Service) Reschedule(r *models.Reminder, now time.Time, loc *time.Location) error {
	if err := s.ensureRemindersSessionLoaded(r.UserID); err != nil {
		return err
	}

	if r.Schedule == models.ReminderScheduleOnce {
		r.Enabled = false
		r.NextRun = time.Time{}
		if err := s.reminderRepo.Update(r); err != nil {
			return err
		}
		s.upsertReminderInStore(r.UserID, *r)
		return nil
	}

	next, ok := s.scheduler.ComputeNext(*r, now, loc)
	if !ok {
		r.Enabled = false
		if err := s.reminderRepo.Update(r); err != nil {
			return err
		}
		s.upsertReminderInStore(r.UserID, *r)
		return nil
	}
	r.NextRun = next
	if err := s.reminderRepo.Update(r); err != nil {
		return err
	}
	s.upsertReminderInStore(r.UserID, *r)
	return nil
}

func (s *Service) SetNextRun(id uint, userID int64, next time.Time) error {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return err
	}

	if err := s.reminderRepo.UpdateNextRunAt(id, userID, next); err != nil {
		return err
	}
	s.updateNextRunInStore(userID, id, next)
	return nil
}

func (s *Service) GetDue(now time.Time) ([]models.Reminder, error) {
	return s.reminderRepo.GetDue(now)
}

func (s *Service) GetByUser(userID int64) ([]models.Reminder, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return nil, err
	}
	return s.store.GetReminderList(userID), nil
}

func (s *Service) GetByID(id uint) (*models.Reminder, bool, error) {
	return s.reminderRepo.TryGet(id)
}

func (s *Service) upsertReminderInStore(userID int64, reminder models.Reminder) {
	reminders := append([]models.Reminder(nil), s.store.GetReminderList(userID)...)
	replaced := false
	for i := range reminders {
		if reminders[i].ID == reminder.ID {
			reminders[i] = reminder
			replaced = true
			break
		}
	}
	if !replaced {
		reminders = append(reminders, reminder)
	}
	s.store.SetReminderList(userID, reminders)
}

func (s *Service) removeReminderFromStore(userID int64, id uint) {
	reminders := s.store.GetReminderList(userID)
	filtered := make([]models.Reminder, 0, len(reminders))
	for _, reminder := range reminders {
		if reminder.ID != id {
			filtered = append(filtered, reminder)
		}
	}
	s.store.SetReminderList(userID, filtered)
}

func (s *Service) updateNextRunInStore(userID int64, id uint, next time.Time) {
	reminders := append([]models.Reminder(nil), s.store.GetReminderList(userID)...)
	for i := range reminders {
		if reminders[i].ID == id {
			reminders[i].NextRun = next
			s.store.SetReminderList(userID, reminders)
			return
		}
	}
}

func (s *Service) ensureRemindersSessionLoaded(userID int64) error {
	if s.store.IsRemindersLoaded(userID) {
		s.logger.Debug(fmt.Sprintf("Reminders already loaded for userID=%d", userID))
		return nil
	}

	s.logger.Debug(fmt.Sprintf("Loading reminders into session for userID=%d", userID))
	reminders, err := s.reminderRepo.GetByUser(userID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error loading reminders from DB for userID=%d: %v", userID, err))
		return err
	}

	s.store.SetReminderList(userID, reminders)
	s.logger.Debug(fmt.Sprintf("Reminders loaded into session for userID=%d", userID))
	return nil
}

func (s *Service) isDuplicateName(userID int64, name string) bool {
	for _, r := range s.store.GetReminderList(userID) {
		if r.Name == name {
			return true
		}
	}
	return false
}

func (s *Service) IsDuplicateName(userID int64, name string) (bool, error) {
	if err := s.ensureRemindersSessionLoaded(userID); err != nil {
		return false, err
	}
	return s.isDuplicateName(userID, name), nil
}

// ClampToActiveWindow adjusts time-based reminders to fit within user's active window and recomputes NextRun.
func (s *Service) ClampToActiveWindow(user models.User, now time.Time, loc *time.Location) error {
	if err := s.ensureRemindersSessionLoaded(user.TelegramID); err != nil {
		return err
	}
	reminders := s.store.GetReminderList(user.TelegramID)
	updated := make([]models.Reminder, 0, len(reminders))

	for _, r := range reminders {
		if r.Schedule == models.ReminderScheduleInterval || r.TimeOfDayMinutes == nil {
			updated = append(updated, r)
			continue
		}

		adjusted, clamped := ClampMinutesToWindow(int(*r.TimeOfDayMinutes), int(user.DayStart), int(user.DayEnd))
		if clamped {
			val := int16(adjusted)
			r.TimeOfDayMinutes = &val
		}

		if next, ok := s.scheduler.ComputeNext(r, now, loc); ok {
			r.NextRun = next
		}
		if err := s.reminderRepo.Update(&r); err != nil {
			return err
		}
		updated = append(updated, r)
	}

	s.store.SetReminderList(user.TelegramID, updated)
	return nil
}

func ClampMinutesToWindow(minutes, start, end int) (int, bool) {
	if start == end {
		return start, true
	}

	inWindow := func(min int) bool {
		if start <= end {
			return min >= start && min <= end
		}
		return min >= start || min <= end
	}
	if inWindow(minutes) {
		return minutes, false
	}

	distance := func(a, b int) int {
		diff := a - b
		if diff < 0 {
			diff = -diff
		}
		if diff > constants.MinutesInDay/2 {
			diff = constants.MinutesInDay - diff
		}
		return diff
	}
	dStart := distance(minutes, start)
	dEnd := distance(minutes, end)
	if dStart <= dEnd {
		return start, true
	}
	return end, true
}
