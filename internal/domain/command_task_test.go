package domain_test

import (
	"testing"

	"taskflow/internal/domain"
)

type stubTaskService struct {
	tasks  map[int64]*domain.Task
	nextID int64
}

func newStubTaskService() *stubTaskService {
	return &stubTaskService{tasks: make(map[int64]*domain.Task), nextID: 1}
}

func (s *stubTaskService) List(userID int64) ([]*domain.Task, error)       { return nil, nil }
func (s *stubTaskService) ListByProject(pid int64) ([]*domain.Task, error) { return nil, nil }
func (s *stubTaskService) Iterate(userID int64) domain.Iterator            { return domain.NewSliceIterator(nil) }
func (s *stubTaskService) Create(t *domain.Task) error {
	s.nextID++
	t.ID = s.nextID
	s.tasks[t.ID] = t
	return nil
}
func (s *stubTaskService) Update(t *domain.Task) error {
	s.tasks[t.ID] = t
	return nil
}
func (s *stubTaskService) Delete(id int64) error {
	delete(s.tasks, id)
	return nil
}

func TestCompleteTaskCommand_ExecuteAndUndo(t *testing.T) {
	svc := newStubTaskService()
	tk := &domain.Task{ID: 1, Status: domain.TaskStatusTodo}
	svc.tasks[1] = tk

	cmd := domain.NewCompleteTaskCommand(svc, tk)
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
	h := &domain.CommandHistory{}
	if err := h.Undo(); err != nil {
		t.Errorf("Undo on empty history should not error, got %v", err)
	}
}

func TestDeleteTaskCommand_ExecuteAndUndo(t *testing.T) {
	svc := newStubTaskService()
	tk := &domain.Task{ID: 5, UserID: 1, Content: "X", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	svc.tasks[5] = tk

	cmd := domain.NewDeleteTaskCommand(svc, tk)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, ok := svc.tasks[5]; ok {
		t.Error("expected task deleted after Execute")
	}

	if err := cmd.Undo(); err != nil {
		t.Fatal(err)
	}
	if len(svc.tasks) == 0 {
		t.Error("expected task restored after Undo")
	}
}
