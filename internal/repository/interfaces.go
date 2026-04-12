// internal/repository/interfaces.go
package repository

import "taskflow/internal/domain"

type TaskRepository interface {
	Create(task *domain.Task) error
	GetByID(id int64) (*domain.Task, error)
	GetAllByUser(userID int64) ([]*domain.Task, error)
	GetByProject(projectID int64) ([]*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id int64) error
}

type ProjectRepository interface {
	Create(p *domain.Project) error
	GetByID(id int64) (*domain.Project, error)
	GetAllByUser(userID int64) ([]*domain.Project, error)
	Update(p *domain.Project) error
	Delete(id int64) error
}

type NoteRepository interface {
	Create(n *domain.Note) error
	GetByID(id int64) (*domain.Note, error)
	GetAllByUser(userID int64) ([]*domain.Note, error)
	Update(n *domain.Note) error
	Delete(id int64) error
}

type ReminderRepository interface {
	Create(r *domain.Reminder) error
	GetByID(id int64) (*domain.Reminder, error)
	GetAllByUser(userID int64) ([]*domain.Reminder, error)
	GetPending() ([]*domain.Reminder, error)
	Update(r *domain.Reminder) error
	Delete(id int64) error
}

type PomodoroRepository interface {
	Create(s *domain.PomodoroSession) error
	GetByID(id int64) (*domain.PomodoroSession, error)
	GetAllByUser(userID int64) ([]*domain.PomodoroSession, error)
	GetCompletedByProject(projectID int64) ([]*domain.PomodoroSession, error)
	Update(s *domain.PomodoroSession) error
}

type TagRepository interface {
	Create(t *domain.Tag) error
	GetByID(id int64) (*domain.Tag, error)
	GetAllByUser(userID int64) ([]*domain.Tag, error)
	Delete(id int64) error
}

type UserRepository interface {
	Create(u *domain.User) error
	GetByID(id int64) (*domain.User, error)
	GetFirst() (*domain.User, error)
}
