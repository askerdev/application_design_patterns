package reminder

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// Notifier is satisfied by telegram.ReminderService (or any type with these two methods).
type Notifier interface {
	IsConfigured() bool
	CheckAndNotify() error
}

// Service defines operations on reminders.
type Service interface {
	List(userID int64) ([]*domain.Reminder, error)
	Create(r *domain.Reminder) error
	Delete(id int64) error
	// Notifier returns the push-notification sender; may return nil when not configured.
	Notifier() Notifier
}

type service struct {
	repo     repository.ReminderRepository
	notifier Notifier
}

// NewService returns a Service backed by repo. n may be nil when telegram is not configured.
func NewService(repo repository.ReminderRepository, n Notifier) Service {
	return &service{repo: repo, notifier: n}
}

func (s *service) List(userID int64) ([]*domain.Reminder, error) {
	return s.repo.GetAllByUser(userID)
}
func (s *service) Create(r *domain.Reminder) error { return s.repo.Create(r) }
func (s *service) Delete(id int64) error            { return s.repo.Delete(id) }
func (s *service) Notifier() Notifier               { return s.notifier }
