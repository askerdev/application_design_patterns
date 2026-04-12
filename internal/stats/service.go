package stats

import (
	"taskflow/internal/repository"
)

// Summary is an aggregated overview of a user's task state.
type Summary struct {
	Total      int
	Done       int
	Todo       int
	InProgress int
	Projects   []ProjectETA
}

// ProjectETA holds per-project task progress and estimated time to completion.
type ProjectETA struct {
	Name      string
	Remaining int
	ETAMins   float64
	HasETA    bool
}

// Service computes stats aggregates.
type Service interface {
	Summarize(userID int64) (*Summary, error)
}

type statsSvc struct {
	tasks    repository.TaskRepository
	projects repository.ProjectRepository
	pomodoro repository.PomodoroRepository
}

// NewService returns a Service backed by the provided repositories.
func NewService(
	tasks repository.TaskRepository,
	projects repository.ProjectRepository,
	pomodoro repository.PomodoroRepository,
) Service {
	return &statsSvc{tasks: tasks, projects: projects, pomodoro: pomodoro}
}

func (s *statsSvc) Summarize(userID int64) (*Summary, error) {
	tasks, err := s.tasks.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}
	sum := &Summary{}
	for _, t := range tasks {
		sum.Total++
		switch t.Status {
		case "DONE":
			sum.Done++
		case "TODO":
			sum.Todo++
		case "IN_PROGRESS":
			sum.InProgress++
		}
	}

	projects, err := s.projects.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		ptasks, _ := s.tasks.GetByProject(p.ID)
		remaining := 0
		for _, t := range ptasks {
			if t.Status != "DONE" {
				remaining++
			}
		}
		sessions, _ := s.pomodoro.GetCompletedByProject(p.ID)
		completed := make([]int, len(sessions))
		for i := range completed {
			completed[i] = 1
		}
		eta, err := CalculateETA(sessions, completed, remaining)
		peta := ProjectETA{Name: p.Name, Remaining: remaining}
		if err == nil {
			peta.ETAMins = eta
			peta.HasETA = true
		}
		sum.Projects = append(sum.Projects, peta)
	}
	return sum, nil
}
