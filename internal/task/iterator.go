package task

import "taskflow/internal/domain"

// Iterator is the Iterator pattern: sequential access to a collection
// without exposing the underlying slice.
type Iterator interface {
	HasNext() bool
	Next() *domain.Task
}

// SliceIterator walks a []*domain.Task slice.
type SliceIterator struct {
	tasks []*domain.Task
	index int
}

func NewSliceIterator(tasks []*domain.Task) *SliceIterator {
	return &SliceIterator{tasks: tasks}
}

func (it *SliceIterator) HasNext() bool {
	return it.index < len(it.tasks)
}

func (it *SliceIterator) Next() *domain.Task {
	t := it.tasks[it.index]
	it.index++
	return t
}
