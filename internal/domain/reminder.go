package domain

import "time"

type ReminderStatus string

const (
	ReminderStatusPending ReminderStatus = "PENDING"
	ReminderStatusSent    ReminderStatus = "SENT"
	ReminderStatusFailed  ReminderStatus = "FAILED"
)

type Reminder struct {
	ID           int64
	UserID       int64
	ProjectID    *int64
	TagID        *int64
	Content      string
	ReminderTime time.Time
	Status       ReminderStatus
	CreatedAt    time.Time
}

func (r *Reminder) IsReady() bool {
	return r.Status == ReminderStatusPending && time.Now().After(r.ReminderTime)
}

func (r *Reminder) MarkAsSent() {
	r.Status = ReminderStatusSent
}

func (r *Reminder) MarkAsFailed() {
	r.Status = ReminderStatusFailed
}
