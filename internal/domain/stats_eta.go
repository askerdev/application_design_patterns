package domain

import (
	"errors"
)

func CalculateETA(sessions []*PomodoroSession, completedPerSession []int, remaining int) (float64, error) {
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
