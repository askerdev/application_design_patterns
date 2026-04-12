package domain

import "time"

// EntityFactory provides methods to create domain entities with default values.
// Implements the Factory pattern for consistent object creation.
type EntityFactory struct{}

// NewEntityFactory creates a new instance of EntityFactory.
func NewEntityFactory() *EntityFactory {
	return &EntityFactory{}
}

// CreateTask creates a new Task with the given parameters and sets default values.
// The factory ensures consistent initialization across the application.
func (f *EntityFactory) CreateTask(userID int64, content string, priority Priority, projectID, tagID *int64) *Task {
	return &Task{
		UserID:    userID,
		Content:   content,
		Status:    TaskStatusTodo,
		Priority:  priority,
		ProjectID: projectID,
		TagID:     tagID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateNote creates a new Note with the given parameters and sets default values.
// The factory ensures consistent initialization across the application.
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

// CreateReminder creates a new Reminder with the given parameters and sets default values.
// The factory ensures consistent initialization across the application.
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

// CreateTag creates a new Tag with the given parameters and sets default values.
// The factory ensures consistent initialization across the application.
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
