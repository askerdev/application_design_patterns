package domain

type TagService interface {
	List(userID int64) ([]*Tag, error)
	Create(t *Tag) error
	Delete(id int64) error
}

type tagService struct{ repo TagRepository }

func NewTagService(repo TagRepository) TagService { return &tagService{repo: repo} }

func (s *tagService) List(userID int64) ([]*Tag, error) { return s.repo.GetAllByUser(userID) }
func (s *tagService) Create(t *Tag) error               { return s.repo.Create(t) }
func (s *tagService) Delete(id int64) error             { return s.repo.Delete(id) }
