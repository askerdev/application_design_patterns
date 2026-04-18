package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"taskflow/internal/app"
	"taskflow/internal/config"
	"taskflow/internal/db"
	domain "taskflow/internal/domain"
	tgnot "taskflow/internal/notifications/telegram"
	pomodorosvc "taskflow/internal/pomodoro"
	sqliterepo "taskflow/internal/repository/sqlite"
	"taskflow/internal/tui"
)

func main() {
	cfg := config.Instance()
	conn := db.Instance(cfg.DBPath)
	defer conn.Close()

	taskRepo := sqliterepo.NewTaskRepo(conn)
	projectRepo := sqliterepo.NewProjectRepo(conn)
	noteRepo := sqliterepo.NewNoteRepo(conn)
	reminderRepo := sqliterepo.NewReminderRepo(conn)
	tagRepo := sqliterepo.NewTagRepo(conn)
	pomodoroRepo := sqliterepo.NewPomodoroRepo(conn)
	userRepo := sqliterepo.NewUserRepo(conn)

	tgClient := tgnot.NewClient(cfg.TelegramBotToken, cfg.TelegramChatID)
	tgSender := tgnot.NewClientAdapter(tgClient)
	tgCoordinator := tgnot.NewReminderCoordinator(reminderRepo)
	tgCoordinator.SetSender(tgSender)
	tgCoordinator.Register(tgnot.NewTelegramNotifier(tgSender))

	cachedTaskRepo := domain.NewCachingTaskRepo(taskRepo)
	taskSvc := domain.NewTaskService(cachedTaskRepo)
	projectSvc := domain.NewProjectService(projectRepo)
	noteSvc := domain.NewNoteService(noteRepo)

	svcs := tui.Services{
		Tasks:     taskSvc,
		Projects:  projectSvc,
		Notes:     noteSvc,
		Reminders: domain.NewReminderService(reminderRepo, tgCoordinator),
		Tags:      domain.NewTagService(tagRepo),
		Pomodoro:  pomodorosvc.NewService(pomodoroRepo),
		Stats:     domain.NewStatsService(taskRepo, projectRepo, pomodoroRepo),
	}

	user, err := userRepo.GetFirst()
	if err != nil {
		user = &domain.User{Username: "default"}
		if err := userRepo.Create(user); err != nil {
			log.Fatal("create user:", err)
		}
	}

	facade := app.NewAppFacade(taskSvc, projectSvc, noteSvc, user)

	rootCmd := &cobra.Command{
		Use:   "taskflow",
		Short: "TaskFlow — task manager with TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tui.New(svcs, user).Run()
			return err
		},
	}

	var priority string
	taskAddCmd := &cobra.Command{
		Use:   "add [content]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := facade.AddTask(args[0], priority)
			if err != nil {
				return err
			}
			fmt.Printf("Created task #%d: %s\n", t.ID, t.Content)
			return nil
		},
	}
	taskAddCmd.Flags().StringVarP(&priority, "priority", "p", "MEDIUM", "Priority: HIGH, MEDIUM, LOW")

	taskListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := facade.ListTasks()
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Println("No tasks.")
				return nil
			}
			for _, t := range tasks {
				status := "[ ]"
				if t.Status == domain.TaskStatusDone {
					status = "[x]"
				}
				fmt.Printf("%s #%d %s (%s)\n", status, t.ID, t.Content, t.Priority)
			}
			return nil
		},
	}

	taskCmd := &cobra.Command{Use: "task", Short: "Manage tasks"}
	taskCmd.AddCommand(taskAddCmd, taskListCmd)

	projectAddCmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := facade.AddProject(args[0], "")
			if err != nil {
				return err
			}
			fmt.Printf("Created project #%d: %s\n", p.ID, p.Name)
			return nil
		},
	}

	projectListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := facade.ListProjects()
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				fmt.Println("No projects.")
				return nil
			}
			for _, p := range projects {
				fmt.Printf("#%d %s [%s]\n", p.ID, p.Name, p.Status)
			}
			return nil
		},
	}

	projectCmd := &cobra.Command{Use: "project", Short: "Manage projects"}
	projectCmd.AddCommand(projectAddCmd, projectListCmd)

	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Populate DB with demo data (≥20 records)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := db.RunSeed(conn); err != nil {
				return err
			}
			fmt.Println("Seeded: 4 projects, 4 tags, 10 tasks, 4 notes, 4 reminders, 5 pomodoro sessions")
			return nil
		},
	}

	rootCmd.AddCommand(taskCmd, projectCmd, seedCmd)

	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("TaskFlow v0.1.0")
		return
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
