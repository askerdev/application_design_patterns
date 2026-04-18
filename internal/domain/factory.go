package domain

import "time"

type EntityFactory struct{}

func NewEntityFactory() *EntityFactory {
	return &EntityFactory{}
}

func (f *EntityFactory) CreateTask(userID int64, content string, priority Priority, projectID, tagID *int64, dueDate *time.Time) *Task {
	return &Task{
		UserID:    userID,
		Content:   content,
		Status:    TaskStatusTodo,
		Priority:  priority,
		ProjectID: projectID,
		TagID:     tagID,
		DueDate:   dueDate,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (f *EntityFactory) CreateNote(userID int64, title, content string, projectID, tagID *int64) *Note {
	return &Note{
		UserID:    userID,
		Title:     title,
		Content:   content,
		ProjectID: projectID,
		TagID:     tagID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (f *EntityFactory) CreateReminder(userID int64, content string, reminderTime time.Time, projectID, tagID *int64) *Reminder {
	return &Reminder{
		UserID:       userID,
		Content:      content,
		ReminderTime: reminderTime,
		Status:       ReminderStatusPending,
		ProjectID:    projectID,
		TagID:        tagID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (f *EntityFactory) CreateProject(userID int64, name, description string) *Project {
	return &Project{
		UserID:      userID,
		Name:        name,
		Description: description,
		Status:      ProjectStatusActive,
		CreatedAt:   time.Now(),
	}
}

func (f *EntityFactory) CreateTag(userID int64, name, color string) *Tag {
	if color == "" {
		color = "#ffffff"
	}
	return &Tag{
		UserID:    userID,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now(),
	}
}
