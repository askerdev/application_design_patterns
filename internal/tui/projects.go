package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/domain"
)

type projectItem struct{ project *domain.Project }

func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string { return string(i.project.Status) }
func (i projectItem) FilterValue() string { return i.project.Name }

type projectsModel struct {
	list   list.Model
	detail viewport.Model
	form   *huh.Form
	fName  string
	fDesc  string

	repos  *repos
	user   *domain.User
	status string
	width  int
	height int
}

func newProjectsModel(r *repos, user *domain.User) projectsModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Projects"
	l.Styles.Title = titleStyle
	l.SetFilteringEnabled(false)
	return projectsModel{repos: r, user: user, list: l, detail: viewport.New(0, 0)}
}

func (m projectsModel) reload() tea.Cmd {
	return func() tea.Msg { return projectsLoadMsg{} }
}

type projectsLoadMsg struct{}

func (m projectsModel) Init() tea.Cmd { return m.reload() }

func (m projectsModel) Update(msg tea.Msg) (projectsModel, tea.Cmd) {
	if m.form != nil {
		f, cmd := m.form.Update(msg)
		if form, ok := f.(*huh.Form); ok {
			m.form = form
		}
		if m.form.State == huh.StateCompleted {
			m.form = nil
			if err := m.repos.projects.Create(&domain.Project{
				UserID:      m.user.ID,
				Name:        m.fName,
				Description: m.fDesc,
				Status:      domain.ProjectStatusActive,
			}); err != nil {
				m.status = errorStyle.Render("Error: " + err.Error())
			} else {
				m.status = "Project created."
			}
			return m, m.reload()
		}
		if m.form.State == huh.StateAborted {
			m.form = nil
			m.status = "Cancelled."
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case projectsLoadMsg:
		projects, _ := m.repos.projects.GetAllByUser(m.user.ID)
		items := make([]list.Item, len(projects))
		for i, p := range projects {
			items[i] = projectItem{project: p}
		}
		m.list.SetItems(items)
		m.refreshDetail()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "h", "backspace":
			return m, func() tea.Msg { return backMsg{} }
		case "a":
			m.fName, m.fDesc = "", ""
			m.form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Project name").Value(&m.fName),
					huh.NewInput().Title("Description (optional)").Value(&m.fDesc),
				),
			)
			return m, m.form.Init()
		case "d":
			if item, ok := m.list.SelectedItem().(projectItem); ok {
				if err := m.repos.projects.Delete(item.project.ID); err != nil {
					m.status = errorStyle.Render("Error: " + err.Error())
				} else {
					m.status = "Deleted."
				}
				return m, m.reload()
			}
		case "j", "down":
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			m.refreshDetail()
			return m, cmd
		case "k", "up":
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			m.refreshDetail()
			return m, cmd
		}
	}

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	var vpCmd tea.Cmd
	m.detail, vpCmd = m.detail.Update(msg)
	return m, tea.Batch(listCmd, vpCmd)
}

func (m *projectsModel) refreshDetail() {
	item, ok := m.list.SelectedItem().(projectItem)
	if !ok {
		m.detail.SetContent("(select a project)")
		return
	}
	p := item.project
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name:   %s\n", p.Name)
	fmt.Fprintf(&sb, "Status: %s\n", p.Status)
	if p.Description != "" {
		fmt.Fprintf(&sb, "\n%s\n", p.Description)
	}

	tasks, _ := m.repos.tasks.GetByProject(p.ID)
	fmt.Fprintf(&sb, "\nTasks (%d):\n", len(tasks))
	for _, t := range tasks {
		overdue := ""
		if t.IsOverdue() {
			overdue = " OVERDUE"
		}
		fmt.Fprintf(&sb, "  [%s] %s%s\n", t.Status, t.Content, overdue)
	}

	sessions, _ := m.repos.pomodoro.GetCompletedByProject(p.ID)
	fmt.Fprintf(&sb, "\nPomodoro sessions (%d):\n", len(sessions))
	for _, s := range sessions {
		start := ""
		if s.StartTime != nil {
			start = " @ " + s.StartTime.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(&sb, "  %dmin%s\n", s.WorkDuration, start)
	}

	m.detail.SetContent(sb.String())
	m.detail.GotoTop()
}

func (m projectsModel) View() string {
	if m.form != nil {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.form.View())
	}

	leftW := m.width / 3
	rightW := m.width - leftW - 3

	left := activePaneStyle.Width(leftW).Height(m.height - 4).Render(m.list.View())
	right := inactivePaneStyle.Width(rightW).Height(m.height - 4).Render(m.detail.View())
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	status := ""
	if m.status != "" {
		status = "\n" + m.status
	}
	help := statusBarStyle.Render("a add  d delete  j/k navigate  esc back")
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 1).Render(body+status),
		help,
	)
}

func (m projectsModel) setSize(w, h int) projectsModel {
	m.width, m.height = w, h
	listW := w/3 - 4
	vpW := w - w/3 - 6
	m.list.SetSize(listW, h-6)
	m.detail = viewport.New(vpW, h-6)
	return m
}
