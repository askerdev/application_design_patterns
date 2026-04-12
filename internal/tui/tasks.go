package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/domain"
)

type taskItem struct{ task *domain.Task }

func (i taskItem) Title() string {
	check := "[ ]"
	if i.task.Status == domain.TaskStatusDone {
		check = "[x]"
	}
	title := fmt.Sprintf("%s %s", check, i.task.Content)
	if i.task.IsOverdue() {
		title += overdueStyle.Render(" OVERDUE")
	}
	return title
}

func (i taskItem) Description() string {
	parts := []string{string(i.task.Priority), string(i.task.Status)}
	return strings.Join(parts, " · ")
}

func (i taskItem) FilterValue() string { return i.task.Content }

type tasksModel struct {
	list      list.Model
	form      *huh.Form
	fContent  string
	fPriority string
	fProject  string
	fTag      string

	repos  *repos
	user   *domain.User
	status string
	width  int
	height int
}

func newTasksModel(r *repos, user *domain.User) tasksModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Tasks"
	l.Styles.Title = titleStyle
	return tasksModel{repos: r, user: user, list: l}
}

func (m tasksModel) reload() tea.Cmd {
	return func() tea.Msg { return tasksLoadMsg{} }
}

type tasksLoadMsg struct{}

func (m tasksModel) Init() tea.Cmd { return m.reload() }

func (m tasksModel) Update(msg tea.Msg) (tasksModel, tea.Cmd) {
	if m.form != nil {
		f, cmd := m.form.Update(msg)
		if form, ok := f.(*huh.Form); ok {
			m.form = form
		}
		if m.form.State == huh.StateCompleted {
			m.form = nil
			if err := m.saveTask(); err != nil {
				m.status = errorStyle.Render("Error: " + err.Error())
			} else {
				m.status = "Task created."
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
	case tasksLoadMsg:
		tasks, _ := m.repos.tasks.GetAllByUser(m.user.ID)
		items := make([]list.Item, len(tasks))
		for i, t := range tasks {
			items[i] = taskItem{task: t}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "h", "backspace":
			return m, func() tea.Msg { return backMsg{} }
		case "a":
			m.fContent, m.fPriority, m.fProject, m.fTag = "", "MEDIUM", "", ""
			m.form = m.buildAddForm()
			return m, m.form.Init()
		case "d":
			if item, ok := m.list.SelectedItem().(taskItem); ok {
				if err := m.repos.tasks.Delete(item.task.ID); err != nil {
					m.status = errorStyle.Render("Error: " + err.Error())
				} else {
					m.status = "Deleted."
				}
				return m, m.reload()
			}
		case "c":
			if item, ok := m.list.SelectedItem().(taskItem); ok {
				t := item.task
				t.Complete()
				if err := m.repos.tasks.Update(t); err != nil {
					m.status = errorStyle.Render("Error: " + err.Error())
				} else {
					m.status = "Marked done."
				}
				return m, m.reload()
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m tasksModel) View() string {
	if m.form != nil {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.form.View())
	}
	help := statusBarStyle.Render("a add  c complete  d delete  esc back")
	body := m.list.View()
	if m.status != "" {
		body += "\n" + m.status
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(body),
		help,
	)
}

func (m tasksModel) setSize(w, h int) tasksModel {
	m.width, m.height = w, h
	m.list.SetSize(w-4, h-4)
	return m
}

func (m *tasksModel) buildAddForm() *huh.Form {
	projects, _ := m.repos.projects.GetAllByUser(m.user.ID)
	projOpts := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, p := range projects {
		projOpts = append(projOpts, huh.NewOption(p.Name, fmt.Sprintf("%d", p.ID)))
	}

	tags, _ := m.repos.tags.GetAllByUser(m.user.ID)
	tagOpts := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, t := range tags {
		tagOpts = append(tagOpts, huh.NewOption(t.Name, fmt.Sprintf("%d", t.ID)))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Content").
				Value(&m.fContent),
			huh.NewSelect[string]().
				Title("Priority").
				Options(
					huh.NewOption("High", "HIGH"),
					huh.NewOption("Medium", "MEDIUM"),
					huh.NewOption("Low", "LOW"),
				).
				Value(&m.fPriority),
			huh.NewSelect[string]().
				Title("Project (optional)").
				Options(projOpts...).
				Value(&m.fProject),
			huh.NewSelect[string]().
				Title("Tag (optional)").
				Options(tagOpts...).
				Value(&m.fTag),
		),
	)
}

func (m *tasksModel) saveTask() error {
	task := &domain.Task{
		UserID:   m.user.ID,
		Content:  m.fContent,
		Status:   domain.TaskStatusTodo,
		Priority: domain.Priority(m.fPriority),
	}
	if m.fProject != "" {
		var id int64
		fmt.Sscanf(m.fProject, "%d", &id)
		task.ProjectID = &id
	}
	if m.fTag != "" {
		var id int64
		fmt.Sscanf(m.fTag, "%d", &id)
		task.TagID = &id
	}
	return m.repos.tasks.Create(task)
}
