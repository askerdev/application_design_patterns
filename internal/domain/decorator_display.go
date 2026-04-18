package domain

import "fmt"

type Renderer interface {
	Render(t *Task) string
}

type BaseRenderer struct{}

func (r *BaseRenderer) Render(t *Task) string {
	return fmt.Sprintf("%s · %s", t.Priority, t.Status)
}

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
