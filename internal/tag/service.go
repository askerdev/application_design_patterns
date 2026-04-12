package tag

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// Service defines operations on tags.
type Service interface {
	List(userID int64) ([]*domain.Tag, error)
	Create(t *domain.Tag) error
	Delete(id int64) error
}

type service struct{ repo repository.TagRepository }

// NewService returns a Service backed by repo.
func NewService(repo repository.TagRepository) Service { return &service{repo: repo} }

func (s *service) List(userID int64) ([]*domain.Tag, error) { return s.repo.GetAllByUser(userID) }
func (s *service) Create(t *domain.Tag) error               { return s.repo.Create(t) }
func (s *service) Delete(id int64) error                    { return s.repo.Delete(id) }
