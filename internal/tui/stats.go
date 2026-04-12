package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/domain"
	taskmath "taskflow/internal/math"
)

type statsModel struct {
	viewport viewport.Model
	repos    *repos
	user     *domain.User
	width    int
	height   int
}

func newStatsModel(r *repos, user *domain.User) statsModel {
	return statsModel{repos: r, user: user, viewport: viewport.New(0, 0)}
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
		case "esc", "h", "backspace":
			return m, func() tea.Msg { return backMsg{} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m statsModel) View() string {
	help := statusBarStyle.Render("j/k scroll  esc back")
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
	var sb strings.Builder

	tasks, _ := m.repos.tasks.GetAllByUser(m.user.ID)
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
	fmt.Fprintf(&sb, "Tasks: %d total  |  %d done  |  %d todo  |  %d in-progress\n\n", total, done, todo, inprog)

	projects, _ := m.repos.projects.GetAllByUser(m.user.ID)
	if len(projects) > 0 {
		sb.WriteString("Project ETA:\n")
		for _, p := range projects {
			ptasks, _ := m.repos.tasks.GetByProject(p.ID)
			remaining := 0
			for _, t := range ptasks {
				if t.Status != "DONE" {
					remaining++
				}
			}
			sessions, _ := m.repos.pomodoro.GetCompletedByProject(p.ID)
			completed := make([]int, len(sessions))
			for i := range completed {
				completed[i] = 1
			}
			eta, err := taskmath.CalculateETA(sessions, completed, remaining)
			if err != nil {
				fmt.Fprintf(&sb, "  %s: %d remaining (no history)\n", p.Name, remaining)
			} else {
				hours := int(eta) / 60
				mins := int(eta) % 60
				fmt.Fprintf(&sb, "  %s: %d remaining — ETA ~%dh%dm\n", p.Name, remaining, hours, mins)
			}
		}
	}
	return sb.String()
}
