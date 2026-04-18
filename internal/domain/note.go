package domain

import "time"

type Note struct {
	ID        int64
	UserID    int64
	ProjectID *int64
	TagID     *int64
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
