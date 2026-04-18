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

type reminderTickMsg time.Time

func reminderTickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg { return reminderTickMsg(t) })
}

type reminderItem struct {
	reminder    *domain.Reminder
	projectName string
	tagName     string
}

func (i reminderItem) Title() string {
	due := ""
	if i.reminder.IsReady() {
		due = overdueStyle.Render(" DUE")
	}
	return fmt.Sprintf("[%s] %s%s", i.reminder.Status, i.reminder.Content, due)
}

func (i reminderItem) Description() string {
	s := i.reminder.ReminderTime.Format("2006-01-02 15:04") + " · " + i.projectName
	if i.tagName != "" {
		s += " #" + i.tagName
	}
	return s
}

func (i reminderItem) FilterValue() string {
	return i.reminder.Content + " " + i.projectName + " " + i.tagName
}

type remindersModel struct {
	list     list.Model
	form     *huh.Form
	fContent *string
	fTime    *string
	fProject *string
	fTag     *string

	svcs   Services
	user   *domain.User
	status string
	width  int
	height int
}

func newRemindersModel(svcs Services, user *domain.User) remindersModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Reminders"
	l.Styles.Title = titleStyle
	return remindersModel{svcs: svcs, user: user, list: l}
}

func (m remindersModel) reload() tea.Cmd {
	return func() tea.Msg { return remindersLoadMsg{} }
}

type remindersLoadMsg struct{}

func (m remindersModel) Init() tea.Cmd { return m.reload() }

func (m remindersModel) Update(msg tea.Msg) (remindersModel, tea.Cmd) {
	if m.form != nil {
		f, cmd := m.form.Update(msg)
		if form, ok := f.(*huh.Form); ok {
			m.form = form
		}
		if m.form.State == huh.StateCompleted {
			m.form = nil
			if err := m.saveReminder(); err != nil {
				m.status = errorStyle.Render("Error: " + err.Error())
			} else {
				m.status = "Reminder created."
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
	case reminderTickMsg:
		if err := m.svcs.Reminders.Tick(); err != nil {
			m.status = errorStyle.Render("Reminder error: " + err.Error())
		}
		return m, tea.Batch(m.reload(), reminderTickCmd())

	case remindersLoadMsg:
		reminders, _ := m.svcs.Reminders.List(m.user.ID)
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
		items := make([]list.Item, len(reminders))
		for i, r := range reminders {
			proj, tag := "", ""
			if r.ProjectID != nil {
				proj = projMap[*r.ProjectID]
			}
			if r.TagID != nil {
				tag = tagMap[*r.TagID]
			}
			items[i] = reminderItem{reminder: r, projectName: proj, tagName: tag}
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
				content, rtime, proj, tag := "", "", fmt.Sprintf("%d", projects[0].ID), ""
				m.fContent, m.fTime, m.fProject, m.fTag = &content, &rtime, &proj, &tag
				m.form = m.buildAddForm(projects)
				return m, m.form.Init()
			case "d":
				if item, ok := m.list.SelectedItem().(reminderItem); ok {
					if err := m.svcs.Reminders.Delete(item.reminder.ID); err != nil {
						m.status = errorStyle.Render("Error: " + err.Error())
					} else {
						m.status = "Deleted."
					}
					return m, m.reload()
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m remindersModel) View() string {
	if m.form != nil {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.form.View())
	}
	tgStatus := "Telegram: NOT CONFIGURED"
	if m.svcs.Reminders.IsNotifierConfigured() {
		tgStatus = "Telegram: configured"
	}
	help := statusBarStyle.Render("a add  d delete  / filter  tab switch  q quit")
	body := tgStatus + "\n\n" + m.list.View()
	if m.status != "" {
		body += "\n" + m.status
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(body),
		help,
	)
}

func (m remindersModel) setSize(w, h int) remindersModel {
	m.width, m.height = w, h
	m.list.SetSize(w-4, h-6)
	return m
}

func (m *remindersModel) buildAddForm(projects []*domain.Project) *huh.Form {
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
			huh.NewInput().Title("Content").Value(m.fContent),
			huh.NewInput().Title("Time (YYYY-MM-DD HH:MM)").Value(m.fTime),
			huh.NewSelect[string]().Title("Project").Options(projOpts...).Value(m.fProject),
			huh.NewSelect[string]().Title("Tag (optional)").Options(tagOpts...).Value(m.fTag),
		),
	)
}

func (m *remindersModel) saveReminder() error {
	t, err := time.ParseInLocation("2006-01-02 15:04", *m.fTime, time.Local)
	if err != nil {
		return fmt.Errorf("invalid time format — use YYYY-MM-DD HH:MM")
	}
	var projID int64
	fmt.Sscanf(*m.fProject, "%d", &projID)
	var tagID *int64
	if *m.fTag != "" {
		id := int64(0)
		fmt.Sscanf(*m.fTag, "%d", &id)
		tagID = &id
	}
	factory := domain.NewEntityFactory()
	r := factory.CreateReminder(m.user.ID, *m.fContent, t, &projID, tagID)
	return m.svcs.Reminders.Create(r)
}
