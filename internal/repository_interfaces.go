package domain

type TaskRepository interface {
	Create(task *Task) error
	GetByID(id int64) (*Task, error)
	GetAllByUser(userID int64) ([]*Task, error)
	GetByProject(projectID int64) ([]*Task, error)
	Update(task *Task) error
	Delete(id int64) error
}

type ProjectRepository interface {
	Create(p *Project) error
	GetByID(id int64) (*Project, error)
	GetAllByUser(userID int64) ([]*Project, error)
	Update(p *Project) error
	Delete(id int64) error
}

type NoteRepository interface {
	Create(n *Note) error
	GetByID(id int64) (*Note, error)
	GetAllByUser(userID int64) ([]*Note, error)
	Update(n *Note) error
	Delete(id int64) error
}

type ReminderRepository interface {
	Create(r *Reminder) error
	GetByID(id int64) (*Reminder, error)
	GetAllByUser(userID int64) ([]*Reminder, error)
	GetPending() ([]*Reminder, error)
	Update(r *Reminder) error
	Delete(id int64) error
}

type PomodoroRepository interface {
	Create(s *PomodoroSession) error
	GetByID(id int64) (*PomodoroSession, error)
	GetAllByUser(userID int64) ([]*PomodoroSession, error)
	GetCompletedByProject(projectID int64) ([]*PomodoroSession, error)
	Update(s *PomodoroSession) error
}

type TagRepository interface {
	Create(t *Tag) error
	GetByID(id int64) (*Tag, error)
	GetAllByUser(userID int64) ([]*Tag, error)
	Delete(id int64) error
}

type UserRepository interface {
	Create(u *User) error
	GetByID(id int64) (*User, error)
	GetFirst() (*User, error)
}
