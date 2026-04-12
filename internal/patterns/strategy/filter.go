package strategy

import (
	"time"

	"taskflow/internal/domain"
)

// FilterStrategy — стратегия фильтрации задач (паттерн Стратегия)
type FilterStrategy interface {
	Filter(tasks []*domain.Task) []*domain.Task
}

// ByStatusFilter фильтрует задачи по статусу
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

// ByTagFilter фильтрует задачи по тегу
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

// ByDateFilter фильтрует задачи по диапазону due_date
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
