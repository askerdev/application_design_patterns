package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type ProjectRepo struct{ db *sql.DB }

func NewProjectRepo(db *sql.DB) *ProjectRepo { return &ProjectRepo{db: db} }

func (r *ProjectRepo) Create(p *domain.Project) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO projects (user_id, name, description, status, due_date, created_at) VALUES (?,?,?,?,?,?)`,
		p.UserID, p.Name, p.Description, p.Status, p.DueDate, now,
	)
	if err != nil {
		return err
	}
	p.ID, err = res.LastInsertId()
	p.CreatedAt = now
	return err
}

func (r *ProjectRepo) GetByID(id int64) (*domain.Project, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, name, description, status, due_date, created_at FROM projects WHERE id=?`, id,
	)
	p := &domain.Project{}
	err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.DueDate, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	return p, err
}

func (r *ProjectRepo) GetAllByUser(userID int64) ([]*domain.Project, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, description, status, due_date, created_at FROM projects WHERE user_id=? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []*domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.DueDate, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Update(p *domain.Project) error {
	_, err := r.db.Exec(
		`UPDATE projects SET name=?, description=?, status=?, due_date=? WHERE id=?`,
		p.Name, p.Description, p.Status, p.DueDate, p.ID,
	)
	return err
}

func (r *ProjectRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id=?`, id)
	return err
}
