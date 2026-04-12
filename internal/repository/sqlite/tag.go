package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type TagRepo struct{ db *sql.DB }

func NewTagRepo(db *sql.DB) *TagRepo { return &TagRepo{db: db} }

func (r *TagRepo) Create(t *domain.Tag) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO tags (user_id, name, color, created_at) VALUES (?,?,?,?)`,
		t.UserID, t.Name, t.Color, now,
	)
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	t.CreatedAt = now
	return err
}

func (r *TagRepo) GetByID(id int64) (*domain.Tag, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, name, color, created_at FROM tags WHERE id=?`, id,
	)
	t := &domain.Tag{}
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found")
	}
	return t, err
}

func (r *TagRepo) GetAllByUser(userID int64) ([]*domain.Tag, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, color, created_at FROM tags WHERE user_id=?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*domain.Tag
	for rows.Next() {
		t := &domain.Tag{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *TagRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM tags WHERE id=?`, id)
	return err
}
