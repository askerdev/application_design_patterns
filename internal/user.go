package domain

import "time"

type User struct {
	ID             int64
	Username       string
	TelegramID     int64
	TelegramChatID int64
	CreatedAt      time.Time
}
