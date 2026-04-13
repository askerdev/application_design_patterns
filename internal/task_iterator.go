package domain

// Iterator is the Iterator pattern: sequential access to a collection
// without exposing the underlying slice.
type Iterator interface {
	HasNext() bool
	Next() *Task
}

// SliceIterator walks a []*Task slice.
type SliceIterator struct {
	tasks []*Task
	index int
}

func NewSliceIterator(tasks []*Task) *SliceIterator {
	return &SliceIterator{tasks: tasks}
}

func (it *SliceIterator) HasNext() bool {
	return it.index < len(it.tasks)
}

func (it *SliceIterator) Next() *Task {
	t := it.tasks[it.index]
	it.index++
	return t
}
