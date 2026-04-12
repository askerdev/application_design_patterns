package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/domain"
	"taskflow/internal/telegram"
)

type reminderItem struct{ reminder *domain.Reminder }

func (i reminderItem) Title() string {
	due := ""
	if i.reminder.IsReady() {
		due = overdueStyle.Render(" DUE")
	}
	return fmt.Sprintf("[%s] %s%s", i.reminder.Status, i.reminder.Content, due)
}

func (i reminderItem) Description() string {
	return i.reminder.ReminderTime.Format("2006-01-02 15:04")
}

func (i reminderItem) FilterValue() string { return i.reminder.Content }

type remindersModel struct {
	list     list.Model
	form     *huh.Form
	fContent string
	fTime    string
	fProject string
	fTag     string

	repos  *repos
	user   *domain.User
	svc    *telegram.ReminderService
	status string
	width  int
	height int
}

func newRemindersModel(r *repos, user *domain.User, svc *telegram.ReminderService) remindersModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Reminders"
	l.Styles.Title = titleStyle
	return remindersModel{repos: r, user: user, svc: svc, list: l}
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
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case remindersLoadMsg:
		reminders, _ := m.repos.reminders.GetAllByUser(m.user.ID)
		items := make([]list.Item, len(reminders))
		for i, r := range reminders {
			items[i] = reminderItem{reminder: r}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "h", "backspace":
			return m, func() tea.Msg { return backMsg{} }
		case "a":
			m.fContent, m.fTime, m.fProject, m.fTag = "", "", "", ""
			m.form = m.buildAddForm()
			return m, m.form.Init()
		case "s":
			if m.svc != nil {
				if err := m.svc.CheckAndNotify(); err != nil {
					m.status = errorStyle.Render("Error: " + err.Error())
				} else {
					m.status = "Pending reminders processed."
				}
				return m, m.reload()
			}
		case "d":
			if item, ok := m.list.SelectedItem().(reminderItem); ok {
				if err := m.repos.reminders.Delete(item.reminder.ID); err != nil {
					m.status = errorStyle.Render("Error: " + err.Error())
				} else {
					m.status = "Deleted."
				}
				return m, m.reload()
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
	if m.svc != nil && m.svc.IsConfigured() {
		tgStatus = "Telegram: configured"
	}
	help := statusBarStyle.Render("a add  s send pending  d delete  esc back")
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

func (m *remindersModel) buildAddForm() *huh.Form {
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
			huh.NewInput().Title("Content").Value(&m.fContent),
			huh.NewInput().Title("Time (YYYY-MM-DD HH:MM)").Value(&m.fTime),
			huh.NewSelect[string]().Title("Project (optional)").Options(projOpts...).Value(&m.fProject),
			huh.NewSelect[string]().Title("Tag (optional)").Options(tagOpts...).Value(&m.fTag),
		),
	)
}

func (m *remindersModel) saveReminder() error {
	t, err := time.ParseInLocation("2006-01-02 15:04", m.fTime, time.Local)
	if err != nil {
		return fmt.Errorf("invalid time format — use YYYY-MM-DD HH:MM")
	}
	r := &domain.Reminder{
		UserID:       m.user.ID,
		Content:      m.fContent,
		ReminderTime: t,
		Status:       domain.ReminderStatusPending,
	}
	if m.fProject != "" {
		var id int64
		fmt.Sscanf(m.fProject, "%d", &id)
		r.ProjectID = &id
	}
	if m.fTag != "" {
		var id int64
		fmt.Sscanf(m.fTag, "%d", &id)
		r.TagID = &id
	}
	return m.repos.reminders.Create(r)
}
