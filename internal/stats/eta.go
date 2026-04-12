package stats

import (
	"errors"

	"taskflow/internal/domain"
)

// CalculateETA estimates time to complete remaining tasks based on pomodoro history.
//
// Formula:
//   avg_pomodoro_min      = sum(session.WorkDuration) / count(sessions)
//   avg_tasks_per_session = sum(completedPerSession[i]) / count(sessions)
//   ETA_minutes           = (remaining / avg_tasks_per_session) * avg_pomodoro_min
//
// If avg_tasks_per_session == 0, falls back to 1 task/session.
// Returns ETA in minutes, or error if no session history exists.
func CalculateETA(sessions []*domain.PomodoroSession, completedPerSession []int, remaining int) (float64, error) {
	if len(sessions) == 0 {
		return 0, errors.New("eta: no session history — cannot estimate")
	}
	if remaining == 0 {
		return 0, nil
	}

	totalDuration := 0
	totalCompleted := 0
	for i, s := range sessions {
		totalDuration += s.WorkDuration
		if i < len(completedPerSession) {
			totalCompleted += completedPerSession[i]
		}
	}

	n := float64(len(sessions))
	avgPomodoro := float64(totalDuration) / n
	avgTasksPerSession := float64(totalCompleted) / n

	if avgTasksPerSession == 0 {
		avgTasksPerSession = 1
	}

	return (float64(remaining) / avgTasksPerSession) * avgPomodoro, nil
}
