package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	domain "taskflow/internal/domain"
)

type taskItem struct {
	task        *domain.Task
	projectName string
	tagName     string
}

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
	r := domain.NewOverdueDecorator(domain.NewPriorityDecorator(&domain.BaseRenderer{}))
	s := r.Render(i.task)
	if i.projectName != "" {
		s += " · " + i.projectName
	}
	if i.tagName != "" {
		s += " #" + i.tagName
	}
	return s
}

func (i taskItem) FilterValue() string {
	return i.task.Content + " " + i.projectName + " " + i.tagName
}

type namedFilter struct {
	label    string
	strategy domain.FilterStrategy
}

func taskFilters() []namedFilter {
	return []namedFilter{
		{"все", nil},
		{"todo", domain.ByStatusFilter{Status: domain.TaskStatusTodo}},
		{"done", domain.ByStatusFilter{Status: domain.TaskStatusDone}},
		{"просроченные", domain.ByDateFilter{From: time.Time{}, To: time.Now()}},
	}
}

type tasksModel struct {
	list      list.Model
	form      *huh.Form
	fContent  *string
	fPriority *string
	fProject  *string
	fTag      *string
	fDueDate  *string

	svcs      Services
	user      *domain.User
	history   domain.CommandHistory
	filterIdx int
	status    string
	width     int
	height    int
}

func newTasksModel(svcs Services, user *domain.User) tasksModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Tasks"
	l.Styles.Title = titleStyle
	return tasksModel{svcs: svcs, user: user, list: l}
}

func (m tasksModel) reload() tea.Cmd {
	return func() tea.Msg { return tasksLoadMsg{} }
}

type tasksLoadMsg struct{}

func (m tasksModel) Init() tea.Cmd { return m.reload() }

func (m tasksModel) Update(msg tea.Msg) (tasksModel, tea.Cmd) {
	if m.form != nil {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			m.form = nil
			m.status = "Cancelled."
			return m, nil
		}
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
		projects, _ := m.svcs.Projects.List(m.user.ID)
		projMap := make(map[int64]string, len(projects))
		for _, p := range projects {
			projMap[p.ID] = p.Name
		}
		tags, _ := m.svcs.Tags.List(m.user.ID)
		tagMap := make(map[int64]string, len(tags))
		for _, t := range tags {
			tagMap[t.ID] = t.Name
		}
		iter := m.svcs.Tasks.Iterate(m.user.ID)
		var all []*domain.Task
		for iter.HasNext() {
			all = append(all, iter.Next())
		}
		if f := taskFilters()[m.filterIdx].strategy; f != nil {
			all = f.Filter(all)
		}
		var items []list.Item
		for _, t := range all {
			proj, tag := "", ""
			if t.ProjectID != nil {
				proj = projMap[*t.ProjectID]
			}
			if t.TagID != nil {
				tag = tagMap[*t.TagID]
			}
			items = append(items, taskItem{task: t, projectName: proj, tagName: tag})
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Unfiltered {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "a":
				projects, _ := m.svcs.Projects.List(m.user.ID)
				if len(projects) == 0 {
					m.status = errorStyle.Render("Create a project first.")
					return m, nil
				}
				content, priority, proj, tag, due := "", "MEDIUM", fmt.Sprintf("%d", projects[0].ID), "", ""
				m.fContent, m.fPriority, m.fProject, m.fTag, m.fDueDate = &content, &priority, &proj, &tag, &due
				m.form = m.buildAddForm(projects)
				return m, m.form.Init()
			case "d":
				if item, ok := m.list.SelectedItem().(taskItem); ok {
					cmd := domain.NewDeleteTaskCommand(m.svcs.Tasks, item.task)
					if err := cmd.Execute(); err != nil {
						m.status = errorStyle.Render("Error: " + err.Error())
					} else {
						m.history.Push(cmd)
						m.status = "Deleted."
					}
					return m, m.reload()
				}
			case "x":
				if item, ok := m.list.SelectedItem().(taskItem); ok {
					cmd := domain.NewCompleteTaskCommand(m.svcs.Tasks, item.task)
					if err := cmd.Execute(); err != nil {
						m.status = errorStyle.Render("Error: " + err.Error())
					} else {
						m.history.Push(cmd)
						m.status = "Marked done."
					}
					return m, m.reload()
				}
			case "f":
				m.filterIdx = (m.filterIdx + 1) % len(taskFilters())
				m.status = "Фильтр: " + taskFilters()[m.filterIdx].label
				return m, m.reload()
			case "u":
				if err := m.history.Undo(); err != nil {
					m.status = errorStyle.Render("Undo failed: " + err.Error())
				} else {
					m.status = "Undone."
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
	help := statusBarStyle.Render(fmt.Sprintf("a add  x complete  d delete  u undo  f фильтр[%s]  / поиск  tab switch  q quit", taskFilters()[m.filterIdx].label))
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

func (m *tasksModel) buildAddForm(projects []*domain.Project) *huh.Form {
	projOpts := make([]huh.Option[string], len(projects))
	for i, p := range projects {
		projOpts[i] = huh.NewOption(p.Name, fmt.Sprintf("%d", p.ID))
	}

	tags, _ := m.svcs.Tags.List(m.user.ID)
	tagOpts := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, t := range tags {
		tagOpts = append(tagOpts, huh.NewOption(t.Name, fmt.Sprintf("%d", t.ID)))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Content").
				Value(m.fContent),
			huh.NewSelect[string]().
				Title("Priority").
				Options(
					huh.NewOption("High", "HIGH"),
					huh.NewOption("Medium", "MEDIUM"),
					huh.NewOption("Low", "LOW"),
				).
				Value(m.fPriority),
			huh.NewSelect[string]().
				Title("Project").
				Options(projOpts...).
				Value(m.fProject),
			huh.NewSelect[string]().
				Title("Tag (optional)").
				Options(tagOpts...).
				Value(m.fTag),
			huh.NewInput().
				Title("Due Date (YYYY-MM-DD, optional)").
				Placeholder("2026-05-01").
				Value(m.fDueDate),
		),
	)
}

func (m *tasksModel) saveTask() error {
	var projID int64
	fmt.Sscanf(*m.fProject, "%d", &projID)
	var tagID *int64
	if *m.fTag != "" {
		id := int64(0)
		fmt.Sscanf(*m.fTag, "%d", &id)
		tagID = &id
	}
	var dueDate *time.Time
	if *m.fDueDate != "" {
		t, err := time.ParseInLocation("2006-01-02", *m.fDueDate, time.Local)
		if err == nil {
			dueDate = &t
		}
	}
	factory := domain.NewEntityFactory()
	t := factory.CreateTask(m.user.ID, *m.fContent, domain.Priority(*m.fPriority), &projID, tagID, dueDate)
	if err := m.svcs.Tasks.Create(t); err != nil {
		return err
	}
	if dueDate != nil {
		reminderTime := dueDate.Add(9 * time.Hour)
		r := factory.CreateReminder(m.user.ID, "Дедлайн задачи: "+t.Content, reminderTime, &projID, tagID)
		_ = m.svcs.Reminders.Create(r)
	}
	return nil
}
