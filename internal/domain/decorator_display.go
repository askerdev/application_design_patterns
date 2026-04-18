package domain

import "fmt"

// Renderer is the component interface for the Decorator pattern.
type Renderer interface {
	Render(t *Task) string
}

// BaseRenderer returns the basic description of a task.
type BaseRenderer struct{}

func (r *BaseRenderer) Render(t *Task) string {
	return fmt.Sprintf("%s · %s", t.Priority, t.Status)
}

// PriorityDecorator wraps a Renderer and prepends a priority badge.
type PriorityDecorator struct {
	wrapped Renderer
}

func NewPriorityDecorator(r Renderer) *PriorityDecorator {
	return &PriorityDecorator{wrapped: r}
}

func (d *PriorityDecorator) Render(t *Task) string {
	badge := ""
	switch t.Priority {
	case PriorityHigh:
		badge = "!!! "
	case PriorityMedium:
		badge = "!!  "
	case PriorityLow:
		badge = "·   "
	}
	return badge + d.wrapped.Render(t)
}

// OverdueDecorator wraps a Renderer and appends an overdue warning.
type OverdueDecorator struct {
	wrapped Renderer
}

func NewOverdueDecorator(r Renderer) *OverdueDecorator {
	return &OverdueDecorator{wrapped: r}
}

func (d *OverdueDecorator) Render(t *Task) string {
	s := d.wrapped.Render(t)
	if t.IsOverdue() {
		s += " [OVERDUE]"
	}
	return s
}
