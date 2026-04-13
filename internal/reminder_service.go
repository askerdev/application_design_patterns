package domain

// NotificationCoordinator is the port interface for reminder notification delivery.
// Implemented by the telegram coordinator in internal/notifications/telegram/.
type NotificationCoordinator interface {
	IsConfigured() bool
	CheckAndNotify() error
}

// ReminderService defines operations on reminders.
type ReminderService interface {
	List(userID int64) ([]*Reminder, error)
	Create(r *Reminder) error
	Delete(id int64) error
	// Tick checks for due reminders and sends notifications. Safe to call when no coordinator is configured.
	Tick() error
	// IsNotifierConfigured reports whether a push-notification backend is ready.
	IsNotifierConfigured() bool
}

type reminderService struct {
	repo        ReminderRepository
	coordinator NotificationCoordinator
}

// NewReminderService returns a ReminderService. coordinator may be nil when notifications are disabled.
func NewReminderService(repo ReminderRepository, coordinator NotificationCoordinator) ReminderService {
	return &reminderService{repo: repo, coordinator: coordinator}
}

func (s *reminderService) List(userID int64) ([]*Reminder, error) {
	return s.repo.GetAllByUser(userID)
}
func (s *reminderService) Create(r *Reminder) error { return s.repo.Create(r) }
func (s *reminderService) Delete(id int64) error    { return s.repo.Delete(id) }

func (s *reminderService) Tick() error {
	if s.coordinator == nil || !s.coordinator.IsConfigured() {
		return nil
	}
	return s.coordinator.CheckAndNotify()
}

func (s *reminderService) IsNotifierConfigured() bool {
	return s.coordinator != nil && s.coordinator.IsConfigured()
}
