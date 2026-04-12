package domain

import "time"

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "ACTIVE"
	ProjectStatusArchived ProjectStatus = "ARCHIVED"
	ProjectStatusDone     ProjectStatus = "DONE"
)

type Project struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	Status      ProjectStatus
	DueDate     *time.Time
	CreatedAt   time.Time
}
