package domain

type Summary struct {
	Total      int
	Done       int
	Todo       int
	InProgress int
	Projects   []ProjectETA
}

type ProjectETA struct {
	Name      string
	Remaining int
	ETAMins   float64
	HasETA    bool
}

type StatsService interface {
	Summarize(userID int64) (*Summary, error)
}

type statsSvc struct {
	tasks    TaskRepository
	projects ProjectRepository
	pomodoro PomodoroRepository
}

func NewStatsService(tasks TaskRepository, projects ProjectRepository, pomodoro PomodoroRepository) StatsService {
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
		case TaskStatusDone:
			sum.Done++
		case TaskStatusTodo:
			sum.Todo++
		case TaskStatusInProgress:
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
			if t.Status != TaskStatusDone {
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
