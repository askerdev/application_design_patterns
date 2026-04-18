package domain

import "time"

type SessionState string

const (
	SessionStateIdle      SessionState = "IDLE"
	SessionStateRunning   SessionState = "RUNNING"
	SessionStatePaused    SessionState = "PAUSED"
	SessionStateCompleted SessionState = "COMPLETED"
	SessionStateCancelled SessionState = "CANCELLED"
)

type PomodoroSession struct {
	ID            int64
	UserID        int64
	ProjectID     *int64
	StartTime     *time.Time
	WorkDuration  int
	FinishTime    *time.Time
	RemainingTime int
	State         SessionState
	CreatedAt     time.Time
}
