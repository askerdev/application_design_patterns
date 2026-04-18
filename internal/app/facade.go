package app

import (
	domain "taskflow/internal/domain"
)

// AppFacade is the Facade pattern: single simplified entry point
// for common operations across multiple subsystems.
type AppFacade struct {
	tasks    domain.TaskService
	projects domain.ProjectService
	notes    domain.NoteService
	user     *domain.User
}

func NewAppFacade(tasks domain.TaskService, projects domain.ProjectService, notes domain.NoteService, user *domain.User) *AppFacade {
	return &AppFacade{tasks: tasks, projects: projects, notes: notes, user: user}
}

func (f *AppFacade) AddTask(content, priority string) (*domain.Task, error) {
	factory := domain.NewEntityFactory()
	t := factory.CreateTask(f.user.ID, content, domain.Priority(priority), nil, nil)
	if err := f.tasks.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (f *AppFacade) ListTasks() ([]*domain.Task, error) {
	return f.tasks.List(f.user.ID)
}

func (f *AppFacade) AddProject(name, description string) (*domain.Project, error) {
	factory := domain.NewEntityFactory()
	p := factory.CreateProject(f.user.ID, name, description)
	if err := f.projects.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (f *AppFacade) ListProjects() ([]*domain.Project, error) {
	return f.projects.List(f.user.ID)
}
