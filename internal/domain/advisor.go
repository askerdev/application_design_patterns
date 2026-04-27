package domain

import (
	"fmt"
	"strings"
	"time"
)

type LLMGenerator interface {
	Generate(prompt string) (string, error)
}

type AdvisorService interface {
	Advise(userID int64, focusProjectID int64) (string, error)
}

type advisorSvc struct {
	tasks    TaskRepository
	projects ProjectRepository
	pomodoro PomodoroRepository
	llm      LLMGenerator
}

func NewAdvisorService(tasks TaskRepository, projects ProjectRepository, pomodoro PomodoroRepository, llm LLMGenerator) AdvisorService {
	return &advisorSvc{tasks: tasks, projects: projects, pomodoro: pomodoro, llm: llm}
}

func (s *advisorSvc) Advise(userID int64, focusProjectID int64) (string, error) {
	projects, err := s.projects.GetAllByUser(userID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	var sb strings.Builder
	sb.WriteString("You are a productivity advisor. Analyze the project data below and give concise advice (3-5 bullet points) on which tasks to tackle first to avoid missing deadlines.\n\n")
	fmt.Fprintf(&sb, "Current date: %s\n\n", now.Format("2006-01-02"))

	for _, p := range projects {
		if p.Status == ProjectStatusDone || p.Status == ProjectStatusArchived {
			continue
		}

		ptasks, _ := s.tasks.GetByProject(p.ID)
		remaining := 0
		totalSP := 0
		var taskLines []string
		for _, t := range ptasks {
			if t.Status == TaskStatusDone {
				continue
			}
			remaining++
			totalSP += t.StoryPoints
			due := ""
			if t.DueDate != nil {
				due = fmt.Sprintf(" due:%s", t.DueDate.Format("2006-01-02"))
			}
			overdue := ""
			if t.IsOverdue() {
				overdue = " OVERDUE"
			}
			focus := ""
			if p.ID == focusProjectID {
				focus = " [CURRENT FOCUS]"
			}
			taskLines = append(taskLines, fmt.Sprintf("    - %s | %dSP | %s%s%s%s",
				t.Content, t.StoryPoints, t.Priority, due, overdue, focus))
		}

		sessions, _ := s.pomodoro.GetCompletedByProject(p.ID)
		completed := make([]int, len(sessions))
		for i := range completed {
			completed[i] = 1
		}
		eta, etaErr := CalculateETA(sessions, completed, remaining)

		fmt.Fprintf(&sb, "Project: %s\n", p.Name)
		if p.DueDate != nil {
			daysLeft := int(p.DueDate.Sub(now).Hours() / 24)
			fmt.Fprintf(&sb, "  Due: %s (%d days left)\n", p.DueDate.Format("2006-01-02"), daysLeft)
		}
		if etaErr == nil && eta > 0 {
			fmt.Fprintf(&sb, "  ETA to finish: %.0f min\n", eta)
		}
		fmt.Fprintf(&sb, "  Remaining: %d tasks, %d total SP\n", remaining, totalSP)
		if len(taskLines) > 0 {
			sb.WriteString("  Open tasks:\n")
			for _, l := range taskLines {
				sb.WriteString(l + "\n")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Reply in the same language the project/task names are written in. Be direct and actionable.\n")

	return s.llm.Generate(sb.String())
}
