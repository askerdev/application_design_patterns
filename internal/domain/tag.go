package domain

import "time"

type Tag struct {
	ID        int64
	UserID    int64
	Name      string
	Color     string
	CreatedAt time.Time
}
