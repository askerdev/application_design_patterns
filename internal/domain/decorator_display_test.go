package domain_test

import (
	"testing"
	"time"

	"taskflow/internal/domain"
)

func TestBaseRenderer(t *testing.T) {
	r := &domain.BaseRenderer{}
	tk := &domain.Task{Priority: domain.PriorityHigh, Status: domain.TaskStatusTodo}
	got := r.Render(tk)
	want := "HIGH · TODO"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPriorityDecorator_AddsHighBadge(t *testing.T) {
	r := domain.NewPriorityDecorator(&domain.BaseRenderer{})
	tk := &domain.Task{Priority: domain.PriorityHigh, Status: domain.TaskStatusTodo}
	got := r.Render(tk)
	if got[:4] != "!!! " {
		t.Errorf("expected high badge prefix, got %q", got)
	}
}

func TestOverdueDecorator_AppendsWhenOverdue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	r := domain.NewOverdueDecorator(&domain.BaseRenderer{})
	tk := &domain.Task{
		Priority: domain.PriorityLow,
		Status:   domain.TaskStatusTodo,
		DueDate:  &past,
	}
	got := r.Render(tk)
	if got[len(got)-9:] != "[OVERDUE]" {
		t.Errorf("expected overdue suffix, got %q", got)
	}
}

func TestOverdueDecorator_NoSuffixWhenDone(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	r := domain.NewOverdueDecorator(&domain.BaseRenderer{})
	tk := &domain.Task{
		Priority: domain.PriorityLow,
		Status:   domain.TaskStatusDone,
		DueDate:  &past,
	}
	got := r.Render(tk)
	if len(got) >= 9 && got[len(got)-9:] == " [OVERDUE]" {
		t.Errorf("done task should not show overdue, got %q", got)
	}
}
