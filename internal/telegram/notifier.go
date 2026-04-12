package telegram

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// NotificationObserver — интерфейс наблюдателя (паттерн Наблюдатель)
type NotificationObserver interface {
	Notify(r *domain.Reminder) error
}

// ReminderService — субъект наблюдения (паттерн Наблюдатель)
type ReminderService struct {
	observers []NotificationObserver
	repo      repository.ReminderRepository
	sender    MessageSender
}

// IsConfigured returns true when the sender is set and ready to send.
func (s *ReminderService) IsConfigured() bool {
	return s.sender != nil && s.sender.IsConfigured()
}

func NewReminderService(repo repository.ReminderRepository) *ReminderService {
	return &ReminderService{repo: repo}
}

// SetSender stores the MessageSender used for config status checks.
func (s *ReminderService) SetSender(sender MessageSender) {
	s.sender = sender
}

func (s *ReminderService) Register(o NotificationObserver) {
	s.observers = append(s.observers, o)
}

// CheckAndNotify fetches pending reminders and notifies all observers
func (s *ReminderService) CheckAndNotify() error {
	pending, err := s.repo.GetPending()
	if err != nil {
		return err
	}
	for _, r := range pending {
		allOK := true
		for _, o := range s.observers {
			if err := o.Notify(r); err != nil {
				allOK = false
			}
		}
		if allOK {
			r.MarkAsSent()
		} else {
			r.MarkAsFailed()
		}
		s.repo.Update(r)
	}
	return nil
}

// TelegramNotifier — конкретный наблюдатель (паттерн Наблюдатель).
// Depends on the MessageSender interface (Adapter pattern) rather than *Client directly.
type TelegramNotifier struct {
	sender MessageSender
}

func NewTelegramNotifier(sender MessageSender) *TelegramNotifier {
	return &TelegramNotifier{sender: sender}
}

func (n *TelegramNotifier) Notify(r *domain.Reminder) error {
	text := "⏰ Reminder: " + r.Content
	return n.sender.Send(text)
}
