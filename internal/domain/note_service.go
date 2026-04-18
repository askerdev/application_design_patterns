package domain

type NoteService interface {
	List(userID int64) ([]*Note, error)
	Create(n *Note) error
	Update(n *Note) error
	Delete(id int64) error
}

type noteService struct{ repo NoteRepository }

func NewNoteService(repo NoteRepository) NoteService { return &noteService{repo: repo} }

func (s *noteService) List(userID int64) ([]*Note, error) { return s.repo.GetAllByUser(userID) }
func (s *noteService) Create(n *Note) error               { return s.repo.Create(n) }
func (s *noteService) Update(n *Note) error               { return s.repo.Update(n) }
func (s *noteService) Delete(id int64) error              { return s.repo.Delete(id) }
