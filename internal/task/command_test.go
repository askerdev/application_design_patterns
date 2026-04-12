package task_test

import (
	"testing"

	"taskflow/internal/domain"
	"taskflow/internal/task"
)

// stubService is a minimal in-memory Service for testing commands.
type stubService struct {
	tasks  map[int64]*domain.Task
	nextID int64
}

func newStubService() *stubService {
	return &stubService{tasks: make(map[int64]*domain.Task), nextID: 1}
}

func (s *stubService) List(userID int64) ([]*domain.Task, error) { return nil, nil }
func (s *stubService) ListByProject(projectID int64) ([]*domain.Task, error) {
	return nil, nil
}
func (s *stubService) Iterate(userID int64) task.Iterator { return task.NewSliceIterator(nil) }
func (s *stubService) Create(t *domain.Task) error {
	s.nextID++
	t.ID = s.nextID
	s.tasks[t.ID] = t
	return nil
}
func (s *stubService) Update(t *domain.Task) error {
	s.tasks[t.ID] = t
	return nil
}
func (s *stubService) Delete(id int64) error {
	delete(s.tasks, id)
	return nil
}

func TestCompleteTaskCommand_ExecuteAndUndo(t *testing.T) {
	svc := newStubService()
	tk := &domain.Task{ID: 1, Status: domain.TaskStatusTodo}
	svc.tasks[1] = tk

	cmd := task.NewCompleteTaskCommand(svc, tk)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if tk.Status != domain.TaskStatusDone {
		t.Errorf("expected DONE after Execute, got %s", tk.Status)
	}

	if err := cmd.Undo(); err != nil {
		t.Fatal(err)
	}
	if tk.Status != domain.TaskStatusTodo {
		t.Errorf("expected TODO after Undo, got %s", tk.Status)
	}
}

func TestCommandHistory_UndoEmpty(t *testing.T) {
	h := &task.CommandHistory{}
	if err := h.Undo(); err != nil {
		t.Errorf("Undo on empty history should not error, got %v", err)
	}
}

func TestDeleteTaskCommand_ExecuteAndUndo(t *testing.T) {
	svc := newStubService()
	tk := &domain.Task{ID: 5, UserID: 1, Content: "X", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	svc.tasks[5] = tk

	cmd := task.NewDeleteTaskCommand(svc, tk)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, ok := svc.tasks[5]; ok {
		t.Error("expected task deleted after Execute")
	}

	if err := cmd.Undo(); err != nil {
		t.Fatal(err)
	}
	// After undo re-create, task should be back (with new ID assigned by stubService)
	if len(svc.tasks) == 0 {
		t.Error("expected task restored after Undo")
	}
}
