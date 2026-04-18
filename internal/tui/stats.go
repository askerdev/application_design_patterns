package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	domain "taskflow/internal/domain"
)

type statsModel struct {
	viewport viewport.Model
	svcs     Services
	user     *domain.User
	width    int
	height   int
}

func newStatsModel(svcs Services, user *domain.User) statsModel {
	return statsModel{svcs: svcs, user: user, viewport: viewport.New(0, 0)}
}

func (m statsModel) reload() tea.Cmd {
	return func() tea.Msg { return statsLoadMsg{} }
}

type statsLoadMsg struct{}

func (m statsModel) Init() tea.Cmd { return m.reload() }

func (m statsModel) Update(msg tea.Msg) (statsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case statsLoadMsg:
		m.viewport.SetContent(m.buildContent())
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m statsModel) View() string {
	help := statusBarStyle.Render("j/k scroll  tab switch  q quit")
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Padding(1, 2).Render("Stats"),
		lipgloss.NewStyle().Padding(0, 2).Render(m.viewport.View()),
		help,
	)
}

func (m statsModel) setSize(w, h int) statsModel {
	m.width, m.height = w, h
	m.viewport = viewport.New(w-4, h-6)
	return m
}

func (m *statsModel) buildContent() string {
	taskReport := domain.NewTaskCountReport(m.svcs.Stats)
	etaReport := domain.NewProjectETAReport(m.svcs.Stats)

	taskSection := domain.Generate(taskReport, m.user.ID)
	etaSection := domain.Generate(etaReport, m.user.ID)

	if etaSection == "" {
		return taskSection
	}
	return taskSection + "\n\n" + etaSection
}
