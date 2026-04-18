package domain

type TaskService interface {
	List(userID int64) ([]*Task, error)
	ListByProject(projectID int64) ([]*Task, error)
	Iterate(userID int64) Iterator
	Create(t *Task) error
	Update(t *Task) error
	Delete(id int64) error
}

type taskService struct{ repo TaskRepository }

func NewTaskService(repo TaskRepository) TaskService { return &taskService{repo: repo} }

func (s *taskService) List(userID int64) ([]*Task, error)       { return s.repo.GetAllByUser(userID) }
func (s *taskService) ListByProject(pid int64) ([]*Task, error) { return s.repo.GetByProject(pid) }
func (s *taskService) Iterate(userID int64) Iterator {
	tasks, _ := s.repo.GetAllByUser(userID)
	return NewSliceIterator(tasks)
}
func (s *taskService) Create(t *Task) error  { return s.repo.Create(t) }
func (s *taskService) Update(t *Task) error  { return s.repo.Update(t) }
func (s *taskService) Delete(id int64) error { return s.repo.Delete(id) }
