package sqlite_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"taskflow/internal/db"
	"taskflow/internal/domain"
	"taskflow/internal/repository/sqlite"
)

func setupTestDB(t *testing.T) *sqlite.TaskRepo {
	t.Helper()
	conn, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	if err := db.RunMigrations(conn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	conn.Exec("INSERT INTO users (id, username) VALUES (1, 'testuser')")
	return sqlite.NewTaskRepo(conn)
}

func TestTaskRepoCreate(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{
		UserID:   1,
		Content:  "Buy milk",
		Status:   domain.TaskStatusTodo,
		Priority: domain.PriorityMedium,
	}
	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.ID <= 0 {
		t.Errorf("expected task.ID > 0, got %d", task.ID)
	}
}

func TestTaskRepoGetByID(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{UserID: 1, Content: "Test task", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(task.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	opts := cmpopts.IgnoreFields(domain.Task{}, "CreatedAt", "UpdatedAt")
	if diff := cmp.Diff(task, got, opts); diff != "" {
		t.Errorf("GetByID mismatch (-want +got):\n%s", diff)
	}
}

func TestTaskRepoGetAllByUser(t *testing.T) {
	repo := setupTestDB(t)
	repo.Create(&domain.Task{UserID: 1, Content: "A", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow})
	repo.Create(&domain.Task{UserID: 1, Content: "B", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow})

	tasks, err := repo.GetAllByUser(1)
	if err != nil {
		t.Fatalf("GetAllByUser: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTaskRepoUpdate(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{UserID: 1, Content: "Old", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	task.Content = "New"
	task.Status = domain.TaskStatusDone
	if err := repo.Update(task); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(task.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	opts := cmpopts.IgnoreFields(domain.Task{}, "CreatedAt", "UpdatedAt")
	if diff := cmp.Diff(task, got, opts); diff != "" {
		t.Errorf("after Update mismatch (-want +got):\n%s", diff)
	}
}

func TestTaskRepoDelete(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{UserID: 1, Content: "Delete me", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(task.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := repo.GetByID(task.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
