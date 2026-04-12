package domain_test

import (
	"testing"
	"time"

	"taskflow/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestTaskIsOverdue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusTodo}
	assert.True(t, task.IsOverdue())
}

func TestTaskIsNotOverdueWhenDone(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusDone}
	assert.False(t, task.IsOverdue())
}

func TestTaskIsNotOverdueWhenNoDueDate(t *testing.T) {
	task := domain.Task{Status: domain.TaskStatusTodo}
	assert.False(t, task.IsOverdue())
}
