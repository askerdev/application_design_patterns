package domain_test

import (
	"testing"
	"time"

	"taskflow/internal/domain"
)

func TestTaskIsOverdue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusTodo}
	if !task.IsOverdue() {
		t.Error("expected task to be overdue")
	}
}

func TestTaskIsNotOverdueWhenDone(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusDone}
	if task.IsOverdue() {
		t.Error("completed task should not be overdue")
	}
}

func TestTaskIsNotOverdueWhenNoDueDate(t *testing.T) {
	task := domain.Task{Status: domain.TaskStatusTodo}
	if task.IsOverdue() {
		t.Error("task without due date should not be overdue")
	}
}
