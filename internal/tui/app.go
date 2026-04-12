package tui

import (
	"database/sql"
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"taskflow/internal/domain"
	"taskflow/internal/repository"
	"taskflow/internal/telegram"
)

// ── Screen enum ──────────────────────────────────────────────────────────────

type screen int

const (
	screenHome screen = iota
	screenTasks
	screenProjects
	screenNotes
	screenReminders
	screenTags
	screenPomodoro
	screenStats
)

// ── Navigation messages ───────────────────────────────────────────────────────

type navigateMsg struct{ to screen }
type backMsg struct{}

// ── Shared repo bundle ────────────────────────────────────────────────────────

type repos struct {
	users     *repository.UserRepo
	tasks     *repository.TaskRepo
	projects  *repository.ProjectRepo
	notes     *repository.NoteRepo
	reminders *repository.ReminderRepo
	tags      *repository.TagRepo
	pomodoro  *repository.PomodoroRepo
}

// ── Root model ────────────────────────────────────────────────────────────────

type model struct {
	screen    screen
	home      homeModel
	tasks     tasksModel
	projects  projectsModel
	notes     notesModel
	reminders remindersModel
	tags      tagsModel
	pomodoro  pomodoroModel
	stats     statsModel
	width     int
	height    int
}

func newModel(r *repos, user *domain.User, svc *telegram.ReminderService) model {
	return model{
		screen:    screenHome,
		home:      newHomeModel(),
		tasks:     newTasksModel(r, user),
		projects:  newProjectsModel(r, user),
		notes:     newNotesModel(r, user),
		reminders: newRemindersModel(r, user, svc),
		tags:      newTagsModel(r, user),
		pomodoro:  newPomodoroModel(r, user),
		stats:     newStatsModel(r, user),
	}
}

func (m model) Init() tea.Cmd {
	return m.home.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.home = m.home.setSize(msg.Width, msg.Height)
		m.tasks = m.tasks.setSize(msg.Width, msg.Height)
		m.projects = m.projects.setSize(msg.Width, msg.Height)
		m.notes = m.notes.setSize(msg.Width, msg.Height)
		m.reminders = m.reminders.setSize(msg.Width, msg.Height)
		m.tags = m.tags.setSize(msg.Width, msg.Height)
		m.pomodoro = m.pomodoro.setSize(msg.Width, msg.Height)
		m.stats = m.stats.setSize(msg.Width, msg.Height)
		return m, nil

	case navigateMsg:
		m.screen = msg.to
		switch m.screen {
		case screenTasks:
			return m, m.tasks.reload()
		case screenProjects:
			return m, m.projects.reload()
		case screenNotes:
			return m, m.notes.reload()
		case screenReminders:
			return m, m.reminders.reload()
		case screenTags:
			return m, m.tags.reload()
		case screenPomodoro:
			return m, m.pomodoro.reload()
		case screenStats:
			return m, m.stats.reload()
		}
		return m, nil

	case backMsg:
		m.screen = screenHome
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Delegate to active screen
	switch m.screen {
	case screenHome:
		updated, cmd := m.home.Update(msg)
		m.home = updated
		return m, cmd
	case screenTasks:
		updated, cmd := m.tasks.Update(msg)
		m.tasks = updated
		return m, cmd
	case screenProjects:
		updated, cmd := m.projects.Update(msg)
		m.projects = updated
		return m, cmd
	case screenNotes:
		updated, cmd := m.notes.Update(msg)
		m.notes = updated
		return m, cmd
	case screenReminders:
		updated, cmd := m.reminders.Update(msg)
		m.reminders = updated
		return m, cmd
	case screenTags:
		updated, cmd := m.tags.Update(msg)
		m.tags = updated
		return m, cmd
	case screenPomodoro:
		updated, cmd := m.pomodoro.Update(msg)
		m.pomodoro = updated
		return m, cmd
	case screenStats:
		updated, cmd := m.stats.Update(msg)
		m.stats = updated
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenHome:
		return m.home.View()
	case screenTasks:
		return m.tasks.View()
	case screenProjects:
		return m.projects.View()
	case screenNotes:
		return m.notes.View()
	case screenReminders:
		return m.reminders.View()
	case screenTags:
		return m.tags.View()
	case screenPomodoro:
		return m.pomodoro.View()
	case screenStats:
		return m.stats.View()
	}
	return ""
}

// ── Public entry point ────────────────────────────────────────────────────────

// New returns a tea.Program. Call .Run() from main.
func New(db *sql.DB, svc *telegram.ReminderService) *tea.Program {
	r := &repos{
		users:     repository.NewUserRepo(db),
		tasks:     repository.NewTaskRepo(db),
		projects:  repository.NewProjectRepo(db),
		notes:     repository.NewNoteRepo(db),
		reminders: repository.NewReminderRepo(db),
		tags:      repository.NewTagRepo(db),
		pomodoro:  repository.NewPomodoroRepo(db),
	}
	u, err := r.users.GetFirst()
	if err != nil {
		u = &domain.User{Username: "default"}
		if err := r.users.Create(u); err != nil {
			log.Fatal("create user:", err)
		}
	}
	return tea.NewProgram(newModel(r, u, svc), tea.WithAltScreen())
}
