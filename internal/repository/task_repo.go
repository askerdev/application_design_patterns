// internal/repository/task_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type TaskRepo struct{ db *sql.DB }

func NewTaskRepo(db *sql.DB) *TaskRepo { return &TaskRepo{db: db} }

func (r *TaskRepo) Create(t *domain.Task) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO tasks (user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.UserID, t.ProjectID, t.TagID, t.Content, t.Status, t.Priority, t.DueDate, now, now,
	)
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	t.CreatedAt = now
	t.UpdatedAt = now
	return err
}

func (r *TaskRepo) GetByID(id int64) (*domain.Task, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE id = ?`, id,
	)
	return scanTask(row)
}

func (r *TaskRepo) GetAllByUser(userID int64) ([]*domain.Task, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) GetByProject(projectID int64) ([]*domain.Task, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE project_id = ? ORDER BY created_at DESC`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) Update(t *domain.Task) error {
	t.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE tasks SET project_id=?, tag_id=?, content=?, status=?, priority=?, due_date=?, updated_at=? WHERE id=?`,
		t.ProjectID, t.TagID, t.Content, t.Status, t.Priority, t.DueDate, t.UpdatedAt, t.ID,
	)
	return err
}

func (r *TaskRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func scanTask(row *sql.Row) (*domain.Task, error) {
	t := &domain.Task{}
	err := row.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.TagID, &t.Content, &t.Status, &t.Priority,
		&t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	return t, err
}

func scanTasks(rows *sql.Rows) ([]*domain.Task, error) {
	var tasks []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.TagID, &t.Content, &t.Status, &t.Priority,
			&t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
