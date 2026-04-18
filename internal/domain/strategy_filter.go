package domain

import "time"

type FilterStrategy interface {
	Filter(tasks []*Task) []*Task
}

type ByStatusFilter struct {
	Status TaskStatus
}

func (f ByStatusFilter) Filter(tasks []*Task) []*Task {
	var result []*Task
	for _, t := range tasks {
		if t.Status == f.Status {
			result = append(result, t)
		}
	}
	return result
}

type ByTagFilter struct {
	TagID int64
}

func (f ByTagFilter) Filter(tasks []*Task) []*Task {
	var result []*Task
	for _, t := range tasks {
		if t.TagID != nil && *t.TagID == f.TagID {
			result = append(result, t)
		}
	}
	return result
}

type ByDateFilter struct {
	From time.Time
	To   time.Time
}

func (f ByDateFilter) Filter(tasks []*Task) []*Task {
	var result []*Task
	for _, t := range tasks {
		if t.DueDate != nil && !t.DueDate.Before(f.From) && !t.DueDate.After(f.To) {
			result = append(result, t)
		}
	}
	return result
}
