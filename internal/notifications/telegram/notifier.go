package telegram

import (
	domain "taskflow/internal/domain"
	notifications "taskflow/internal/notifications"
)

type NotificationObserver interface {
	// ID() string
	Notify(r *domain.Reminder) error
}

type ReminderCoordinator struct {
	observers []NotificationObserver
	repo      domain.ReminderRepository
	sender    notifications.MessageSender
	enabled   bool
}

func NewReminderCoordinator(repo domain.ReminderRepository) *ReminderCoordinator {
	return &ReminderCoordinator{repo: repo, enabled: true}
}

func (c *ReminderCoordinator) SetSender(sender notifications.MessageSender) {
	c.sender = sender
}

// SetEnabled управляет глобальным флагом отправки уведомлений.
// Когда disabled, CheckAndNotify ничего не делает (но всё равно может пометить — см. ниже).
func (c *ReminderCoordinator) SetEnabled(enabled bool) {
	c.enabled = enabled
}

func (c *ReminderCoordinator) IsEnabled() bool {
	return c.enabled
}

func (c *ReminderCoordinator) Register(o NotificationObserver) {
	c.observers = append(c.observers, o)
}

// func (c *ReminderCoordinator) Unregister(o NotificationObserver) {
// 	observers = []NotificationObserver{}
// 	for _, r := range c.observers {
// 		if r.ID() != o.ID() {
// 			observers = append(observers, r)
// 		}
// 	}
// 	c.observers = observers
// }

func (c *ReminderCoordinator) IsConfigured() bool {
	return c.sender != nil && c.sender.IsConfigured()
}

func (c *ReminderCoordinator) CheckAndNotify() error {
	if !c.enabled {
		return nil
	}
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
