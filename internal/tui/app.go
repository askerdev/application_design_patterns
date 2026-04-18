package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	domain "taskflow/internal/domain"
	pomodorosvc "taskflow/internal/pomodoro"
)

// ── Tab definitions ───────────────────────────────────────────────────────────

var tabNames = []string{"Tasks", "Projects", "Notes", "Reminders", "Tags", "Pomodoro", "Stats"}

const (
	tabTasks = iota
	tabProjects
	tabNotes
	tabReminders
	tabTags
	tabPomodoro
	tabStats
)

// ── Service bundle ────────────────────────────────────────────────────────────

// Services bundles all domain service interfaces used by the TUI.
type Services struct {
	Tasks     domain.TaskService
	Projects  domain.ProjectService
	Notes     domain.NoteService
	Reminders domain.ReminderService
	Tags      domain.TagService
	Pomodoro  pomodorosvc.Service
	Stats     domain.StatsService
}

// ── Root model ────────────────────────────────────────────────────────────────

type model struct {
	activeTab int
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

const tabBarHeight = 2

func newModel(svcs Services, user *domain.User) model {
	return model{
		activeTab: tabTasks,
		tasks:     newTasksModel(svcs, user),
		projects:  newProjectsModel(svcs, user),
		notes:     newNotesModel(svcs, user),
		reminders: newRemindersModel(svcs, user),
		tags:      newTagsModel(svcs, user),
		pomodoro:  newPomodoroModel(svcs, user),
		stats:     newStatsModel(svcs, user),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.tasks.reload(),
		m.projects.reload(),
		m.notes.reload(),
		m.reminders.reload(),
		m.tags.reload(),
		m.pomodoro.reload(),
		m.stats.reload(),
		reminderTickCmd(),
	)
}

func (m model) tabBar() string {
	tabs := make([]string, len(tabNames))
	for i, name := range tabNames {
		if i == m.activeTab {
			tabs[i] = activeTabStyle.Render(name)
		} else {
			tabs[i] = inactiveTabStyle.Render(name)
		}
	}
	bar := strings.Join(tabs, " ")
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorMuted).
		Width(m.width).
		Render(bar)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always route tick to pomodoro so timer keeps running on any tab
	if _, ok := msg.(tickMsg); ok {
		updated, cmd := m.pomodoro.Update(msg)
		m.pomodoro = updated
		return m, cmd
	}
	// Always route reminder tick regardless of active tab
	if _, ok := msg.(reminderTickMsg); ok {
		updated, cmd := m.reminders.Update(msg)
		m.reminders = updated
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		contentH := msg.Height - tabBarHeight
		m.tasks = m.tasks.setSize(msg.Width, contentH)
		m.projects = m.projects.setSize(msg.Width, contentH)
		m.notes = m.notes.setSize(msg.Width, contentH)
		m.reminders = m.reminders.setSize(msg.Width, contentH)
		m.tags = m.tags.setSize(msg.Width, contentH)
		m.pomodoro = m.pomodoro.setSize(msg.Width, contentH)
		m.stats = m.stats.setSize(msg.Width, contentH)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if !m.activeHasForm() {
				m.activeTab = (m.activeTab + 1) % len(tabNames)
				return m, m.reloadActive()
			}
		case "shift+tab":
			if !m.activeHasForm() {
				m.activeTab = (m.activeTab - 1 + len(tabNames)) % len(tabNames)
				return m, m.reloadActive()
			}
		}
	}

	// Delegate to active tab
	switch m.activeTab {
	case tabTasks:
		updated, cmd := m.tasks.Update(msg)
		m.tasks = updated
		return m, cmd
	case tabProjects:
		updated, cmd := m.projects.Update(msg)
		m.projects = updated
		return m, cmd
	case tabNotes:
		updated, cmd := m.notes.Update(msg)
		m.notes = updated
		return m, cmd
	case tabReminders:
		updated, cmd := m.reminders.Update(msg)
		m.reminders = updated
		return m, cmd
	case tabTags:
		updated, cmd := m.tags.Update(msg)
		m.tags = updated
		return m, cmd
	case tabPomodoro:
		updated, cmd := m.pomodoro.Update(msg)
		m.pomodoro = updated
		return m, cmd
	case tabStats:
		updated, cmd := m.stats.Update(msg)
		m.stats = updated
		return m, cmd
	}
	return m, nil
}

func (m model) activeHasForm() bool {
	switch m.activeTab {
	case tabTasks:
		return m.tasks.form != nil
	case tabProjects:
		return m.projects.form != nil
	case tabNotes:
		return m.notes.form != nil
	case tabReminders:
		return m.reminders.form != nil
	case tabTags:
		return m.tags.form != nil
	case tabPomodoro:
		return m.pomodoro.form != nil
	}
	return false
}

func (m model) reloadActive() tea.Cmd {
	switch m.activeTab {
	case tabTasks:
		return m.tasks.reload()
	case tabProjects:
		return m.projects.reload()
	case tabNotes:
		return m.notes.reload()
	case tabReminders:
		return m.reminders.reload()
	case tabTags:
		return m.tags.reload()
	case tabPomodoro:
		cmd := m.pomodoro.reload()
		if m.pomodoro.machine != nil && m.pomodoro.machine.StateName() == "RUNNING" {
			cmd = tea.Batch(cmd, tickCmd())
		}
		return cmd
	case tabStats:
		return m.stats.reload()
	}
	return nil
}

func (m model) View() string {
	bar := m.tabBar()
	var content string
	switch m.activeTab {
	case tabTasks:
		content = m.tasks.View()
	case tabProjects:
		content = m.projects.View()
	case tabNotes:
		content = m.notes.View()
	case tabReminders:
		content = m.reminders.View()
	case tabTags:
		content = m.tags.View()
	case tabPomodoro:
		content = m.pomodoro.View()
	case tabStats:
		content = m.stats.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, bar, content)
}

// ── Public entry point ────────────────────────────────────────────────────────

// New returns a tea.Program. Call .Run() from main.
func New(svcs Services, user *domain.User) *tea.Program {
	return tea.NewProgram(newModel(svcs, user), tea.WithAltScreen())
}
