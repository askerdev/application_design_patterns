package telegram

import (
	domain "taskflow/internal/domain"
	notifications "taskflow/internal/notifications"
)

// NotificationObserver is the Observer pattern interface for reminder notifications.
type NotificationObserver interface {
	Notify(r *domain.Reminder) error
}

// ReminderCoordinator implements the notification coordinator role.
// It fetches pending reminders and dispatches to registered observers (Observer pattern).
type ReminderCoordinator struct {
	observers []NotificationObserver
	repo      domain.ReminderRepository
	sender    notifications.MessageSender
}

func NewReminderCoordinator(repo domain.ReminderRepository) *ReminderCoordinator {
	return &ReminderCoordinator{repo: repo}
}

// SetSender stores the MessageSender used for IsConfigured checks.
func (c *ReminderCoordinator) SetSender(sender notifications.MessageSender) {
	c.sender = sender
}

func (c *ReminderCoordinator) Register(o NotificationObserver) {
	c.observers = append(c.observers, o)
}

// IsConfigured returns true when the sender is set and ready.
func (c *ReminderCoordinator) IsConfigured() bool {
	return c.sender != nil && c.sender.IsConfigured()
}

// CheckAndNotify fetches pending reminders and notifies all observers.
func (c *ReminderCoordinator) CheckAndNotify() error {
	pending, err := c.repo.GetPending()
	if err != nil {
		return err
	}
	for _, r := range pending {
		allOK := true
		for _, o := range c.observers {
			if err := o.Notify(r); err != nil {
				allOK = false
			}
		}
		if allOK {
			r.MarkAsSent()
		} else {
			r.MarkAsFailed()
		}
		c.repo.Update(r)
	}
	return nil
}

// TelegramNotifier is a concrete Observer that sends reminders via Telegram.
type TelegramNotifier struct {
	sender notifications.MessageSender
}

func NewTelegramNotifier(sender notifications.MessageSender) *TelegramNotifier {
	return &TelegramNotifier{sender: sender}
}

func (n *TelegramNotifier) Notify(r *domain.Reminder) error {
	text := "⏰ Reminder: " + r.Content
	return n.sender.Send(text)
}
