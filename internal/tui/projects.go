package tui

import (
	"fmt"
	"strings"

	"taskflow/internal/domain"
)

func (a *App) projectsMenu() {
	for {
		fmt.Printf("\n%s-- Projects --%s\n", colorCyan, colorReset)
		projects, _ := a.projectRepo.GetAllByUser(a.currentUser.ID)
		if len(projects) == 0 {
			fmt.Println("  (no projects)")
		}
		for _, p := range projects {
			fmt.Printf("  %d. [%s] %s\n", p.ID, p.Status, p.Name)
		}
		fmt.Println("\n  a.Add  v.View  d.Delete  b.Back")
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "a":
			name := strings.TrimSpace(a.readLine("Name: "))
			desc := strings.TrimSpace(a.readLine("Description (optional): "))
			p := &domain.Project{
				UserID:      a.currentUser.ID,
				Name:        name,
				Description: desc,
				Status:      domain.ProjectStatusActive,
			}
			if err := a.projectRepo.Create(p); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sProject created (id=%d)%s\n", colorGreen, p.ID, colorReset)
			}
		case "v":
			idStr := a.readLine("Project ID: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			a.viewProject(id)
		case "d":
			idStr := a.readLine("Project ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.projectRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) viewProject(id int64) {
	p, err := a.projectRepo.GetByID(id)
	if err != nil {
		fmt.Println("Project not found")
		return
	}
	fmt.Printf("\n%s== Project: %s ==%s  [%s]\n", colorBold, p.Name, colorReset, p.Status)
	if p.Description != "" {
		fmt.Println("  " + p.Description)
	}

	// Tasks
	tasks, _ := a.taskRepo.GetByProject(id)
	fmt.Printf("\n%sTasks (%d):%s\n", colorYellow, len(tasks), colorReset)
	for _, t := range tasks {
		overdue := ""
		if t.IsOverdue() {
			overdue = colorRed + " OVERDUE" + colorReset
		}
		fmt.Printf("  %d. [%s] %s%s\n", t.ID, t.Status, t.Content, overdue)
	}

	// Notes
	notes, _ := a.noteRepo.GetAllByUser(a.currentUser.ID)
	fmt.Printf("\n%sNotes:%s\n", colorYellow, colorReset)
	count := 0
	for _, n := range notes {
		if n.ProjectID != nil && *n.ProjectID == id {
			fmt.Printf("  %d. %s\n", n.ID, n.Title)
			count++
		}
	}
	if count == 0 {
		fmt.Println("  (none)")
	}

	// Reminders
	reminders, _ := a.reminderRepo.GetAllByUser(a.currentUser.ID)
	fmt.Printf("\n%sReminders:%s\n", colorYellow, colorReset)
	count = 0
	for _, r := range reminders {
		if r.ProjectID != nil && *r.ProjectID == id {
			fmt.Printf("  %d. [%s] %s @ %s\n", r.ID, r.Status, r.Content, r.ReminderTime.Format("2006-01-02 15:04"))
			count++
		}
	}
	if count == 0 {
		fmt.Println("  (none)")
	}

	// Pomodoro sessions
	sessions, _ := a.pomodoroRepo.GetCompletedByProject(id)
	fmt.Printf("\n%sCompleted Pomodoro Sessions (%d):%s\n", colorYellow, len(sessions), colorReset)
	for _, s := range sessions {
		fmt.Printf("  %d. %d min", s.ID, s.WorkDuration)
		if s.StartTime != nil {
			fmt.Printf(" started %s", s.StartTime.Format("2006-01-02 15:04"))
		}
		fmt.Println()
	}

	a.readLine("\nPress Enter to go back...")
}
