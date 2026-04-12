package task

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// Service defines operations on tasks.
type Service interface {
	List(userID int64) ([]*domain.Task, error)
	ListByProject(projectID int64) ([]*domain.Task, error)
	Create(t *domain.Task) error
	Update(t *domain.Task) error
	Delete(id int64) error
}

type service struct{ repo repository.TaskRepository }

// NewService returns a Service backed by repo.
func NewService(repo repository.TaskRepository) Service { return &service{repo: repo} }

func (s *service) List(userID int64) ([]*domain.Task, error) {
	return s.repo.GetAllByUser(userID)
}
func (s *service) ListByProject(projectID int64) ([]*domain.Task, error) {
	return s.repo.GetByProject(projectID)
}
func (s *service) Create(t *domain.Task) error { return s.repo.Create(t) }
func (s *service) Update(t *domain.Task) error { return s.repo.Update(t) }
func (s *service) Delete(id int64) error       { return s.repo.Delete(id) }
