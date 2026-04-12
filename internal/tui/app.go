package tui

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorRed    = "\033[31m"
)

type App struct {
	db          *sql.DB
	userRepo    *repository.UserRepo
	taskRepo    *repository.TaskRepo
	projectRepo *repository.ProjectRepo
	noteRepo    *repository.NoteRepo
	currentUser *domain.User
	scanner     *bufio.Scanner
}

func New(db *sql.DB) *App {
	return &App{
		db:          db,
		userRepo:    repository.NewUserRepo(db),
		taskRepo:    repository.NewTaskRepo(db),
		projectRepo: repository.NewProjectRepo(db),
		noteRepo:    repository.NewNoteRepo(db),
		scanner:     bufio.NewScanner(os.Stdin),
	}
}

func (a *App) Run() {
	a.ensureUser()
	a.mainLoop()
}

func (a *App) ensureUser() {
	u, err := a.userRepo.GetFirst()
	if err != nil {
		u = &domain.User{Username: "default"}
		if err := a.userRepo.Create(u); err != nil {
			fmt.Println("Error creating user:", err)
			os.Exit(1)
		}
		fmt.Printf("%sWelcome to TaskFlow!%s Created user 'default'.\n\n", colorGreen, colorReset)
	}
	a.currentUser = u
}

func (a *App) mainLoop() {
	for {
		a.printMenu()
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "1":
			a.tasksMenu()
		case "2":
			a.projectsMenu()
		case "3":
			a.notesMenu()
		case "0":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("Unknown option.")
		}
	}
}

func (a *App) printMenu() {
	fmt.Printf("\n%s=== TaskFlow ===%s  [user: %s]\n", colorBold, colorReset, a.currentUser.Username)
	fmt.Println("  1. Tasks")
	fmt.Println("  2. Projects")
	fmt.Println("  3. Notes")
	fmt.Println("  0. Exit")
}

func (a *App) tasksMenu() {
	for {
		fmt.Printf("\n%s-- Tasks --%s\n", colorCyan, colorReset)
		tasks, _ := a.taskRepo.GetAllByUser(a.currentUser.ID)
		if len(tasks) == 0 {
			fmt.Println("  (no tasks)")
		}
		for _, t := range tasks {
			status := fmt.Sprintf("[%s]", t.Status)
			overdue := ""
			if t.IsOverdue() {
				overdue = colorRed + " OVERDUE" + colorReset
			}
			fmt.Printf("  %d. %s %s%s\n", t.ID, status, t.Content, overdue)
		}
		fmt.Println("\n  a. Add    d. Delete    b. Back")
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "a":
			content := a.readLine("Task content: ")
			task := &domain.Task{
				UserID:   a.currentUser.ID,
				Content:  strings.TrimSpace(content),
				Status:   domain.TaskStatusTodo,
				Priority: domain.PriorityMedium,
			}
			if err := a.taskRepo.Create(task); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sTask created (id=%d)%s\n", colorGreen, task.ID, colorReset)
			}
		case "d":
			idStr := a.readLine("Task ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.taskRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}

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
		fmt.Println("\n  a. Add    b. Back")
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "a":
			name := a.readLine("Project name: ")
			p := &domain.Project{
				UserID: a.currentUser.ID,
				Name:   strings.TrimSpace(name),
				Status: domain.ProjectStatusActive,
			}
			if err := a.projectRepo.Create(p); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sProject created (id=%d)%s\n", colorGreen, p.ID, colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) notesMenu() {
	for {
		fmt.Printf("\n%s-- Notes --%s\n", colorCyan, colorReset)
		notes, _ := a.noteRepo.GetAllByUser(a.currentUser.ID)
		if len(notes) == 0 {
			fmt.Println("  (no notes)")
		}
		for _, n := range notes {
			fmt.Printf("  %d. %s\n", n.ID, n.Title)
		}
		fmt.Println("\n  a. Add    b. Back")
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "a":
			title := a.readLine("Title: ")
			content := a.readLine("Content: ")
			n := &domain.Note{
				UserID:  a.currentUser.ID,
				Title:   strings.TrimSpace(title),
				Content: strings.TrimSpace(content),
			}
			if err := a.noteRepo.Create(n); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sNote created (id=%d)%s\n", colorGreen, n.ID, colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) readLine(prompt string) string {
	fmt.Print(prompt)
	a.scanner.Scan()
	return a.scanner.Text()
}
