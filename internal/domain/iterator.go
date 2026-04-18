package domain

type Iterator interface {
	HasNext() bool
	Next() *Task
}

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
