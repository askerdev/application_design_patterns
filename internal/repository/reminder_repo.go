// internal/repository/reminder_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type ReminderRepo struct{ db *sql.DB }

func NewReminderRepo(db *sql.DB) *ReminderRepo { return &ReminderRepo{db: db} }

func (r *ReminderRepo) Create(rem *domain.Reminder) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO reminders (user_id, project_id, tag_id, content, reminder_time, status, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?)`,
		rem.UserID, rem.ProjectID, rem.TagID, rem.Content, rem.ReminderTime, rem.Status, now, now,
	)
	if err != nil {
		return err
	}
	rem.ID, err = res.LastInsertId()
	rem.CreatedAt = now
	rem.UpdatedAt = now
	return err
}

func (r *ReminderRepo) GetByID(id int64) (*domain.Reminder, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, project_id, tag_id, content, reminder_time, status, created_at, updated_at FROM reminders WHERE id=?`, id,
	)
	rem := &domain.Reminder{}
	err := row.Scan(&rem.ID, &rem.UserID, &rem.ProjectID, &rem.TagID, &rem.Content, &rem.ReminderTime, &rem.Status, &rem.CreatedAt, &rem.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reminder not found")
	}
	return rem, err
}

func (r *ReminderRepo) GetAllByUser(userID int64) ([]*domain.Reminder, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, content, reminder_time, status, created_at, updated_at FROM reminders WHERE user_id=? ORDER BY reminder_time`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (r *ReminderRepo) GetPending() ([]*domain.Reminder, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, content, reminder_time, status, created_at, updated_at FROM reminders WHERE status='PENDING' AND reminder_time <= ?`,
		time.Now(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (r *ReminderRepo) Update(rem *domain.Reminder) error {
	_, err := r.db.Exec(
		`UPDATE reminders SET status=?, updated_at=? WHERE id=?`,
		rem.Status, rem.UpdatedAt, rem.ID,
	)
	return err
}

func (r *ReminderRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM reminders WHERE id=?`, id)
	return err
}

func scanReminders(rows *sql.Rows) ([]*domain.Reminder, error) {
	var result []*domain.Reminder
	for rows.Next() {
		rem := &domain.Reminder{}
		if err := rows.Scan(&rem.ID, &rem.UserID, &rem.ProjectID, &rem.TagID, &rem.Content, &rem.ReminderTime, &rem.Status, &rem.CreatedAt, &rem.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, rem)
	}
	return result, rows.Err()
}
