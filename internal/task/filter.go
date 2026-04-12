package task

import (
	"time"

	"taskflow/internal/domain"
)

// FilterStrategy defines how tasks are filtered.
type FilterStrategy interface {
	Filter(tasks []*domain.Task) []*domain.Task
}

// ByStatusFilter filters tasks by status.
type ByStatusFilter struct {
	Status domain.TaskStatus
}

func (f ByStatusFilter) Filter(tasks []*domain.Task) []*domain.Task {
	var result []*domain.Task
	for _, t := range tasks {
		if t.Status == f.Status {
			result = append(result, t)
		}
	}
	return result
}

// ByTagFilter filters tasks by tag.
type ByTagFilter struct {
	TagID int64
}

func (f ByTagFilter) Filter(tasks []*domain.Task) []*domain.Task {
	var result []*domain.Task
	for _, t := range tasks {
		if t.TagID != nil && *t.TagID == f.TagID {
			result = append(result, t)
		}
	}
	return result
}

// ByDateFilter filters tasks by due date range.
type ByDateFilter struct {
	From time.Time
	To   time.Time
}

func (f ByDateFilter) Filter(tasks []*domain.Task) []*domain.Task {
	var result []*domain.Task
	for _, t := range tasks {
		if t.DueDate != nil && !t.DueDate.Before(f.From) && !t.DueDate.After(f.To) {
			result = append(result, t)
		}
	}
	return result
}
