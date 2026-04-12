package tui

import (
	"fmt"

	taskmath "taskflow/internal/math"
)

func (a *App) statsMenu() {
	fmt.Printf("\n%s-- Stats --%s\n", colorCyan, colorReset)

	// Overall task counts
	tasks, _ := a.taskRepo.GetAllByUser(a.currentUser.ID)
	total, done, todo, inprog := 0, 0, 0, 0
	for _, t := range tasks {
		total++
		switch t.Status {
		case "DONE":
			done++
		case "TODO":
			todo++
		case "IN_PROGRESS":
			inprog++
		}
	}
	fmt.Printf("  Tasks: %d total  |  %s%d done%s  |  %d todo  |  %d in-progress\n",
		total, colorGreen, done, colorReset, todo, inprog)

	// Per-project ETA
	projects, _ := a.projectRepo.GetAllByUser(a.currentUser.ID)
	if len(projects) > 0 {
		fmt.Printf("\n  %sProject ETA:%s\n", colorYellow, colorReset)
		for _, p := range projects {
			ptasks, _ := a.taskRepo.GetByProject(p.ID)
			remaining := 0
			for _, t := range ptasks {
				if t.Status != "DONE" {
					remaining++
				}
			}

			sessions, _ := a.pomodoroRepo.GetCompletedByProject(p.ID)
			// Build completedPerSession: approximate as 1 task per session if no better data
			completedPerSession := make([]int, len(sessions))
			for i := range completedPerSession {
				completedPerSession[i] = 1
			}

			eta, err := taskmath.CalculateETA(sessions, completedPerSession, remaining)
			if err != nil {
				fmt.Printf("    %s: %d tasks remaining (no session history)\n", p.Name, remaining)
			} else {
				hours := int(eta) / 60
				mins := int(eta) % 60
				fmt.Printf("    %s: %d tasks remaining — ETA ~%dh%dm\n", p.Name, remaining, hours, mins)
			}
		}
	}

	a.readLine("\nPress Enter to go back...")
}
