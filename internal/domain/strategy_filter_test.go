package domain_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"taskflow/internal/domain"
)

func makeFilterTasks() []*domain.Task {
	tag1 := int64(1)
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(24 * time.Hour)
	return []*domain.Task{
		{ID: 1, Status: domain.TaskStatusTodo, Priority: domain.PriorityHigh, TagID: &tag1, DueDate: &past},
		{ID: 2, Status: domain.TaskStatusDone, Priority: domain.PriorityLow, DueDate: &future},
		{ID: 3, Status: domain.TaskStatusInProgress, Priority: domain.PriorityMedium, TagID: &tag1},
	}
}

func TestByStatusFilter(t *testing.T) {
	f := domain.ByStatusFilter{Status: domain.TaskStatusTodo}
	result := f.Filter(makeFilterTasks())
	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("expected task ID 1, got %d", result[0].ID)
	}
}

func TestByTagFilter(t *testing.T) {
	f := domain.ByTagFilter{TagID: 1}
	result := f.Filter(makeFilterTasks())
	if len(result) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result))
	}
	wantIDs := []int64{1, 3}
	gotIDs := []int64{result[0].ID, result[1].ID}
	if diff := cmp.Diff(wantIDs, gotIDs); diff != "" {
		t.Errorf("IDs mismatch (-want +got):\n%s", diff)
	}
}

func TestByDateFilter(t *testing.T) {
	from := time.Now().Add(-2 * time.Hour)
	to := time.Now()
	f := domain.ByDateFilter{From: from, To: to}
	result := f.Filter(makeFilterTasks())
	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("expected task ID 1, got %d", result[0].ID)
	}
}

func TestByStatusFilterInProgress(t *testing.T) {
	f := domain.ByStatusFilter{Status: domain.TaskStatusInProgress}
	result := f.Filter(makeFilterTasks())
	if len(result) != 1 {
		t.Fatalf("expected 1 IN_PROGRESS task, got %d", len(result))
	}
}
