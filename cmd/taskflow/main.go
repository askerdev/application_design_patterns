package main

import (
	"fmt"
	"log"
	"os"

	"taskflow/internal/config"
	"taskflow/internal/db"
	"taskflow/internal/domain"
	notesvc "taskflow/internal/note"
	pomodorosvc "taskflow/internal/pomodoro"
	projectsvc "taskflow/internal/project"
	remindersvc "taskflow/internal/reminder"
	sqliterepo "taskflow/internal/repository/sqlite"
	statssvc "taskflow/internal/stats"
	tasksvc "taskflow/internal/task"
	tagsvc "taskflow/internal/tag"
	"taskflow/internal/telegram"
	"taskflow/internal/tui"
)

func main() {
	cfg := config.Instance()
	conn := db.Instance(cfg.DBPath)
	defer conn.Close()

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("TaskFlow v0.1.0")
		return
	}

	// ── Repositories ─────────────────────────────────────────────────────────
	taskRepo := sqliterepo.NewTaskRepo(conn)
	cachedTaskRepo := tasksvc.NewCachingTaskRepo(taskRepo)
	projectRepo := sqliterepo.NewProjectRepo(conn)
	noteRepo := sqliterepo.NewNoteRepo(conn)
	reminderRepo := sqliterepo.NewReminderRepo(conn)
	tagRepo := sqliterepo.NewTagRepo(conn)
	pomodoroRepo := sqliterepo.NewPomodoroRepo(conn)
	userRepo := sqliterepo.NewUserRepo(conn)

	// ── Telegram (Observer + Adapter patterns) ───────────────────────────────
	tgClient := telegram.NewClient(cfg.TelegramBotToken, cfg.TelegramChatID)
	tgReminderSvc := telegram.NewReminderService(reminderRepo)
	tgSender := telegram.NewClientAdapter(tgClient)
	tgReminderSvc.SetSender(tgSender)
	tgReminderSvc.Register(telegram.NewTelegramNotifier(tgSender))

	// ── Services ──────────────────────────────────────────────────────────────
	svcs := tui.Services{
		Tasks:     tasksvc.NewService(cachedTaskRepo),
		Projects:  projectsvc.NewService(projectRepo),
		Notes:     notesvc.NewService(noteRepo),
		Reminders: remindersvc.NewService(reminderRepo, tgReminderSvc),
		Tags:      tagsvc.NewService(tagRepo),
		Pomodoro:  pomodorosvc.NewService(pomodoroRepo),
		Stats:     statssvc.NewService(taskRepo, projectRepo, pomodoroRepo),
	}

	// ── Bootstrap user ────────────────────────────────────────────────────────
	user, err := userRepo.GetFirst()
	if err != nil {
		user = &domain.User{Username: "default"}
		if err := userRepo.Create(user); err != nil {
			log.Fatal("create user:", err)
		}
	}

	// ── Run TUI ───────────────────────────────────────────────────────────────
	if _, err := tui.New(svcs, user).Run(); err != nil {
		log.Fatal(err)
	}
}
