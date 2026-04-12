package note

import (
	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// Service defines operations on notes.
type Service interface {
	List(userID int64) ([]*domain.Note, error)
	Create(n *domain.Note) error
	Update(n *domain.Note) error
	Delete(id int64) error
}

type service struct{ repo repository.NoteRepository }

// NewService returns a Service backed by repo.
func NewService(repo repository.NoteRepository) Service { return &service{repo: repo} }

func (s *service) List(userID int64) ([]*domain.Note, error) { return s.repo.GetAllByUser(userID) }
func (s *service) Create(n *domain.Note) error               { return s.repo.Create(n) }
func (s *service) Update(n *domain.Note) error               { return s.repo.Update(n) }
func (s *service) Delete(id int64) error                     { return s.repo.Delete(id) }
