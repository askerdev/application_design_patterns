package task

import (
	"fmt"
	"taskflow/internal/domain"
)

// Renderer is the component interface for the Decorator pattern.
type Renderer interface {
	Render(t *domain.Task) string
}

// BaseRenderer returns the basic description of a task.
type BaseRenderer struct{}

func (r *BaseRenderer) Render(t *domain.Task) string {
	return fmt.Sprintf("%s · %s", t.Priority, t.Status)
}

// PriorityDecorator wraps a Renderer and prepends a priority badge.
type PriorityDecorator struct {
	wrapped Renderer
}

func NewPriorityDecorator(r Renderer) *PriorityDecorator {
	return &PriorityDecorator{wrapped: r}
}

func (d *PriorityDecorator) Render(t *domain.Task) string {
	badge := ""
	switch t.Priority {
	case domain.PriorityHigh:
		badge = "!!! "
	case domain.PriorityMedium:
		badge = "!!  "
	case domain.PriorityLow:
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

func (d *OverdueDecorator) Render(t *domain.Task) string {
	s := d.wrapped.Render(t)
	if t.IsOverdue() {
		s += " [OVERDUE]"
	}
	return s
}
