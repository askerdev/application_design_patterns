// internal/repository/pomodoro_repo.go
package repository

import (
	"database/sql"
	"time"

	"taskflow/internal/domain"
)

type PomodoroRepo struct{ db *sql.DB }

func NewPomodoroRepo(db *sql.DB) *PomodoroRepo { return &PomodoroRepo{db: db} }

func (r *PomodoroRepo) Create(s *domain.PomodoroSession) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO pomodoro_sessions (user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		s.UserID, s.ProjectID, s.StartTime, s.WorkDuration, s.FinishTime, s.RemainingTime, s.State, now,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	s.CreatedAt = now
	return err
}

func (r *PomodoroRepo) GetByID(id int64) (*domain.PomodoroSession, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at FROM pomodoro_sessions WHERE id=?`, id,
	)
	s := &domain.PomodoroSession{}
	err := row.Scan(&s.ID, &s.UserID, &s.ProjectID, &s.StartTime, &s.WorkDuration, &s.FinishTime, &s.RemainingTime, &s.State, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return s, err
}

func (r *PomodoroRepo) GetAllByUser(userID int64) ([]*domain.PomodoroSession, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at FROM pomodoro_sessions WHERE user_id=? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (r *PomodoroRepo) GetCompletedByProject(projectID int64) ([]*domain.PomodoroSession, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at FROM pomodoro_sessions WHERE project_id=? AND state='COMPLETED'`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (r *PomodoroRepo) Update(s *domain.PomodoroSession) error {
	_, err := r.db.Exec(
		`UPDATE pomodoro_sessions SET start_time=?, finish_time=?, remaining_time=?, state=? WHERE id=?`,
		s.StartTime, s.FinishTime, s.RemainingTime, s.State, s.ID,
	)
	return err
}

func scanSessions(rows *sql.Rows) ([]*domain.PomodoroSession, error) {
	var result []*domain.PomodoroSession
	for rows.Next() {
		s := &domain.PomodoroSession{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.ProjectID, &s.StartTime, &s.WorkDuration, &s.FinishTime, &s.RemainingTime, &s.State, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
