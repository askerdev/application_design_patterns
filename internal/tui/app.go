package tui

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"taskflow/internal/domain"
	"taskflow/internal/repository"
	"taskflow/internal/telegram"
)

const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorCyan    = "\033[36m"
	colorRed     = "\033[31m"
	colorMagenta = "\033[35m"
)

type App struct {
	db              *sql.DB
	userRepo        *repository.UserRepo
	taskRepo        *repository.TaskRepo
	projectRepo     *repository.ProjectRepo
	noteRepo        *repository.NoteRepo
	reminderRepo    *repository.ReminderRepo
	tagRepo         *repository.TagRepo
	pomodoroRepo    *repository.PomodoroRepo
	reminderService *telegram.ReminderService
	currentUser     *domain.User
	scanner         *bufio.Scanner
}

func New(db *sql.DB, reminderService *telegram.ReminderService) *App {
	return &App{
		db:              db,
		userRepo:        repository.NewUserRepo(db),
		taskRepo:        repository.NewTaskRepo(db),
		projectRepo:     repository.NewProjectRepo(db),
		noteRepo:        repository.NewNoteRepo(db),
		reminderRepo:    repository.NewReminderRepo(db),
		tagRepo:         repository.NewTagRepo(db),
		pomodoroRepo:    repository.NewPomodoroRepo(db),
		reminderService: reminderService,
		scanner:         bufio.NewScanner(os.Stdin),
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
		fmt.Printf("%sWelcome to TaskFlow!%s\n\n", colorGreen, colorReset)
	}
	a.currentUser = u
}

func (a *App) mainLoop() {
	for {
		a.printMainMenu()
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "1":
			a.tasksMenu()
		case "2":
			a.projectsMenu()
		case "3":
			a.notesMenu()
		case "4":
			a.remindersMenu()
		case "5":
			a.tagsMenu()
		case "6":
			a.pomodoroMenu()
		case "7":
			a.statsMenu()
		case "0":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("Unknown option.")
		}
	}
}

func (a *App) printMainMenu() {
	fmt.Printf("\n%s=== TaskFlow ===%s  [%s]\n", colorBold, colorReset, a.currentUser.Username)
	fmt.Println("  1. Tasks")
	fmt.Println("  2. Projects")
	fmt.Println("  3. Notes")
	fmt.Println("  4. Reminders")
	fmt.Println("  5. Tags")
	fmt.Println("  6. Pomodoro")
	fmt.Println("  7. Stats")
	fmt.Println("  0. Exit")
}

// readLine prints prompt and reads one line from stdin
func (a *App) readLine(prompt string) string {
	fmt.Print(prompt)
	a.scanner.Scan()
	return a.scanner.Text()
}

// pickProject asks user to optionally pick a project; returns nil if skipped
func (a *App) pickProject() *int64 {
	projects, _ := a.projectRepo.GetAllByUser(a.currentUser.ID)
	if len(projects) == 0 {
		return nil
	}
	fmt.Println("  Projects:")
	for _, p := range projects {
		fmt.Printf("    %d. %s\n", p.ID, p.Name)
	}
	idStr := strings.TrimSpace(a.readLine("  Project ID (empty=none): "))
	if idStr == "" {
		return nil
	}
	var id int64
	fmt.Sscanf(idStr, "%d", &id)
	return &id
}

// pickTag asks user to optionally pick a tag; returns nil if skipped
func (a *App) pickTag() *int64 {
	tags, _ := a.tagRepo.GetAllByUser(a.currentUser.ID)
	if len(tags) == 0 {
		return nil
	}
	fmt.Println("  Tags:")
	for _, t := range tags {
		fmt.Printf("    %d. %s (%s)\n", t.ID, t.Name, t.Color)
	}
	idStr := strings.TrimSpace(a.readLine("  Tag ID (empty=none): "))
	if idStr == "" {
		return nil
	}
	var id int64
	fmt.Sscanf(idStr, "%d", &id)
	return &id
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
