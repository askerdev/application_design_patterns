package domain

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "TODO"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusDone       TaskStatus = "DONE"
)

type Priority string

const (
	PriorityHigh   Priority = "HIGH"
	PriorityMedium Priority = "MEDIUM"
	PriorityLow    Priority = "LOW"
)

type Task struct {
	ID        int64
	UserID    int64
	ProjectID *int64
	TagID     *int64
	Content   string
	Status    TaskStatus
	Priority  Priority
	DueDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == TaskStatusDone {
		return false
	}
	return time.Now().After(*t.DueDate)
}

func (t *Task) Complete() {
	t.Status = TaskStatusDone
	t.UpdatedAt = time.Now()
}
