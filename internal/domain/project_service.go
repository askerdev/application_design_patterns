package domain

// ProjectService defines operations on projects.
type ProjectService interface {
	List(userID int64) ([]*Project, error)
	Create(p *Project) error
	Update(p *Project) error
	Delete(id int64) error
}

type projectService struct{ repo ProjectRepository }

// NewProjectService returns a ProjectService backed by repo.
func NewProjectService(repo ProjectRepository) ProjectService { return &projectService{repo: repo} }

func (s *projectService) List(userID int64) ([]*Project, error) { return s.repo.GetAllByUser(userID) }
func (s *projectService) Create(p *Project) error               { return s.repo.Create(p) }
func (s *projectService) Update(p *Project) error               { return s.repo.Update(p) }
func (s *projectService) Delete(id int64) error                 { return s.repo.Delete(id) }
