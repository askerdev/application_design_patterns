package math

import (
	"errors"

	"taskflow/internal/domain"
)

// CalculateETA вычисляет оценочное время завершения оставшихся задач.
//
// Формула:
//   avg_pomodoro_min        = sum(session.WorkDuration) / count(sessions)
//   avg_tasks_per_session   = sum(completedPerSession[i]) / count(sessions)
//   ETA_minutes             = (remaining / avg_tasks_per_session) * avg_pomodoro_min
//
// Если avg_tasks_per_session == 0 (задачи в сессиях не закрывались), используется
// fallback: 1 задача/сессия.
//
// sessions             — завершённые помодоро-сессии (обычно по проекту)
// completedPerSession  — кол-во задач, закрытых в sessions[i]
// remaining            — кол-во незакрытых задач
// Возвращает ETA в минутах или ошибку если история пуста.
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
		avgTasksPerSession = 1 // fallback: 1 задача/сессия
	}

	return (float64(remaining) / avgTasksPerSession) * avgPomodoro, nil
}
