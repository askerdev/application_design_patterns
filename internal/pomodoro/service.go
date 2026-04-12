package pomodoro

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// Service defines CRUD operations on pomodoro sessions.
type Service interface {
	List(userID int64) ([]*domain.PomodoroSession, error)
	ListCompletedByProject(projectID int64) ([]*domain.PomodoroSession, error)
	Create(s *domain.PomodoroSession) error
	Update(s *domain.PomodoroSession) error
}

type svc struct{ repo repository.PomodoroRepository }

// NewService returns a Service backed by repo.
func NewService(repo repository.PomodoroRepository) Service { return &svc{repo: repo} }

func (s *svc) List(userID int64) ([]*domain.PomodoroSession, error) {
	return s.repo.GetAllByUser(userID)
}
func (s *svc) ListCompletedByProject(projectID int64) ([]*domain.PomodoroSession, error) {
	return s.repo.GetCompletedByProject(projectID)
}
func (s *svc) Create(ps *domain.PomodoroSession) error { return s.repo.Create(ps) }
func (s *svc) Update(ps *domain.PomodoroSession) error { return s.repo.Update(ps) }
