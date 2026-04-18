package domain_test

import (
	"testing"
	"time"

	"taskflow/internal/domain"
)

// --- Task tests ---

func TestTaskIsOverdue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusTodo}
	if !task.IsOverdue() {
		t.Error("expected task with past due date and TODO status to be overdue")
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

func TestTaskComplete(t *testing.T) {
	task := domain.Task{Status: domain.TaskStatusTodo}
	task.Complete()
	if task.Status != domain.TaskStatusDone {
		t.Errorf("expected status DONE, got %s", task.Status)
	}
	if task.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set after Complete()")
	}
}

// --- Reminder tests ---

func TestReminderIsReady(t *testing.T) {
	past := time.Now().Add(-1 * time.Minute)
	r := domain.Reminder{Status: domain.ReminderStatusPending, ReminderTime: past}
	if !r.IsReady() {
		t.Error("expected reminder with past time and PENDING status to be ready")
	}
}

func TestReminderIsNotReadyWhenFuture(t *testing.T) {
	future := time.Now().Add(10 * time.Minute)
	r := domain.Reminder{Status: domain.ReminderStatusPending, ReminderTime: future}
	if r.IsReady() {
		t.Error("reminder with future time should not be ready")
	}
}

func TestReminderIsNotReadyWhenAlreadySent(t *testing.T) {
	past := time.Now().Add(-1 * time.Minute)
	r := domain.Reminder{Status: domain.ReminderStatusSent, ReminderTime: past}
	if r.IsReady() {
		t.Error("already sent reminder should not be ready")
	}
}

func TestReminderMarkAsSent(t *testing.T) {
	r := domain.Reminder{Status: domain.ReminderStatusPending}
	r.MarkAsSent()
	if r.Status != domain.ReminderStatusSent {
		t.Errorf("expected status SENT, got %s", r.Status)
	}
	if r.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set after MarkAsSent()")
	}
}

func TestReminderMarkAsFailed(t *testing.T) {
	r := domain.Reminder{Status: domain.ReminderStatusPending}
	r.MarkAsFailed()
	if r.Status != domain.ReminderStatusFailed {
		t.Errorf("expected status FAILED, got %s", r.Status)
	}
	if r.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set after MarkAsFailed()")
	}
}
