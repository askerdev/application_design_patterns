package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	domain "taskflow/internal"
)

type NoteRepo struct{ db *sql.DB }

func NewNoteRepo(db *sql.DB) *NoteRepo { return &NoteRepo{db: db} }

func (r *NoteRepo) Create(n *domain.Note) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO notes (user_id, project_id, tag_id, title, content, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		n.UserID, n.ProjectID, n.TagID, n.Title, n.Content, now, now,
	)
	if err != nil {
		return err
	}
	n.ID, err = res.LastInsertId()
	n.CreatedAt, n.UpdatedAt = now, now
	return err
}

func (r *NoteRepo) GetByID(id int64) (*domain.Note, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, project_id, tag_id, title, content, created_at, updated_at FROM notes WHERE id=?`, id,
	)
	n := &domain.Note{}
	err := row.Scan(&n.ID, &n.UserID, &n.ProjectID, &n.TagID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note not found")
	}
	return n, err
}

func (r *NoteRepo) GetAllByUser(userID int64) ([]*domain.Note, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, title, content, created_at, updated_at FROM notes WHERE user_id=? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []*domain.Note
	for rows.Next() {
		n := &domain.Note{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.ProjectID, &n.TagID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepo) Update(n *domain.Note) error {
	n.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE notes SET project_id=?, tag_id=?, title=?, content=?, updated_at=? WHERE id=?`,
		n.ProjectID, n.TagID, n.Title, n.Content, n.UpdatedAt, n.ID,
	)
	return err
}

func (r *NoteRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM notes WHERE id=?`, id)
	return err
}
