package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	domain "taskflow/internal"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(u *domain.User) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO users (username, telegram_id, telegram_chat_id, created_at) VALUES (?,?,?,?)`,
		u.Username, u.TelegramID, u.TelegramChatID, now,
	)
	if err != nil {
		return err
	}
	u.ID, err = res.LastInsertId()
	u.CreatedAt = now
	return err
}

func (r *UserRepo) GetByID(id int64) (*domain.User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, telegram_id, telegram_chat_id, created_at FROM users WHERE id=?`, id,
	)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Username, &u.TelegramID, &u.TelegramChatID, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepo) GetFirst() (*domain.User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, telegram_id, telegram_chat_id, created_at FROM users ORDER BY id LIMIT 1`,
	)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Username, &u.TelegramID, &u.TelegramChatID, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no users found")
	}
	return u, err
}
