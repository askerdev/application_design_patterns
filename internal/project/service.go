package project

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// Service defines operations on projects.
type Service interface {
	List(userID int64) ([]*domain.Project, error)
	Create(p *domain.Project) error
	Update(p *domain.Project) error
	Delete(id int64) error
}

type service struct{ repo repository.ProjectRepository }

// NewService returns a Service backed by repo.
func NewService(repo repository.ProjectRepository) Service { return &service{repo: repo} }

func (s *service) List(userID int64) ([]*domain.Project, error) {
	return s.repo.GetAllByUser(userID)
}
func (s *service) Create(p *domain.Project) error { return s.repo.Create(p) }
func (s *service) Update(p *domain.Project) error { return s.repo.Update(p) }
func (s *service) Delete(id int64) error           { return s.repo.Delete(id) }
